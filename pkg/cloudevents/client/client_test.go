package client_test

import (
	"bytes"
	"context"
	"fmt"
	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/client"
	cehttp "github.com/cloudevents/sdk-go/pkg/cloudevents/transport/http"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/types"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

var (
	// Headers that are added to the response, but we don't want to check in our assertions.
	unimportantHeaders = []string{
		"accept-encoding",
		"content-length",
		"user-agent",
		"connection",
	}
)

func simpleBinaryClient(target string) client.Client {
	t, err := cehttp.New(
		cehttp.WithTarget(target),
		cehttp.WithBinaryEncoding(),
	)
	if err != nil {
		return nil
	}

	c, err := client.New(t)
	if err != nil {
		return nil
	}
	return c
}

func simpleStructuredClient(target string) client.Client {
	t, err := cehttp.New(
		cehttp.WithTarget(target),
		cehttp.WithStructuredEncoding(),
	)
	if err != nil {
		return nil
	}

	c, err := client.New(t)
	if err != nil {
		return nil
	}
	return c
}

func TestClientSend(t *testing.T) {
	now := time.Now()

	testCases := map[string]struct {
		c       func(target string) client.Client
		event   cloudevents.Event
		resp    *http.Response
		want    *requestValidation
		wantErr string
	}{
		"binary simple v0.1": {
			c: simpleBinaryClient,
			event: cloudevents.Event{
				Context: cloudevents.EventContextV01{
					EventType: "unit.test.client",
					Source:    *types.ParseURLRef("/unit/test/client"),
					EventTime: &types.Timestamp{Time: now},
					EventID:   "AABBCCDDEE",
				}.AsV01(),
				Data: &map[string]interface{}{
					"sq":  42,
					"msg": "hello",
				},
			},
			resp: &http.Response{
				StatusCode: http.StatusAccepted,
			},
			want: &requestValidation{
				Headers: map[string][]string{
					"ce-cloudeventsversion": {"0.1"},
					"ce-eventid":            {"AABBCCDDEE"},
					"ce-eventtime":          {now.UTC().Format(time.RFC3339Nano)},
					"ce-eventtype":          {"unit.test.client"},
					"ce-source":             {"/unit/test/client"},
					"content-type":          {"application/json"},
				},
				Body: `{"msg":"hello","sq":42}`,
			},
		},
		"binary simple v0.2": {
			c: simpleBinaryClient,
			event: cloudevents.Event{
				Context: cloudevents.EventContextV02{
					Type:   "unit.test.client",
					Source: *types.ParseURLRef("/unit/test/client"),
					Time:   &types.Timestamp{Time: now},
					ID:     "AABBCCDDEE",
				}.AsV02(),
				Data: &map[string]interface{}{
					"sq":  42,
					"msg": "hello",
				},
			},
			resp: &http.Response{
				StatusCode: http.StatusAccepted,
			},
			want: &requestValidation{
				Headers: map[string][]string{
					"ce-specversion": {"0.2"},
					"ce-id":          {"AABBCCDDEE"},
					"ce-time":        {now.UTC().Format(time.RFC3339Nano)},
					"ce-type":        {"unit.test.client"},
					"ce-source":      {"/unit/test/client"},
					"content-type":   {"application/json"},
				},
				Body: `{"msg":"hello","sq":42}`,
			},
		},
		"binary simple v0.3": {
			c: simpleBinaryClient,
			event: cloudevents.Event{
				Context: cloudevents.EventContextV03{
					Type:   "unit.test.client",
					Source: *types.ParseURLRef("/unit/test/client"),
					Time:   &types.Timestamp{Time: now},
					ID:     "AABBCCDDEE",
				}.AsV03(),
				Data: &map[string]interface{}{
					"sq":  42,
					"msg": "hello",
				},
			},
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
				Body: `{"msg":"hello","sq":42}`,
			},
		},
		"structured simple v0.1": {
			c: simpleStructuredClient,
			event: cloudevents.Event{
				Context: cloudevents.EventContextV01{
					EventType: "unit.test.client",
					Source:    *types.ParseURLRef("/unit/test/client"),
					EventTime: &types.Timestamp{Time: now},
					EventID:   "AABBCCDDEE",
				}.AsV01(),
				Data: &map[string]interface{}{
					"sq":  42,
					"msg": "hello",
				},
			},
			resp: &http.Response{
				StatusCode: http.StatusAccepted,
			},
			want: &requestValidation{
				Headers: map[string][]string{
					"content-type": {"application/cloudevents+json"},
				},
				Body: fmt.Sprintf(`{"cloudEventsVersion":"0.1","contentType":"application/json","data":{"msg":"hello","sq":42},"eventID":"AABBCCDDEE","eventTime":%q,"eventType":"unit.test.client","source":"/unit/test/client"}`,
					now.UTC().Format(time.RFC3339Nano),
				),
			},
		},
		"structured simple v0.2": {
			c: simpleStructuredClient,
			event: cloudevents.Event{
				Context: cloudevents.EventContextV02{
					Type:   "unit.test.client",
					Source: *types.ParseURLRef("/unit/test/client"),
					Time:   &types.Timestamp{Time: now},
					ID:     "AABBCCDDEE",
				}.AsV02(),
				Data: &map[string]interface{}{
					"sq":  42,
					"msg": "hello",
				},
			},
			resp: &http.Response{
				StatusCode: http.StatusAccepted,
			},
			want: &requestValidation{
				Headers: map[string][]string{
					"content-type": {"application/cloudevents+json"},
				},
				Body: fmt.Sprintf(`{"contenttype":"application/json","data":{"msg":"hello","sq":42},"id":"AABBCCDDEE","source":"/unit/test/client","specversion":"0.2","time":%q,"type":"unit.test.client"}`,
					now.UTC().Format(time.RFC3339Nano),
				),
			},
		},
		"structured simple v0.3": {
			c: simpleStructuredClient,
			event: cloudevents.Event{
				Context: cloudevents.EventContextV03{
					Type:   "unit.test.client",
					Source: *types.ParseURLRef("/unit/test/client"),
					Time:   &types.Timestamp{Time: now},
					ID:     "AABBCCDDEE",
				}.AsV03(),
				Data: &map[string]interface{}{
					"sq":  42,
					"msg": "hello",
				},
			},
			resp: &http.Response{
				StatusCode: http.StatusAccepted,
			},
			want: &requestValidation{
				Headers: map[string][]string{
					"content-type": {"application/cloudevents+json"},
				},
				Body: fmt.Sprintf(`{"data":{"msg":"hello","sq":42},"datacontenttype":"application/json","id":"AABBCCDDEE","source":"/unit/test/client","specversion":"0.3","time":%q,"type":"unit.test.client"}`,
					now.UTC().Format(time.RFC3339Nano),
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

			_, err := c.Send(context.TODO(), tc.event) // TODO: update test with new returned event
			if tc.wantErr != "" {
				if err == nil {
					t.Fatalf("failed to return expected error, got nil")
				}
				want := tc.wantErr
				got := err.Error()
				if !strings.Contains(got, want) {
					t.Fatalf("failed to return expected error, got %q, want %q", err, want)
				}
				return
			} else {
				if err != nil {
					t.Fatalf("failed to send event %s", err)
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
		cehttp.WithBinaryEncoding(),
	}
	if len(path) > 0 {
		opts = append(opts, cehttp.WithPath(path))
	}
	return opts
}

func simpleStructuredOptions(port int, path string) []cehttp.Option {
	opts := []cehttp.Option{
		cehttp.WithPort(port),
		cehttp.WithStructuredEncoding(),
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
		want    cloudevents.Event
		wantErr string
	}{
		"binary simple v0.1": {
			optsFn: simpleBinaryOptions,
			req: &requestValidation{
				Headers: map[string][]string{
					"ce-cloudeventsversion": {"0.1"},
					"ce-eventid":            {"AABBCCDDEE"},
					"ce-eventtime":          {now.UTC().Format(time.RFC3339Nano)},
					"ce-eventtype":          {"unit.test.client"},
					"ce-source":             {"/unit/test/client"},
					"content-type":          {"application/json"},
				},
				Body: `{"msg":"hello","sq":"42"}`,
			},
			want: cloudevents.Event{
				Context: cloudevents.EventContextV01{
					EventType:   "unit.test.client",
					ContentType: cloudevents.StringOfApplicationJSON(),
					Source:      *types.ParseURLRef("/unit/test/client"),
					EventTime:   &types.Timestamp{Time: now},
					EventID:     "AABBCCDDEE",
				}.AsV01(),
				Data: &map[string]string{
					"sq":  "42",
					"msg": "hello",
				},
			},
		},
		"binary simple v0.2": {
			optsFn: simpleBinaryOptions,
			req: &requestValidation{
				Headers: map[string][]string{
					"ce-specversion": {"0.2"},
					"ce-id":          {"AABBCCDDEE"},
					"ce-time":        {now.UTC().Format(time.RFC3339Nano)},
					"ce-type":        {"unit.test.client"},
					"ce-source":      {"/unit/test/client"},
					"content-type":   {"application/json"},
				},
				Body: `{"msg":"hello","sq":"42"}`,
			},
			want: cloudevents.Event{
				Context: cloudevents.EventContextV02{
					Type:        "unit.test.client",
					ContentType: cloudevents.StringOfApplicationJSON(),
					Source:      *types.ParseURLRef("/unit/test/client"),
					Time:        &types.Timestamp{Time: now},
					ID:          "AABBCCDDEE",
				}.AsV02(),
				Data: &map[string]string{
					"sq":  "42",
					"msg": "hello",
				},
			},
		},
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
				Body: `{"msg":"hello","sq":"42"}`,
			},
			want: cloudevents.Event{
				Context: cloudevents.EventContextV03{
					Type:            "unit.test.client",
					DataContentType: cloudevents.StringOfApplicationJSON(),
					Source:          *types.ParseURLRef("/unit/test/client"),
					Time:            &types.Timestamp{Time: now},
					ID:              "AABBCCDDEE",
				}.AsV03(),
				Data: &map[string]string{
					"sq":  "42",
					"msg": "hello",
				},
			},
		},
		"structured simple v0.1": {
			optsFn: simpleStructuredOptions,
			req: &requestValidation{
				Headers: map[string][]string{
					"content-type": {"application/cloudevents+json"},
				},
				Body: fmt.Sprintf(`{"cloudEventsVersion":"0.1","contentType":"application/json","data":{"msg":"hello","sq":"42"},"eventID":"AABBCCDDEE","eventTime":%q,"eventType":"unit.test.client","source":"/unit/test/client"}`,
					now.UTC().Format(time.RFC3339Nano),
				),
			},
			want: cloudevents.Event{
				Context: cloudevents.EventContextV01{
					EventType:   "unit.test.client",
					ContentType: cloudevents.StringOfApplicationJSON(),
					Source:      *types.ParseURLRef("/unit/test/client"),
					EventTime:   &types.Timestamp{Time: now},
					EventID:     "AABBCCDDEE",
				}.AsV01(),
				Data: &map[string]string{
					"sq":  "42",
					"msg": "hello",
				},
			},
		},
		"structured simple v0.2": {
			optsFn: simpleStructuredOptions,
			req: &requestValidation{
				Headers: map[string][]string{
					"content-type": {"application/cloudevents+json"},
				},
				Body: fmt.Sprintf(`{"contenttype":"application/json","data":{"msg":"hello","sq":"42"},"id":"AABBCCDDEE","source":"/unit/test/client","specversion":"0.2","time":%q,"type":"unit.test.client"}`,
					now.UTC().Format(time.RFC3339Nano),
				),
			},
			want: cloudevents.Event{
				Context: cloudevents.EventContextV02{
					Type:        "unit.test.client",
					ContentType: cloudevents.StringOfApplicationJSON(),
					Source:      *types.ParseURLRef("/unit/test/client"),
					Time:        &types.Timestamp{Time: now},
					ID:          "AABBCCDDEE",
				}.AsV02(),
				Data: &map[string]string{
					"sq":  "42",
					"msg": "hello",
				},
			},
		},
		"structured simple v0.3": {
			optsFn: simpleStructuredOptions,
			req: &requestValidation{
				Headers: map[string][]string{
					"content-type": {"application/cloudevents+json"},
				},
				Body: fmt.Sprintf(`{"data":{"msg":"hello","sq":"42"},"datacontenttype":"application/json","id":"AABBCCDDEE","source":"/unit/test/client","specversion":"0.3","time":%q,"type":"unit.test.client"}`,
					now.UTC().Format(time.RFC3339Nano),
				),
			},
			want: cloudevents.Event{
				Context: cloudevents.EventContextV03{
					Type:            "unit.test.client",
					DataContentType: cloudevents.StringOfApplicationJSON(),
					Source:          *types.ParseURLRef("/unit/test/client"),
					Time:            &types.Timestamp{Time: now},
					ID:              "AABBCCDDEE",
				}.AsV03(),
				Data: &map[string]string{
					"sq":  "42",
					"msg": "hello",
				},
			},
		},
	}

	type startFn func(events chan cloudevents.Event, opts ...client.Option) (context.Context, client.Client, error)

	for n, tc := range testCases {
		for _, path := range []string{"", "/", "/unittest/"} {
			t.Run(n+" at path "+path, func(t *testing.T) {

				events := make(chan cloudevents.Event)

				tp, err := cehttp.New(tc.optsFn(0, path)...)
				if err != nil {
					t.Errorf("failed to make http transport %s", err.Error())
				}

				c, err := client.New(tp)
				if err != nil {
					t.Errorf("failed to make client %s", err.Error())
				}

				ctx, cancel := context.WithCancel(context.TODO())
				go func() {
					err = c.StartReceiver(ctx, func(ctx context.Context, event cloudevents.Event, resp *cloudevents.EventResponse) error {
						go func() {
							events <- event
						}()
						return nil
					})
					if err != nil {
						t.Errorf("failed to start receiver %s", err.Error())
					}
				}()
				time.Sleep(5 * time.Millisecond) // let the server start

				target, _ := url.Parse(fmt.Sprintf("http://localhost:%d%s", tp.GetPort(), tp.GetPath()))

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
					Body:          ioutil.NopCloser(strings.NewReader(tc.req.Body)),
					ContentLength: int64(len([]byte(tc.req.Body))),
				}

				_, err = http.DefaultClient.Do(req)

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

				data := &map[string]string{}
				err = got.DataAs(data)
				if err != nil {
					t.Fatalf("returned unexpected error, got %s", err.Error())
				}

				if diff := cmp.Diff(tc.want.Data, data); diff != "" {
					t.Errorf("unexpected events.Data (-want, +got) = %v", diff)
				}

				// Now stop the client
				cancel()

				// try the request again, expecting an error:

				if _, err = http.DefaultClient.Do(req); err == nil {
					t.Fatalf("expected error to when sending request to stopped client")
				}
			})
		}
	}
}

type requestValidation struct {
	Host    string
	Headers http.Header
	Body    string
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
		Body:    string(body),
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
			buf.ReadFrom(f.response.Body)
			w.Write(buf.Bytes())
		}
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(""))
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
