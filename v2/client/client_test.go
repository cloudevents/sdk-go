package client_test

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/lightstep/tracecontext.go/traceparent"
	"go.opencensus.io/trace"

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

func simpleTracingBinaryClient(target string) client.Client {
	p, err := cehttp.New(cehttp.WithTarget(target))
	if err != nil {
		log.Printf("failed to create protocol, %v", err)
		return nil
	}

	c, err := client.NewObserved(p, client.WithTracePropagation())
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
			c: simpleStructuredClient,
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
					"content-type": {"application/cloudevents+json"},
				},
				Body: []byte(fmt.Sprintf(`{"data":{"msg":"hello","sq":42},"datacontenttype":"application/json","id":"AABBCCDDEE","source":"/unit/test/client","specversion":"0.3","time":%q,"type":"unit.test.client"}`,
					now.UTC().Format(time.RFC3339Nano)),
				),
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

func TestTracingClientSend(t *testing.T) {
	now := time.Now()

	testCases := map[string]struct {
		c        func(target string) client.Client
		event    event.Event
		resp     *http.Response
		tpHeader string
		sample   bool
	}{
		"send unsampled": {
			c: simpleTracingBinaryClient,
			event: func() event.Event {
				e := event.Event{
					Context: event.EventContextV1{
						Type:   "unit.test.client",
						Source: *types.ParseURIRef("/unit/test/client"),
						Time:   &types.Timestamp{Time: now},
						ID:     "AABBCCDDEE",
					}.AsV1(),
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
			tpHeader: "ce-traceparent",
		},
		"send sampled": {
			c: simpleTracingBinaryClient,
			event: func() event.Event {
				e := event.Event{
					Context: event.EventContextV1{
						Type:   "unit.test.client",
						Source: *types.ParseURIRef("/unit/test/client"),
						Time:   &types.Timestamp{Time: now},
						ID:     "AABBCCDDEE",
					}.AsV1(),
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
			sample:   true,
			tpHeader: "ce-traceparent",
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

			var sampler trace.Sampler
			if tc.sample {
				sampler = trace.AlwaysSample()
			} else {
				sampler = trace.NeverSample()
			}
			ctx, span := trace.StartSpan(context.TODO(), "test-span", trace.WithSampler(sampler))
			sc := span.SpanContext()

			result := c.Send(ctx, tc.event)
			span.End()

			if !protocol.IsACK(result) {
				t.Fatalf("failed to send event: %s", result)
			}

			rv := handler.popRequest(t)

			var got traceparent.TraceParent
			if tp := rv.Headers.Get(tc.tpHeader); tp == "" {
				t.Fatal("missing traceparent header")
			} else {
				var err error
				got, err = traceparent.ParseString(tp)
				if err != nil {
					t.Fatalf("invalid traceparent: %s", err)
				}
			}
			if got.TraceID != sc.TraceID {
				t.Errorf("unexpected trace id: want %s got %s", sc.TraceID, got.TraceID)
			}
			if got.Flags.Recorded != tc.sample {
				t.Errorf("unexpected recorded flag: want %t got %t", tc.sample, got.Flags.Recorded)
			}
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

func TestTracedClientReceive(t *testing.T) {
	now := time.Now()

	testCases := map[string]struct {
		optsFn func(port int, path string) []cehttp.Option
		event  event.Event
	}{
		"simple binary v0.3": {
			optsFn: simpleBinaryOptions,
			event: func() event.Event {
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
		t.Run(n, func(t *testing.T) {
			spanContexts := make(chan trace.SpanContext)

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
				err = c.StartReceiver(ctx, func(ctx context.Context, e event.Event) (*event.Event, protocol.Result) {
					go func() {
						_, span := client.TraceSpan(ctx, e)
						defer span.End()
						spanContexts <- span.SpanContext()
					}()
					return nil, nil
				})
				if err != nil {
					t.Errorf("failed to start receiver %s", err.Error())
				}
			}()
			time.Sleep(5 * time.Millisecond) // let the server start

			target := fmt.Sprintf("http://localhost:%d", p.GetListeningPort())
			sender := simpleTracingBinaryClient(target)

			ctx, span := trace.StartSpan(context.TODO(), "test-span")
			result := sender.Send(ctx, tc.event)
			span.End()

			if !protocol.IsACK(result) {
				t.Fatalf("failed to send event: %s", result)
			}

			got := <-spanContexts

			if span.SpanContext().TraceID != got.TraceID {
				t.Errorf("unexpected traceID. want: %s, got %s", span.SpanContext().TraceID, got.TraceID)
			}

			// Now stop the client
			cancel()
		})
	}
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
