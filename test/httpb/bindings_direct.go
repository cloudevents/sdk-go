package httpb

import (
	"context"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport/httpb"
	"github.com/cloudevents/sdk-go/test/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	cloudevents "github.com/cloudevents/sdk-go"
)

// Direct Test:

//         Obj -> Send -> Wire Format -> Receive -> Got
// Given:   ^                 ^                      ^==Want
// Obj is an event of a version.
// Client is a set to binary or

func ClientBindingsDirect(t *testing.T, tc http.DirectTapTest, topts ...httpb.Option) {
	tap := http.NewTap()
	server := httptest.NewServer(tap)
	defer server.Close()

	if len(topts) == 0 {
		topts = append(topts, httpb.WithBinaryEncoding())
	}
	topts = append(topts, httpb.WithTarget(server.URL))
	topts = append(topts, httpb.WithPort(0)) // random port
	transport, err := httpb.New(
		topts...,
	)
	if err != nil {
		t.Fatal(err)
	}

	tap.Handler = transport

	ce, err := cloudevents.NewClient(
		transport,
		cloudevents.WithEventDefaulter(http.AlwaysThen(tc.Now)),
		cloudevents.WithoutTracePropagation(),
	)
	if err != nil {
		t.Fatal(err)
	}

	testID := uuid.New().String()
	ctx := cloudevents.ContextWithHeader(context.Background(), http.UnitTestIDKey, testID)

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

	http.AssertEventEqualityExact(t, "event", tc.Want, got)

	if req, ok := tap.Req[testID]; ok {
		http.AssertTappedEquality(t, "http request", tc.AsSent, &req)
	}
}
