package http

import (
	"bytes"
	"context"
	"fmt"
	"github.com/google/uuid"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/cloudevents/sdk-go"
	cehttp "github.com/cloudevents/sdk-go/pkg/cloudevents/transport/http"
)

// Test should do the following:
// raw http -> ce receiver -> http invoke
//    A                           B
// this tests that A ~= B
func TestClientInvoke_v03(t *testing.T) {
	now := time.Now()

	testCases := map[string]struct {
		now  time.Time
		send *TapValidation
		recv *TapValidation
	}{
		"Structured Base64 With Extensions v0.3 -> v0.3": {
			now: now,
			send: &TapValidation{
				Method: "POST",
				URI:    "/",
				Header: map[string][]string{
					"content-type": {"application/cloudevents+json"},
				},
				Body: fmt.Sprintf(`{"data":"eyJoZWxsbyI6InVuaXR0ZXN0In0=","datacontentencoding":"base64","datacontenttype":"application/json","id":"ABC-123","source":"/unit/test/client","specversion":"0.3","time":%q,"type":"unit.test.client.sent"}`, now.UTC().Format(time.RFC3339Nano)),
			},
			recv: &TapValidation{
				Method: "POST",
				URI:    "/",
				Header: map[string][]string{
					"ce-specversion":         {"0.3"},
					"ce-id":                  {"ABC-123"},
					"ce-time":                {now.UTC().Format(time.RFC3339Nano)},
					"ce-type":                {"unit.test.client.sent.out"},
					"ce-source":              {"/unit/test/client"},
					"ce-datacontentencoding": {"base64"},
					"content-type":           {"application/json"},
				},
				Body:          `"eyJoZWxsbyI6InVuaXR0ZXN0In0="`, // TODO: this seems wrong.
				ContentLength: 30,
			},
		},
	}

	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			topts := []cehttp.Option{cloudevents.WithStructuredEncoding()}

			tap := NewTap()
			server := httptest.NewServer(tap)
			defer server.Close()

			if len(topts) == 0 {
				topts = append(topts, cloudevents.WithBinaryEncoding())
			}
			topts = append(topts, cloudevents.WithTarget(server.URL))
			topts = append(topts, cloudevents.WithPort(0)) // random port
			transport, err := cloudevents.NewHTTPTransport(
				topts...,
			)
			if err != nil {
				t.Fatal(err)
			}

			tap.handler = transport

			ce, err := cloudevents.NewClient(
				transport,
				cloudevents.WithEventDefaulter(AlwaysThen(tc.now)),
			)
			if err != nil {
				t.Fatal(err)
			}

			testID := uuid.New().String()
			destTestID := uuid.New().String()
			ctx := cloudevents.ContextWithHeader(context.Background(), unitTestIDKey, testID)
			_ = ctx

			recvCtx, recvCancel := context.WithCancel(context.Background())

			go func() {
				t.Log(ce.StartReceiver(recvCtx, func(event cloudevents.Event) {
					// t.Log("got: ", event.String())
					if event.Type() == "unit.test.client.sent" {
						ctx := cloudevents.ContextWithHeader(recvCtx, unitTestIDKey, destTestID)
						event.SetType("unit.test.client.sent.out")
						_, _ = ce.Send(ctx, event)
					}
				}))
			}()

			u, _ := url.Parse(server.URL)

			req := &http.Request{
				Method: tc.send.Method,
				URL:    u,
				Body:   ioutil.NopCloser(bytes.NewBuffer([]byte(tc.send.Body))),
				Header: tc.send.Header,
			}
			req.Header.Add(unitTestIDKey, testID)

			resp, err := server.Client().Do(req)
			if err != nil {
				t.Fatal(err)
			}
			_ = resp // TODO

			recvCancel()

			if req, ok := tap.req[destTestID]; ok {
				assertTappedEquality(t, "http request", tc.recv, &req)
			} else {
				t.Fatalf("failed to find test id %q", testID)
			}
		})
	}
}
