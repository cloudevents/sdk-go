/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package pubsub

import (
	"context"
	"fmt"
	"io"
	"sync"

	"cloud.google.com/go/pubsub"
	"github.com/cloudevents/sdk-go/protocol/pubsub/v2/internal"
	"github.com/cloudevents/sdk-go/v2/binding"
	cecontext "github.com/cloudevents/sdk-go/v2/context"
	"github.com/cloudevents/sdk-go/v2/protocol"
	"golang.org/x/sync/errgroup"
)

const (
	ProtocolName = "Pub/Sub"
)

type subscriptionWithTopic struct {
	topicID        string
	subscriptionID string
	filter         string
}

// Protocol acts as both a pubsub topic and a pubsub subscription .
type Protocol struct {
	// PubSub

	// ReceiveSettings is used to configure Pubsub pull subscription.
	ReceiveSettings *pubsub.ReceiveSettings

	// AllowCreateTopic controls if the transport can create a topic if it does
	// not exist.
	AllowCreateTopic bool

	// AllowCreateSubscription controls if the transport can create a
	// subscription if it does not exist.
	AllowCreateSubscription bool

	// MessageOrdering enables message ordering for all topics and subscriptions.
	MessageOrdering bool

	projectID string
	topicID   string

	gccMux sync.Mutex

	subscriptions []subscriptionWithTopic
	client        *pubsub.Client

	connectionsBySubscription map[string]*internal.Connection
	connectionsByTopic        map[string]*internal.Connection

	incoming chan pubsub.Message
}

// New creates a new pubsub transport.
func New(ctx context.Context, opts ...Option) (*Protocol, error) {
	t := &Protocol{}
	t.incoming = make(chan pubsub.Message)
	if err := t.applyOptions(opts...); err != nil {
		return nil, err
	}

	if t.client == nil {
		// Auth to pubsub.
		client, err := pubsub.NewClient(ctx, t.projectID)
		if err != nil {
			return nil, err
		}
		// Success.
		t.client = client
	}

	if t.connectionsBySubscription == nil {
		t.connectionsBySubscription = make(map[string]*internal.Connection)
	}

	if t.connectionsByTopic == nil {
		t.connectionsByTopic = make(map[string]*internal.Connection)
	}
	return t, nil
}

func (t *Protocol) applyOptions(opts ...Option) error {
	for _, fn := range opts {
		if err := fn(t); err != nil {
			return err
		}
	}
	return nil
}

// Send implements Sender.Send
func (t *Protocol) Send(ctx context.Context, in binding.Message, transformers ...binding.Transformer) error {
	var err error
	defer func() { _ = in.Finish(err) }()

	topic := cecontext.TopicFrom(ctx)
	if topic == "" {
		topic = t.topicID
	}

	conn := t.getOrCreateConnection(ctx, topic, "", "")

	msg := &pubsub.Message{}

	if key, ok := ctx.Value(withOrderingKey{}).(string); ok {
		if !t.MessageOrdering {
			return fmt.Errorf("ordering key cannot be used when message ordering is disabled")
		}
		msg.OrderingKey = key
	}

	if err := WritePubSubMessage(ctx, in, msg, transformers...); err != nil {
		return err
	}

	if _, err := conn.Publish(ctx, msg); err != nil {
		return err
	}
	return nil
}

func (t *Protocol) getConnection(ctx context.Context, topic, subscription string) *internal.Connection {
	if subscription != "" {
		if conn, ok := t.connectionsBySubscription[subscription]; ok {
			return conn
		}
	}
	if topic != "" {
		if conn, ok := t.connectionsByTopic[topic]; ok {
			return conn
		}
	}

	return nil
}

func (t *Protocol) getOrCreateConnection(ctx context.Context, topic, subscription, filter string) *internal.Connection {
	t.gccMux.Lock()
	defer t.gccMux.Unlock()

	// Get.
	if conn := t.getConnection(ctx, topic, subscription); conn != nil {
		return conn
	}
	// Create.
	conn := &internal.Connection{
		AllowCreateSubscription: t.AllowCreateSubscription,
		AllowCreateTopic:        t.AllowCreateTopic,
		ReceiveSettings:         t.ReceiveSettings,
		Client:                  t.client,
		ProjectID:               t.projectID,
		MessageOrdering:         t.MessageOrdering,
		TopicID:                 topic,
		SubscriptionID:          subscription,
		Filter:                  filter,
	}
	// Save for later.
	if subscription != "" {
		t.connectionsBySubscription[subscription] = conn
	}
	if topic != "" {
		t.connectionsByTopic[topic] = conn
	}

	return conn
}

// Receive implements Receiver.Receive
func (t *Protocol) Receive(ctx context.Context) (binding.Message, error) {
	select {
	case m, ok := <-t.incoming:
		if !ok {
			return nil, io.EOF
		}

		msg := NewMessage(&m)
		return msg, nil
	case <-ctx.Done():
		return nil, io.EOF
	}
}

func (t *Protocol) startSubscriber(ctx context.Context, sub subscriptionWithTopic) error {
	logger := cecontext.LoggerFrom(ctx)
	logger.Infof("starting subscriber for Topic %q, Subscription %q", sub.topicID, sub.subscriptionID)
	conn := t.getOrCreateConnection(ctx, sub.topicID, sub.subscriptionID, sub.filter)

	logger.Info("conn is", conn)
	if conn == nil {
		return fmt.Errorf("failed to find connection for Topic: %q, Subscription: %q", sub.topicID, sub.subscriptionID)
	}
	// Ok, ready to start pulling.
	return conn.Receive(ctx, func(ctx context.Context, m *pubsub.Message) {
		t.incoming <- *m
	})
}

func (t *Protocol) OpenInbound(ctx context.Context) error {
	eg, ctx := errgroup.WithContext(ctx)
	// Start up each subscription.
	for _, sub := range t.subscriptions {
		ctx, sub := ctx, sub
		eg.Go(func() error {
			return t.startSubscriber(ctx, sub)
		})
	}

	return eg.Wait()
}

// Close implements Closer.Close
func (t *Protocol) Close(ctx context.Context) error {
	// TODO: Implement this.
	return nil
}

// pubsub protocol implements Sender, Receiver, Closer, Opener
var _ protocol.Opener = (*Protocol)(nil)
var _ protocol.Sender = (*Protocol)(nil)
var _ protocol.Receiver = (*Protocol)(nil)
var _ protocol.Closer = (*Protocol)(nil)

type withOrderingKey struct{}

// WithOrderingKey allows to set the Pub/Sub ordering key for publishing events.
func WithOrderingKey(ctx context.Context, key string) context.Context {
	return context.WithValue(ctx, withOrderingKey{}, key)
}
