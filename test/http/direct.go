package http

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	cloudevents "github.com/cloudevents/sdk-go"
	cehttp "github.com/cloudevents/sdk-go/pkg/cloudevents/transport/http"
)

// Direct Test:

//         Obj -> Send -> Wire Format -> Receive -> Got
// Given:   ^                 ^                      ^==Want
// Obj is an event of a version.
// Client is a set to binary or

type DirectTapTest struct {
	Now    time.Time
	Event  *cloudevents.Event
	Want   *cloudevents.Event
	AsSent *TapValidation
}

type DirectTapTestCases map[string]DirectTapTest

func ClientDirect(t *testing.T, tc DirectTapTest, topts ...cehttp.Option) {
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
		cloudevents.WithEventDefaulter(AlwaysThen(tc.Now)),
		cloudevents.WithoutTracePropagation(),
	)
	if err != nil {
		t.Fatal(err)
	}

	testID := uuid.New().String()
	ctx := cloudevents.ContextWithHeader(context.Background(), UnitTestIDKey, testID)

	recvCtx, recvCancel := context.WithTimeout(ctx, time.Second*5)
	defer recvCancel()

	var got *cloudevents.Event
	go func() {
		if err := ce.StartReceiver(recvCtx, func(event cloudevents.Event) {
			got = &event
			recvCancel()
		}); err != nil {
			t.Log(err)
		}
	}()

	err = ce.Send(ctx, *tc.Event)
	if err != nil {
		t.Fatal(err)
	}

	// Wait until the receiver is done.
	<-recvCtx.Done()

	AssertEventEqualityExact(t, "event", tc.Want, got)

	if req, ok := tap.Req[testID]; ok {
		AssertTappedEquality(t, "http request", tc.AsSent, &req)
	}
}
