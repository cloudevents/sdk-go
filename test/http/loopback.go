package http

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/client"
	cehttp "github.com/cloudevents/sdk-go/pkg/cloudevents/transport/http"
)

// Loopback Test:

//         Obj -> Send -> Wire Format -> Receive -> Got
// Given:   ^                 ^                      ^==Want
// Obj is an event of a version.
// Client is a set to binary or

func AlwaysThen(then time.Time) client.EventDefaulter {
	return func(ctx context.Context, event cloudevents.Event) cloudevents.Event {
		if event.Context != nil {
			_ = event.Context.SetTime(then)
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
		if err := ce.StartReceiver(recvCtx, func(resp *cloudevents.EventResponse) {
			if tc.resp != nil {
				resp.RespondWith(200, tc.resp)
			}
		}); err != nil {
			t.Log(err)
		}
	}()

	_, got, err := ce.Send(ctx, *tc.event)
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
