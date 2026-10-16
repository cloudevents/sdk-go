/*
 Copyright 2026 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package http

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/cloudevents/sdk-go/v2/protocol"
)

func TestMakeRequestHeaderIsolation(t *testing.T) {
	for _, tc := range []struct {
		name   string
		custom http.Header
		set    bool
	}{
		{name: "no custom headers"},
		{name: "nil custom headers", set: true},
		{name: "empty custom headers", custom: http.Header{}, set: true},
		{
			name: "multiple custom values",
			custom: http.Header{
				"X-Shared": {"context", "second"},
				"X-Keep":   {"original"},
			},
			set: true,
		},
	} {
		for _, withTemplate := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/template=%t", tc.name, withTemplate), func(t *testing.T) {
				ctx := context.Background()
				custom := tc.custom.Clone()
				original := custom.Clone()
				if tc.set {
					ctx = WithCustomHeader(ctx, custom)
				}
				p, err := New()
				require.NoError(t, err)
				want := custom.Clone()
				if want == nil {
					want = http.Header{}
				}
				var template http.Header
				if withTemplate {
					template = http.Header{
						"X-Shared":   {"template"},
						"X-Template": {"unchanged"},
					}
					p.RequestTemplate = &http.Request{Header: template}
					want.Add("X-Shared", "template")
					want.Set("X-Template", "unchanged")
				}
				templateOriginal := template.Clone()

				first := p.makeRequest(ctx)
				require.NotNil(t, first.Header)
				assert.Equal(t, want, first.Header)
				assert.Equal(t, original, custom, "constructing a request must not change context headers")
				second := p.makeRequest(ctx)
				assert.Equal(t, want, second.Header)

				// Mutate both a slice element and the map to catch shallow copies.
				if len(first.Header["X-Shared"]) > 0 {
					first.Header["X-Shared"][0] = "changed"
				}
				first.Header.Add("X-Shared", "added")
				first.Header.Set("X-New", "first request only")
				first.Header.Del("X-Keep")
				assert.Equal(t, want, second.Header)
				assert.Equal(t, original, custom)
				assert.Equal(t, templateOriginal, template)
				assert.Equal(t, want, p.makeRequest(ctx).Header)
			})
		}
	}
}

func TestSendWithSharedCustomHeaders(t *testing.T) {
	for _, tc := range []struct {
		name       string
		structured []bool
	}{
		{name: "binary", structured: []bool{false, false, false}},
		{name: "structured", structured: []bool{true, true, true}},
		{name: "switch encoding", structured: []bool{false, true, false}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			type received struct {
				header http.Header
				body   []byte
				err    error
			}
			requests := make(chan received, len(tc.structured))
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, err := io.ReadAll(r.Body)
				requests <- received{header: r.Header.Clone(), body: body, err: err}
				w.WriteHeader(http.StatusAccepted)
			}))
			defer server.Close()
			p, err := New(WithTarget(server.URL))
			require.NoError(t, err)
			p.RequestTemplate = &http.Request{
				Method: http.MethodPost,
				Header: http.Header{"X-Custom": {"template"}},
			}
			custom := http.Header{"X-Custom": {"context", "second"}}
			original := custom.Clone()
			ctx := WithCustomHeader(context.Background(), custom)

			for i, structured := range tc.structured {
				e := event.New()
				e.SetID(strconv.Itoa(i))
				e.SetSource("/header-isolation")
				e.SetType("com.example.headers")
				if i == 0 {
					e.SetSubject("first event only")
					e.SetExtension("first", "extension")
				}
				require.NoError(t, e.SetData("text/plain", "payload"))
				sendCtx := binding.WithForceBinary(ctx)
				if structured {
					sendCtx = binding.WithForceStructured(ctx)
				}
				require.True(t, protocol.IsACK(p.Send(sendCtx, binding.ToMessage(&e))))
				got := <-requests
				require.NoError(t, got.err)
				assert.Equal(t, []string{"context", "second", "template"}, got.header.Values("X-Custom"))
				assert.Equal(t, original, custom)

				if structured {
					assert.Equal(t, event.ApplicationCloudEventsJSON, got.header.Get(ContentType))
					assert.Empty(t, got.header.Get("Ce-Id"))
					assert.Empty(t, got.header.Get("Ce-Subject"))
					assert.Empty(t, got.header.Get("Ce-First"))
					var decoded event.Event
					require.NoError(t, decoded.UnmarshalJSON(got.body))
					assert.Equal(t, e.ID(), decoded.ID())
					assert.Equal(t, e.Subject(), decoded.Subject())
				} else {
					assert.Equal(t, []string{e.ID()}, got.header.Values("Ce-Id"))
					assert.Equal(t, []string{e.Type()}, got.header.Values("Ce-Type"))
					assert.Equal(t, []string{e.Source()}, got.header.Values("Ce-Source"))
					assert.Equal(t, []string{e.SpecVersion()}, got.header.Values("Ce-Specversion"))
					assert.Equal(t, []string{"text/plain"}, got.header.Values(ContentType))
					assert.Equal(t, e.Subject(), got.header.Get("Ce-Subject"))
					if i > 0 {
						assert.Empty(t, got.header.Get("Ce-First"))
					}
					assert.Equal(t, "payload", string(got.body))
				}
			}
		})
	}
}

func TestSendWithSharedCustomHeadersConcurrent(t *testing.T) {
	const count = 16
	requests := make(chan http.Header, count)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		requests <- r.Header.Clone()
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()
	p, err := New(WithTarget(server.URL))
	require.NoError(t, err)
	p.RequestTemplate = &http.Request{
		Method: http.MethodPost,
		Header: http.Header{"X-Custom": {"template"}},
	}
	custom := http.Header{"X-Custom": {"context"}}
	original := custom.Clone()
	ctx := WithCustomHeader(context.Background(), custom)
	results := make(chan error, count)
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < count; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			e := event.New()
			e.SetID(strconv.Itoa(id))
			e.SetSource("/concurrent")
			e.SetType("com.example.headers")
			<-start
			results <- p.Send(binding.WithForceBinary(ctx), binding.ToMessage(&e))
		}(i)
	}
	close(start)
	wg.Wait()
	close(results)
	for result := range results {
		require.True(t, protocol.IsACK(result), "%v", result)
	}
	close(requests)
	seen := make(map[string]bool, count)
	for header := range requests {
		require.Len(t, header.Values("Ce-Id"), 1)
		id := header.Get("Ce-Id")
		assert.False(t, seen[id], "duplicate event ID %s", id)
		seen[id] = true
		assert.Equal(t, []string{"context", "template"}, header.Values("X-Custom"))
	}
	assert.Len(t, seen, count)
	assert.Equal(t, original, custom)
}
