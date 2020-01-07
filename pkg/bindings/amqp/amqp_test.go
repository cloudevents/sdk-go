package amqp

import (
	"io"
	"net/url"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"pack.ag/amqp"

	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/binding/test"
	ce "github.com/cloudevents/sdk-go/pkg/cloudevents"
)

func TestSendReceiveBinary(t *testing.T) {
	c, s, r := testSenderReceiver(t, ForceBinary())
	defer c.Close()
	test.EachEvent(t, test.Events(), func(t *testing.T, e ce.Event) {
		eventIn := test.ExToStr(t, e)
		in := binding.NewMockBinaryMessage(eventIn)
		test.SendReceive(t, in, s, r, func(out binding.Message) {
			eventOut, _, isBinary := test.MustToEvent(out)
			assert.True(t, isBinary)
			test.AssertEventEquals(t, eventIn, test.ExToStr(t, eventOut))
		})
	})
}

func TestSendReceiveStruct(t *testing.T) {
	c, s, r := testSenderReceiver(t, ForceStructured())
	defer c.Close()
	test.EachEvent(t, test.Events(), func(t *testing.T, e ce.Event) {
		eventIn := test.ExToStr(t, e)
		in := binding.NewMockStructuredMessage(eventIn)
		test.SendReceive(t, in, s, r, func(out binding.Message) {
			eventOut, isStructured, _ := test.MustToEvent(out)
			assert.True(t, isStructured)
			test.AssertEventEquals(t, eventIn, test.ExToStr(t, eventOut))
		})
	})
}

func TestSendEventReceiveBinary(t *testing.T) {
	c, s, r := testSenderReceiver(t, ForceBinary())
	defer c.Close()
	test.EachEvent(t, test.Events(), func(t *testing.T, e ce.Event) {
		eventIn := test.ExToStr(t, e)
		in := binding.EventMessage(eventIn)
		test.SendReceive(t, in, s, r, func(out binding.Message) {
			eventOut, _, isBinary := test.MustToEvent(out)
			assert.True(t, isBinary)
			test.AssertEventEquals(t, eventIn, test.ExToStr(t, eventOut))
		})
	})
}

func TestSendEventReceiveStruct(t *testing.T) {
	c, s, r := testSenderReceiver(t, ForceStructured())
	defer c.Close()
	test.EachEvent(t, test.Events(), func(t *testing.T, e ce.Event) {
		eventIn := test.ExToStr(t, e)
		in := binding.EventMessage(eventIn)
		test.SendReceive(t, in, s, r, func(out binding.Message) {
			eventOut, isStructured, _ := test.MustToEvent(out)
			assert.True(t, isStructured)
			test.AssertEventEquals(t, eventIn, test.ExToStr(t, eventOut))
		})
	})
}

// TODO(alanconway) Need better self-test without external dependency.
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
