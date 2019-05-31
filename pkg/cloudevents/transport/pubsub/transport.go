package pubsub

import (
	"context"
	"errors"
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

const (
	TransportName = "Pub/Sub"
)

// Transport acts as both a pubsub topic and a pubsub subscription .
type Transport struct {
	// Encoding
	Encoding Encoding

	// DefaultEncodingSelectionFn allows for other encoding selection strategies to be injected.
	DefaultEncodingSelectionFn EncodingSelector

	codec transport.Codec
	// Codec Mutex
	coMu sync.Mutex

	// PubSub

	// AllowCreateTopic controls if the transport can create a topic if it does
	// not exist.
	AllowCreateTopic bool

	// AllowCreateSubscription controls if the transport can create a
	// subscription if it does not exist.
	AllowCreateSubscription bool

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

	// Converter is invoked if the incoming transport receives an undecodable
	// message.
	Converter transport.Converter
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

func (t *Transport) loadCodec(ctx context.Context) bool {
	if t.codec == nil {
		t.coMu.Lock()
		if t.DefaultEncodingSelectionFn != nil && t.Encoding != Default {
			logger := cecontext.LoggerFrom(ctx)
			logger.Warn("transport has a DefaultEncodingSelectionFn set but Encoding is not Default. DefaultEncodingSelectionFn will be ignored.")

			t.codec = &Codec{
				Encoding: t.Encoding,
			}
		} else {
			t.codec = &Codec{
				Encoding:                   t.Encoding,
				DefaultEncodingSelectionFn: t.DefaultEncodingSelectionFn,
			}
		}
		t.coMu.Unlock()
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
			_ = t.client.Close()
			return
		}
		// If the topic does not exist, create a new topic with the given name.
		if !ok {
			if !t.AllowCreateTopic {
				err = fmt.Errorf("transport not allowed to create topic %q", t.topicID)
				return
			}
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
}

func (t *Transport) DeleteTopic(ctx context.Context) error {
	if t.topicWasCreated {
		if err := t.topic.Delete(ctx); err != nil {
			return err
		}
		t.topic = nil
		t.topicWasCreated = false
		t.topicOnce = sync.Once{}
	}
	return errors.New("topic was not created by pubsub transport")
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
			_ = t.client.Close()
			return
		}
		// If subscription doesn't exist, create it.
		if !ok {
			if !t.AllowCreateSubscription {
				err = fmt.Errorf("transport not allowed to create subscription %q", t.subscriptionID)
				return
			}
			// Create a new subscription to the previously created topic
			// with the given name.
			// TODO: allow to use push config + allow setting the SubscriptionConfig.
			sub, err = t.client.CreateSubscription(ctx, t.subscriptionID, pubsub.SubscriptionConfig{
				Topic:             topic,
				AckDeadline:       30 * time.Second,
				RetentionDuration: 25 * time.Hour,
			})
			if err != nil {
				_ = t.client.Close()
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

func (t *Transport) DeleteSubscription(ctx context.Context) error {
	if t.subWasCreated {
		if err := t.sub.Delete(ctx); err != nil {
			return err
		}
		t.sub = nil
		t.subWasCreated = false
		t.subOnce = sync.Once{}
	}
	return errors.New("subscription was not created by pubsub transport")
}

// Send implements Transport.Send
func (t *Transport) Send(ctx context.Context, event cloudevents.Event) (*cloudevents.Event, error) {
	if ok := t.loadCodec(ctx); !ok {
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
			Data:       m.Data,
		})

		_, err := r.Get(ctx)
		if err != nil {
			return nil, err
		}
		return nil, nil
	}

	return nil, fmt.Errorf("failed to encode Event into a Message")
}

// SetReceiver implements Transport.SetReceiver
func (t *Transport) SetReceiver(r transport.Receiver) {
	t.Receiver = r
}

// SetConverter implements Transport.SetConverter
func (t *Transport) SetConverter(c transport.Converter) {
	t.Converter = c
}

// HasConverter implements Transport.HasConverter
func (t *Transport) HasConverter() bool {
	return t.Converter != nil
}

// StartReceiver implements Transport.StartReceiver
// NOTE: This is a blocking call.
func (t *Transport) StartReceiver(ctx context.Context) error {
	logger := cecontext.LoggerFrom(ctx)

	// Load the codec.
	if ok := t.loadCodec(ctx); !ok {
		return fmt.Errorf("unknown encoding set on transport: %d", t.Encoding)
	}

	sub, err := t.getOrCreateSubscription(ctx)
	if err != nil {
		return err
	}
	// Ok, ready to start pulling.
	err = sub.Receive(ctx, func(ctx context.Context, m *pubsub.Message) {
		ctx = WithTransportContext(ctx, NewTransportContext(t.projectID, t.topicID, t.subscriptionID, "pull", m))

		msg := &Message{
			Attributes: m.Attributes,
			Data:       m.Data,
		}
		event, err := t.codec.Decode(msg)
		// If codec returns and error, try with the converter if it is set.
		if err != nil && t.HasConverter() {
			event, err = t.Converter.Convert(ctx, msg, err)
		}
		if err != nil {
			logger.Errorw("failed to decode message", zap.Error(err))
			m.Nack()
			return
		}

		if err := t.Receiver.Receive(ctx, *event, nil); err != nil {
			logger.Warnw("pubsub receiver return err", zap.Error(err))
			m.Nack()
			return
		}
		m.Ack()
	})

	return err
}
