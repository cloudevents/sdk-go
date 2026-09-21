/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package client_test

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/event"
	cehttp "github.com/cloudevents/sdk-go/v2/protocol/http"
)

// Reproduces the receiver stall under concurrent request cancellation.
func TestReceiverConcurrentCancelStall(t *testing.T) {
	p, err := cehttp.New()
	if err != nil {
		t.Fatal(err)
	}
	h, err := cloudevents.NewHTTPReceiveHandler(context.Background(), p, func(_ event.Event) {})
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(h)
	defer srv.Close()

	client := &http.Client{Transport: &http.Transport{
		MaxIdleConns: 300, MaxIdleConnsPerHost: 300, MaxConnsPerHost: 300,
	}}
	defer client.CloseIdleConnections()

	post := func(ctx context.Context) error {
		req, _ := http.NewRequestWithContext(ctx, http.MethodPost, srv.URL, bytes.NewReader([]byte("x")))
		req.Header.Set("Ce-Specversion", "1.0")
		req.Header.Set("Ce-Id", "id")
		req.Header.Set("Ce-Source", "example/uri")
		req.Header.Set("Ce-Type", "example.type")
		req.Header.Set("Content-Type", "application/octet-stream")
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		return resp.Body.Close()
	}

	// A healthy request uses a generous deadline against an instant handler, so
	// the only way it can exceed that deadline is the receiver stall this test
	// guards against: a concurrent request's cancellation orphaning this
	// request's delivery. Ordinary CPU contention adds milliseconds, not
	// seconds, so the full-deadline classifier does not flake under parallel
	// `go test ./...` load.
	const (
		workers  = 12
		duration = 3 * time.Second
		slack    = 30 * time.Second
	)
	var stalled, ok, connErr int64
	var wg sync.WaitGroup
	deadline := time.Now().Add(duration)
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			for i := 0; time.Now().Before(deadline); i++ {
				if w%4 == 0 && i%4 == 0 { // force mid-flight cancellation
					ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
					_ = post(ctx)
					cancel()
					time.Sleep(time.Millisecond)
					continue
				}
				ctx, cancel := context.WithTimeout(context.Background(), slack)
				err := post(ctx)
				cancel()
				switch {
				case errors.Is(err, context.DeadlineExceeded):
					atomic.AddInt64(&stalled, 1)
				case err != nil:
					atomic.AddInt64(&connErr, 1)
				default:
					atomic.AddInt64(&ok, 1)
				}
				time.Sleep(time.Millisecond)
			}
		}(w)
	}
	wg.Wait()
	t.Logf("ok=%d stalled=%d connErr=%d", ok, stalled, connErr)
	if stalled > 0 {
		t.Fatalf("%d healthy instant-handler requests stalled (hit their full deadline) under concurrent cancellation", stalled)
	}
	// Guard against a vacuous pass: the receiver must actually deliver healthy
	// requests successfully, not merely avoid deadline-exceeded errors.
	if ok == 0 {
		t.Fatal("no healthy request completed successfully")
	}
	// A few connection/transport errors are tolerated: forcing mid-flight
	// cancellation churns pooled TCP connections, so a healthy request can
	// occasionally observe a reset from connection-reuse timing under parallel
	// load. That is not the stall this test guards against. Only fail if such
	// errors dominate, which would mean the receiver is fundamentally broken
	// rather than merely stalling.
	if connErr >= ok {
		t.Fatalf("connection/transport errors dominate: connErr=%d >= ok=%d", connErr, ok)
	}
}
