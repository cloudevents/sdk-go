package http_binding_test

import (
	"context"
	nethttp "net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/event"
	bindings "github.com/cloudevents/sdk-go/pkg/transport"
	http "github.com/cloudevents/sdk-go/pkg/transport/http"
	test "github.com/cloudevents/sdk-go/pkg/transport/test"

	. "github.com/cloudevents/sdk-go/pkg/binding/test"
)

func TestSendSkipBinary(t *testing.T) {
	close, s, r := testSenderReceiver(t)
	defer close()
	EachEvent(t, Events(), func(t *testing.T, eventIn event.Event) {
		eventIn = ExToStr(t, eventIn)
		in := MustCreateMockBinaryMessage(eventIn)
		test.SendReceive(t, binding.WithSkipDirectBinaryEncoding(binding.WithPreferredEventEncoding(context.Background(), binding.EncodingStructured), true), in, s, r, func(out binding.Message) {
			eventOut := MustToEvent(t, context.Background(), out)
			assert.Equal(t, binding.EncodingStructured, out.ReadEncoding())
			AssertEventEquals(t, eventIn, ExToStr(t, eventOut))
		})
	})
}

func TestSendSkipStructured(t *testing.T) {
	close, s, r := testSenderReceiver(t)
	defer close()
	EachEvent(t, Events(), func(t *testing.T, eventIn event.Event) {
		eventIn = ExToStr(t, eventIn)
		in := MustCreateMockStructuredMessage(eventIn)
		test.SendReceive(t, binding.WithSkipDirectStructuredEncoding(context.Background(), true), in, s, r, func(out binding.Message) {
			eventOut := MustToEvent(t, context.Background(), out)
			assert.Equal(t, binding.EncodingBinary, out.ReadEncoding())
			AssertEventEquals(t, eventIn, ExToStr(t, eventOut))
		})
	})
}

func TestSendBinaryReceiveBinary(t *testing.T) {
	close, s, r := testSenderReceiver(t)
	defer close()
	EachEvent(t, Events(), func(t *testing.T, eventIn event.Event) {
		eventIn = ExToStr(t, eventIn)
		in := MustCreateMockBinaryMessage(eventIn)
		test.SendReceive(t, context.Background(), in, s, r, func(out binding.Message) {
			eventOut := MustToEvent(t, context.Background(), out)
			assert.Equal(t, binding.EncodingBinary, out.ReadEncoding())
			AssertEventEquals(t, eventIn, ExToStr(t, eventOut))
		})
	})
}

func TestSendStructReceiveStruct(t *testing.T) {
	close, s, r := testSenderReceiver(t)
	defer close()
	EachEvent(t, Events(), func(t *testing.T, eventIn event.Event) {
		eventIn = ExToStr(t, eventIn)
		in := MustCreateMockStructuredMessage(eventIn)
		test.SendReceive(t, context.Background(), in, s, r, func(out binding.Message) {
			eventOut := MustToEvent(t, context.Background(), out)
			require.Equal(t, binding.EncodingStructured, out.ReadEncoding())
			AssertEventEquals(t, eventIn, ExToStr(t, eventOut))
		})
	})
}

func TestSendEventReceiveBinary(t *testing.T) {
	close, s, r := testSenderReceiver(t)
	defer close()
	EachEvent(t, Events(), func(t *testing.T, eventIn event.Event) {
		eventIn = ExToStr(t, eventIn)
		in := binding.EventMessage(eventIn)
		test.SendReceive(t, context.Background(), in, s, r, func(out binding.Message) {
			eventOut := MustToEvent(t, context.Background(), out)
			require.Equal(t, binding.EncodingBinary, out.ReadEncoding())
			AssertEventEquals(t, eventIn, ExToStr(t, eventOut))
		})
	})
}

func testSenderReceiver(t testing.TB, options ...http.SenderOptionFunc) (func(), bindings.Sender, bindings.Receiver) {
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
