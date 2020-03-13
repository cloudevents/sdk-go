package http

import (
	"context"
	"github.com/cloudevents/sdk-go/pkg/client"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	cloudevents "github.com/cloudevents/sdk-go"
	cehttp "github.com/cloudevents/sdk-go/pkg/transport/http"
)

func ClientMiddleware(t *testing.T, tc TapTest, copts ...client.Option) {
	tap := NewTap()
	server := httptest.NewServer(tap)
	defer server.Close()

	opts := make([]cehttp.Option, 0)
	opts = append(opts, cloudevents.WithTarget(server.URL))
	opts = append(opts, cloudevents.WithPort(0))

	protocol, err := cloudevents.NewHTTP(opts...)
	if err != nil {
		t.Fatal(err)
	}

	tap.handler = protocol

	copts = append(copts, cloudevents.WithEventDefaulter(AlwaysThen(tc.now)))
	copts = append(copts, cloudevents.WithoutTracePropagation())

	ce, err := cloudevents.NewClient(protocol, copts...)
	if err != nil {
		t.Fatal(err)
	}

	testID := uuid.New().String()
	tc.event.SetExtension(unitTestIDKey, testID)

	recvCtx, recvCancel := context.WithCancel(context.Background())

	go func() {
		if err := ce.StartReceiver(recvCtx, func(event cloudevents.Event) *cloudevents.Event {
			return &event
		}); err != nil {
			t.Log(err)
		}
	}()

	got, err := ce.Request(context.Background(), *tc.event)
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
