/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package nats

import (
	"context"
	"github.com/nats-io/nats.go"
	"os"
	"testing"

	ce_nats "github.com/cloudevents/sdk-go/protocol/nats/v2"
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

func TestSendStructuredMessagedToStructures(t *testing.T) {
	conn := testConn(t)
	defer conn.Close()

	type args struct {
		opts []ce_nats.ProtocolOption
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "regular subscriber",
			args: args{},
		}, {
			name: "queue subscriber",
			args: args{
				opts: []ce_nats.ProtocolOption{
					ce_nats.WithConsumerOptions(
						ce_nats.WithQueueSubscriber(uuid.New().String()),
					),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup, s, r := testProtocol(t, conn, tt.args.opts...)
			defer cleanup()
			EachEvent(t, Events(), func(t *testing.T, eventIn event.Event) {
				eventIn = ConvertEventExtensionsToString(t, eventIn)

				in := MustCreateMockStructuredMessage(t, eventIn)

				test.SendReceive(t, binding.WithPreferredEventEncoding(context.TODO(), binding.EncodingStructured), in, s, r, func(out binding.Message) {
					eventOut := MustToEvent(t, context.Background(), out)
					assert.Equal(t, binding.EncodingStructured, out.ReadEncoding())
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
		s = "nats://localhost:4222"
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
		s = "nats://localhost:4222"
	}

	subject := "test-ce-client-" + uuid.New().String()

	// use NewProtocol rather than individual Consumer and Sender since this gives us more coverage
	p, err := ce_nats.NewProtocol(s, subject, subject, ce_nats.NatsOptions(), opts...)
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
	c, s, r := testProtocol(b, conn)
	defer c() // Cleanup
	test.BenchmarkSendReceive(b, s, r)
}
