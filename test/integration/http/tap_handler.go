/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package http

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

type TapValidation struct {
	Method        string
	URI           string
	Header        http.Header
	Body          string
	BodyContains  []string
	Status        string
	ContentLength int64
}

type tapHandler struct {
	handler    http.Handler
	statusCode int

	req  map[string]TapValidation
	resp map[string]TapValidation
}

func NewTap() *tapHandler {
	return &tapHandler{
		req:  make(map[string]TapValidation, 10),
		resp: make(map[string]TapValidation, 10),
	}
}

// To help with debug, if needed.
func printTap(t *testing.T, tap *tapHandler, testID string) {
	if r, ok := tap.req[testID]; ok {
		t.Log("tap request ", r.URI, r.Method)
		if r.ContentLength > 0 {
			t.Log(" .body: ", r.Body)
		} else {
			t.Log("tap request had no body.")
		}

		if len(r.Header) > 0 {
			for h, vs := range r.Header {
				for _, v := range vs {
					t.Logf(" .header %s: %s", h, v)
				}
			}
		} else {
			t.Log("tap request had no headers.")
		}
	}

	if r, ok := tap.resp[testID]; ok {
		t.Log("tap response.status: ", r.Status)
		if r.ContentLength > 0 {
			t.Log(" .body: ", r.Body)
		} else {
			t.Log("tap response had no body.")
		}

		if len(r.Header) > 0 {
			for h, vs := range r.Header {
				for _, v := range vs {
					t.Logf(" .header %s: %s", h, v)
				}
			}
		} else {
			t.Log("tap response had no headers.")
		}
	}
}

// Don't consider printTap as dead code even if not currently in use.
// We want to keep it for possible future debugging.
var _ = printTap

const (
	unitTestIDKey = "unittestid"
)

func (t *tapHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	id := r.Header.Get("ce-" + unitTestIDKey)
	r.Header.Del("ce-" + unitTestIDKey)

	// Make a copy of the request.
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("failed to read the request body")
	}
	// Set the body back
	r.Body = ioutil.NopCloser(bytes.NewReader(body))

	t.req[id] = TapValidation{
		Method:        r.Method,
		URI:           r.RequestURI,
		Header:        copyHeaders(r.Header),
		Body:          string(body),
		ContentLength: r.ContentLength,
	}

	if t.handler == nil {
		w.WriteHeader(500)
		return
	}
	if t.statusCode > 299 {
		w.WriteHeader(t.statusCode)
		return
	}

	rec := httptest.NewRecorder()
	t.handler.ServeHTTP(rec, r)

	resp := rec.Result()
	for k, vs := range resp.Header {
		for _, v := range vs {
			w.Header().Set(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("failed to read the resp body")
	}
	_, _ = w.Write(body)
	_ = resp.Body.Close()

	t.resp[id] = TapValidation{
		Status:        resp.Status,
		Header:        copyHeaders(resp.Header),
		Body:          string(body),
		ContentLength: resp.ContentLength,
	}
}

func copyHeaders(from http.Header) http.Header {
	to := http.Header{}
	if from == nil {
		return to
	}
	for header, values := range from {
		for _, value := range values {
			to.Add(header, value)
		}
	}
	return to
}
