package http

import (
	"context"
	"github.com/cloudevents/sdk-go/pkg/event"
	"github.com/google/uuid"
	"net/http/httptest"
	"testing"

	cloudevents "github.com/cloudevents/sdk-go"
	cehttp "github.com/cloudevents/sdk-go/pkg/cloudevents/transport/http"
)

func ClientMiddleware(t *testing.T, tc TapTest, topts ...cehttp.Option) {
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

	tap.Handler = transport

	ce, err := cloudevents.NewClient(
		transport,
		cloudevents.WithEventDefaulter(AlwaysThen(tc.now)),
		cloudevents.WithoutTracePropagation(),
	)
	if err != nil {
		t.Fatal(err)
	}

	testID := uuid.New().String()
	ctx := cloudevents.ContextWithHeader(context.Background(), UnitTestIDKey, testID)

	recvCtx, recvCancel := context.WithCancel(context.Background())

	go func() {
		if err := ce.StartReceiver(recvCtx, func(event cloudevents.Event, resp *cloudevents.EventResponse) {
			resp.RespondWith(200, &event)
		}); err != nil {
			t.Log(err)
		}
	}()

	var got *cloudevents.Event
	err = ce.Send(ctx, *tc.event, func(e *event.Event) {
		got = e
	})
	if err != nil {
		t.Fatal(err)
	}

	recvCancel()

	AssertEventEquality(t, "response event", tc.want, got)

	if req, ok := tap.Req[testID]; ok {
		AssertTappedEquality(t, "http request", tc.asSent, &req)
	}

	if resp, ok := tap.Resp[testID]; ok {
		AssertTappedEquality(t, "http response", tc.asRecv, &resp)
	}
}
