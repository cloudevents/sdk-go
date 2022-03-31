/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package http

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cloudevents/sdk-go/v2/client"

	"github.com/google/uuid"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	cehttp "github.com/cloudevents/sdk-go/v2/protocol/http"
)

// Direct Test:

//         Obj -> Send -> Wire Format -> Receive -> Got
// Given:   ^                 ^                      ^==Want
// Obj is an event of a version.
// Client is a set to binary or

type DirectTapTest struct {
	now                      time.Time
	event                    *cloudevents.Event
	serverReturnedStatusCode int
	want                     *cloudevents.Event
	wantResult               cloudevents.Result
	asSent                   *TapValidation
}

type DirectTapTestCases map[string]DirectTapTest

func ClientDirect(t *testing.T, tc DirectTapTest, copts ...client.Option) {
	tap := NewTap()
	tap.statusCode = tc.serverReturnedStatusCode

	server := httptest.NewServer(tap)
	defer server.Close()

	opts := make([]cehttp.Option, 0)
	opts = append(opts, cloudevents.WithTarget(server.URL))
	opts = append(opts, cloudevents.WithPort(0)) // random port

	protocol, err := cloudevents.NewHTTP(opts...)
	if err != nil {
		t.Fatal(err)
	}

	tap.handler = protocol

	copts = append(copts, cloudevents.WithEventDefaulter(AlwaysThen(tc.now)))

	ce, err := cloudevents.NewClient(protocol, copts...)
	if err != nil {
		t.Fatal(err)
	}

	testID := uuid.New().String()
	tc.event.SetExtension(unitTestIDKey, testID)

	recvCtx, recvCancel := context.WithTimeout(context.Background(), time.Second*5)
	defer recvCancel()

	var got *cloudevents.Event
	go func() {
		if err := ce.StartReceiver(recvCtx, func(event cloudevents.Event) {
			event.SetExtension(unitTestIDKey, nil)
			got = &event
			recvCancel()
		}); err != nil {
			t.Log(err)
		}
	}()

	result := ce.Send(context.Background(), *tc.event)
	if result != nil {
		if tc.wantResult == nil {
			if !cloudevents.IsACK(result) {
				t.Errorf("expected ACK, got %s", result)
			}
		} else if !cloudevents.ResultIs(result, tc.wantResult) {
			t.Errorf("result.IsUndelivered = %v", cloudevents.IsUndelivered(result))
			t.Fatalf("expected %s, got %s", tc.wantResult, result)
		}
	}

	// Wait until the receiver is done.
	<-recvCtx.Done()

	assertEventEqualityExact(t, "event", tc.want, got)

	if req, ok := tap.req[testID]; ok {
		assertTappedEquality(t, "http request", tc.asSent, &req)
	}
}
