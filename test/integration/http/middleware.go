/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package http

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/cloudevents/sdk-go/v2/client"

	"github.com/google/uuid"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	cehttp "github.com/cloudevents/sdk-go/v2/protocol/http"
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

	got, result := ce.Request(context.Background(), *tc.event)
	if result != nil {
		if tc.wantResult == nil {
			if !cloudevents.IsACK(result) {
				t.Errorf("expected ACK, got %s", result)
			}
		} else if !cloudevents.ResultIs(result, tc.wantResult) {
			t.Fatalf("expected %s, got %s", tc.wantResult, result)
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
