package nats

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	natsd "github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
)

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

func TestWithConnOptions(t *testing.T) {
	opts := []nats.Option{
		nats.DontRandomize(),
		nats.NoEcho(),
	}

	srv, err := natsd.NewServer(&natsd.Options{Port: -1})
	if err != nil {
		t.Errorf("could not start nats server: %s", err)
	}
	go srv.Start()
	defer srv.Shutdown()

	if !srv.ReadyForConnections(10 * time.Second) {
		t.Errorf("nats server did not start")
	}

	tr, err := New(srv.Addr().String(), "testing", WithConnOptions(opts...))
	if err != nil {
		t.Errorf("connection failed: %s", err)
	}

	if !tr.Conn.Opts.NoRandomize {
		t.Errorf("NoRandomize option was not set")
	}

	if !tr.Conn.Opts.NoEcho {
		t.Errorf("NoEcho option was not set")
	}
}
