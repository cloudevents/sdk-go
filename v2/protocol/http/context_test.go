/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package http

import (
	"context"
	nethttp "net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	tMethod = nethttp.MethodPost
)

func TestWithRequest(t *testing.T) {
	testCases := map[string]struct {
		request *nethttp.Request

		expectedRequest *RequestData
	}{
		"request": {
			request: newRequest("http://testhost:8080/test/path.json"),
			expectedRequest: &RequestData{
				Host:   "testhost:8080",
				URL:    newURL("http://testhost:8080/test/path.json"),
				Header: nethttp.Header{},
			},
		},
		"request with headers": {
			request: newRequest("http://testhost:8080/test/path.json",
				requestOptionAddHeader("key1", "value1"),
				requestOptionAddHeader("key2", "value2.1"),
				requestOptionAddHeader("key2", "value2.2"),
			),
			expectedRequest: &RequestData{
				Host: "testhost:8080",
				URL:  newURL("http://testhost:8080/test/path.json"),
				Header: nethttp.Header{
					"Key1": []string{"value1"},
					"Key2": []string{"value2.1", "value2.2"},
				},
			},
		},
		"request with host header": {
			request: newRequest("http://testhost:8080/test/path.json",
				requestOptionHostHeader("alternative.host"),
			),
			expectedRequest: &RequestData{
				Host:   "alternative.host",
				URL:    newURL("http://testhost:8080/test/path.json"),
				Header: nethttp.Header{},
			},
		},
		"request with remote address": {
			request: newRequest("http://testhost:8080/test/path.json",
				requestOptionRemoteAddr("requester.address"),
			),
			expectedRequest: &RequestData{
				Host:       "testhost:8080",
				URL:        newURL("http://testhost:8080/test/path.json"),
				Header:     nethttp.Header{},
				RemoteAddr: "requester.address",
			},
		},
		"nil request": {
			request:         nil,
			expectedRequest: nil,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ctx := WithRequestDataAtContext(context.TODO(), tc.request)

			req := RequestDataFromContext(ctx)
			assert.Equal(t, req, tc.expectedRequest)
		})
	}
}

type requestOption func(*nethttp.Request)

func newRequest(url string, opts ...requestOption) *nethttp.Request {
	r, err := nethttp.NewRequest(tMethod, url, nil)
	if err != nil {
		panic(err)
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

func requestOptionAddHeader(key, value string) requestOption {
	return func(r *nethttp.Request) {
		r.Header.Add(key, value)
	}
}

func requestOptionHostHeader(host string) requestOption {
	return func(r *nethttp.Request) {
		r.Host = host
	}
}

func requestOptionRemoteAddr(addr string) requestOption {
	return func(r *nethttp.Request) {
		r.RemoteAddr = addr
	}
}

func newURL(u string) *url.URL {
	parsed, err := url.Parse(u)
	if err != nil {
		panic(err)
	}
	return parsed
}
