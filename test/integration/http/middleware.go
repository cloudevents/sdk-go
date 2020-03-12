package http

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/cloudevents/sdk-go/pkg/transport/http"
)

func ClientMiddleware(t *testing.T, tc TapTest, opts ...interface{}) {
	tap := NewTap()
	server := httptest.NewServer(tap)
	defer server.Close()

	topts := make([]http.Option, 0)
	popts := make([]http.ProtocolOption, 0)
	for _, v := range opts {
		switch o := v.(type) {
		case http.Option:
			topts = append(topts, o)
		case http.ProtocolOption:
			popts = append(popts, o)
		}
	}

	popts = append(popts, cloudevents.WithTarget(server.URL))

	protocol, err := cloudevents.NewHTTPProtocol(popts...)
	if err != nil {
		t.Fatal(err)
	}

	if len(topts) == 0 {
		topts = append(topts, cloudevents.WithEncoding(cloudevents.HTTPBinaryEncoding))
	}
	topts = append(topts, cloudevents.WithPort(0)) // random port
	transport, err := cloudevents.NewHTTPTransport(protocol, topts...)
	if err != nil {
		t.Fatal(err)
	}

	tap.handler = protocol

	ce, err := cloudevents.NewClient(
		transport,
		cloudevents.WithEventDefaulter(AlwaysThen(tc.now)),
		cloudevents.WithoutTracePropagation(),
	)
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
