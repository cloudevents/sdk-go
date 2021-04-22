/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package http_binding_test

import (
	"context"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/event"
	bindings "github.com/cloudevents/sdk-go/v2/protocol"
	"github.com/cloudevents/sdk-go/v2/protocol/http"
	"github.com/cloudevents/sdk-go/v2/protocol/test"

	. "github.com/cloudevents/sdk-go/v2/binding/test"
	. "github.com/cloudevents/sdk-go/v2/test"
)

func TestSendSkipBinary(t *testing.T) {
	close, s, r := testSenderReceiver(t)
	defer close()
	EachEvent(t, Events(), func(t *testing.T, eventIn event.Event) {
		eventIn = ConvertEventExtensionsToString(t, eventIn)
		in := MustCreateMockBinaryMessage(eventIn)
		test.SendReceive(t, binding.WithSkipDirectBinaryEncoding(binding.WithPreferredEventEncoding(context.Background(), binding.EncodingStructured), true), in, s, r, func(out binding.Message) {
			eventOut := MustToEvent(t, context.Background(), out)
			assert.Equal(t, binding.EncodingStructured, out.ReadEncoding())
			AssertEventEquals(t, eventIn, ConvertEventExtensionsToString(t, eventOut))
		})
	})
}

func TestSendSkipStructured(t *testing.T) {
	close, s, r := testSenderReceiver(t)
	defer close()
	EachEvent(t, Events(), func(t *testing.T, eventIn event.Event) {
		eventIn = ConvertEventExtensionsToString(t, eventIn)
		in := MustCreateMockStructuredMessage(t, eventIn)
		test.SendReceive(t, binding.WithSkipDirectStructuredEncoding(context.Background(), true), in, s, r, func(out binding.Message) {
			eventOut := MustToEvent(t, context.Background(), out)
			assert.Equal(t, binding.EncodingBinary, out.ReadEncoding())
			AssertEventEquals(t, eventIn, ConvertEventExtensionsToString(t, eventOut))
		})
	})
}

func TestSendBinaryReceiveBinary(t *testing.T) {
	close, s, r := testSenderReceiver(t)
	defer close()
	EachEvent(t, Events(), func(t *testing.T, eventIn event.Event) {
		eventIn = ConvertEventExtensionsToString(t, eventIn)
		in := MustCreateMockBinaryMessage(eventIn)
		test.SendReceive(t, context.Background(), in, s, r, func(out binding.Message) {
			eventOut := MustToEvent(t, context.Background(), out)
			assert.Equal(t, binding.EncodingBinary, out.ReadEncoding())
			AssertEventEquals(t, eventIn, ConvertEventExtensionsToString(t, eventOut))
		})
	})
}

func TestSendStructReceiveStruct(t *testing.T) {
	close, s, r := testSenderReceiver(t)
	defer close()
	EachEvent(t, Events(), func(t *testing.T, eventIn event.Event) {
		eventIn = ConvertEventExtensionsToString(t, eventIn)
		in := MustCreateMockStructuredMessage(t, eventIn)
		test.SendReceive(t, context.Background(), in, s, r, func(out binding.Message) {
			eventOut := MustToEvent(t, context.Background(), out)
			require.Equal(t, binding.EncodingStructured, out.ReadEncoding())
			AssertEventEquals(t, eventIn, ConvertEventExtensionsToString(t, eventOut))
		})
	})
}

func TestSendEventReceiveBinary(t *testing.T) {
	close, s, r := testSenderReceiver(t)
	defer close()
	EachEvent(t, Events(), func(t *testing.T, eventIn event.Event) {
		eventIn = ConvertEventExtensionsToString(t, eventIn)
		in := (*binding.EventMessage)(&eventIn)
		test.SendReceive(t, context.Background(), in, s, r, func(out binding.Message) {
			eventOut := MustToEvent(t, context.Background(), out)
			require.Equal(t, binding.EncodingBinary, out.ReadEncoding())
			AssertEventEquals(t, eventIn, ConvertEventExtensionsToString(t, eventOut))
		})
	})
}

func testSenderReceiver(t testing.TB, options ...http.Option) (func(), bindings.Sender, bindings.Receiver) {
	p, err := http.New(options...)
	require.NoError(t, err)
	srv := httptest.NewServer(p)
	u, err := url.Parse(srv.URL)
	require.NoError(t, err)
	p.Target = u
	return func() { srv.Close() }, p, p
}

func BenchmarkSendReceive(b *testing.B) {
	c, s, r := testSenderReceiver(b)
	defer c() // Cleanup
	test.BenchmarkSendReceive(b, s, r)
}
