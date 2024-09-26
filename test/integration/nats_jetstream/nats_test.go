/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package nats_jetstream

import (
	"context"
	"os"
	"testing"

	"github.com/nats-io/nats.go"

	ce_nats "github.com/cloudevents/sdk-go/protocol/nats_jetstream/v2"
	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/event"
	bindings "github.com/cloudevents/sdk-go/v2/protocol"
	"github.com/cloudevents/sdk-go/v2/protocol/test"
	. "github.com/cloudevents/sdk-go/v2/test"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "github.com/cloudevents/sdk-go/v2/binding/test"
)

func TestSendReceiveStructuredAndBinary(t *testing.T) {
	conn := testConn(t)
	defer conn.Close()

	type args struct {
		opts            []ce_nats.ProtocolOption
		bindingEncoding binding.Encoding
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "regular subscriber - structured",
			args: args{
				bindingEncoding: binding.EncodingStructured,
			},
		},
		{
			name: "queue subscriber - structured",
			args: args{
				opts: []ce_nats.ProtocolOption{
					ce_nats.WithConsumerOptions(
						ce_nats.WithQueueSubscriber(uuid.New().String()),
					),
				},
				bindingEncoding: binding.EncodingStructured,
			},
		},
		{
			name: "pull subscriber - structured",
			args: args{
				opts: []ce_nats.ProtocolOption{
					ce_nats.WithConsumerOptions(
						ce_nats.WithPullSubscriber(uuid.New().String(), fetchCallback),
					),
				},
				bindingEncoding: binding.EncodingStructured,
			},
		},
		{
			name: "regular subscriber - binary",
			args: args{
				bindingEncoding: binding.EncodingBinary,
			},
		}, {
			name: "queue subscriber - binary",
			args: args{
				opts: []ce_nats.ProtocolOption{
					ce_nats.WithConsumerOptions(
						ce_nats.WithQueueSubscriber(uuid.New().String()),
					),
				},
				bindingEncoding: binding.EncodingBinary,
			},
		},
		{
			name: "pull subscriber - binary",
			args: args{
				opts: []ce_nats.ProtocolOption{
					ce_nats.WithConsumerOptions(
						ce_nats.WithPullSubscriber(uuid.New().String(), fetchCallback),
					),
				},
				bindingEncoding: binding.EncodingBinary,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup, s, r := testProtocol(t, conn, tt.args.opts...)
			defer cleanup()
			EachEvent(t, Events(), func(t *testing.T, eventIn event.Event) {
				eventIn = ConvertEventExtensionsToString(t, eventIn)

				var in binding.Message
				switch tt.args.bindingEncoding {
				case binding.EncodingStructured:
					in = MustCreateMockStructuredMessage(t, eventIn)
				case binding.EncodingBinary:
					in = MustCreateMockBinaryMessage(eventIn)
				}

				test.SendReceive(t, binding.WithPreferredEventEncoding(context.TODO(), tt.args.bindingEncoding), in, s, r, func(out binding.Message) {
					eventOut := MustToEvent(t, context.Background(), out)
					assert.Equal(t, tt.args.bindingEncoding, out.ReadEncoding())
					AssertEventEquals(t, eventIn, ConvertEventExtensionsToString(t, eventOut))
				})
			})
		})
	}
}

func testConn(t testing.TB) *nats.Conn {
	t.Helper()
	// STAN connections actually connect to NATS, so the env var is named appropriately
	s := os.Getenv("TEST_NATS_SERVER")
	if s == "" {
		s = "nats://localhost:4223"
	}

	conn, err := nats.Connect(s)
	if err != nil {
		t.Skipf("Cannot create STAN client to NATS server [%s]: %v", s, err)
	}

	return conn
}

func testProtocol(t testing.TB, natsConn *nats.Conn, opts ...ce_nats.ProtocolOption) (func(), bindings.Sender,
	bindings.Receiver) {
	// STAN connections actually connect to NATS, so the env var is named appropriately
	s := os.Getenv("TEST_NATS_SERVER")
	if s == "" {
		s = "nats://localhost:4223"
	}

	stream := "test-ce-client-" + uuid.New().String()
	subject := stream + ".test"

	// use NewProtocol rather than individual Consumer and Sender since this gives us more coverage
	p, err := ce_nats.NewProtocol(s, stream, subject, subject, ce_nats.NatsOptions(), []nats.JSOpt{}, []nats.SubOpt{}, opts...)
	require.NoError(t, err)

	go func() {
		require.NoError(t, p.OpenInbound(context.TODO()))
	}()

	return func() {
		err = p.Close(context.TODO())
		require.NoError(t, err)
	}, p.Sender, p.Consumer
}

func fetchCallback(natsSub *nats.Subscription) ([]*nats.Msg, error) {
	return natsSub.Fetch(1)
}

func BenchmarkSendReceive(b *testing.B) {
	conn := testConn(b)
	defer conn.Close()
	c, s, r := testProtocol(b, conn)
	defer c() // Cleanup
	test.BenchmarkSendReceive(b, s, r)
}
