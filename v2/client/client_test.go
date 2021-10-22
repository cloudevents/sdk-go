/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package client_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/cloudevents/sdk-go/v2/client"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/cloudevents/sdk-go/v2/protocol"
	cehttp "github.com/cloudevents/sdk-go/v2/protocol/http"
	"github.com/cloudevents/sdk-go/v2/types"
)

var (
	// Headers that are added to the response, but we don't want to check in our assertions.
	unimportantHeaders = []string{
		"accept-encoding",
		"content-length",
		"user-agent",
		"connection",
		"traceparent",
		"tracestate",
	}
)

func simpleBinaryClient(target string) client.Client {
	p, err := cehttp.New(cehttp.WithTarget(target))
	if err != nil {
		log.Printf("failed to create protocol, %v", err)
		return nil
	}

	c, err := client.New(p, client.WithForceBinary())
	if err != nil {
		return nil
	}
	return c
}

func simpleStructuredClient(target string) client.Client {
	p, err := cehttp.New(cehttp.WithTarget(target))
	if err != nil {
		log.Printf("failed to create protocol, %v", err)
		return nil
	}

	c, err := client.New(p, client.WithForceStructured())
	if err != nil {
		return nil
	}
	return c
}

func TestClientSend(t *testing.T) {
	now := time.Now()

	structuredEvent := event.Event{
		Context: event.EventContextV03{
			Type:   "unit.test.client",
			Source: *types.ParseURIRef("/unit/test/client"),
			Time:   &types.Timestamp{Time: now},
			ID:     "AABBCCDDEE",
		}.AsV03(),
	}
	_ = structuredEvent.SetData(event.ApplicationJSON, &map[string]interface{}{
		"sq":  42,
		"msg": "hello",
	})

	testCases := map[string]struct {
		c       func(target string) client.Client
		event   event.Event
		resp    *http.Response
		want    *requestValidation
		wantRes string
	}{
		"binary simple v0.3": {
			c: simpleBinaryClient,
			event: func() event.Event {
				e := event.Event{
					Context: event.EventContextV03{
						Type:   "unit.test.client",
						Source: *types.ParseURIRef("/unit/test/client"),
						Time:   &types.Timestamp{Time: now},
						ID:     "AABBCCDDEE",
					}.AsV03(),
				}
				_ = e.SetData(event.ApplicationJSON, &map[string]interface{}{
					"sq":  42,
					"msg": "hello",
				})
				return e
			}(),
			resp: &http.Response{
				StatusCode: http.StatusAccepted,
			},
			want: &requestValidation{
				Headers: map[string][]string{
					"ce-specversion": {"0.3"},
					"ce-id":          {"AABBCCDDEE"},
					"ce-time":        {now.UTC().Format(time.RFC3339Nano)},
					"ce-type":        {"unit.test.client"},
					"ce-source":      {"/unit/test/client"},
					"content-type":   {"application/json"},
				},
				Body: []byte(`{"msg":"hello","sq":42}`),
			},
		},
		"structured simple v0.3": {
			c:     simpleStructuredClient,
			event: structuredEvent,
			resp: &http.Response{
				StatusCode: http.StatusAccepted,
			},
			want: &requestValidation{
				Headers: map[string][]string{
					"content-type": {"application/cloudevents+json"},
				},
				Body: func() []byte {
					b, _ := json.Marshal(structuredEvent)
					return b
				}(),
			},
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			handler := &fakeHandler{
				t:        t,
				response: tc.resp,
				requests: make([]requestValidation, 0),
			}
			server := httptest.NewServer(handler)
			defer server.Close()

			c := tc.c(server.URL)

			result := c.Send(context.TODO(), tc.event)
			if tc.wantRes != "" {
				if result == nil {
					t.Fatalf("failed to return expected error, got nil")
				}
				want := tc.wantRes
				got := result.Error()
				if !strings.Contains(got, want) {
					t.Fatalf("failed to return expected error, got %q, want %q", result, want)
				}
				return
			} else {
				if !protocol.IsACK(result) {
					t.Fatalf("expected ACK, got: %s", result)
				}
			}

			rv := handler.popRequest(t)

			assertEquality(t, server.URL, *tc.want, rv)
		})
	}
}

func simpleBinaryOptions(port int, path string) []cehttp.Option {
	opts := []cehttp.Option{
		cehttp.WithPort(port),
		//cehttp.WithBinaryEncoding(),
	}
	if len(path) > 0 {
		opts = append(opts, cehttp.WithPath(path))
	}
	return opts
}

func simpleStructuredOptions(port int, path string) []cehttp.Option {
	opts := []cehttp.Option{
		cehttp.WithPort(port),
		//cehttp.WithStructuredEncoding(),
	}
	if len(path) > 0 {
		opts = append(opts, cehttp.WithPath(path))
	}
	return opts
}

func TestClientReceive(t *testing.T) {
	now := time.Now()

	testCases := map[string]struct {
		optsFn  func(port int, path string) []cehttp.Option
		req     *requestValidation
		want    event.Event
		wantErr string
	}{
		"binary simple v0.3": {
			optsFn: simpleBinaryOptions,
			req: &requestValidation{
				Headers: map[string][]string{
					"ce-specversion": {"0.3"},
					"ce-id":          {"AABBCCDDEE"},
					"ce-time":        {now.UTC().Format(time.RFC3339Nano)},
					"ce-type":        {"unit.test.client"},
					"ce-source":      {"/unit/test/client"},
					"content-type":   {"application/json"},
				},
				Body: []byte(`{"msg":"hello","sq":"42"}`),
			},
			want: func() event.Event {
				e := event.Event{
					Context: event.EventContextV03{
						Type:   "unit.test.client",
						Source: *types.ParseURIRef("/unit/test/client"),
						Time:   &types.Timestamp{Time: now},
						ID:     "AABBCCDDEE",
					}.AsV03(),
				}
				_ = e.SetData(event.ApplicationJSON, &map[string]string{
					"sq":  "42",
					"msg": "hello",
				})
				return e
			}(),
		},
		"structured simple v0.3": {
			optsFn: simpleStructuredOptions,
			req: &requestValidation{
				Headers: map[string][]string{
					"content-type": {"application/cloudevents+json"},
				},
				Body: []byte(fmt.Sprintf(`{"data":{"msg":"hello","sq":"42"},"datacontenttype":"application/json","id":"AABBCCDDEE","source":"/unit/test/client","specversion":"0.3","time":%q,"type":"unit.test.client"}`,
					now.UTC().Format(time.RFC3339Nano),
				)),
			},
			want: func() event.Event {
				e := event.Event{
					Context: event.EventContextV03{
						Type:   "unit.test.client",
						Source: *types.ParseURIRef("/unit/test/client"),
						Time:   &types.Timestamp{Time: now},
						ID:     "AABBCCDDEE",
					}.AsV03(),
				}
				_ = e.SetData(event.ApplicationJSON, &map[string]string{
					"sq":  "42",
					"msg": "hello",
				})
				return e
			}(),
		},
	}
	for n, tc := range testCases {
		for _, path := range []string{"", "/", "/unittest/"} {
			t.Run(n+" at path "+path, func(t *testing.T) {

				events := make(chan event.Event)

				p, err := cehttp.New(tc.optsFn(0, "")...)
				if err != nil {
					t.Fatal(err)
				}

				c, err := client.New(p)
				if err != nil {
					t.Errorf("failed to make client %s", err.Error())
				}

				ctx, cancel := context.WithCancel(context.TODO())
				go func() {
					err := c.StartReceiver(ctx, func(ctx context.Context, event event.Event) error {
						go func() {
							events <- event
						}()
						return nil
					})
					if err != nil {
						t.Errorf("failed to start receiver %s", err.Error())
					}
				}()
				time.Sleep(1 * time.Second) // let the server start

				target, _ := url.Parse(fmt.Sprintf("http://localhost:%d%s", p.GetListeningPort(), p.GetPath()))

				if tc.wantErr != "" {
					if err == nil {
						t.Fatalf("failed to return expected error, got nil")
					}
					want := tc.wantErr
					got := err.Error()
					if !strings.Contains(got, want) {
						t.Fatalf("failed to return expected error, got %q, want %q", err, want)
					}
					cancel()
					return
				} else {
					if err != nil {
						t.Fatalf("failed to send event %s", err)
					}
				}

				req := &http.Request{
					Method:        "POST",
					URL:           target,
					Header:        tc.req.Headers,
					Body:          ioutil.NopCloser(bytes.NewReader(tc.req.Body)),
					ContentLength: int64(len(tc.req.Body)),
				}

				_, _ = http.DefaultClient.Do(req)

				//// Make a copy of the request.
				//body, err := ioutil.ReadAll(resp.Body)
				//if err != nil {
				//	t.Error("failed to read the request body")
				//}
				//gotResp := requestValidation{
				//	Headers: resp.Header,
				//	Body:    string(body),
				//}
				//
				//_ = gotResp // TODO: check response

				got := <-events

				if diff := cmp.Diff(tc.want.Context, got.Context); diff != "" {
					t.Errorf("unexpected events.Context (-want, +got) = %v", diff)
				}

				if diff := cmp.Diff(tc.want.Data(), got.Data()); diff != "" {
					t.Errorf("unexpected events.Data (-want, +got) = %v", diff)
				}

				// Now stop the client
				cancel()

				// try the request again, expecting an error:

				if _, err := http.DefaultClient.Do(req); err == nil {
					t.Fatalf("expected error to when sending request to stopped client")
				}
			})
		}
	}
}

func TestClientContext(t *testing.T) {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("error creating listener: %v", err)
	}
	defer listener.Close()
	type key string

	c, err := client.NewHTTP(cehttp.WithListener(listener), cehttp.WithMiddleware(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), key("inner"), "bar")
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}))

	if err != nil {
		t.Fatalf("error creating client: %v", err)
	}
	var wg sync.WaitGroup
	wg.Add(1)
	handler := func(ctx context.Context) {
		if v := ctx.Value(key("outer")); v != "foo" {
			t.Errorf("expected context to have outer value, got %v", v)
		}
		if v := ctx.Value(key("inner")); v != "bar" {
			t.Errorf("expected context to have inner value, got %v", v)
		}
		wg.Done()
	}
	go func() {
		c.StartReceiver(context.WithValue(context.Background(), key("outer"), "foo"), handler)
	}()

	body := strings.NewReader(`{"data":{"msg":"hello","sq":"42"},"datacontenttype":"application/json","id":"AABBCCDDEE","source":"/unit/test/client","specversion":"0.3","time":%q,"type":"unit.test.client"}`)
	resp, err := http.Post(fmt.Sprintf("http://%s", listener.Addr().String()), "application/cloudevents+json", body)
	if err != nil {
		t.Errorf("err sending request, response: %v, err: %v", resp, err)
	}

	wg.Wait()
}

type requestValidation struct {
	Host    string
	Headers http.Header
	Body    []byte
}

type fakeHandler struct {
	t        *testing.T
	response *http.Response
	requests []requestValidation
}

func (f *fakeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	// Make a copy of the request.
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		f.t.Error("failed to read the request body")
	}
	f.requests = append(f.requests, requestValidation{
		Host:    r.Host,
		Headers: r.Header,
		Body:    body,
	})

	// Write the response.
	if f.response != nil {
		for h, vs := range f.response.Header {
			for _, v := range vs {
				w.Header().Add(h, v)
			}
		}
		w.WriteHeader(f.response.StatusCode)
		var buf bytes.Buffer
		if f.response.ContentLength > 0 {
			_, _ = buf.ReadFrom(f.response.Body)
			_, _ = w.Write(buf.Bytes())
		}
	} else {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(""))
	}
}

func (f *fakeHandler) popRequest(t *testing.T) requestValidation {
	if len(f.requests) == 0 {
		t.Error("Unable to pop request")
	}
	rv := f.requests[0]
	f.requests = f.requests[1:]
	return rv
}

func assertEquality(t *testing.T, replacementURL string, expected, actual requestValidation) {
	server, err := url.Parse(replacementURL)
	if err != nil {
		t.Errorf("Bad replacement URL: %q", replacementURL)
	}
	expected.Host = server.Host
	canonicalizeHeaders(expected, actual)
	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Errorf("Unexpected difference (-want, +got): %v", diff)
	}
}

func canonicalizeHeaders(rvs ...requestValidation) {
	// HTTP header names are case-insensitive, so normalize them to lower case for comparison.
	for _, rv := range rvs {
		headers := rv.Headers
		for n, v := range headers {
			delete(headers, n)
			ln := strings.ToLower(n)

			if isImportantHeader(ln) {
				headers[ln] = v
			}
		}
	}
}

func isImportantHeader(h string) bool {
	for _, v := range unimportantHeaders {
		if v == h {
			return false
		}
	}
	return true
}
