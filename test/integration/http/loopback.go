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

	"github.com/google/uuid"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/client"
	cehttp "github.com/cloudevents/sdk-go/v2/protocol/http"
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
	now        time.Time
	event      *cloudevents.Event
	resp       *cloudevents.Event
	result     cloudevents.Result
	want       *cloudevents.Event
	wantResult cloudevents.Result
	asSent     *TapValidation
	asRecv     *TapValidation
}

type TapTestCases map[string]TapTest

func ClientLoopback(t *testing.T, tc TapTest, copts ...client.Option) {
	tap := NewTap()
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
	recvCtx, recvCancel := context.WithCancel(context.Background())

	go func() {
		if err := ce.StartReceiver(recvCtx, func() (*cloudevents.Event, cloudevents.Result) {
			return tc.resp, tc.result
		}); err != nil {
			t.Log(err)
		}
	}()

	tc.event.SetExtension(unitTestIDKey, testID)
	got, result := ce.Request(context.Background(), *tc.event)
	if result != nil {
		if tc.wantResult == nil {
			if !cloudevents.IsACK(result) {
				t.Errorf("expected ACK, got %s", result)
			}
		} else if !cloudevents.ResultIs(result, tc.wantResult) {
			t.Fatalf("expected %q, got %q", tc.wantResult, result)
		}
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
