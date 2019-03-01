package client

import (
	"fmt"
	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport/http"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport/nats"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	nethttp "net/http"
	"net/http/httptest"
	"net/url"

	"testing"
)

func TestWithTarget(t *testing.T) {
	testCases := map[string]struct {
		c       *ceClient
		target  string
		want    *ceClient
		wantErr string
	}{
		"valid url": {
			c: &ceClient{
				transport: &http.Transport{
					Req: &nethttp.Request{},
				},
			},
			target: "http://localhost:8080/",
			want: &ceClient{transport: &http.Transport{
				Req: &nethttp.Request{
					URL: func() *url.URL {
						u, _ := url.Parse("http://localhost:8080/")
						return u
					}(),
				},
			}},
		},
		"valid url, unset req": {
			c: &ceClient{
				transport: &http.Transport{},
			},
			target: "http://localhost:8080/",
			want: &ceClient{transport: &http.Transport{
				Req: &nethttp.Request{
					Method: nethttp.MethodPost,
					URL: func() *url.URL {
						u, _ := url.Parse("http://localhost:8080/")
						return u
					}(),
				},
			}},
		},
		"invalid url": {
			c: &ceClient{
				transport: &http.Transport{
					Req: &nethttp.Request{},
				},
			},
			target:  "%",
			wantErr: `client option failed to parse target url: parse %: invalid URL escape "%"`,
		},
		"empty target": {
			c: &ceClient{
				transport: &http.Transport{
					Req: &nethttp.Request{},
				},
			},
			target:  "",
			wantErr: `target option was empty string`,
		},
		"whitespace target": {
			c: &ceClient{
				transport: &http.Transport{
					Req: &nethttp.Request{},
				},
			},
			target:  " \t\n",
			wantErr: `target option was empty string`,
		},
		"empty transport": {
			c:       &ceClient{},
			wantErr: `invalid target client option received for transport type`,
		},
		"wrong transport": {
			c:       &ceClient{transport: &nats.Transport{}},
			wantErr: `invalid target client option received for transport type`,
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			err := tc.c.applyOptions(WithTarget(tc.target))

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

			got := tc.c

			if diff := cmp.Diff(tc.want.transport, got.transport,
				cmpopts.IgnoreUnexported(http.Transport{}), cmpopts.IgnoreUnexported(nethttp.Request{})); diff != "" {
				t.Errorf("unexpected (-want, +got) = %v", diff)
			}
		})
	}
}

func TestWithHTTPMethod(t *testing.T) {
	testCases := map[string]struct {
		c       *ceClient
		method  string
		want    *ceClient
		wantErr string
	}{
		"valid method": {
			c: &ceClient{
				transport: &http.Transport{
					Req: &nethttp.Request{},
				},
			},
			method: "GET",
			want: &ceClient{transport: &http.Transport{
				Req: &nethttp.Request{
					Method: nethttp.MethodGet,
				},
			}},
		},
		"valid method, unset req": {
			c: &ceClient{
				transport: &http.Transport{},
			},
			method: "PUT",
			want: &ceClient{transport: &http.Transport{
				Req: &nethttp.Request{
					Method: nethttp.MethodPut,
				},
			}},
		},
		"empty method": {
			c: &ceClient{
				transport: &http.Transport{
					Req: &nethttp.Request{},
				},
			},
			method:  "",
			wantErr: `method option was empty string`,
		},
		"whitespace method": {
			c: &ceClient{
				transport: &http.Transport{
					Req: &nethttp.Request{},
				},
			},
			method:  " \t\n",
			wantErr: `method option was empty string`,
		},
		"empty transport": {
			c:       &ceClient{},
			wantErr: `invalid HTTP method client option received for transport type`,
		},
		"wrong transport": {
			c:       &ceClient{transport: &nats.Transport{}},
			wantErr: `invalid HTTP method client option received for transport type`,
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			err := tc.c.applyOptions(WithHTTPMethod(tc.method))

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

			got := tc.c

			if diff := cmp.Diff(tc.want.transport, got.transport,
				cmpopts.IgnoreUnexported(http.Transport{}), cmpopts.IgnoreUnexported(nethttp.Request{})); diff != "" {
				t.Errorf("unexpected (-want, +got) = %v", diff)
			}
		})
	}
}

func TestWithHTTPClient(t *testing.T) {
	testCases := map[string]struct {
		c       *ceClient
		netc    *nethttp.Client
		want    *ceClient
		wantErr string
	}{
		"valid client": {
			c:    &ceClient{transport: &http.Transport{}},
			netc: httptest.NewServer(nil).Client(),
			want: &ceClient{transport: &http.Transport{
				Client: httptest.NewServer(nil).Client(),
			}},
		},
		"nil client": {
			c:       &ceClient{transport: &http.Transport{}},
			wantErr: `client option was given an nil HTTP client`,
		},
		"empty transport": {
			c:       &ceClient{},
			wantErr: `invalid HTTP client client option received for transport type`,
		},
		"wrong transport": {
			c:       &ceClient{transport: &nats.Transport{}},
			wantErr: `invalid HTTP client client option received for transport type`,
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			err := tc.c.applyOptions(WithHTTPClient(tc.netc))

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

			got := tc.c

			if diff := cmp.Diff(tc.want.transport, got.transport,
				cmpopts.IgnoreUnexported(http.Transport{}),
				cmpopts.IgnoreUnexported(nethttp.Transport{})); diff != "" {
				t.Errorf("unexpected (-want, +got) = %v", diff)
			}
		})
	}
}

func TestWithPort(t *testing.T) {
	testCases := map[string]struct {
		c       *ceClient
		port    int
		want    *ceClient
		wantErr string
	}{
		"valid port": {
			c: &ceClient{
				transport: &http.Transport{},
			},
			port: 8181,
			want: &ceClient{transport: &http.Transport{
				Port: 8181,
			}},
		},
		"invalid port": {
			c: &ceClient{
				transport: &http.Transport{},
			},
			port:    0,
			wantErr: `client option was given an invalid port: 0`,
		},
		"empty transport": {
			c:       &ceClient{},
			wantErr: `port: invalid client option received for non-HTTP transport type`,
		},
		"wrong transport": {
			c:       &ceClient{transport: &nats.Transport{}},
			wantErr: `port: invalid client option received for non-HTTP transport type`,
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			err := tc.c.applyOptions(WithHTTPPort(tc.port))

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

			got := tc.c

			if diff := cmp.Diff(tc.want.transport, got.transport,
				cmpopts.IgnoreUnexported(http.Transport{}), cmpopts.IgnoreUnexported(nethttp.Request{})); diff != "" {
				t.Errorf("unexpected (-want, +got) = %v", diff)
			}
		})
	}
}

func TestWithPath(t *testing.T) {
	testCases := map[string]struct {
		c       *ceClient
		path    string
		want    *ceClient
		wantErr string
	}{
		"valid path": {
			c: &ceClient{
				transport: &http.Transport{},
			},
			path: "/test",
			want: &ceClient{transport: &http.Transport{
				Path: "/test",
			}},
		},
		"invalid path": {
			c: &ceClient{
				transport: &http.Transport{},
			},
			path:    "",
			wantErr: `client option was given an invalid path: ""`,
		},
		"empty transport": {
			c:       &ceClient{},
			wantErr: `path: invalid client option received for non-HTTP transport type`,
		},
		"wrong transport": {
			c:       &ceClient{transport: &nats.Transport{}},
			wantErr: `path: invalid client option received for non-HTTP transport type`,
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			err := tc.c.applyOptions(WithHTTPPath(tc.path))

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

			got := tc.c

			if diff := cmp.Diff(tc.want.transport, got.transport,
				cmpopts.IgnoreUnexported(http.Transport{}), cmpopts.IgnoreUnexported(nethttp.Request{})); diff != "" {
				t.Errorf("unexpected (-want, +got) = %v", diff)
			}
		})
	}
}

func TestWithHTTPEncoding(t *testing.T) {
	testCases := map[string]struct {
		c        *ceClient
		encoding http.Encoding
		want     *ceClient
		wantErr  string
	}{
		"valid encoding": {
			c: &ceClient{
				transport: &http.Transport{},
			},
			encoding: http.StructuredV03,
			want: &ceClient{transport: &http.Transport{
				Encoding: http.StructuredV03,
			}},
		},
		"empty transport": {
			c:       &ceClient{},
			wantErr: `invalid HTTP encoding client option received for transport type`,
		},
		"wrong transport": {
			c:       &ceClient{transport: &nats.Transport{}},
			wantErr: `invalid HTTP encoding client option received for transport type`,
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			err := tc.c.applyOptions(WithHTTPEncoding(tc.encoding))

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

			got := tc.c

			if diff := cmp.Diff(tc.want.transport, got.transport,
				cmpopts.IgnoreUnexported(http.Transport{})); diff != "" {
				t.Errorf("unexpected (-want, +got) = %v", diff)
			}
		})
	}
}

func TestWithHTTPDefaultEncodingSelector(t *testing.T) {

	fn := func(e cloudevents.Event) http.Encoding {
		return http.Default
	}

	testCases := map[string]struct {
		c       *ceClient
		fn      http.EncodingSelector
		want    *ceClient
		wantErr string
	}{
		"valid fn": {
			c: &ceClient{
				transport: &http.Transport{},
			},
			fn: fn,
			want: &ceClient{transport: &http.Transport{
				DefaultEncodingSelectionFn: fn,
			}},
		},
		"invalid fn": {
			c: &ceClient{
				transport: &http.Transport{},
			},
			fn:      nil,
			wantErr: "fn for DefaultEncodingSelector was nil",
		},
		"empty transport": {
			c:       &ceClient{},
			wantErr: `invalid HTTP default encoding selector client option received for transport type`,
		},
		"wrong transport": {
			c:       &ceClient{transport: &nats.Transport{}},
			wantErr: `invalid HTTP default encoding selector client option received for transport type`,
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			err := tc.c.applyOptions(WithHTTPDefaultEncodingSelector(tc.fn))

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

			got := tc.c

			if diff := cmp.Diff(tc.want.transport, got.transport,
				cmpopts.IgnoreUnexported(http.Transport{}),
				cmpopts.IgnoreFields(http.Transport{}, "DefaultEncodingSelectionFn")); diff != "" {
				t.Errorf("unexpected (-want, +got) = %v", diff)
			}
			if tt, ok := got.transport.(*http.Transport); ok {
				if tc.fn == nil {
					if tt.DefaultEncodingSelectionFn != nil {
						t.Errorf("expected nil DefaultEncodingSelectionFn")
					}
				} else {
					want := fmt.Sprintf("%v", tc.fn)
					got := fmt.Sprintf("%v", tt.DefaultEncodingSelectionFn)
					if got != want {
						t.Errorf("unexpected DefaultEncodingSelectionFn; want: %v; got: %v", want, got)
					}
				}
			}
		})
	}
}

func TestWithHTTPBinaryEncoding(t *testing.T) {

	fn := http.DefaultBinaryEncodingSelectionStrategy

	testCases := map[string]struct {
		c       *ceClient
		fn      http.EncodingSelector
		want    *ceClient
		wantErr string
	}{
		"valid": {
			c: &ceClient{
				transport: &http.Transport{},
			},
			fn: fn,
			want: &ceClient{transport: &http.Transport{
				DefaultEncodingSelectionFn: fn,
			}},
		},
		"empty transport": {
			c:       &ceClient{},
			wantErr: `invalid HTTP binary encoding client option received for transport type`,
		},
		"wrong transport": {
			c:       &ceClient{transport: &nats.Transport{}},
			wantErr: `invalid HTTP binary encoding client option received for transport type`,
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			err := tc.c.applyOptions(WithHTTPBinaryEncoding())

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

			got := tc.c

			if diff := cmp.Diff(tc.want.transport, got.transport,
				cmpopts.IgnoreUnexported(http.Transport{}),
				cmpopts.IgnoreFields(http.Transport{}, "DefaultEncodingSelectionFn")); diff != "" {
				t.Errorf("unexpected (-want, +got) = %v", diff)
			}
			if tt, ok := got.transport.(*http.Transport); ok {
				if tc.fn == nil {
					if tt.DefaultEncodingSelectionFn != nil {
						t.Errorf("expected nil DefaultEncodingSelectionFn")
					}
				} else {
					want := fmt.Sprintf("%v", tc.fn)
					got := fmt.Sprintf("%v", tt.DefaultEncodingSelectionFn)
					if got != want {
						t.Errorf("unexpected DefaultEncodingSelectionFn; want: %v; got: %v", want, got)
					}
				}
			}
		})
	}
}

func TestWithHTTPStructuredEncoding(t *testing.T) {

	fn := http.DefaultStructuredEncodingSelectionStrategy

	testCases := map[string]struct {
		c       *ceClient
		fn      http.EncodingSelector
		want    *ceClient
		wantErr string
	}{
		"valid": {
			c: &ceClient{
				transport: &http.Transport{},
			},
			fn: fn,
			want: &ceClient{transport: &http.Transport{
				DefaultEncodingSelectionFn: fn,
			}},
		},
		"empty transport": {
			c:       &ceClient{},
			wantErr: `invalid HTTP structured encoding client option received for transport type`,
		},
		"wrong transport": {
			c:       &ceClient{transport: &nats.Transport{}},
			wantErr: `invalid HTTP structured encoding client option received for transport type`,
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			err := tc.c.applyOptions(WithHTTPStructuredEncoding())

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

			got := tc.c

			if diff := cmp.Diff(tc.want.transport, got.transport,
				cmpopts.IgnoreUnexported(http.Transport{}),
				cmpopts.IgnoreFields(http.Transport{}, "DefaultEncodingSelectionFn")); diff != "" {
				t.Errorf("unexpected (-want, +got) = %v", diff)
			}
			if tt, ok := got.transport.(*http.Transport); ok {
				if tc.fn == nil {
					if tt.DefaultEncodingSelectionFn != nil {
						t.Errorf("expected nil DefaultEncodingSelectionFn")
					}
				} else {
					want := fmt.Sprintf("%v", tc.fn)
					got := fmt.Sprintf("%v", tt.DefaultEncodingSelectionFn)
					if got != want {
						t.Errorf("unexpected DefaultEncodingSelectionFn; want: %v; got: %v", want, got)
					}
				}
			}
		})
	}
}

func TestWithNATSEncoding(t *testing.T) {
	testCases := map[string]struct {
		c        *ceClient
		encoding nats.Encoding
		want     *ceClient
		wantErr  string
	}{
		"valid encoding": {
			c: &ceClient{
				transport: &nats.Transport{},
			},
			encoding: nats.StructuredV03,
			want: &ceClient{transport: &nats.Transport{
				Encoding: nats.StructuredV03,
			}},
		},
		"empty transport": {
			c:       &ceClient{},
			wantErr: `invalid NATS encoding client option received for transport type`,
		},
		"wrong transport": {
			c:       &ceClient{transport: &http.Transport{}},
			wantErr: `invalid NATS encoding client option received for transport type`,
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			err := tc.c.applyOptions(WithNATSEncoding(tc.encoding))

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

			got := tc.c

			if diff := cmp.Diff(tc.want.transport, got.transport,
				cmpopts.IgnoreUnexported(nats.Transport{})); diff != "" {
				t.Errorf("unexpected (-want, +got) = %v", diff)
			}
		})
	}
}

func TestWithEventDefaulter(t *testing.T) {

	v1 := func(event cloudevents.Event) cloudevents.Event {
		event.Context = event.Context.AsV01()
		return event
	}

	v2 := func(event cloudevents.Event) cloudevents.Event {
		event.Context = event.Context.AsV02()
		return event
	}

	v3 := func(event cloudevents.Event) cloudevents.Event {
		event.Context = event.Context.AsV03()
		return event
	}

	testCases := map[string]struct {
		c       *ceClient
		fns     []EventDefaulter
		want    int // number of defaulters
		wantErr string
	}{
		"none": {
			c:    &ceClient{},
			want: 0,
		},
		"one": {
			c:    &ceClient{},
			fns:  []EventDefaulter{v1},
			want: 1,
		},
		"three": {
			c:    &ceClient{},
			fns:  []EventDefaulter{v1, v2, v3},
			want: 3,
		},
		"nil fn": {
			c:       &ceClient{},
			fns:     []EventDefaulter{nil},
			wantErr: "client option was given an nil event defaulter",
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			var err error
			for _, fn := range tc.fns {
				err = tc.c.applyOptions(WithEventDefaulter(fn))
				if err != nil {
					break
				}
			}

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

			got := len(tc.c.eventDefaulterFns)

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("unexpected (-want, +got) = %v", diff)
			}
		})
	}
}

func TestWith_Defaulters(t *testing.T) {

	testCases := map[string]struct {
		c       *ceClient
		opts    []Option
		want    int // number of defaulters
		wantErr string
	}{
		"none": {
			c:    &ceClient{},
			want: 0,
		},
		"uuid": {
			c:    &ceClient{},
			opts: []Option{WithUUIDs()},
			want: 1,
		},
		"time": {
			c:    &ceClient{},
			opts: []Option{WithTimeNow()},
			want: 1,
		},
		"uuid and time": {
			c:    &ceClient{},
			opts: []Option{WithUUIDs(), WithTimeNow()},
			want: 2,
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			var err error
			if len(tc.opts) > 0 {
				err = tc.c.applyOptions(tc.opts...)
			}

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

			got := len(tc.c.eventDefaulterFns)

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("unexpected (-want, +got) = %v", diff)
			}
		})
	}
}
