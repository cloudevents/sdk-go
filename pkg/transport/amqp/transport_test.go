// +build amqp

package amqp_test

import (
	"context"
	"net/url"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cloudevents/sdk-go/pkg/binding/spec"
	"github.com/cloudevents/sdk-go/pkg/event"
	"github.com/cloudevents/sdk-go/pkg/transport"
	"github.com/cloudevents/sdk-go/pkg/transport/amqp"
	"github.com/cloudevents/sdk-go/pkg/types"

	. "github.com/cloudevents/sdk-go/pkg/binding/test"
)

// Requires an external AMQP broker or router, skip if not available.
// The env variable TEST_AMQP_URL provides the URL, default is "/test"
//
// One option is http://qpid.apache.org/components/dispatch-router/indext.html.
// It can be installed from source or from RPMs, see https://qpid.apache.org/packages.html
// Run `qdrouterd` and the tests will work with no further config.
func testTransport(t *testing.T, opts ...amqp.Option) *amqp.Transport {
	t.Helper()
	addr := "test"
	s := os.Getenv("TEST_AMQP_URL")
	if u, err := url.Parse(s); err == nil && u.Path != "" {
		addr = u.Path
	}
	tx, err := amqp.New(s, addr, opts...)
	if err != nil {
		t.Skipf("ampq.New(%#v): %v", s, err)
	}
	return tx
}

type tester struct {
	s, r transport.Transport
	got  chan interface{} // ce.Event or error
}

func (t *tester) Delivery(_ context.Context, e event.Event) (*event.Event, event.Result) {
	t.got <- e
	return nil, nil
}

func (t *tester) Close() {
	_ = t.s.(*amqp.Transport).Close()
	_ = t.r.(*amqp.Transport).Close()
}

func newTester(t *testing.T, sendOpts, recvOpts []amqp.Option) *tester {
	t.Helper()
	tester := &tester{
		s:   testTransport(t, sendOpts...),
		r:   testTransport(t, recvOpts...),
		got: make(chan interface{}),
	}
	got := make(chan interface{}, 100)
	go func() {
		defer func() { close(got) }()
		tester.r.SetDelivery(tester)
		err := tester.r.StartReceiver(context.Background())
		if err != nil {
			got <- err
		}
	}()
	return tester
}

func exurl(e event.Event) event.Event {
	// Flatten exurl to string, AMQP doesn't preserve the URL type.
	// It should preserve other attribute types.
	if s, _ := types.Format(e.Extensions()["exurl"]); s != "" {
		e.SetExtension("exurl", s)
	}
	return e
}

func TestSendReceive(t *testing.T) {
	ctx := context.Background()
	tester := newTester(t, nil, nil)
	defer tester.Close()
	EachEvent(t, Events(), func(t *testing.T, e event.Event) {
		err := tester.s.Send(ctx, e)
		require.NoError(t, err)
		got := <-tester.got
		AssertEventEquals(t, exurl(e), got.(event.Event))
	})
}

func TestWithEncoding(t *testing.T) {
	ctx := context.Background()
	tester := newTester(t, []amqp.Option{amqp.WithEncoding(amqp.StructuredV03)}, nil)
	defer tester.Close()
	// FIXME(alanconway) problem with JSON round-tripping extensions
	events := NoExtensions(Events())
	EachEvent(t, events, func(t *testing.T, e event.Event) {
		err := tester.s.Send(ctx, e)
		require.NoError(t, err)
		got := <-tester.got
		e.Context = spec.V03.Convert(e.Context)
		AssertEventEquals(t, exurl(e), got.(event.Event))
	})
}
