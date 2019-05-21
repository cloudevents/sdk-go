package pubsub

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	cecontext "github.com/cloudevents/sdk-go/pkg/cloudevents/context"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport"
)

// Transport adheres to transport.Transport.
var _ transport.Transport = (*Transport)(nil)

// Transport acts as both a http client and a http handler.
type Transport struct {
	// Encoding
	Encoding Encoding
	codec    transport.Codec

	// PubSub
	client          *pubsub.Client
	topic           *pubsub.Topic
	topicWasCreated bool
	sub             *pubsub.Subscription
	subWasCreated   bool

	// Receiver
	Receiver transport.Receiver
}

// New creates a new pubsub transport.
func New(ctx context.Context, projectID, topicID, subscriptionID string, opts ...Option) (*Transport, error) {
	t := &Transport{}
	if err := t.applyOptions(opts...); err != nil {
		return nil, err
	}

	// Auth to pubsub.
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}

	// Load the topic.
	topic := client.Topic(topicID)
	ok, err := topic.Exists(ctx)
	if err != nil {
		_ = client.Close() // TODO return the error.
		return nil, err
	}
	// If the topic does not exist, create a new topic with the given name.
	if !ok {
		topic, err = client.CreateTopic(ctx, topicID)
		if err != nil {
			return nil, err
		}
		t.topicWasCreated = true
	}

	sub := client.Subscription(subscriptionID)
	ok, err = sub.Exists(ctx)
	if err != nil {
		if t.topicWasCreated {
			_ = topic.Delete(ctx) // TODO return the error.
		}
		_ = client.Close() // TODO return the error.
		return nil, err
	}
	// If subscription doesn't exist, create it.
	if !ok {
		// Create a new subscription to the previously created topic
		// with the given name.
		sub, err = client.CreateSubscription(ctx, subscriptionID, pubsub.SubscriptionConfig{ // TODO: allow for SubscriptionConfig to be an option.
			Topic:             topic,
			AckDeadline:       30 * time.Second,
			RetentionDuration: 25 * time.Hour,
		})
		if err != nil {
			if t.topicWasCreated {
				_ = topic.Delete(ctx) // TODO return the error.
			}
			_ = client.Close() // TODO return the error.
			return nil, err
		}
		t.subWasCreated = true
	}

	t.client = client
	t.topic = topic
	t.sub = sub

	// TODO: call - topic.Stop()
	// TODO: call - client.Stop()

	return t, nil
}

func (t *Transport) applyOptions(opts ...Option) error {
	for _, fn := range opts {
		if err := fn(t); err != nil {
			return err
		}
	}
	return nil
}

func (t *Transport) loadCodec() bool {
	if t.codec == nil {
		switch t.Encoding {
		case Default, BinaryV03, StructuredV03:
			t.codec = &Codec{Encoding: t.Encoding}
		default:
			return false
		}
	}
	return true
}

// Send implements Transport.Send
func (t *Transport) Send(ctx context.Context, event cloudevents.Event) (*cloudevents.Event, error) {
	if ok := t.loadCodec(); !ok {
		return nil, fmt.Errorf("unknown encoding set on transport: %d", t.Encoding)
	}

	msg, err := t.codec.Encode(event)
	if err != nil {
		return nil, err
	}

	if m, ok := msg.(*Message); ok {

		r := t.topic.Publish(ctx, &pubsub.Message{
			Attributes: m.Attributes,
			Data:       m.Body,
		})

		id, err := r.Get(ctx)
		if err != nil {
			return nil, err
		}
		fmt.Printf("Published a message with a message ID: %s\n", id) // TODO: remove

		return nil, nil
	}

	return nil, fmt.Errorf("failed to encode Event into a Message")
}

// SetReceiver implements Transport.SetReceiver
func (t *Transport) SetReceiver(r transport.Receiver) {
	t.Receiver = r
}

// StartReceiver implements Transport.StartReceiver
// NOTE: This is a blocking call.
func (t *Transport) StartReceiver(ctx context.Context) error {
	logger := cecontext.LoggerFrom(ctx)

	if ok := t.loadCodec(); !ok {
		return fmt.Errorf("unknown encoding set on transport: %d", t.Encoding)
	}

	ctx, cancel := context.WithCancel(ctx)

	_ = logger
	// TODO

	<-ctx.Done()

	cancel()
	return nil
}
