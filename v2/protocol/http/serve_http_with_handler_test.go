/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package http

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/protocol"
)

// ceRequest builds a minimal valid binary-mode CloudEvent POST request.
func ceRequest(t *testing.T, url string) *http.Request {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader([]byte("x")))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Ce-Specversion", "1.0")
	req.Header.Set("Ce-Id", "id")
	req.Header.Set("Ce-Source", "example/uri")
	req.Header.Set("Ce-Type", "example.type")
	req.Header.Set("Content-Type", "application/octet-stream")
	return req
}

func serveWith(p *Protocol, handle func(context.Context, binding.Message, protocol.ResponseFn) error) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		p.ServeHTTPWithHandler(rw, req, handle)
	}))
}

// A handler that reports an error without ever invoking the response callback
// must not wedge the request goroutine on wg.Wait forever; the request must
// still complete.
func TestServeHTTPWithHandler_ErrorWithoutRespondCompletes(t *testing.T) {
	p, err := New()
	if err != nil {
		t.Fatal(err)
	}
	srv := serveWith(p, func(ctx context.Context, m binding.Message, respFn protocol.ResponseFn) error {
		_ = m.Finish(nil)
		return errors.New("boom") // never invokes respFn
	})
	defer srv.Close()

	done := make(chan int, 1)
	go func() {
		resp, err := http.DefaultClient.Do(ceRequest(t, srv.URL))
		if err != nil {
			done <- -1
			return
		}
		_ = resp.Body.Close()
		done <- resp.StatusCode
	}()

	select {
	case status := <-done:
		if status != http.StatusInternalServerError {
			t.Fatalf("want 500 for handler error without response, got %d", status)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("request hung: an error returned without responding deadlocked ServeHTTPWithHandler")
	}
}

// A handler that responds and then also returns an error must yield exactly one
// response (no panic from a double wg.Done / double write).
func TestServeHTTPWithHandler_RespondThenErrorSingleResponse(t *testing.T) {
	p, err := New()
	if err != nil {
		t.Fatal(err)
	}
	srv := serveWith(p, func(ctx context.Context, m binding.Message, respFn protocol.ResponseFn) error {
		_ = m.Finish(nil)
		_ = respFn(ctx, nil, nil) // 200
		return errors.New("late error")
	})
	defer srv.Close()

	resp, err := http.DefaultClient.Do(ceRequest(t, srv.URL))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("want 200, got %d", resp.StatusCode)
	}
}

// A handler that responds asynchronously must still work: ServeHTTPWithHandler
// blocks until the callback runs, so the client observes the written response
// rather than an empty one finalized early by net/http.
func TestServeHTTPWithHandler_AsyncRespond(t *testing.T) {
	p, err := New()
	if err != nil {
		t.Fatal(err)
	}
	srv := serveWith(p, func(ctx context.Context, m binding.Message, respFn protocol.ResponseFn) error {
		go func() {
			time.Sleep(20 * time.Millisecond)
			_ = m.Finish(nil)
			_ = respFn(ctx, nil, nil) // 200
		}()
		return nil
	})
	defer srv.Close()

	resp, err := http.DefaultClient.Do(ceRequest(t, srv.URL))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("want 200 from async response, got %d", resp.StatusCode)
	}
}
