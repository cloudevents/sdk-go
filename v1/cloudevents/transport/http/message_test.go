package http_test

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"

	cehttp "github.com/cloudevents/sdk-go/v1/cloudevents/transport/http"
	"github.com/google/go-cmp/cmp"
)

func TestNewMessage(t *testing.T) {
	h := http.Header{"A": []string{"b"}, "X": []string{"y"}}
	b := ioutil.NopCloser(bytes.NewBuffer([]byte("hello")))
	m, err := cehttp.NewMessage(h, b)
	if err != nil {
		t.Error(err)
	}
	if s := cmp.Diff(h, m.Header); s != "" {
		t.Error(s)
	}
	if s := cmp.Diff("hello", string(m.Body)); s != "" {
		t.Error(s)
	}
	// Make sure the Message map is an independent copy
	h.Set("a", "A")
	if s := cmp.Diff(m.Header.Get("a"), "b"); s != "" {
		t.Error(s)
	}
}

func TestNewResponse(t *testing.T) {
	h := http.Header{"A": []string{"b"}, "X": []string{"y"}}
	b := ioutil.NopCloser(bytes.NewBuffer([]byte("hello")))
	m, err := cehttp.NewResponse(h, b, 42)
	if err != nil {
		t.Error(err)
	}
	if s := cmp.Diff(42, m.StatusCode); s != "" {
		if s := cmp.Diff(h, m.Header); s != "" {
			t.Error(s)
		}
		if s := cmp.Diff("hello", string(m.Body)); s != "" {
			t.Error(s)
		}
	}
}

func TestToRequest(t *testing.T) {
	h := http.Header{"a": []string{"b"}, "x": []string{"y"}}
	b := ioutil.NopCloser(bytes.NewBuffer([]byte("hello")))
	m, err := cehttp.NewMessage(h, b)
	if err != nil {
		t.Error(err)
	}
	var req http.Request
	m.ToRequest(&req)
	if s := cmp.Diff(m.Header, req.Header); s != "" {
		t.Error(s)
	}
	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		t.Error(err)
	}
	if s := cmp.Diff("hello", string(data)); s != "" {
		t.Error(s)
	}
	if s := cmp.Diff(len(data), int(req.ContentLength)); s != "" {
		t.Error(s)
	}
	if s := cmp.Diff("POST", req.Method); s != "" {
		t.Error(s)
	}
}

func TestToResponse(t *testing.T) {
	h := http.Header{"a": []string{"b"}, "x": []string{"y"}}
	b := ioutil.NopCloser(bytes.NewBuffer([]byte("hello")))
	m, err := cehttp.NewResponse(h, b, 42)
	if err != nil {
		t.Error(err)
	}
	var resp http.Response
	m.ToResponse(&resp)
	if s := cmp.Diff(m.Header, resp.Header); s != "" {
		t.Error(s)
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
	}
	if s := cmp.Diff("hello", string(data)); s != "" {
		t.Error(s)
	}
	if s := cmp.Diff(len(data), int(resp.ContentLength)); s != "" {
		t.Error(s)
	}
}
