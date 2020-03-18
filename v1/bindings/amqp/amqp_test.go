// +build amqp

package amqp

import (
	"context"
	"io"
	"net/url"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"pack.ag/amqp"

	"github.com/cloudevents/sdk-go/v1/binding"
	"github.com/cloudevents/sdk-go/v1/binding/test"
	ce "github.com/cloudevents/sdk-go/v1/cloudevents"
)

func TestSendSkipBinary(t *testing.T) {
	c, s, r := testSenderReceiver(t)
	defer c.Close()
	test.EachEvent(t, test.Events(), func(t *testing.T, e ce.Event) {
		eventIn := test.ExToStr(t, e)
		in := test.NewMockBinaryMessage(eventIn)
		test.SendReceive(t, binding.WithSkipDirectBinaryEncoding(binding.WithPreferredEventEncoding(context.Background(), binding.EncodingStructured), true), in, s, r, func(out binding.Message) {
			eventOut, encoding := test.MustToEvent(context.TODO(), out)
			assert.Equal(t, encoding, binding.EncodingStructured)
			test.AssertEventEquals(t, eventIn, test.ExToStr(t, eventOut))
		})
	})
}

func TestSendSkipStructured(t *testing.T) {
	c, s, r := testSenderReceiver(t)
	defer c.Close()
	test.EachEvent(t, test.Events(), func(t *testing.T, e ce.Event) {
		eventIn := test.ExToStr(t, e)
		in := test.NewMockStructuredMessage(eventIn)
		test.SendReceive(t, binding.WithSkipDirectStructuredEncoding(context.Background(), true), in, s, r, func(out binding.Message) {
			eventOut, encoding := test.MustToEvent(context.Background(), out)
			assert.Equal(t, encoding, binding.EncodingBinary)
			test.AssertEventEquals(t, eventIn, test.ExToStr(t, eventOut))
		})
	})
}

func TestSendEventReceiveStruct(t *testing.T) {
	c, s, r := testSenderReceiver(t)
	defer c.Close()
	test.EachEvent(t, test.Events(), func(t *testing.T, e ce.Event) {
		eventIn := test.ExToStr(t, e)
		in := binding.EventMessage(eventIn)
		test.SendReceive(t, binding.WithPreferredEventEncoding(context.TODO(), binding.EncodingStructured), in, s, r, func(out binding.Message) {
			eventOut, encoding := test.MustToEvent(context.Background(), out)
			assert.Equal(t, encoding, binding.EncodingStructured)
			test.AssertEventEquals(t, eventIn, test.ExToStr(t, eventOut))
		})
	})
}

func TestSendEventReceiveBinary(t *testing.T) {
	c, s, r := testSenderReceiver(t)
	defer c.Close()
	test.EachEvent(t, test.Events(), func(t *testing.T, e ce.Event) {
		eventIn := test.ExToStr(t, e)
		in := binding.EventMessage(eventIn)
		test.SendReceive(t, context.Background(), in, s, r, func(out binding.Message) {
			eventOut, encoding := test.MustToEvent(context.Background(), out)
			assert.Equal(t, encoding, binding.EncodingBinary)
			test.AssertEventEquals(t, eventIn, test.ExToStr(t, eventOut))
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

func testSenderReceiver(t testing.TB, senderOptions ...SenderOptionFunc) (io.Closer, binding.Sender, binding.Receiver) {
	c, ss, a := testClient(t)
	r, err := ss.NewReceiver(amqp.LinkSourceAddress(a))
	require.NoError(t, err)
	s, err := ss.NewSender(amqp.LinkTargetAddress(a))
	require.NoError(t, err)
	return c, NewSender(s, senderOptions...), &Receiver{r}
}

func BenchmarkSendReceive(b *testing.B) {
	c, s, r := testSenderReceiver(b)
	defer func() { require.NoError(b, c.Close()) }()
	test.BenchmarkSendReceive(b, s, r)
}
