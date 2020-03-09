package http_binding_test

import (
	"context"
	nethttp "net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/cloudevents/sdk-go/pkg/event"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/binding/test"
	"github.com/cloudevents/sdk-go/pkg/bindings"
	"github.com/cloudevents/sdk-go/pkg/bindings/http"

	tests "github.com/cloudevents/sdk-go/pkg/bindings/test"
)

func TestSendSkipBinary(t *testing.T) {
	close, s, r := testSenderReceiver(t)
	defer close()
	test.EachEvent(t, test.Events(), func(t *testing.T, eventIn event.Event) {
		eventIn = test.ExToStr(t, eventIn)
		in := test.MustCreateMockBinaryMessage(eventIn)
		tests.SendReceive(t, binding.WithSkipDirectBinaryEncoding(binding.WithPreferredEventEncoding(context.Background(), binding.EncodingStructured), true), in, s, r, func(out binding.Message) {
			eventOut := test.MustToEvent(t, context.Background(), out)
			assert.Equal(t, binding.EncodingStructured, out.ReadEncoding())
			test.AssertEventEquals(t, eventIn, test.ExToStr(t, eventOut))
		})
	})
}

func TestSendSkipStructured(t *testing.T) {
	close, s, r := testSenderReceiver(t)
	defer close()
	test.EachEvent(t, test.Events(), func(t *testing.T, eventIn event.Event) {
		eventIn = test.ExToStr(t, eventIn)
		in := test.MustCreateMockStructuredMessage(eventIn)
		tests.SendReceive(t, binding.WithSkipDirectStructuredEncoding(context.Background(), true), in, s, r, func(out binding.Message) {
			eventOut := test.MustToEvent(t, context.Background(), out)
			assert.Equal(t, binding.EncodingBinary, out.ReadEncoding())
			test.AssertEventEquals(t, eventIn, test.ExToStr(t, eventOut))
		})
	})
}

func TestSendBinaryReceiveBinary(t *testing.T) {
	close, s, r := testSenderReceiver(t)
	defer close()
	test.EachEvent(t, test.Events(), func(t *testing.T, eventIn event.Event) {
		eventIn = test.ExToStr(t, eventIn)
		in := test.MustCreateMockBinaryMessage(eventIn)
		tests.SendReceive(t, context.Background(), in, s, r, func(out binding.Message) {
			eventOut := test.MustToEvent(t, context.Background(), out)
			assert.Equal(t, binding.EncodingBinary, out.ReadEncoding())
			test.AssertEventEquals(t, eventIn, test.ExToStr(t, eventOut))
		})
	})
}

func TestSendStructReceiveStruct(t *testing.T) {
	close, s, r := testSenderReceiver(t)
	defer close()
	test.EachEvent(t, test.Events(), func(t *testing.T, eventIn event.Event) {
		eventIn = test.ExToStr(t, eventIn)
		in := test.MustCreateMockStructuredMessage(eventIn)
		tests.SendReceive(t, context.Background(), in, s, r, func(out binding.Message) {
			eventOut := test.MustToEvent(t, context.Background(), out)
			require.Equal(t, binding.EncodingStructured, out.ReadEncoding())
			test.AssertEventEquals(t, eventIn, test.ExToStr(t, eventOut))
		})
	})
}

func TestSendEventReceiveBinary(t *testing.T) {
	close, s, r := testSenderReceiver(t)
	defer close()
	test.EachEvent(t, test.Events(), func(t *testing.T, eventIn event.Event) {
		eventIn = test.ExToStr(t, eventIn)
		in := binding.EventMessage(eventIn)
		tests.SendReceive(t, context.Background(), in, s, r, func(out binding.Message) {
			eventOut := test.MustToEvent(t, context.Background(), out)
			require.Equal(t, binding.EncodingBinary, out.ReadEncoding())
			test.AssertEventEquals(t, eventIn, test.ExToStr(t, eventOut))
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
	tests.BenchmarkSendReceive(b, s, r)
}
