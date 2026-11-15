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
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// finishIfReceived replies to one event queued by ServeHTTP.
// HEAD with no GET handler blocks in ServeHTTP until that reply, so a test
// that expects the GET handler still has to drain the queue. Otherwise the
// old receive-path behavior hangs the test instead of failing it.
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

func TestServeHTTP_HEAD_WithGetHandler(t *testing.T) {
	var sawMethod string
	p, err := New(WithGetHandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		sawMethod = req.Method
		rw.Header().Set("X-Health", "ok")
		_, _ = rw.Write([]byte("healthy"))
	}))
	require.NoError(t, err)

	received, stop := finishIfReceived(p, http.StatusAccepted)
	defer stop()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodHead, "http://unittest/health", nil)
	p.ServeHTTP(rec, req)

	require.Equal(t, http.MethodHead, sawMethod, "HEAD request should be dispatched to GetHandlerFn")
	require.False(t, received(), "HEAD request must not be enqueued on the event receive path")
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "ok", rec.Header().Get("X-Health"))
	// ResponseRecorder stores the handler write. net/http's Server drops that
	// body for HEAD; TestServeHTTP_HEAD_WithGetHandler_Server covers the wire.
	require.Equal(t, "healthy", rec.Body.String())
}

func TestServeHTTP_HEAD_WithoutGetHandler_ReceivePath(t *testing.T) {
	p, err := New()
	require.NoError(t, err)

	gotMsg := make(chan *Message, 1)
	go func() {
		in := <-p.incoming
		gotMsg <- in.msg
		if in.respFn != nil {
			_ = in.respFn(context.Background(), nil, NewResult(http.StatusAccepted, "accepted"))
		}
	}()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodHead, "http://unittest/health", nil)
	p.ServeHTTP(rec, req)

	select {
	case msg := <-gotMsg:
		require.NotNil(t, msg, "HEAD without a GET handler must reach the event receive path")
	case <-time.After(2 * time.Second):
		t.Fatal("HEAD without a GET handler was not delivered to the event receive path")
	}
	require.Equal(t, http.StatusAccepted, rec.Code)
	require.Equal(t, "accepted", rec.Body.String())
}

func TestServeHTTP_HEAD_WithGetHandler_Server(t *testing.T) {
	p, err := New(WithGetHandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodHead {
			http.Error(rw, "unexpected method", http.StatusBadRequest)
			return
		}
		rw.Header().Set("X-Health", "ok")
		_, _ = rw.Write([]byte("healthy"))
	}))
	require.NoError(t, err)

	received, stop := finishIfReceived(p, http.StatusAccepted)
	defer stop()

	srv := httptest.NewServer(p)
	defer srv.Close()

	req, err := http.NewRequest(http.MethodHead, srv.URL+"/health", nil)
	require.NoError(t, err)
	resp, err := srv.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, resp.StatusCode, "HEAD request should be dispatched to GetHandlerFn")
	require.Equal(t, "ok", resp.Header.Get("X-Health"))
	require.Empty(t, body, "net/http Server must not send a HEAD body")
	// chunkWriter.Write eats the body, then writeHeader sets Content-Length
	// from those bytes when the handler did not set one itself.
	require.Equal(t, "7", resp.Header.Get("Content-Length"))
	require.Equal(t, int64(7), resp.ContentLength)
	require.Equal(t, "text/plain; charset=utf-8", resp.Header.Get("Content-Type"))
	require.False(t, received(), "HEAD request must not be enqueued on the event receive path")
}

func TestServeHTTP_HEAD_WithoutGetHandler_ServerReceivePath(t *testing.T) {
	p, err := New()
	require.NoError(t, err)

	gotMsg := make(chan *Message, 1)
	go func() {
		in := <-p.incoming
		gotMsg <- in.msg
		if in.respFn != nil {
			_ = in.respFn(context.Background(), nil, NewResult(http.StatusAccepted, "accepted"))
		}
	}()

	srv := httptest.NewServer(p)
	defer srv.Close()

	req, err := http.NewRequest(http.MethodHead, srv.URL+"/health", nil)
	require.NoError(t, err)
	resp, err := srv.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	select {
	case msg := <-gotMsg:
		require.NotNil(t, msg, "HEAD without a GET handler must reach the event receive path")
	default:
		t.Fatal("HEAD without a GET handler was not delivered to the event receive path")
	}
	require.Equal(t, http.StatusAccepted, resp.StatusCode)
	require.Empty(t, body)
}
