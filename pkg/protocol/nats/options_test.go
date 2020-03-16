package nats

import (
	"testing"
	"time"

	natsd "github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
)

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
