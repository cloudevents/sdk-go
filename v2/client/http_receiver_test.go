/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package client_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	cehttp "github.com/cloudevents/sdk-go/v2/protocol/http"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/client"
	"github.com/stretchr/testify/require"
)

func TestEventReceiverServeHTTP_WithContext(t *testing.T) {
	type ctxKey string
	const ctxKeyTest ctxKey = "testKey"

	middleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			ctx = context.WithValue(ctx, ctxKeyTest, "testValue")
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}

	eventReceiver := func(ctx context.Context) error {
		v, ok := ctx.Value(ctxKeyTest).(string)
		if !ok {
			t.Errorf("invalid context value type: %v", v)
			return errors.New("invalid context")
		}
		if v != "testValue" {
			t.Errorf("invalid context value: %s", v)
			return errors.New("invalid context")
		}
		return nil
	}

	p, err := cloudevents.NewHTTP()
	if err != nil {
		t.Fatal(err)
	}
	httpHandler, err := client.NewHTTPReceiveHandler(context.Background(), p, eventReceiver)
	if err != nil {
		t.Fatal(err)
	}
	c, err := cloudevents.NewClientHTTP()
	if err != nil {
		t.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.Handle("/test", middleware(httpHandler))
	ts := httptest.NewServer(mux)
	defer ts.Close()

	event := cloudevents.NewEvent()
	event.SetSource("testSource")
	event.SetType("testType")
	ctx := context.Background()
	ctx = cloudevents.ContextWithTarget(ctx, ts.URL+"/test")

	result := c.Send(ctx, event)
	require.True(t, cloudevents.IsACK(result))
}

func TestEventReceiverServeHTTP_Options(t *testing.T) {
	p, err := cloudevents.NewHTTP()
	if err != nil {
		t.Fatal(err)
	}
	httpHandler, err := client.NewHTTPReceiveHandler(context.Background(), p, func() {})
	if err != nil {
		t.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.Handle("/test", httpHandler)
	ts := httptest.NewServer(mux)
	defer ts.Close()

	req, err := http.NewRequest(http.MethodOptions, ts.URL+"/test", nil)
	require.NoError(t, err)
	res, err := ts.Client().Do(req)
	t.Logf("foo")
	require.NoError(t, err)
	require.NotNil(t, res)
}

func TestEventReceiverServeHTTP_Webhook(t *testing.T) {
	p, err := cloudevents.NewHTTP(cloudevents.WithDefaultOptionsHandlerFunc([]string{http.MethodPost}, cehttp.DefaultAllowedRate, []string{"*"}, false))
	if err != nil {
		t.Fatal(err)
	}
	p.OptionsHandlerFn = p.OptionsHandler
	httpHandler, err := client.NewHTTPReceiveHandler(context.Background(), p, func() {})
	if err != nil {
		t.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.Handle("/test", httpHandler)
	ts := httptest.NewServer(mux)
	defer ts.Close()

	req, err := http.NewRequest(http.MethodOptions, ts.URL+"/test", nil)
	require.NoError(t, err)
	res, err := ts.Client().Do(req)
	t.Logf("foo")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.StatusCode)
	require.Equal(t, strconv.Itoa(cehttp.DefaultAllowedRate), res.Header.Get("WebHook-Allowed-Rate"))
}

// TestEventReceiverServeHTTP_ConsumerPairing is a regression test for #1224.
//
// Protocol.ServeHTTP hands the inbound message over on an unbuffered channel and blocks
// until a consumer picks it up. NewHTTPReceiveHandler starts exactly one consumer per
// ServeHTTP call, so the two must stay paired.
//
// If a consumer is allowed to exit on request cancellation while the message it was meant
// to pick up is still pending, that pairing is permanently offset: from then on every
// request is served by the *next* request's consumer, so each one blocks until further
// traffic arrives. On a low-traffic endpoint this shows up as requests hanging until the
// client or gateway times out.
//
// The first case below is where the offset is introduced -- and without the fix the very
// first ServeHTTP never returns, because its own consumer is gone and no other request
// follows to take the message.
//
// Note the cancellation cannot simply be dropped either: Protocol.ServeHTTP has paths that
// return without sending anything (rate limiting, OPTIONS, GET -- see the tests above) and
// Protocol.incoming is never closed, so an unconditionally waiting consumer would leak.
func TestEventReceiverServeHTTP_ConsumerPairing(t *testing.T) {
	p, err := cloudevents.NewHTTP()
	require.NoError(t, err)

	var mu sync.Mutex
	var handled []string

	httpHandler, err := client.NewHTTPReceiveHandler(context.Background(), p,
		func(_ context.Context, e cloudevents.Event) {
			mu.Lock()
			handled = append(handled, e.ID())
			mu.Unlock()
		})
	require.NoError(t, err)

	newRequest := func(ctx context.Context, id string) *http.Request {
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"hello":"world"}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("ce-specversion", "1.0")
		req.Header.Set("ce-type", "test.type")
		req.Header.Set("ce-source", "test-source")
		req.Header.Set("ce-id", id)
		return req.WithContext(ctx)
	}

	serve := func(ctx context.Context, id string) <-chan struct{} {
		done := make(chan struct{})
		go func() {
			defer close(done)
			httpHandler.ServeHTTP(httptest.NewRecorder(), newRequest(ctx, id))
		}()
		return done
	}

	requireReturns := func(id string, done <-chan struct{}) {
		t.Helper()
		select {
		case <-done:
		case <-time.After(10 * time.Second):
			t.Fatalf("ServeHTTP did not return for request %q", id)
		}
	}

	// A request whose client has already gone away.
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	requireReturns("cancelled", serve(cancelled, "cancelled"))

	// Healthy requests must each complete on their own, without a following request
	// arriving to unblock them.
	requireReturns("first", serve(context.Background(), "first"))
	requireReturns("second", serve(context.Background(), "second"))

	mu.Lock()
	defer mu.Unlock()
	require.ElementsMatch(t, []string{"cancelled", "first", "second"}, handled)
}
