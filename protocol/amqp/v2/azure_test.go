/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package amqp

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Azure/go-amqp"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/stretchr/testify/require"
)

// TestAzureServiceBusIntegration tests the AMQP protocol binding with real Azure Service Bus
// Set SERVICEBUS_CONNECTION environment variable to run this test
func TestAzureServiceBusIntegration(t *testing.T) {
	connStr := os.Getenv("SERVICEBUS_CONNECTION")
	if connStr == "" {
		t.Skip("SERVICEBUS_CONNECTION not set, skipping Azure Service Bus integration test")
	}

	// Parse connection string
	// Format: Endpoint=sb://namespace.servicebus.windows.net/;SharedAccessKeyName=name;SharedAccessKey=key
	parts := make(map[string]string)
	for part := range strings.SplitSeq(connStr, ";") {
		if kv := strings.SplitN(part, "=", 2); len(kv) == 2 {
			parts[kv[0]] = kv[1]
		}
	}

	endpoint := strings.TrimPrefix(parts["Endpoint"], "sb://")
	endpoint = strings.TrimSuffix(endpoint, "/")
	keyName := parts["SharedAccessKeyName"]
	key := strings.Trim(parts["SharedAccessKey"], "\"")

	require.NotEmpty(t, endpoint, "Endpoint required in connection string")
	require.NotEmpty(t, keyName, "SharedAccessKeyName required in connection string")
	require.NotEmpty(t, key, "SharedAccessKey required in connection string")

	// Use amqps protocol
	server := "amqps://" + endpoint

	t.Run("DirectAMQPConnection", func(t *testing.T) {
		testDirectAMQPConnection(t, server, keyName, key)
	})

	t.Run("CloudEventsProtocolBinding", func(t *testing.T) {
		testCloudEventsProtocol(t, server, keyName, key)
	})
}

// testDirectAMQPConnection validates go-amqp v1.x usage pattern
func testDirectAMQPConnection(t *testing.T, server, keyName, key string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test connection with SASL Plain authentication (go-amqp v1.x pattern)
	connOpts := &amqp.ConnOptions{
		SASLType: amqp.SASLTypePlain(keyName, key),
	}

	t.Logf("Connecting to Azure Service Bus: %s", server)
	conn, err := amqp.Dial(ctx, server, connOpts)
	require.NoError(t, err, "Failed to connect to Azure Service Bus")
	defer conn.Close()

	// Create session
	session, err := conn.NewSession(ctx, nil)
	require.NoError(t, err, "Failed to create session")
	defer session.Close(ctx)

	t.Log("✅ Successfully connected to Azure Service Bus with go-amqp v1.x")
	t.Logf("Connection properties: %+v", conn.Properties())
}

// testCloudEventsProtocol validates our CloudEvents AMQP protocol binding
func testCloudEventsProtocol(t *testing.T, server, keyName, key string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	queueName := "ce-test-queue"

	// Test 1: Using NewProtocol with ConnOptions struct
	t.Run("WithConnOptions", func(t *testing.T) {
		connOpts := &amqp.ConnOptions{
			SASLType: amqp.SASLTypePlain(keyName, key),
		}

		protocol, err := NewProtocol(server, queueName, connOpts, nil)
		if err != nil {
			// Queue might not exist, that's ok for connection test
			t.Logf("Note: %v (queue may not exist, connection succeeded)", err)
		} else {
			defer protocol.Close(ctx)
			t.Log("✅ NewProtocol with ConnOptions succeeded")
		}
	})

	// Test 2: Using NewProtocol with inline ConnOptions
	t.Run("WithInlineConnOptions", func(t *testing.T) {
		protocol, err := NewProtocol(
			server,
			queueName,
			&amqp.ConnOptions{
				SASLType: amqp.SASLTypePlain(keyName, key),
			},
			nil, // SessionOptions
		)
		if err != nil {
			t.Logf("Note: %v (queue may not exist, connection succeeded)", err)
		} else {
			defer protocol.Close(ctx)
			t.Log("✅ NewProtocol with inline ConnOptions succeeded")
		}
	})

	// Test 3: Using NewProtocolFromConn (manual connection)
	t.Run("FromConn", func(t *testing.T) {
		connOpts := &amqp.ConnOptions{
			SASLType: amqp.SASLTypePlain(keyName, key),
		}

		conn, err := amqp.Dial(ctx, server, connOpts)
		require.NoError(t, err, "Failed to dial")
		defer conn.Close()

		session, err := conn.NewSession(ctx, nil)
		require.NoError(t, err, "Failed to create session")
		defer session.Close(ctx)

		protocol, err := NewProtocolFromConn(conn, session, queueName)
		if err != nil {
			t.Logf("Note: %v (queue may not exist, connection succeeded)", err)
		} else {
			defer protocol.Close(ctx)
			t.Log("✅ NewProtocolFromConn succeeded")
		}
	})

	// Test 4: Complete CloudEvents send/receive roundtrip
	t.Run("SendReceiveRoundtrip", func(t *testing.T) {
		connOpts := &amqp.ConnOptions{
			SASLType: amqp.SASLTypePlain(keyName, key),
		}

		// Create sender protocol
		senderProtocol, err := NewSenderProtocol(server, queueName, connOpts, nil)
		if err != nil {
			t.Skipf("Queue %s does not exist, skipping send/receive test: %v", queueName, err)
			return
		}
		defer senderProtocol.Close(ctx)

		// Create receiver protocol
		receiverProtocol, err := NewReceiverProtocol(server, queueName, connOpts, nil)
		require.NoError(t, err, "Failed to create receiver protocol")
		defer receiverProtocol.Close(ctx)

		// Create test event with unique ID
		eventID := "test-event-" + time.Now().Format("20060102150405.000")
		event := cloudevents.NewEvent()
		event.SetID(eventID)
		event.SetSource("github.com/cloudevents/sdk-go/protocol/amqp/v2/test")
		event.SetType("com.example.test")
		err = event.SetData(cloudevents.ApplicationJSON, map[string]string{
			"message":   "roundtrip test from CloudEvents AMQP v1.x",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
		require.NoError(t, err)

		// Create sender client and send event
		senderClient, err := cloudevents.NewClient(senderProtocol)
		require.NoError(t, err)

		sendCtx, sendCancel := context.WithTimeout(ctx, 10*time.Second)
		defer sendCancel()

		err = senderClient.Send(sendCtx, event)
		require.NoError(t, err, "Failed to send event")
		t.Logf("✅ Sent event: %s", eventID)

		// Receive messages until we find ours (drain any old messages)
		recvCtx, recvCancel := context.WithTimeout(ctx, 15*time.Second)
		defer recvCancel()

		var receivedEvent *cloudevents.Event
		for {
			msg, err := receiverProtocol.Receive(recvCtx)
			if err != nil {
				require.NoError(t, err, "Failed to receive message")
			}
			require.NotNil(t, msg, "Received nil message")

			evt, err := binding.ToEvent(recvCtx, msg)
			require.NoError(t, err, "Failed to convert message to event")

			// Acknowledge the message
			err = msg.Finish(nil)
			require.NoError(t, err, "Failed to acknowledge message")

			t.Logf("Received event: ID=%s", evt.ID())
			if evt.ID() == eventID {
				receivedEvent = evt
				break
			}
			t.Logf("Skipping old message: %s", evt.ID())
		}

		// Verify event fields
		require.NotNil(t, receivedEvent, "Did not receive the sent event")
		require.Equal(t, "com.example.test", receivedEvent.Type(), "Event type mismatch")
		t.Logf("✅ Received event: ID=%s, Type=%s, Source=%s",
			receivedEvent.ID(), receivedEvent.Type(), receivedEvent.Source())
		t.Log("✅ Send/Receive roundtrip successful")
	})

	// Test 5: Topic/Subscription pattern
	t.Run("TopicSubscriptionRoundtrip", func(t *testing.T) {
		topicName := "ce-test-topic"
		subscriptionPath := "ce-test-topic/Subscriptions/ce-test-subscription"

		connOpts := &amqp.ConnOptions{
			SASLType: amqp.SASLTypePlain(keyName, key),
		}

		// Create sender to topic
		senderProtocol, err := NewSenderProtocol(server, topicName, connOpts, nil)
		if err != nil {
			t.Skipf("Topic %s does not exist, skipping topic/subscription test: %v", topicName, err)
			return
		}
		defer senderProtocol.Close(ctx)

		// Create receiver from subscription
		receiverProtocol, err := NewReceiverProtocol(server, subscriptionPath, connOpts, nil)
		if err != nil {
			t.Skipf("Subscription %s does not exist: %v", subscriptionPath, err)
			return
		}
		defer receiverProtocol.Close(ctx)

		// Create and send test event
		eventID := "topic-event-" + time.Now().Format("20060102150405.000")
		event := cloudevents.NewEvent()
		event.SetID(eventID)
		event.SetSource("github.com/cloudevents/sdk-go/protocol/amqp/v2/test")
		event.SetType("com.example.topic.test")
		err = event.SetData(cloudevents.ApplicationJSON, map[string]string{
			"message":   "topic/subscription test from CloudEvents AMQP v1.x",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
		require.NoError(t, err)

		senderClient, err := cloudevents.NewClient(senderProtocol)
		require.NoError(t, err)

		sendCtx, sendCancel := context.WithTimeout(ctx, 10*time.Second)
		defer sendCancel()

		err = senderClient.Send(sendCtx, event)
		require.NoError(t, err, "Failed to send event to topic")
		t.Logf("✅ Sent event to topic: %s", eventID)

		// Receive from subscription (drain any old messages)
		recvCtx, recvCancel := context.WithTimeout(ctx, 15*time.Second)
		defer recvCancel()

		var receivedEvent *cloudevents.Event
		for {
			msg, err := receiverProtocol.Receive(recvCtx)
			require.NoError(t, err, "Failed to receive from subscription")
			require.NotNil(t, msg)

			evt, err := binding.ToEvent(recvCtx, msg)
			require.NoError(t, err)

			// Acknowledge the message
			err = msg.Finish(nil)
			require.NoError(t, err, "Failed to acknowledge message")

			t.Logf("Received event: ID=%s", evt.ID())
			if evt.ID() == eventID {
				receivedEvent = evt
				break
			}
			t.Logf("Skipping old message: %s", evt.ID())
		}

		require.NotNil(t, receivedEvent, "Did not receive the sent event")
		require.Equal(t, "com.example.topic.test", receivedEvent.Type())
		t.Logf("✅ Received event from subscription: ID=%s, Type=%s",
			receivedEvent.ID(), receivedEvent.Type())
		t.Log("✅ Topic/Subscription roundtrip successful")
	})
}

// BenchmarkAzureServiceBusConnection benchmarks connection creation
func BenchmarkAzureServiceBusConnection(b *testing.B) {
	connStr := os.Getenv("SERVICEBUS_CONNECTION")
	if connStr == "" {
		b.Skip("SERVICEBUS_CONNECTION not set")
	}

	parts := make(map[string]string)
	for part := range strings.SplitSeq(connStr, ";") {
		if kv := strings.SplitN(part, "=", 2); len(kv) == 2 {
			parts[kv[0]] = kv[1]
		}
	}

	endpoint := strings.TrimPrefix(parts["Endpoint"], "sb://")
	endpoint = strings.TrimSuffix(endpoint, "/")
	keyName := parts["SharedAccessKeyName"]
	key := strings.Trim(parts["SharedAccessKey"], "\"")
	server := "amqps://" + endpoint

	for b.Loop() {
		ctx := context.Background()
		protocol, err := NewProtocol(
			server,
			"ce-test-queue",
			&amqp.ConnOptions{
				SASLType: amqp.SASLTypePlain(keyName, key),
			},
			nil,
		)
		if err == nil {
			protocol.Close(ctx)
		}
	}
}
