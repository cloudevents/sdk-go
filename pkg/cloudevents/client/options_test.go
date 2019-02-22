package client

import (
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport/http"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport/nats"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	nethttp "net/http"
	"net/url"

	"testing"
)

func TestWithTarget(t *testing.T) {
	testCases := map[string]struct {
		c       *Client
		target  string
		want    *Client
		wantErr string
	}{
		"valid url": {
			c: &Client{
				transport: &http.Transport{
					Req: &nethttp.Request{},
				},
			},
			target: "http://localhost:8080/",
			want: &Client{transport: &http.Transport{
				Req: &nethttp.Request{
					URL: func() *url.URL {
						u, _ := url.Parse("http://localhost:8080/")
						return u
					}(),
				},
			}},
		},
		"valid url, unset req": {
			c: &Client{
				transport: &http.Transport{},
			},
			target: "http://localhost:8080/",
			want: &Client{transport: &http.Transport{
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
			c: &Client{
				transport: &http.Transport{
					Req: &nethttp.Request{},
				},
			},
			target:  "%",
			wantErr: `client option failed to parse target url: parse %: invalid URL escape "%"`,
		},
		"empty target": {
			c: &Client{
				transport: &http.Transport{
					Req: &nethttp.Request{},
				},
			},
			target:  "",
			wantErr: `target option was empty string`,
		},
		"whitespace target": {
			c: &Client{
				transport: &http.Transport{
					Req: &nethttp.Request{},
				},
			},
			target:  " \t\n",
			wantErr: `target option was empty string`,
		},
		"empty transport": {
			c:       &Client{},
			wantErr: `invalid target client option received for transport type`,
		},
		"wrong transport": {
			c:       &Client{transport: &nats.Transport{}},
			wantErr: `invalid target client option received for transport type`,
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			err := tc.c.applyClientOptions(WithTarget(tc.target))

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
		c       *Client
		method  string
		want    *Client
		wantErr string
	}{
		"valid method": {
			c: &Client{
				transport: &http.Transport{
					Req: &nethttp.Request{},
				},
			},
			method: "GET",
			want: &Client{transport: &http.Transport{
				Req: &nethttp.Request{
					Method: nethttp.MethodGet,
				},
			}},
		},
		"valid method, unset req": {
			c: &Client{
				transport: &http.Transport{},
			},
			method: "PUT",
			want: &Client{transport: &http.Transport{
				Req: &nethttp.Request{
					Method: nethttp.MethodPut,
				},
			}},
		},
		"empty method": {
			c: &Client{
				transport: &http.Transport{
					Req: &nethttp.Request{},
				},
			},
			method:  "",
			wantErr: `method option was empty string`,
		},
		"whitespace method": {
			c: &Client{
				transport: &http.Transport{
					Req: &nethttp.Request{},
				},
			},
			method:  " \t\n",
			wantErr: `method option was empty string`,
		},
		"empty transport": {
			c:       &Client{},
			wantErr: `invalid HTTP method client option received for transport type`,
		},
		"wrong transport": {
			c:       &Client{transport: &nats.Transport{}},
			wantErr: `invalid HTTP method client option received for transport type`,
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			err := tc.c.applyClientOptions(WithHTTPMethod(tc.method))

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

func TestWithPort(t *testing.T) {
	testCases := map[string]struct {
		c       *Client
		port    int
		want    *Client
		wantErr string
	}{
		"valid port": {
			c: &Client{
				transport: &http.Transport{},
			},
			port: 8181,
			want: &Client{transport: &http.Transport{
				Port: 8181,
			}},
		},
		"invalid port": {
			c: &Client{
				transport: &http.Transport{},
			},
			port:    0,
			wantErr: `client option was given an invalid port: 0`,
		},
		"empty transport": {
			c:       &Client{},
			wantErr: `invalid HTTP port client option received for transport type`,
		},
		"wrong transport": {
			c:       &Client{transport: &nats.Transport{}},
			wantErr: `invalid HTTP port client option received for transport type`,
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			err := tc.c.applyClientOptions(WithHTTPPort(tc.port))

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
		c        *Client
		encoding http.Encoding
		want     *Client
		wantErr  string
	}{
		"valid encoding": {
			c: &Client{
				transport: &http.Transport{},
			},
			encoding: http.StructuredV03,
			want: &Client{transport: &http.Transport{
				Encoding: http.StructuredV03,
			}},
		},
		"empty transport": {
			c:       &Client{},
			wantErr: `invalid HTTP encoding client option received for transport type`,
		},
		"wrong transport": {
			c:       &Client{transport: &nats.Transport{}},
			wantErr: `invalid HTTP encoding client option received for transport type`,
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			err := tc.c.applyClientOptions(WithHTTPEncoding(tc.encoding))

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

func TestWithNATSEncoding(t *testing.T) {
	testCases := map[string]struct {
		c        *Client
		encoding nats.Encoding
		want     *Client
		wantErr  string
	}{
		"valid encoding": {
			c: &Client{
				transport: &nats.Transport{},
			},
			encoding: nats.StructuredV03,
			want: &Client{transport: &nats.Transport{
				Encoding: nats.StructuredV03,
			}},
		},
		"empty transport": {
			c:       &Client{},
			wantErr: `invalid NATS encoding client option received for transport type`,
		},
		"wrong transport": {
			c:       &Client{transport: &http.Transport{}},
			wantErr: `invalid NATS encoding client option received for transport type`,
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			err := tc.c.applyClientOptions(WithNATSEncoding(tc.encoding))

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
