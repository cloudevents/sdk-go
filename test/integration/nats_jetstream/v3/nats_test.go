/*
 Copyright 2024 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package nats_jetstream

import (
	"context"
	"os"
	"testing"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"

	ce_nats "github.com/cloudevents/sdk-go/protocol/nats_jetstream/v3"
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
	conn := createTestConnection(t)
	defer conn.Close()

	type args struct {
		opts            []ce_nats.ProtocolOption
		bindingEncoding binding.Encoding
		consumerConfig  any
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "regular consumer - structured",
			args: args{
				consumerConfig:  &jetstream.ConsumerConfig{},
				bindingEncoding: binding.EncodingStructured,
			},
		},
		{
			name: "ordered consumer - structured",
			args: args{
				consumerConfig:  &jetstream.OrderedConsumerConfig{},
				bindingEncoding: binding.EncodingStructured,
			},
		},
		{
			name: "regular consumer - binary",
			args: args{
				consumerConfig:  &jetstream.ConsumerConfig{},
				bindingEncoding: binding.EncodingBinary,
			},
		}, {
			name: "ordered consumer - binary",
			args: args{
				consumerConfig:  &jetstream.OrderedConsumerConfig{},
				bindingEncoding: binding.EncodingBinary,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			cleanup, s, r := executeProtocol(ctx, t, conn, tt.args.consumerConfig, tt.args.opts...)
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

func createTestConnection(t testing.TB) *nats.Conn {
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

func executeProtocol(ctx context.Context, t testing.TB, natsConn *nats.Conn, consumerConfig any, opts ...ce_nats.ProtocolOption) (func(), bindings.Sender,
	bindings.Receiver) {
	t.Helper()
	// STAN connections actually connect to NATS, so the env var is named appropriately
	s := os.Getenv("TEST_NATS_SERVER")
	if s == "" {
		s = "nats://localhost:4223"
	}

	stream := "test-ce-client-" + uuid.New().String()
	subject := stream + ".test"

	var js jetstream.JetStream
	var err error
	js, err = jetstream.New(natsConn)
	require.NoError(t, err)

	streamConfig := jetstream.StreamConfig{Name: stream, Subjects: []string{subject}}
	_, err = js.CreateOrUpdateStream(ctx, streamConfig)
	require.NoError(t, err)

	if normalConsumerConfig, ok := consumerConfig.(*jetstream.ConsumerConfig); ok {
		normalConsumerConfig.FilterSubjects = []string{subject}
		opts = append(opts, ce_nats.WithConsumerConfig(normalConsumerConfig))
	}
	if orderedConsumerConfig, ok := consumerConfig.(*jetstream.OrderedConsumerConfig); ok {
		orderedConsumerConfig.FilterSubjects = []string{subject}
		opts = append(opts, ce_nats.WithOrderedConsumerConfig(orderedConsumerConfig))
	}

	opts = append(opts, ce_nats.WithURL(s), ce_nats.WithSendSubject(subject))
	// use NewProtocol rather than individual Consumer and Sender since this gives us more coverage
	p, err := ce_nats.New(ctx, opts...)
	require.NoError(t, err)

	go func() {
		require.NoError(t, p.OpenInbound(context.TODO()))
	}()

	return func() {
		err = p.Close(context.TODO())
		require.NoError(t, err)
	}, p, p
}

func BenchmarkSendReceive(b *testing.B) {
	ctx := context.Background()
	conn := createTestConnection(b)
	defer conn.Close()
	c, s, r := executeProtocol(ctx, b, conn, &jetstream.ConsumerConfig{})
	defer c() // Cleanup
	test.BenchmarkSendReceive(b, s, r)
}
