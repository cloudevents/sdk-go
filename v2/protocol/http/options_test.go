/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package http

import (
	"context"
	"net"
	"net/http"
	"net/url"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/require"
)

func TestWithTarget(t *testing.T) {
	testCases := map[string]struct {
		t       *Protocol
		target  string
		want    *Protocol
		wantErr string
	}{
		"valid url": {
			t: &Protocol{
				RequestTemplate: &http.Request{},
			},
			target: "http://localhost:8080/",
			want: &Protocol{
				Target: func() *url.URL {
					u, _ := url.Parse("http://localhost:8080/")
					return u
				}(),
				RequestTemplate: &http.Request{
					URL: func() *url.URL {
						u, _ := url.Parse("http://localhost:8080/")
						return u
					}(),
				},
			},
		},
		"valid url, unset req": {
			t:      &Protocol{},
			target: "http://localhost:8080/",
			want: &Protocol{
				Target: func() *url.URL {
					u, _ := url.Parse("http://localhost:8080/")
					return u
				}(),
				RequestTemplate: &http.Request{
					Method: http.MethodPost,
					URL: func() *url.URL {
						u, _ := url.Parse("http://localhost:8080/")
						return u
					}(),
				},
			},
		},
		"invalid url": {
			t: &Protocol{
				RequestTemplate: &http.Request{},
			},
			target: "%",
			wantErr: "http target option failed to parse target url: " + func() string {
				//lint:ignore SA1007 I want to create a bad url to test the error message
				_, err := url.Parse("%") //nolint // I want to create a bad url to test the error message
				return err.Error()
			}(),
		},
		"empty target": {
			t: &Protocol{
				RequestTemplate: &http.Request{},
			},
			target:  "",
			wantErr: `http target option was empty string`,
		},
		"whitespace target": {
			t: &Protocol{
				RequestTemplate: &http.Request{},
			},
			target:  " \t\n",
			wantErr: `http target option was empty string`,
		},
		"nil protocol": {
			wantErr: `http target option can not set nil protocol`,
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			err := tc.t.applyOptions(WithTarget(tc.target))

			if tc.wantErr != "" || err != nil {
				var gotErr string
				if err != nil {
					gotErr = err.Error()
				}
				if diff := cmp.Diff(tc.wantErr, gotErr); diff != "" {
					t.Errorf("unexpected error (-want, +got) = %v", diff)
				}
				return
			}

			got := tc.t

			if diff := cmp.Diff(tc.want, got,
				cmpopts.IgnoreUnexported(Protocol{}), cmpopts.IgnoreUnexported(http.Request{})); diff != "" {
				t.Errorf("unexpected (-want, +got) = %v", diff)
			}
		})
	}
}

func TestWithMethod(t *testing.T) {
	testCases := map[string]struct {
		t       *Protocol
		method  string
		want    *Protocol
		wantErr string
	}{
		"valid method": {
			t: &Protocol{
				RequestTemplate: &http.Request{},
			},
			method: "GET",
			want: &Protocol{
				RequestTemplate: &http.Request{
					Method: http.MethodGet,
				},
			},
		},
		"valid method, unset req": {
			t:      &Protocol{},
			method: "PUT",
			want: &Protocol{
				RequestTemplate: &http.Request{
					Method: http.MethodPut,
				},
			},
		},
		"empty method": {
			t: &Protocol{
				RequestTemplate: &http.Request{},
			},
			method:  "",
			wantErr: `http method option was empty string`,
		},
		"whitespace method": {
			t: &Protocol{
				RequestTemplate: &http.Request{},
			},
			method:  " \t\n",
			wantErr: `http method option was empty string`,
		},
		"nil protocol": {
			wantErr: `http method option can not set nil protocol`,
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			err := tc.t.applyOptions(WithMethod(tc.method))

			if tc.wantErr != "" || err != nil {
				var gotErr string
				if err != nil {
					gotErr = err.Error()
				}
				if diff := cmp.Diff(tc.wantErr, gotErr); diff != "" {
					t.Errorf("unexpected error (-want, +got) = %v", diff)
				}
				return
			}

			got := tc.t

			if diff := cmp.Diff(tc.want, got,
				cmpopts.IgnoreUnexported(Protocol{}), cmpopts.IgnoreUnexported(http.Request{})); diff != "" {
				t.Errorf("unexpected (-want, +got) = %v", diff)
			}
		})
	}
}

func TestWithHeader(t *testing.T) {
	testCases := map[string]struct {
		t       *Protocol
		key     string
		value   string
		want    *Protocol
		wantErr string
	}{
		"valid header": {
			t: &Protocol{
				RequestTemplate: &http.Request{},
			},
			key:   "unit",
			value: "test",
			want: &Protocol{
				RequestTemplate: &http.Request{
					Header: http.Header{
						"Unit": {
							"test",
						},
					},
				},
			},
		},
		"valid header, unset req": {
			t:     &Protocol{},
			key:   "unit",
			value: "test",
			want: &Protocol{
				RequestTemplate: &http.Request{
					Method: http.MethodPost,
					Header: http.Header{
						"Unit": {
							"test",
						},
					},
				},
			},
		},
		"empty header key": {
			t: &Protocol{
				RequestTemplate: &http.Request{},
			},
			value:   "test",
			wantErr: `http header option was empty string`,
		},
		"whitespace key": {
			t: &Protocol{
				RequestTemplate: &http.Request{},
			},
			key:     " \t\n",
			value:   "test",
			wantErr: `http header option was empty string`,
		},
		"nil protocol": {
			wantErr: `http header option can not set nil protocol`,
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			err := tc.t.applyOptions(WithHeader(tc.key, tc.value))

			if tc.wantErr != "" || err != nil {
				var gotErr string
				if err != nil {
					gotErr = err.Error()
				}
				if diff := cmp.Diff(tc.wantErr, gotErr); diff != "" {
					t.Errorf("unexpected error (-want, +got) = %v", diff)
				}
				return
			}

			got := tc.t

			if diff := cmp.Diff(tc.want, got,
				cmpopts.IgnoreUnexported(Protocol{}), cmpopts.IgnoreUnexported(http.Request{})); diff != "" {
				t.Errorf("unexpected (-want, +got) = %v", diff)
			}
		})
	}
}

func TestWithShutdownTimeout(t *testing.T) {
	testCases := map[string]struct {
		t       *Protocol
		timeout time.Duration
		want    *Protocol
		wantErr string
	}{
		"valid timeout": {
			t:       &Protocol{},
			timeout: time.Minute * 4,
			want: &Protocol{
				ShutdownTimeout: time.Minute * 4,
			},
		},
		"nil protocol": {
			wantErr: `http shutdown timeout option can not set nil protocol`,
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			err := tc.t.applyOptions(WithShutdownTimeout(tc.timeout))

			if tc.wantErr != "" || err != nil {
				var gotErr string
				if err != nil {
					gotErr = err.Error()
				}
				if diff := cmp.Diff(tc.wantErr, gotErr); diff != "" {
					t.Errorf("unexpected error (-want, +got) = %v", diff)
				}
				return
			}

			got := tc.t

			if diff := cmp.Diff(tc.want, got,
				cmpopts.IgnoreUnexported(Protocol{})); diff != "" {
				t.Errorf("unexpected (-want, +got) = %v", diff)
			}
		})
	}
}

func TestWithPort(t *testing.T) {
	testCases := map[string]struct {
		t       *Protocol
		port    int
		want    *Protocol
		wantErr string
	}{
		"valid port": {
			t:    &Protocol{},
			port: 8181,
			want: &Protocol{
				Port: 8181,
			},
		},
		"invalid port, low": {
			t:       &Protocol{},
			port:    -1,
			wantErr: `http port option was given an invalid port: -1`,
		},
		"invalid port, high": {
			t:       &Protocol{},
			port:    65536,
			wantErr: `http port option was given an invalid port: 65536`,
		},
		"listener already set": {
			t: &Protocol{
				listener: func() atomic.Value {
					l, _ := net.Listen("tcp", ":0")
					v := atomic.Value{}
					v.Store(l)
					return v
				}(),
			},
			port:    8181,
			wantErr: `error setting http port option: listener already set`,
		},
		"nil protocol": {
			wantErr: `http port option can not set nil protocol`,
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			err := tc.t.applyOptions(WithPort(tc.port))

			if tc.wantErr != "" || err != nil {
				var gotErr string
				if err != nil {
					gotErr = err.Error()
				}
				if diff := cmp.Diff(tc.wantErr, gotErr); diff != "" {
					t.Errorf("unexpected error (-want, +got) = %v", diff)
				}
				return
			}

			got := tc.t

			if diff := cmp.Diff(tc.want, got,
				cmpopts.IgnoreUnexported(Protocol{}), cmpopts.IgnoreUnexported(http.Request{})); diff != "" {
				t.Errorf("unexpected (-want, +got) = %v", diff)
			}
		})
	}
}

// Force a transport to close its server/listener by cancelling StartReceiver
func forceClose(tr *Protocol) {
	ctx, cancel := context.WithCancel(context.Background())
	go func() { _ = tr.OpenInbound(ctx) }()
	cancel()
}

func TestWithPort0(t *testing.T) {
	testCases := map[string]func() (*Protocol, error){
		"WithPort0": func() (*Protocol, error) { return New(WithPort(0)) },
		"SetPort0":  func() (*Protocol, error) { return &Protocol{Port: 0}, nil },
	}
	for name, f := range testCases {
		t.Run(name, func(t *testing.T) {
			tr, err := f()
			if err != nil {
				t.Fatal(err)
			}
			defer func() { forceClose(tr) }()
			// Start listening
			listener, err := tr.listen()
			require.NoError(t, err)
			require.NotNil(t, listener)

			// Check the listening port is correct
			port := tr.GetListeningPort()
			require.Greater(t, port, 0)
		})
	}
}

func TestWithListener_forcecloser(t *testing.T) {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal(err)
	}
	tr, err := New(WithListener(l))
	defer func() { forceClose(tr) }()
	if err != nil {
		t.Fatal(err)
	}
	port := tr.GetListeningPort()
	if port <= 0 {
		t.Error("no dynamic port")
	}
	if d := cmp.Diff(port, l.Addr().(*net.TCPAddr).Port); d != "" {
		t.Error(d)
	}
}

func TestWithListener(t *testing.T) {
	testCases := map[string]struct {
		t        *Protocol
		listener net.Listener
		want     *Protocol
		wantErr  string
	}{
		"valid listener": {
			t: &Protocol{},
			listener: func() net.Listener {
				l, _ := net.Listen("tcp", ":0")
				return l
			}(),
			want: &Protocol{
				Port: 0,
			},
		},
		"listener already set": {
			t: &Protocol{
				listener: func() atomic.Value {
					l, _ := net.Listen("tcp", ":0")
					v := atomic.Value{}
					v.Store(l)
					return v
				}(),
			},
			listener: func() net.Listener {
				l, _ := net.Listen("tcp", ":0")
				return l
			}(),
			wantErr: `error setting http listener: listener already set`,
		},
		"nil protocol": {
			wantErr: `http listener option can not set nil protocol`,
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			err := tc.t.applyOptions(WithListener(tc.listener))

			if tc.wantErr != "" || err != nil {
				var gotErr string
				if err != nil {
					gotErr = err.Error()
				}
				if diff := cmp.Diff(tc.wantErr, gotErr); diff != "" {
					t.Errorf("unexpected error (-want, +got) = %v", diff)
				}
				return
			}

			got := tc.t

			if diff := cmp.Diff(tc.want, got,
				cmpopts.IgnoreUnexported(Protocol{}), cmpopts.IgnoreFields(Protocol{}, "Port"), cmpopts.IgnoreUnexported(http.Request{})); diff != "" {
				t.Errorf("unexpected (-want, +got) = %v", diff)
			}
		})
	}
}

func TestWithPath(t *testing.T) {
	testCases := map[string]struct {
		t       *Protocol
		path    string
		want    *Protocol
		wantErr string
	}{
		"valid path": {
			t:    &Protocol{},
			path: "/test",
			want: &Protocol{
				Path: "/test",
			},
		},
		"invalid path": {
			t:       &Protocol{},
			path:    "",
			wantErr: `http path option was given an invalid path: ""`,
		},
		"nil protocol": {
			wantErr: `http path option can not set nil protocol`,
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			err := tc.t.applyOptions(WithPath(tc.path))

			if tc.wantErr != "" || err != nil {
				var gotErr string
				if err != nil {
					gotErr = err.Error()
				}
				if diff := cmp.Diff(tc.wantErr, gotErr); diff != "" {
					t.Errorf("unexpected error (-want, +got) = %v", diff)
				}
				return
			}

			got := tc.t

			if diff := cmp.Diff(tc.want, got,
				cmpopts.IgnoreUnexported(Protocol{}), cmpopts.IgnoreUnexported(http.Request{})); diff != "" {
				t.Errorf("unexpected (-want, +got) = %v", diff)
			}
		})
	}
}

func TestWithMiddleware(t *testing.T) {
	testCases := map[string]struct {
		t       *Protocol
		wantErr string
	}{
		"nil protocol": {
			wantErr: "http middleware option can not set nil protocol",
		},
		"non-nil protocol": {
			t: &Protocol{},
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			err := tc.t.applyOptions(WithMiddleware(func(next http.Handler) http.Handler {
				return next
			}))
			if tc.wantErr != "" {
				if err == nil || err.Error() != tc.wantErr {
					t.Fatalf("Expected error '%s'. Actual '%v'", tc.wantErr, err)
				}
			} else if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
		})
	}
}

type customTransport struct{}

func (c *customTransport) RoundTrip(*http.Request) (*http.Response, error) {
	panic("implement me")
}

var _ http.RoundTripper = (*customTransport)(nil)

func TestWithRoundTripper(t *testing.T) {
	testCases := map[string]struct {
		t            *Protocol
		roundTripper http.RoundTripper
		wantErr      string
	}{
		"nil protocol": {
			wantErr: "http round tripper option can not set nil protocol",
		},
		"non-nil protocol": {
			t:            &Protocol{},
			roundTripper: &customTransport{},
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			err := tc.t.applyOptions(WithRoundTripper(tc.roundTripper))
			if tc.wantErr != "" {
				if err == nil || err.Error() != tc.wantErr {
					t.Fatalf("Expected error '%s'. Actual '%v'", tc.wantErr, err)
				}
			} else if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
		})
	}
}

func TestWithGetHandlerFunc(t *testing.T) {
	testCases := map[string]struct {
		t       *Protocol
		fn      http.HandlerFunc
		wantErr string
	}{
		"nil protocol": {
			wantErr: "http GET handler func can not set nil protocol",
		},
		"non-nil protocol": {
			t:  &Protocol{},
			fn: func(http.ResponseWriter, *http.Request) {},
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			err := tc.t.applyOptions(WithGetHandlerFunc(tc.fn))
			if tc.wantErr != "" {
				if err == nil || err.Error() != tc.wantErr {
					t.Fatalf("Expected error '%s'. Actual '%v'", tc.wantErr, err)
				}
			} else if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
		})
	}
}

func TestWithOptionsHandlerFunc(t *testing.T) {
	testCases := map[string]struct {
		t       *Protocol
		fn      http.HandlerFunc
		wantErr string
	}{
		"nil protocol": {
			wantErr: "http OPTIONS handler func can not set nil protocol",
		},
		"non-nil protocol": {
			t:  &Protocol{},
			fn: func(http.ResponseWriter, *http.Request) {},
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			err := tc.t.applyOptions(WithOptionsHandlerFunc(tc.fn))
			if tc.wantErr != "" {
				if err == nil || err.Error() != tc.wantErr {
					t.Fatalf("Expected error '%s'. Actual '%v'", tc.wantErr, err)
				}
			} else if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
		})
	}
}

func TestWithDefaultOptionsHandlerFunc(t *testing.T) {
	testCases := map[string]struct {
		t       *Protocol
		fn      http.HandlerFunc
		wantErr string
	}{
		"nil protocol": {
			wantErr: "http OPTIONS handler func can not set nil protocol",
		},
		"non-nil protocol": {
			t:  &Protocol{},
			fn: func(http.ResponseWriter, *http.Request) {},
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			err := tc.t.applyOptions(WithOptionsHandlerFunc(tc.fn))
			if tc.wantErr != "" {
				if err == nil || err.Error() != tc.wantErr {
					t.Fatalf("Expected error '%s'. Actual '%v'", tc.wantErr, err)
				}
			} else if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
		})
	}
}

func TestWithIsRetriableFunc(t *testing.T) {
	testCases := map[string]struct {
		p           *Protocol
		isRetriable IsRetriable
		wantErr     string
	}{
		"nil protocol": {
			wantErr: "isRetriable handler func can not set nil protocol",
		},
		"non-nil protocol, nil handler": {
			p:           &Protocol{},
			isRetriable: nil,
			wantErr:     "isRetriable handler can not be nil",
		},
		"non-nil protocol": {
			p:           &Protocol{},
			isRetriable: func(int) bool { return false },
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			err := tc.p.applyOptions(WithIsRetriableFunc(tc.isRetriable))
			if tc.wantErr != "" {
				if err == nil || err.Error() != tc.wantErr {
					t.Fatalf("Expected error '%s'. Actual '%v'", tc.wantErr, err)
				}
			} else if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
		})
	}
}

func TestWithRequestDataAtContextMiddleware(t *testing.T) {
	const tURL = "https://testhost:8080/test/path"
	const tRemoteAddr = "remote.address:1234"
	const tHost = "testhost:8080"

	tHeader := http.Header{
		"Key": []string{"value"},
	}

	u, err := url.Parse(tURL)
	if err != nil {
		t.Fatal(err)
	}

	tRequestData := &RequestData{
		Host:       tHost,
		URL:        u,
		RemoteAddr: tRemoteAddr,
		Header:     tHeader,
	}

	testCases := map[string]struct {
		t                     *Protocol
		options               []Option
		wantApplyOptionsErr   string
		expectMiddlewareCount int
		expectRequestData     *RequestData
	}{
		"nil protocol": {
			wantApplyOptionsErr: "http middleware option can not set nil protocol",
			options:             []Option{WithRequestDataAtContextMiddleware()},
		},
		"protocol with RequestData middleware": {
			t:                     &Protocol{},
			options:               []Option{WithRequestDataAtContextMiddleware()},
			expectMiddlewareCount: 1,
			expectRequestData:     tRequestData,
		},
		"protocol without RequestData middleware": {
			t:                     &Protocol{},
			expectMiddlewareCount: 0,
			expectRequestData:     nil,
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			err := tc.t.applyOptions(tc.options...)

			if tc.wantApplyOptionsErr != "" {
				if err == nil || err.Error() != tc.wantApplyOptionsErr {
					t.Fatalf("Expected error '%s'. Actual '%v'", tc.wantApplyOptionsErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
				return
			}

			if len(tc.t.middleware) != tc.expectMiddlewareCount {
				t.Fatalf("Expected number of registered middleware %d. Actual '%v'", tc.expectMiddlewareCount, len(tc.t.middleware))
				return
			}

			ms := mockOptionsServer{
				handler: func(w http.ResponseWriter, r *http.Request) {
					rd := RequestDataFromContext(r.Context())
					require.Equal(t, rd, tc.expectRequestData)
				},
			}

			handler := attachMiddleware(ms, tc.t.middleware)

			req, err := http.NewRequest("POST", tURL, nil)
			if err != nil {
				t.Fatal(err)
			}
			req.RemoteAddr = tRemoteAddr
			req.Header = tHeader

			handler.ServeHTTP(nil, req)
		})
	}
}

// mockOptionsServer implements http.Handler passing
// unmodified calls to the internal handler.
type mockOptionsServer struct {
	handler http.HandlerFunc
}

func (m mockOptionsServer) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	m.handler(res, req)
}
