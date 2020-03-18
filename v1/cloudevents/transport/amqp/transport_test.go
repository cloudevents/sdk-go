// +build amqp

package amqp_test

import (
	"context"
	"net/url"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cloudevents/sdk-go/v1/binding/spec"
	"github.com/cloudevents/sdk-go/v1/binding/test"
	ce "github.com/cloudevents/sdk-go/v1/cloudevents"
	"github.com/cloudevents/sdk-go/v1/cloudevents/transport"
	"github.com/cloudevents/sdk-go/v1/cloudevents/transport/amqp"
	"github.com/cloudevents/sdk-go/v1/cloudevents/types"
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
	transport, err := amqp.New(s, addr, opts...)
	if err != nil {
		t.Skipf("ampq.New(%#v): %v", s, err)
	}
	return transport
}

type tester struct {
	s, r transport.Transport
	got  chan interface{} // ce.Event or error
}

func (t *tester) Receive(_ context.Context, e ce.Event, _ *ce.EventResponse) error {
	t.got <- e
	return nil
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
		tester.r.SetReceiver(tester)
		err := tester.r.StartReceiver(context.Background())
		if err != nil {
			got <- err
		}
	}()
	return tester
}

func exurl(e ce.Event) ce.Event {
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
	test.EachEvent(t, test.Events(), func(t *testing.T, e ce.Event) {
		_, _, err := tester.s.Send(ctx, e)
		require.NoError(t, err)
		got := <-tester.got
		test.AssertEventEquals(t, exurl(e), got.(ce.Event))
	})
}

func TestWithEncoding(t *testing.T) {
	ctx := context.Background()
	tester := newTester(t, []amqp.Option{amqp.WithEncoding(amqp.StructuredV03)}, nil)
	defer tester.Close()
	// FIXME(alanconway) problem with JSON round-tripping extensions
	events := test.NoExtensions(test.Events())
	test.EachEvent(t, events, func(t *testing.T, e ce.Event) {
		_, _, err := tester.s.Send(ctx, e)
		require.NoError(t, err)
		got := <-tester.got
		e.Context = spec.V03.Convert(e.Context)
		test.AssertEventEquals(t, exurl(e), got.(ce.Event))
	})
}
