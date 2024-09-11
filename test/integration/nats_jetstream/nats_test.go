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
	"github.com/nats-io/nats.go/jetstream"

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
		version         int
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "regular subscriber - structured",
			args: args{
				bindingEncoding: binding.EncodingStructured,
				version:         1,
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
				version:         1,
			},
		},
		{
			name: "pull consumer config - structured",
			args: args{
				opts: []ce_nats.ProtocolOption{
					ce_nats.WithConsumerOptions(
						ce_nats.WithConsumerConfig(&jetstream.ConsumerConfig{
							Name: uuid.New().String(),
						}),
					),
				},
				bindingEncoding: binding.EncodingStructured,
				version:         2,
			},
		},
		{
			name: "ordered consumer config - structured",
			args: args{
				opts: []ce_nats.ProtocolOption{
					ce_nats.WithConsumerOptions(
						ce_nats.WithOrderedConsumerConfig(&jetstream.OrderedConsumerConfig{}),
					),
				},
				bindingEncoding: binding.EncodingStructured,
				version:         2,
			},
		},
		{
			name: "regular subscriber - binary",
			args: args{
				bindingEncoding: binding.EncodingBinary,
				version:         1,
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
				version:         1,
			},
		}, {
			name: "pull consumer config - binary",
			args: args{
				opts: []ce_nats.ProtocolOption{
					ce_nats.WithConsumerOptions(
						ce_nats.WithConsumerConfig(&jetstream.ConsumerConfig{
							Name: uuid.New().String(),
						}),
					),
				},
				bindingEncoding: binding.EncodingBinary,
				version:         2,
			},
		},
		{
			name: "ordered consumer config - binary",
			args: args{
				opts: []ce_nats.ProtocolOption{
					ce_nats.WithConsumerOptions(
						ce_nats.WithOrderedConsumerConfig(&jetstream.OrderedConsumerConfig{}),
					),
				},
				bindingEncoding: binding.EncodingBinary,
				version:         2,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup, s, r := testProtocol(t, conn, tt.args.version, tt.args.opts...)
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

func testProtocol(t testing.TB, natsConn *nats.Conn, version int, opts ...ce_nats.ProtocolOption) (func(), bindings.Sender,
	bindings.Receiver) {
	// STAN connections actually connect to NATS, so the env var is named appropriately
	s := os.Getenv("TEST_NATS_SERVER")
	if s == "" {
		s = "nats://localhost:4223"
	}

	stream := "test-ce-client-" + uuid.New().String()
	subject := stream + ".test"

	// use NewProtocol rather than individual Consumer and Sender since this gives us more coverage
	var p *ce_nats.Protocol
	var err error
	if version == 1 {
		p, err = ce_nats.NewProtocol(s, stream, subject, subject, ce_nats.NatsOptions(), []nats.JSOpt{}, []nats.SubOpt{}, opts...)
	} else {
		ctx := context.Background()
		p, err = ce_nats.NewProtocolV2(ctx, s, stream, subject, ce_nats.NatsOptions(), []jetstream.JetStreamOpt{}, opts...)
		require.NoError(t, err)
		if p.Consumer.ConsumerConfig != nil {
			p.Consumer.ConsumerConfig.FilterSubjects = []string{subject}
		} else if p.Consumer.OrderedConsumerConfig != nil {
			p.Consumer.OrderedConsumerConfig.FilterSubjects = []string{subject}
		}
	}
	require.NoError(t, err)

	go func() {
		require.NoError(t, p.OpenInbound(context.TODO()))
	}()

	return func() {
		err = p.Close(context.TODO())
		require.NoError(t, err)
	}, p.Sender, p.Consumer
}

func BenchmarkSendReceive(b *testing.B) {
	conn := testConn(b)
	defer conn.Close()
	c, s, r := testProtocol(b, conn, 1)
	defer c() // Cleanup
	test.BenchmarkSendReceive(b, s, r)
}

func BenchmarkSendReceiveV2(b *testing.B) {
	conn := testConn(b)
	defer conn.Close()
	c, s, r := testProtocol(b, conn, 2)
	defer c() // Cleanup
	test.BenchmarkSendReceive(b, s, r)
}
