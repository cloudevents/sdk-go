package amqp_binding

import (
	"context"
	"io"
	"net/url"
	"os"
	"testing"

	amqp2 "github.com/cloudevents/sdk-go/v2/protocol/amqp"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"pack.ag/amqp"

	"github.com/cloudevents/sdk-go/v2/binding"
	. "github.com/cloudevents/sdk-go/v2/binding/test"
	"github.com/cloudevents/sdk-go/v2/event"
	bindings "github.com/cloudevents/sdk-go/v2/protocol"
	"github.com/cloudevents/sdk-go/v2/protocol/test"
)

func TestSendSkipBinary(t *testing.T) {
	c, s, r := testSenderReceiver(t)
	defer c.Close()
	EachEvent(t, Events(), func(t *testing.T, e event.Event) {
		eventIn := ExToStr(t, e)
		in := MustCreateMockBinaryMessage(eventIn)
		test.SendReceive(t, binding.WithSkipDirectBinaryEncoding(binding.WithPreferredEventEncoding(context.Background(), binding.EncodingStructured), true), in, s, r, func(out binding.Message) {
			eventOut := MustToEvent(t, context.TODO(), out)
			assert.Equal(t, out.ReadEncoding(), binding.EncodingStructured)
			AssertEventEquals(t, eventIn, ExToStr(t, eventOut))
		})
	})
}

func TestSendSkipStructured(t *testing.T) {
	c, s, r := testSenderReceiver(t)
	defer c.Close()
	EachEvent(t, Events(), func(t *testing.T, e event.Event) {
		eventIn := ExToStr(t, e)
		in := MustCreateMockStructuredMessage(eventIn)
		test.SendReceive(t, binding.WithSkipDirectStructuredEncoding(context.Background(), true), in, s, r, func(out binding.Message) {
			eventOut := MustToEvent(t, context.Background(), out)
			assert.Equal(t, out.ReadEncoding(), binding.EncodingBinary)
			AssertEventEquals(t, eventIn, ExToStr(t, eventOut))
		})
	})
}

func TestSendEventReceiveStruct(t *testing.T) {
	c, s, r := testSenderReceiver(t)
	defer c.Close()
	EachEvent(t, Events(), func(t *testing.T, e event.Event) {
		eventIn := ExToStr(t, e)
		in := (*binding.EventMessage)(&eventIn)
		test.SendReceive(t, binding.WithPreferredEventEncoding(context.TODO(), binding.EncodingStructured), in, s, r, func(out binding.Message) {
			eventOut := MustToEvent(t, context.Background(), out)
			assert.Equal(t, out.ReadEncoding(), binding.EncodingStructured)
			AssertEventEquals(t, eventIn, ExToStr(t, eventOut))
		})
	})
}

func TestSendEventReceiveBinary(t *testing.T) {
	c, s, r := testSenderReceiver(t)
	defer c.Close()
	EachEvent(t, Events(), func(t *testing.T, e event.Event) {
		eventIn := ExToStr(t, e)
		in := (*binding.EventMessage)(&eventIn)
		test.SendReceive(t, context.Background(), in, s, r, func(out binding.Message) {
			eventOut := MustToEvent(t, context.Background(), out)
			assert.Equal(t, out.ReadEncoding(), binding.EncodingBinary)
			AssertEventEquals(t, eventIn, ExToStr(t, eventOut))
		})
	})
}

// Ideally add AMQP server support to the binding.

// Some test require an AMQP broker or router. If the connection fails
// the tests are skipped. The env variable TEST_AMQP_URL can be set to the
// test URL, otherwise the default is "/test"
//
// On option is http://qpid.apache.org/components/dispatch-router/indexthtml.
// It can be installed from source or from RPMs, see https://qpid.apache.org/packages.html
// Run `qdrouterd` and the tests will work with no further config.
func testClient(t testing.TB) (client *amqp.Client, session *amqp.Session, addr string) {
	t.Helper()
	addr = "test"
	s := os.Getenv("TEST_AMQP_URL")
	if u, err := url.Parse(s); err == nil && u.Path != "" {
		addr = u.Path
	}
	client, err := amqp.Dial(s)
	if err != nil {
		t.Skipf("ampq.Dial(%#v): %v", s, err)
	}
	session, err = client.NewSession()
	require.NoError(t, err)
	return client, session, addr
}

func testSenderReceiver(t testing.TB, senderOptions ...amqp2.SenderOptionFunc) (io.Closer, bindings.Sender, bindings.Receiver) {
	c, ss, a := testClient(t)
	r, err := ss.NewReceiver(amqp.LinkSourceAddress(a))
	require.NoError(t, err)
	s, err := ss.NewSender(amqp.LinkTargetAddress(a))
	require.NoError(t, err)
	return c, amqp2.NewSender(s, senderOptions...), amqp2.NewReceiver(r)
}

func BenchmarkSendReceive(b *testing.B) {
	c, s, r := testSenderReceiver(b)
	defer func() { require.NoError(b, c.Close()) }()
	test.BenchmarkSendReceive(b, s, r)
}
