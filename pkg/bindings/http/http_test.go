package http_test

import (
	nethttp "net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/binding/test"
	"github.com/cloudevents/sdk-go/pkg/bindings/http"
	ce "github.com/cloudevents/sdk-go/pkg/cloudevents"
)

func TestForceSendStructured(t *testing.T) {
	close, s, r := testSenderReceiver(t, http.ForceStructured())
	defer close()
	test.EachEvent(t, test.Events(), func(t *testing.T, eventIn ce.Event) {
		eventIn = test.ExToStr(t, eventIn)
		in := binding.NewMockBinaryMessage(eventIn)
		test.SendReceive(t, in, s, r, func(out binding.Message) {
			eventOut, isStructured, _ := test.MustToEvent(out)
			assert.True(t, isStructured)
			test.AssertEventEquals(t, eventIn, test.ExToStr(t, eventOut))
		})
	})
}

func TestForceSendBinary(t *testing.T) {
	close, s, r := testSenderReceiver(t, http.ForceBinary())
	defer close()
	test.EachEvent(t, test.Events(), func(t *testing.T, eventIn ce.Event) {
		eventIn = test.ExToStr(t, eventIn)
		in := binding.NewMockStructuredMessage(eventIn)
		test.SendReceive(t, in, s, r, func(out binding.Message) {
			eventOut, _, isBinary := test.MustToEvent(out)
			assert.True(t, isBinary)
			test.AssertEventEquals(t, eventIn, test.ExToStr(t, eventOut))
		})
	})
}

func TestSendBinaryReceiveBinary(t *testing.T) {
	close, s, r := testSenderReceiver(t)
	defer close()
	test.EachEvent(t, test.Events(), func(t *testing.T, eventIn ce.Event) {
		eventIn = test.ExToStr(t, eventIn)
		in := binding.NewMockBinaryMessage(eventIn)
		test.SendReceive(t, in, s, r, func(out binding.Message) {
			eventOut, _, isBinary := test.MustToEvent(out)
			assert.True(t, isBinary)
			test.AssertEventEquals(t, eventIn, test.ExToStr(t, eventOut))
		})
	})
}

func TestSendStructReceiveStruct(t *testing.T) {
	close, s, r := testSenderReceiver(t)
	defer close()
	test.EachEvent(t, test.Events(), func(t *testing.T, eventIn ce.Event) {
		eventIn = test.ExToStr(t, eventIn)
		in := binding.NewMockStructuredMessage(eventIn)
		test.SendReceive(t, in, s, r, func(out binding.Message) {
			eventOut, isStructured, _ := test.MustToEvent(out)
			assert.True(t, isStructured)
			test.AssertEventEquals(t, eventIn, test.ExToStr(t, eventOut))
		})
	})
}

func TestSendEventReceiveBinary(t *testing.T) {
	close, s, r := testSenderReceiver(t)
	defer close()
	test.EachEvent(t, test.Events(), func(t *testing.T, eventIn ce.Event) {
		eventIn = test.ExToStr(t, eventIn)
		in := binding.EventMessage(eventIn)
		test.SendReceive(t, in, s, r, func(out binding.Message) {
			eventOut, _, isBinary := test.MustToEvent(out)
			assert.True(t, isBinary)
			test.AssertEventEquals(t, eventIn, test.ExToStr(t, eventOut))
		})
	})
}

func testSenderReceiver(t testing.TB, options ...http.SenderOptionFunc) (func(), binding.Sender, binding.Receiver) {
	r := http.NewReceiver() // Parameters? Capacity, sync.
	srv := httptest.NewServer(r)
	u, err := url.Parse(srv.URL)
	require.NoError(t, err)
	s := http.NewSender(&nethttp.Client{}, u, options...) // Capacity, sync etc.
	return func() { srv.Close() }, s, r
}

func BenchmarkSendReceive(b *testing.B) {
	c, s, r := testSenderReceiver(b)
	defer c() // Cleanup
	test.BenchmarkSendReceive(b, s, r)
}
