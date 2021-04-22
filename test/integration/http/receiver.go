/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

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

type ReceiverTapTest struct {
	now                 time.Time
	request             func(url string) *http.Request
	asRecv              *TapValidation
	receiverFuncFactory func(context.CancelFunc) interface{}
	opts                []client.Option
}

type ReceiverTapTestCases map[string]ReceiverTapTest

func ClientReceiver(t *testing.T, tc ReceiverTapTest, copts ...client.Option) {
	tap := NewTap()
	server := httptest.NewServer(tap)
	client := http.Client{}
	defer server.Close()

	opts := make([]cehttp.Option, 0)
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

	recvCtx, recvCancel := context.WithTimeout(context.Background(), time.Second*5)
	defer recvCancel()

	go func() {
		if err := ce.StartReceiver(recvCtx, tc.receiverFuncFactory(recvCancel)); err != nil {
			t.Log(err)
		}
	}()

	testID := uuid.New().String()

	req := tc.request(server.URL)
	req.Header.Set("ce-"+unitTestIDKey, testID)
	_, err = client.Do(req)

	// Wait until the receiver is done.
	<-recvCtx.Done()

	require.NoError(t, err)
	require.Contains(t, tap.resp, testID)

	res := tap.resp[testID]
	assertTappedEquality(t, "http response", tc.asRecv, &res)
}
