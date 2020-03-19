package amqp

import (
	"net/url"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"pack.ag/amqp"

	bindingtest "github.com/cloudevents/sdk-go/v2/binding/test"
	clienttest "github.com/cloudevents/sdk-go/v2/client/test"
	"github.com/cloudevents/sdk-go/v2/event"
	protocolamqp "github.com/cloudevents/sdk-go/v2/protocol/amqp"
)

func TestSendEvent(t *testing.T) {
	bindingtest.EachEvent(t, bindingtest.Events(), func(t *testing.T, eventIn event.Event) {
		eventIn = bindingtest.ExToStr(t, eventIn)
		clienttest.SendReceive(t, func() interface{} {
			return protocolFactory(t)
		}, eventIn, func(e event.Event) {
			bindingtest.AssertEventEquals(t, eventIn, bindingtest.ExToStr(t, e))
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
func testClient(t *testing.T) (client *amqp.Client, session *amqp.Session, addr string) {
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

func protocolFactory(t *testing.T) *protocolamqp.Protocol {
	c, ss, a := testClient(t)

	p, err := protocolamqp.NewProtocolFromClient(c, ss, a)
	require.NoError(t, err)

	return p
}
