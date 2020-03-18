package http

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/cloudevents/sdk-go/v1/cloudevents"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestWithTarget(t *testing.T) {
	testCases := map[string]struct {
		t       *Transport
		target  string
		want    *Transport
		wantErr string
	}{
		"valid url": {
			t: &Transport{
				Req: &http.Request{},
			},
			target: "http://localhost:8080/",
			want: &Transport{
				Req: &http.Request{
					URL: func() *url.URL {
						u, _ := url.Parse("http://localhost:8080/")
						return u
					}(),
				},
			},
		},
		"valid url, unset req": {
			t:      &Transport{},
			target: "http://localhost:8080/",
			want: &Transport{
				Req: &http.Request{
					Method: http.MethodPost,
					URL: func() *url.URL {
						u, _ := url.Parse("http://localhost:8080/")
						return u
					}(),
				},
			},
		},
		"invalid url": {
			t: &Transport{
				Req: &http.Request{},
			},
			target:  "%",
			wantErr: `http target option failed to parse target url: parse %: invalid URL escape "%"`,
		},
		"empty target": {
			t: &Transport{
				Req: &http.Request{},
			},
			target:  "",
			wantErr: `http target option was empty string`,
		},
		"whitespace target": {
			t: &Transport{
				Req: &http.Request{},
			},
			target:  " \t\n",
			wantErr: `http target option was empty string`,
		},
		"nil transport": {
			wantErr: `http target option can not set nil transport`,
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
				cmpopts.IgnoreUnexported(Transport{}), cmpopts.IgnoreUnexported(http.Request{})); diff != "" {
				t.Errorf("unexpected (-want, +got) = %v", diff)
			}
		})
	}
}

func TestWithMethod(t *testing.T) {
	testCases := map[string]struct {
		t       *Transport
		method  string
		want    *Transport
		wantErr string
	}{
		"valid method": {
			t: &Transport{
				Req: &http.Request{},
			},
			method: "GET",
			want: &Transport{
				Req: &http.Request{
					Method: http.MethodGet,
				},
			},
		},
		"valid method, unset req": {
			t:      &Transport{},
			method: "PUT",
			want: &Transport{
				Req: &http.Request{
					Method: http.MethodPut,
				},
			},
		},
		"empty method": {
			t: &Transport{
				Req: &http.Request{},
			},
			method:  "",
			wantErr: `http method option was empty string`,
		},
		"whitespace method": {
			t: &Transport{
				Req: &http.Request{},
			},
			method:  " \t\n",
			wantErr: `http method option was empty string`,
		},
		"nil transport": {
			wantErr: `http method option can not set nil transport`,
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
				cmpopts.IgnoreUnexported(Transport{}), cmpopts.IgnoreUnexported(http.Request{})); diff != "" {
				t.Errorf("unexpected (-want, +got) = %v", diff)
			}
		})
	}
}

func TestWithHeader(t *testing.T) {
	testCases := map[string]struct {
		t       *Transport
		key     string
		value   string
		want    *Transport
		wantErr string
	}{
		"valid header": {
			t: &Transport{
				Req: &http.Request{},
			},
			key:   "unit",
			value: "test",
			want: &Transport{
				Req: &http.Request{
					Header: http.Header{
						"Unit": {
							"test",
						},
					},
				},
			},
		},
		"valid header, unset req": {
			t:     &Transport{},
			key:   "unit",
			value: "test",
			want: &Transport{
				Req: &http.Request{
					Header: http.Header{
						"Unit": {
							"test",
						},
					},
				},
			},
		},
		"empty header key": {
			t: &Transport{
				Req: &http.Request{},
			},
			value:   "test",
			wantErr: `http header option was empty string`,
		},
		"whitespace key": {
			t: &Transport{
				Req: &http.Request{},
			},
			key:     " \t\n",
			value:   "test",
			wantErr: `http header option was empty string`,
		},
		"nil transport": {
			wantErr: `http header option can not set nil transport`,
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
				cmpopts.IgnoreUnexported(Transport{}), cmpopts.IgnoreUnexported(http.Request{})); diff != "" {
				t.Errorf("unexpected (-want, +got) = %v", diff)
			}
		})
	}
}

func TestWithShutdownTimeout(t *testing.T) {
	testCases := map[string]struct {
		t       *Transport
		timeout time.Duration
		want    *Transport
		wantErr string
	}{
		"valid timeout": {
			t:       &Transport{},
			timeout: time.Minute * 4,
			want: &Transport{
				ShutdownTimeout: durationptr(time.Minute * 4),
			},
		},
		"nil transport": {
			wantErr: `http shutdown timeout option can not set nil transport`,
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
				cmpopts.IgnoreUnexported(Transport{})); diff != "" {
				t.Errorf("unexpected (-want, +got) = %v", diff)
			}
		})
	}
}

func durationptr(duration time.Duration) *time.Duration {
	return &duration
}

func intptr(i int) *int {
	return &i
}

func TestWithPort(t *testing.T) {
	testCases := map[string]struct {
		t       *Transport
		port    int
		want    *Transport
		wantErr string
	}{
		"valid port": {
			t:    &Transport{},
			port: 8181,
			want: &Transport{
				Port: intptr(8181),
			},
		},
		"invalid port, low": {
			t:       &Transport{},
			port:    -1,
			wantErr: `http port option was given an invalid port: -1`,
		},
		"invalid port, high": {
			t:       &Transport{},
			port:    65536,
			wantErr: `http port option was given an invalid port: 65536`,
		},
		"nil transport": {
			wantErr: `http port option can not set nil transport`,
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
				cmpopts.IgnoreUnexported(Transport{}), cmpopts.IgnoreUnexported(http.Request{})); diff != "" {
				t.Errorf("unexpected (-want, +got) = %v", diff)
			}
		})
	}
}

// Force a transport to close its server/listener by cancelling StartReceiver
func forceClose(tr *Transport) {
	ctx, cancel := context.WithCancel(context.Background())
	go func() { _ = tr.StartReceiver(ctx) }()
	cancel()
}

func TestWithPort0(t *testing.T) {
	testCases := map[string]func() (*Transport, error){
		"WithPort0": func() (*Transport, error) { return New(WithPort(0)) },
		"SetPort0":  func() (*Transport, error) { return &Transport{Port: new(int)}, nil },
	}
	for name, f := range testCases {
		t.Run(name, func(t *testing.T) {
			tr, err := f()
			if err != nil {
				t.Fatal(err)
			}
			defer func() { forceClose(tr) }()
			port := tr.GetPort()
			if port <= 0 {
				t.Error("no dynamic port")
			}
			if d := cmp.Diff(port, *tr.Port); d != "" {
				t.Error(d)
			}
		})
	}
}

func TestWithListener(t *testing.T) {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal(err)
	}
	tr, err := New(WithListener(l))
	defer func() { forceClose(tr) }()
	if err != nil {
		t.Fatal(err)
	}
	port := tr.GetPort()
	if port <= 0 {
		t.Error("no dynamic port")
	}
	if d := cmp.Diff(port, l.Addr().(*net.TCPAddr).Port); d != "" {
		t.Error(d)
	}
}

func TestWithPath(t *testing.T) {
	testCases := map[string]struct {
		t       *Transport
		path    string
		want    *Transport
		wantErr string
	}{
		"valid path": {
			t:    &Transport{},
			path: "/test",
			want: &Transport{
				Path: "/test",
			},
		},
		"invalid path": {
			t:       &Transport{},
			path:    "",
			wantErr: `http path option was given an invalid path: ""`,
		},
		"nil transport": {
			wantErr: `http path option can not set nil transport`,
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
				cmpopts.IgnoreUnexported(Transport{}), cmpopts.IgnoreUnexported(http.Request{})); diff != "" {
				t.Errorf("unexpected (-want, +got) = %v", diff)
			}
		})
	}
}

func TestWithEncoding(t *testing.T) {
	testCases := map[string]struct {
		t        *Transport
		encoding Encoding
		want     *Transport
		wantErr  string
	}{
		"valid encoding": {
			t:        &Transport{},
			encoding: StructuredV03,
			want: &Transport{
				Encoding: StructuredV03,
			},
		},
		"nil transport": {
			wantErr: `http encoding option can not set nil transport`,
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			err := tc.t.applyOptions(WithEncoding(tc.encoding))

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
				cmpopts.IgnoreUnexported(Transport{})); diff != "" {
				t.Errorf("unexpected (-want, +got) = %v", diff)
			}
		})
	}
}

func TestWithDefaultEncodingSelector(t *testing.T) {

	fn := func(ctx context.Context, e cloudevents.Event) Encoding {
		return Default
	}

	testCases := map[string]struct {
		t       *Transport
		fn      EncodingSelector
		want    *Transport
		wantErr string
	}{
		"valid fn": {
			t:  &Transport{},
			fn: fn,
			want: &Transport{
				DefaultEncodingSelectionFn: fn,
			},
		},
		"invalid fn": {
			t:       &Transport{},
			fn:      nil,
			wantErr: "http fn for DefaultEncodingSelector was nil",
		},
		"nil transport": {
			wantErr: `http default encoding selector option can not set nil transport`,
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			err := tc.t.applyOptions(WithDefaultEncodingSelector(tc.fn))

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
				cmpopts.IgnoreUnexported(Transport{}),
				cmpopts.IgnoreFields(Transport{}, "DefaultEncodingSelectionFn")); diff != "" {
				t.Errorf("unexpected (-want, +got) = %v", diff)
			}
			if tc.fn == nil {
				if got.DefaultEncodingSelectionFn != nil {
					t.Errorf("expected nil DefaultEncodingSelectionFn")
				}
			} else {
				want := fmt.Sprintf("%v", tc.fn)
				got := fmt.Sprintf("%v", got.DefaultEncodingSelectionFn)
				if got != want {
					t.Errorf("unexpected DefaultEncodingSelectionFn; want: %v; got: %v", want, got)
				}

			}
		})
	}
}

func TestWithBinaryEncoding(t *testing.T) {

	fn := DefaultBinaryEncodingSelectionStrategy

	testCases := map[string]struct {
		t       *Transport
		fn      EncodingSelector
		want    *Transport
		wantErr string
	}{
		"valid": {
			t:  &Transport{},
			fn: fn,
			want: &Transport{
				DefaultEncodingSelectionFn: fn,
			},
		},
		"nil transport": {
			wantErr: `http binary encoding option can not set nil transport`,
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			err := tc.t.applyOptions(WithBinaryEncoding())

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
				cmpopts.IgnoreUnexported(Transport{}),
				cmpopts.IgnoreFields(Transport{}, "DefaultEncodingSelectionFn")); diff != "" {
				t.Errorf("unexpected (-want, +got) = %v", diff)
			}

			if tc.fn == nil {
				if got.DefaultEncodingSelectionFn != nil {
					t.Errorf("expected nil DefaultEncodingSelectionFn")
				}
			} else {
				want := fmt.Sprintf("%v", tc.fn)
				got := fmt.Sprintf("%v", got.DefaultEncodingSelectionFn)
				if got != want {
					t.Errorf("unexpected DefaultEncodingSelectionFn; want: %v; got: %v", want, got)
				}
			}
		})
	}
}

func TestWithStructuredEncoding(t *testing.T) {

	fn := DefaultStructuredEncodingSelectionStrategy

	testCases := map[string]struct {
		t       *Transport
		fn      EncodingSelector
		want    *Transport
		wantErr string
	}{
		"valid": {
			t:  &Transport{},
			fn: fn,
			want: &Transport{
				DefaultEncodingSelectionFn: fn,
			},
		},
		"nil transport": {
			wantErr: `http structured encoding option can not set nil transport`,
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			err := tc.t.applyOptions(WithStructuredEncoding())

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
				cmpopts.IgnoreUnexported(Transport{}),
				cmpopts.IgnoreFields(Transport{}, "DefaultEncodingSelectionFn")); diff != "" {
				t.Errorf("unexpected (-want, +got) = %v", diff)
			}
			if tc.fn == nil {
				if got.DefaultEncodingSelectionFn != nil {
					t.Errorf("expected nil DefaultEncodingSelectionFn")
				}
			} else {
				want := fmt.Sprintf("%v", tc.fn)
				got := fmt.Sprintf("%v", got.DefaultEncodingSelectionFn)
				if got != want {
					t.Errorf("unexpected DefaultEncodingSelectionFn; want: %v; got: %v", want, got)
				}
			}
		})
	}
}

func TestWithMiddleware(t *testing.T) {
	testCases := map[string]struct {
		t       *Transport
		wantErr string
	}{
		"nil transport": {
			wantErr: "http middleware option can not set nil transport",
		},
		"non-nil transport": {
			t: &Transport{},
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

func TestWithLongPollTarget(t *testing.T) {
	testCases := map[string]struct {
		t       *Transport
		target  string
		want    *Transport
		wantErr string
	}{
		"valid url": {
			t: &Transport{
				LongPollReq: &http.Request{},
			},
			target: "http://localhost:8080/",
			want: &Transport{
				LongPollReq: &http.Request{
					URL: func() *url.URL {
						u, _ := url.Parse("http://localhost:8080/")
						return u
					}(),
				},
			},
		},
		"valid url, unset req": {
			t:      &Transport{},
			target: "http://localhost:8080/",
			want: &Transport{
				LongPollReq: &http.Request{
					Method: http.MethodGet,
					URL: func() *url.URL {
						u, _ := url.Parse("http://localhost:8080/")
						return u
					}(),
				},
			},
		},
		"invalid url": {
			t: &Transport{
				LongPollReq: &http.Request{},
			},
			target:  "%",
			wantErr: `http long poll target option failed to parse target url: parse %: invalid URL escape "%"`,
		},
		"empty target": {
			t: &Transport{
				LongPollReq: &http.Request{},
			},
			target:  "",
			wantErr: `http long poll target option was empty string`,
		},
		"whitespace target": {
			t: &Transport{
				LongPollReq: &http.Request{},
			},
			target:  " \t\n",
			wantErr: `http long poll target option was empty string`,
		},
		"nil transport": {
			wantErr: `http long poll target option can not set nil transport`,
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			err := tc.t.applyOptions(WithLongPollTarget(tc.target))

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
				cmpopts.IgnoreUnexported(Transport{}), cmpopts.IgnoreUnexported(http.Request{})); diff != "" {
				t.Errorf("unexpected (-want, +got) = %v", diff)
			}
		})
	}
}
