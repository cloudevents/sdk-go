/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package stan

import (
	"context"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/stan.go"
	"os"
	"testing"

	ce_stan "github.com/cloudevents/sdk-go/protocol/stan/v2"
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

const (
	TEST_CLUSTER_ID = "test-cluster"
	TEST_CLIENT_ID  = "my-client"
)

func TestSendStructuredMessagedToStructures(t *testing.T) {
	conn := testConn(t)
	defer conn.Close()

	type args struct {
		opts []ce_stan.ProtocolOption
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "regular subscriber",
			args: args{
				opts: []ce_stan.ProtocolOption{
					ce_stan.WithConsumerOptions(
						ce_stan.WithSubscriptionOptions(
							stan.StartAtSequence(0),
						),
					),
				},
			},
		}, {
			name: "queue subscriber",
			args: args{
				opts: []ce_stan.ProtocolOption{
					ce_stan.WithConsumerOptions(
						ce_stan.WithQueueSubscriber(uuid.New().String()),
						ce_stan.WithSubscriptionOptions(
							stan.StartAtSequence(0),
						),
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

func testProtocol(t testing.TB, natsConn *nats.Conn, opts ...ce_stan.ProtocolOption) (func(), bindings.Sender,
	bindings.Receiver) {
	subject := "test-ce-client-" + uuid.New().String()

	// use NewProtocol rather than individual Consumer and Sender since this gives us more coverage
	p, err := ce_stan.NewProtocol(TEST_CLUSTER_ID, TEST_CLIENT_ID, subject, subject, ce_stan.StanOptions(stan.NatsConn(natsConn)), opts...)
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
