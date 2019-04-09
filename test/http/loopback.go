package http

import (
	"context"
	"github.com/cloudevents/sdk-go"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/client"
	cehttp "github.com/cloudevents/sdk-go/pkg/cloudevents/transport/http"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/types"
	"github.com/google/uuid"
	"net/http/httptest"
	"testing"
	"time"
)

// Loopback Test:

//         Obj -> Send -> Wire Format -> Receive -> Got
// Given:   ^                 ^                      ^==Want
// Obj is an event of a version.
// Client is a set to binary or

func AlwaysThen(then time.Time) client.EventDefaulter {
	return func(event cloudevents.Event) cloudevents.Event {
		if event.Context != nil {
			switch event.Context.GetSpecVersion() {
			case "0.1":
				ec := event.Context.AsV01()
				ec.EventTime = &types.Timestamp{Time: then}
				event.Context = ec
			case "0.2":
				ec := event.Context.AsV02()
				ec.Time = &types.Timestamp{Time: then}
				event.Context = ec
			case "0.3":
				ec := event.Context.AsV03()
				ec.Time = &types.Timestamp{Time: then}
				event.Context = ec
			}
		}
		return event
	}
}

type TapTest struct {
	now    time.Time
	event  *cloudevents.Event
	resp   *cloudevents.Event
	want   *cloudevents.Event
	asSent *TapValidation
	asRecv *TapValidation
}

type TapTestCases map[string]TapTest

func ClientLoopback(t *testing.T, tc TapTest, topts ...cehttp.Option) {
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
	ctx := cloudevents.ContextWithHeader(context.Background(), unitTestIDKey, testID)

	recvCtx, recvCancel := context.WithCancel(context.Background())

	go func() {
		t.Log(ce.StartReceiver(recvCtx, func(resp *cloudevents.EventResponse) {
			if tc.resp != nil {
				resp.RespondWith(200, tc.resp)
			}
		}))
	}()

	got, err := ce.Send(ctx, *tc.event)
	if err != nil {
		t.Fatal(err)
	}

	recvCancel()

	assertEventEquality(t, "response event", tc.want, got)

	if req, ok := tap.req[testID]; ok {
		assertTappedEquality(t, "http request", tc.asSent, &req)
	}

	if resp, ok := tap.resp[testID]; ok {
		assertTappedEquality(t, "http response", tc.asRecv, &resp)
	}
}

// To help with debug, if needed.
func printTap(t *testing.T, tap *tapHandler, testID string) {
	if r, ok := tap.req[testID]; ok {
		t.Log("tap request ", r.URI, r.Method)
		if r.ContentLength > 0 {
			t.Log(" .body: ", r.Body)
		} else {
			t.Log("tap request had no body.")
		}

		if len(r.Header) > 0 {
			for h, vs := range r.Header {
				for _, v := range vs {
					t.Logf(" .header %s: %s", h, v)
				}
			}
		} else {
			t.Log("tap request had no headers.")
		}
	}

	if r, ok := tap.resp[testID]; ok {
		t.Log("tap response.status: ", r.Status)
		if r.ContentLength > 0 {
			t.Log(" .body: ", r.Body)
		} else {
			t.Log("tap response had no body.")
		}

		if len(r.Header) > 0 {
			for h, vs := range r.Header {
				for _, v := range vs {
					t.Logf(" .header %s: %s", h, v)
				}
			}
		} else {
			t.Log("tap response had no headers.")
		}
	}
}
