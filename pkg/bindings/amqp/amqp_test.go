package amqp

import (
	"io"
	"net/url"
	"os"
	"testing"

	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/binding/format"
	"github.com/cloudevents/sdk-go/pkg/binding/test"
	ce "github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"pack.ag/amqp"
)

var (
	testEvent = test.FullEvent()

	testAMQPBinary = &amqp.Message{
		Data:       [][]byte{testEvent.Data.([]byte)},
		Properties: &amqp.MessageProperties{ContentType: testEvent.DataContentType()},
		ApplicationProperties: map[string]interface{}{
			"cloudEvents:specversion": "1.0",
			"cloudEvents:type":        testEvent.Type(),
			"cloudEvents:source":      testEvent.Source(),
			"cloudEvents:id":          testEvent.ID(),
			"cloudEvents:time":        testEvent.Time(),
			"cloudEvents:dataschema":  testEvent.DataSchema(),
			"cloudEvents:subject":     testEvent.Subject(),

			"cloudEvents:exbool":   true,
			"cloudEvents:exint":    int64(42),
			"cloudEvents:exstring": "exstring",
			"cloudEvents:exbinary": []byte{0, 1, 2, 3},
			"cloudEvents:extime":   testEvent.Time(),
			"cloudEvents:exurl":    testEvent.Source(),
		},
	}

	testAMQPStruct = &amqp.Message{
		Data:       [][]byte{test.MustJSON(testEvent)}, // JSON encoded string value.
		Properties: &amqp.MessageProperties{ContentType: format.JSON.MediaType()},
	}
)

func TestNewBinary(t *testing.T) {
	got, err := NewBinary(testEvent)
	assert.NoError(t, err)
	assert.Equal(t, testAMQPBinary, got)
}

func TestDecodeBinary(t *testing.T) {
	got, err := Message{AMQP: testAMQPBinary}.Event()
	assert.NoError(t, err)
	assert.Equal(t, exurl(testEvent), got)
}

func TestNewStruct(t *testing.T) {
	m, err := StructEncoder{Format: format.JSON}.Encode(testEvent)
	assert.NoError(t, err)
	assert.Equal(t, m.(Message).AMQP, testAMQPStruct)
}

func exurl(e ce.Event) ce.Event {
	// Flatten exurl to string, AMQP doesn't preserve the URL type.
	// It should preserve other attribute types.
	if s, _ := types.Format(e.Extensions()["exurl"]); s != "" {
		e.SetExtension("exurl", s)
	}

	return e
}

func TestEncodeDecodeBinary(t *testing.T) {
	test.EachEvent(t, test.Events(), func(t *testing.T, in ce.Event) {
		assert.Equal(t, exurl(in), test.EncodeDecode(t, in, BinaryEncoder{}))
	})
}

func TestEncodeDecodeStruct(t *testing.T) {
	t.Skip("JSON does not round-trip extension attributes")
	test.EachEvent(t, test.Events(), func(t *testing.T, in ce.Event) {
		assert.Equal(t, in, test.EncodeDecode(t, in, StructEncoder{Format: format.JSON}))
	})
}

func TestSendReceiveBinary(t *testing.T) {
	c, s, r := testSenderReceiver(t)
	defer c.Close()
	test.EachEvent(t, test.Events(), func(t *testing.T, e ce.Event) {
		in := binding.EventMessage(exurl(e))
		out := test.SendReceive(t, in, s, r)
		test.AssertMessageEqual(t, in, out)
	})
}

func TestSendReceiveStruct(t *testing.T) {
	c, s, r := testSenderReceiver(t)
	defer c.Close()
	test.EachEvent(t, test.Events(), func(t *testing.T, e ce.Event) {
		in, err := binding.StructEncoder{Format: format.JSON}.Encode(e)
		require.NoError(t, err)
		out := test.SendReceive(t, in, s, r)
		test.AssertMessageEqual(t, in, out)
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

func testSenderReceiver(t testing.TB) (io.Closer, Sender, Receiver) {
	c, ss, a := testClient(t)
	r, err := ss.NewReceiver(amqp.LinkSourceAddress(a))
	require.NoError(t, err)
	s, err := ss.NewSender(amqp.LinkTargetAddress(a))
	require.NoError(t, err)
	return c, Sender{s}, Receiver{r}
}

func BenchmarkSendReceive(b *testing.B) {
	c, s, r := testSenderReceiver(b)
	defer func() { require.NoError(b, c.Close()) }()
	test.BenchmarkSendReceive(b, s, r)
}
