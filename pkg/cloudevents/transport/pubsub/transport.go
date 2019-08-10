package pubsub

import (
	"context"
	"fmt"
	"sync"

	"go.uber.org/zap"

	"cloud.google.com/go/pubsub"
	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	cecontext "github.com/cloudevents/sdk-go/pkg/cloudevents/context"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport/pubsub/internal"
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

	projectID      string
	topicID        string
	subscriptionID string
	client         *pubsub.Client

	connections map[string]*internal.Connection

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

	if t.connections == nil {
		t.connections = make(map[string]*internal.Connection, 0)
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

func (t *Transport) getConnectionKey(ctx context.Context, topic, subscription string) string {
	return fmt.Sprintf("Topic:%s Subscription:%s", topic, subscription)
}

func (t *Transport) getOrCreateConnection(ctx context.Context, topic, subscription string) *internal.Connection {
	// Get.
	key := t.getConnectionKey(ctx, topic, subscription)
	if conn, ok := t.connections[key]; ok {
		return conn
	}
	// Create.
	conn := &internal.Connection{
		AllowCreateSubscription: t.AllowCreateSubscription,
		AllowCreateTopic:        t.AllowCreateTopic,
		Client:                  t.client,
		ProjectID:               t.projectID,
		TopicID:                 topic,
		SubscriptionID:          subscription,
	}
	t.connections[key] = conn
	return conn
}

// Send implements Transport.Send
func (t *Transport) Send(ctx context.Context, event cloudevents.Event) (*cloudevents.Event, error) {
	if ok := t.loadCodec(ctx); !ok {
		return nil, fmt.Errorf("unknown encoding set on transport: %d", t.Encoding)
	}

	conn := t.getOrCreateConnection(ctx, t.topicID, t.subscriptionID)

	msg, err := t.codec.Encode(ctx, event)
	if err != nil {
		return nil, err
	}

	if m, ok := msg.(*Message); ok {
		return conn.Publish(ctx, &pubsub.Message{
			Attributes: m.Attributes,
			Data:       m.Data,
		})
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

	conn := t.getOrCreateConnection(ctx, t.topicID, t.subscriptionID)

	// Ok, ready to start pulling.
	return conn.Receive(ctx, func(ctx context.Context, m *pubsub.Message) {
		msg := &Message{
			Attributes: m.Attributes,
			Data:       m.Data,
		}
		event, err := t.codec.Decode(ctx, msg)
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
}
