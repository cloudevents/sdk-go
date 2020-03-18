package http_test

import (
	"context"
	nethttp "net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cloudevents/sdk-go/v1/binding"
	"github.com/cloudevents/sdk-go/v1/binding/test"
	"github.com/cloudevents/sdk-go/v1/bindings/http"
	ce "github.com/cloudevents/sdk-go/v1/cloudevents"
)

func TestSendSkipBinary(t *testing.T) {
	close, s, r := testSenderReceiver(t)
	defer close()
	test.EachEvent(t, test.Events(), func(t *testing.T, eventIn ce.Event) {
		eventIn = test.ExToStr(t, eventIn)
		in := test.NewMockBinaryMessage(eventIn)
		test.SendReceive(t, binding.WithSkipDirectBinaryEncoding(binding.WithPreferredEventEncoding(context.Background(), binding.EncodingStructured), true), in, s, r, func(out binding.Message) {
			eventOut, encoding := test.MustToEvent(context.Background(), out)
			assert.Equal(t, encoding, binding.EncodingStructured)
			test.AssertEventEquals(t, eventIn, test.ExToStr(t, eventOut))
		})
	})
}

func TestSendSkipStructured(t *testing.T) {
	close, s, r := testSenderReceiver(t)
	defer close()
	test.EachEvent(t, test.Events(), func(t *testing.T, eventIn ce.Event) {
		eventIn = test.ExToStr(t, eventIn)
		in := test.NewMockStructuredMessage(eventIn)
		test.SendReceive(t, binding.WithSkipDirectStructuredEncoding(context.Background(), true), in, s, r, func(out binding.Message) {
			eventOut, encoding := test.MustToEvent(context.Background(), out)
			assert.Equal(t, encoding, binding.EncodingBinary)
			test.AssertEventEquals(t, eventIn, test.ExToStr(t, eventOut))
		})
	})
}

func TestSendBinaryReceiveBinary(t *testing.T) {
	close, s, r := testSenderReceiver(t)
	defer close()
	test.EachEvent(t, test.Events(), func(t *testing.T, eventIn ce.Event) {
		eventIn = test.ExToStr(t, eventIn)
		in := test.NewMockBinaryMessage(eventIn)
		test.SendReceive(t, context.Background(), in, s, r, func(out binding.Message) {
			eventOut, encoding := test.MustToEvent(context.Background(), out)
			assert.Equal(t, encoding, binding.EncodingBinary)
			test.AssertEventEquals(t, eventIn, test.ExToStr(t, eventOut))
		})
	})
}

func TestSendStructReceiveStruct(t *testing.T) {
	close, s, r := testSenderReceiver(t)
	defer close()
	test.EachEvent(t, test.Events(), func(t *testing.T, eventIn ce.Event) {
		eventIn = test.ExToStr(t, eventIn)
		in := test.NewMockStructuredMessage(eventIn)
		test.SendReceive(t, context.Background(), in, s, r, func(out binding.Message) {
			eventOut, encoding := test.MustToEvent(context.Background(), out)
			assert.Equal(t, encoding, binding.EncodingStructured)
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
		test.SendReceive(t, context.Background(), in, s, r, func(out binding.Message) {
			eventOut, encoding := test.MustToEvent(context.Background(), out)
			assert.Equal(t, encoding, binding.EncodingBinary)
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
