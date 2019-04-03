package http

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
)

type TapValidation struct {
	Method        string
	URI           string
	Header        http.Header
	Body          string
	Status        string
	ContentLength int64
}

type tapHandler struct {
	handler http.Handler

	req  map[string]TapValidation
	resp map[string]TapValidation
}

func NewTap() *tapHandler {
	return &tapHandler{
		req:  make(map[string]TapValidation, 10),
		resp: make(map[string]TapValidation, 10),
	}
}

const (
	unitTestIDKey = "Test-Ce-Id"
)

func (t *tapHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	id := r.Header.Get(unitTestIDKey)

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

	rec := httptest.NewRecorder()
	t.handler.ServeHTTP(rec, r)

	resp := rec.Result()
	for k, vs := range resp.Header {
		for _, v := range vs {
			w.Header().Set(k, v)
		}
	}
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
	if from == nil || to == nil {
		return to
	}
	for header, values := range from {
		for _, value := range values {
			to.Add(header, value)
		}
	}
	return to
}
