package pubsub

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"sync"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	cecontext "github.com/cloudevents/sdk-go/pkg/cloudevents/context"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport"
)

// Transport adheres to transport.Transport.
var _ transport.Transport = (*Transport)(nil)

// Transport acts as both a pubsub topic and a pubsub subscription .
type Transport struct {
	// Encoding
	Encoding Encoding
	codec    transport.Codec

	// PubSub

	projectID string

	client *pubsub.Client

	topicID         string
	topic           *pubsub.Topic
	topicWasCreated bool
	topicOnce       sync.Once

	subscriptionID string
	sub            *pubsub.Subscription
	subWasCreated  bool
	subOnce        sync.Once

	// Receiver
	Receiver transport.Receiver
}

// New creates a new pubsub transport.
func New(ctx context.Context, opts ...Option) (*Transport, error) {
	t := &Transport{}
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

func (t *Transport) getOrCreateTopic(ctx context.Context) (*pubsub.Topic, error) {
	var err error
	t.topicOnce.Do(func() {
		var ok bool
		// Load the topic.
		topic := t.client.Topic(t.topicID)
		ok, err = topic.Exists(ctx)
		if err != nil {
			_ = t.client.Close() // TODO return the error.
			return
		}
		// If the topic does not exist, create a new topic with the given name.
		if !ok { // TODO: add a setting that prevents creation of a topic.
			topic, err = t.client.CreateTopic(ctx, t.topicID)
			if err != nil {
				return
			}
			t.topicWasCreated = true
		}
		// Success.
		t.topic = topic
	})
	if t.topic == nil {
		return nil, fmt.Errorf("unable to create topic %q", t.topicID)
	}
	return t.topic, err
	// TODO: call - topic.Stop()
}

func (t *Transport) getOrCreateSubscription(ctx context.Context) (*pubsub.Subscription, error) {
	var err error
	t.subOnce.Do(func() {
		// Load the topic.
		var topic *pubsub.Topic
		topic, err = t.getOrCreateTopic(ctx)
		if err != nil {
			return
		}
		// Load the subscription.
		var ok bool
		sub := t.client.Subscription(t.subscriptionID)
		ok, err = sub.Exists(ctx)
		if err != nil {
			if t.topicWasCreated {
				//_ = t.topic.Delete(ctx) // TODO return the error. Do this?
			}
			_ = t.client.Close() // TODO return the error.
			return
		}
		// If subscription doesn't exist, create it.
		if !ok { // TODO: add a setting that prevents creation of a subscription.
			// Create a new subscription to the previously created topic
			// with the given name.
			// TODO: allow to use push config.
			// TODO: allow setting the SubscriptionConfig ?
			sub, err = t.client.CreateSubscription(ctx, t.subscriptionID, pubsub.SubscriptionConfig{ // TODO: allow for SubscriptionConfig to be an option.
				Topic:             topic,
				AckDeadline:       30 * time.Second,
				RetentionDuration: 25 * time.Hour,
			})
			if err != nil {
				if t.topicWasCreated {
					//_ = t.topic.Delete(ctx) // TODO return the error. Do this?
				}
				_ = t.client.Close() // TODO return the error.
				return
			}
			t.subWasCreated = true
		}
		// Success.
		t.sub = sub
	})
	if t.sub == nil {
		return nil, fmt.Errorf("unable to create sunscription %q", t.subscriptionID)
	}
	return t.sub, err
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

	topic, err := t.getOrCreateTopic(ctx)
	if err != nil {
		return nil, err
	}

	if m, ok := msg.(*Message); ok {

		r := topic.Publish(ctx, &pubsub.Message{
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

	// Load the codec.
	if ok := t.loadCodec(); !ok {
		return fmt.Errorf("unknown encoding set on transport: %d", t.Encoding)
	}

	sub, err := t.getOrCreateSubscription(ctx)
	if err != nil {
		return err
	}
	// Ok, ready to start pulling.
	err = sub.Receive(ctx, func(ctx context.Context, m *pubsub.Message) {

		msg := &Message{
			Attributes: m.Attributes,
			Body:       m.Data,
		}

		event, err := t.codec.Decode(msg)
		if err != nil {
			logger.Errorw("failed to decode message", zap.Error(err))
			m.Nack()
			return
		}

		ctx = WithTransportContext(ctx, NewTransportContext(t.topicID, t.subscriptionID, "pull", m))

		if err := t.Receiver.Receive(ctx, *event, nil); err != nil {
			logger.Warnw("pubsub receiver return err", zap.Error(err))
			m.Nack()
			return
		}
		m.Ack()
	})

	// TODO: Should it clean up the subscription it it was created?
	//if t.subWasCreated {
	//	_= t.sub.Delete(ctx)
	//	t.sub = nil
	//	t.subOnce = sync.Once{}
	//	t.subWasCreated = false
	//}
	return err
}
