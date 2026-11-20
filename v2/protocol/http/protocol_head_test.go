/*
 Copyright 2026 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package http

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// finishIfReceived replies to one event queued by ServeHTTP.
// A HEAD request that stays on the receive path blocks in ServeHTTP until
// that reply, so tests drain p.incoming with a cancellable goroutine so a
// failing assertion neither hangs nor leaks.
func finishIfReceived(p *Protocol, status int) (received func() bool, stop func()) {
	got := make(chan struct{}, 1)
	stopCh := make(chan struct{})
	var once sync.Once
	go func() {
		select {
		case in := <-p.incoming:
			// Signal before respFn. ServeHTTP returns only after respFn, so a
			// check after ServeHTTP observes this send.
			got <- struct{}{}
			if in.respFn != nil {
				_ = in.respFn(context.Background(), nil, NewResult(status, "accepted"))
			}
		case <-stopCh:
		}
	}()
	return func() bool {
			select {
			case <-got:
				return true
			default:
				return false
			}
		}, func() {
			once.Do(func() { close(stopCh) })
		}
}

func serveHeadWithDeadline(t *testing.T, p *Protocol) (*httptest.ResponseRecorder, *Message) {
	t.Helper()
	gotMsg := make(chan *Message, 1)
	done := make(chan struct{})

	go func() {
		select {
		case in := <-p.incoming:
			gotMsg <- in.msg
			if in.respFn != nil {
				_ = in.respFn(context.Background(), nil, NewResult(http.StatusAccepted, "accepted"))
			}
		case <-done:
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodHead, "http://unittest/health", nil).WithContext(ctx)
	go func() {
		defer close(done)
		p.ServeHTTP(rec, req)
	}()

	select {
	case <-done:
	case <-ctx.Done():
		t.Fatal("ServeHTTP did not return before deadline")
	}

	select {
	case msg := <-gotMsg:
		return rec, msg
	default:
		t.Fatalf("HEAD with a nil HeadHandlerFn was not delivered to the event receive path (status=%d)", rec.Code)
		return nil, nil
	}
}

func TestServeHTTP_HEAD_WithHeadHandler_DoesNotCallGet(t *testing.T) {
	var headCalls, getCalls atomic.Int32
	var sawMethod string
	p, err := New(
		WithHeadHandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			headCalls.Add(1)
			sawMethod = req.Method
			rw.Header().Set("X-Health", "ok")
			rw.WriteHeader(http.StatusNoContent)
		}),
		WithGetHandlerFunc(func(http.ResponseWriter, *http.Request) {
			getCalls.Add(1)
		}),
	)
	require.NoError(t, err)

	received, stop := finishIfReceived(p, http.StatusAccepted)
	defer stop()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodHead, "http://unittest/health", nil).WithContext(ctx)
	done := make(chan struct{})
	go func() {
		defer close(done)
		p.ServeHTTP(rec, req)
	}()
	select {
	case <-done:
	case <-ctx.Done():
		t.Fatal("ServeHTTP did not return before deadline")
	}

	require.Equal(t, int32(1), headCalls.Load(), "HEAD request should be dispatched to HeadHandlerFn")
	require.Equal(t, int32(0), getCalls.Load(), "HEAD request must not be dispatched to GetHandlerFn")
	require.Equal(t, http.MethodHead, sawMethod)
	require.False(t, received(), "HEAD request must not be enqueued on the event receive path")
	require.Equal(t, http.StatusNoContent, rec.Code)
	require.Equal(t, "ok", rec.Header().Get("X-Health"))
}

func TestServeHTTP_HEAD_WithHeadHandler_Server(t *testing.T) {
	var getCalls atomic.Int32
	p, err := New(
		WithHeadHandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			if req.Method != http.MethodHead {
				http.Error(rw, "unexpected method", http.StatusBadRequest)
				return
			}
			rw.Header().Set("X-Health", "ok")
			_, _ = rw.Write([]byte("healthy"))
		}),
		WithGetHandlerFunc(func(http.ResponseWriter, *http.Request) {
			getCalls.Add(1)
		}),
	)
	require.NoError(t, err)

	received, stop := finishIfReceived(p, http.StatusAccepted)
	defer stop()

	srv := httptest.NewServer(p)
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, srv.URL+"/health", nil)
	require.NoError(t, err)
	resp, err := srv.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	require.Equal(t, int32(0), getCalls.Load(), "HEAD request must not be dispatched to GetHandlerFn")
	require.Equal(t, http.StatusOK, resp.StatusCode, "HEAD request should be dispatched to HeadHandlerFn")
	require.Equal(t, "ok", resp.Header.Get("X-Health"))
	require.Empty(t, body, "net/http Server must not send a HEAD body")
	// chunkWriter.Write eats the body, then writeHeader sets Content-Length
	// from those bytes when the handler did not set one itself.
	require.Equal(t, "7", resp.Header.Get("Content-Length"))
	require.Equal(t, int64(7), resp.ContentLength)
	require.Equal(t, "text/plain; charset=utf-8", resp.Header.Get("Content-Type"))
	require.False(t, received(), "HEAD request must not be enqueued on the event receive path")
}

func TestServeHTTP_HEAD_NilHeadHandler_WithGetHandler_ReceivePath(t *testing.T) {
	var getCalls atomic.Int32
	p, err := New(WithGetHandlerFunc(func(http.ResponseWriter, *http.Request) {
		getCalls.Add(1)
	}))
	require.NoError(t, err)
	require.Nil(t, p.HeadHandlerFn)

	rec, msg := serveHeadWithDeadline(t, p)

	require.NotNil(t, msg, "HEAD with a nil HeadHandlerFn must reach the event receive path")
	require.Equal(t, int32(0), getCalls.Load(), "HEAD must not call GetHandlerFn when HeadHandlerFn is nil")
	require.Equal(t, http.StatusAccepted, rec.Code)
	require.NotEqual(t, http.StatusMethodNotAllowed, rec.Code)
	require.Equal(t, "accepted", rec.Body.String())
}

func TestServeHTTP_HEAD_NilHeadHandler_WithoutGetHandler_ReceivePath(t *testing.T) {
	p, err := New()
	require.NoError(t, err)
	require.Nil(t, p.HeadHandlerFn)
	require.Nil(t, p.GetHandlerFn)

	rec, msg := serveHeadWithDeadline(t, p)

	require.NotNil(t, msg, "HEAD without a HEAD handler must reach the event receive path")
	require.Equal(t, http.StatusAccepted, rec.Code)
	require.NotEqual(t, http.StatusMethodNotAllowed, rec.Code)
	require.Equal(t, "accepted", rec.Body.String())
}

func TestServeHTTP_HEAD_NilHeadHandler_WithoutGetHandler_ServerReceivePath(t *testing.T) {
	p, err := New()
	require.NoError(t, err)

	gotMsg := make(chan *Message, 1)
	stopCh := make(chan struct{})
	defer close(stopCh)
	go func() {
		select {
		case in := <-p.incoming:
			gotMsg <- in.msg
			if in.respFn != nil {
				_ = in.respFn(context.Background(), nil, NewResult(http.StatusAccepted, "accepted"))
			}
		case <-stopCh:
		}
	}()

	srv := httptest.NewServer(p)
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, srv.URL+"/health", nil)
	require.NoError(t, err)
	resp, err := srv.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	select {
	case msg := <-gotMsg:
		require.NotNil(t, msg, "HEAD without a HEAD handler must reach the event receive path")
	default:
		t.Fatal("HEAD without a HEAD handler was not delivered to the event receive path")
	}
	require.Equal(t, http.StatusAccepted, resp.StatusCode)
	require.NotEqual(t, http.StatusMethodNotAllowed, resp.StatusCode)
	require.Empty(t, body)
}
