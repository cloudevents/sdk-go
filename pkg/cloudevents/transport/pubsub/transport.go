package pubsub

import (
	"context"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"strings"
	"sync"

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

type subscriptionWithTopic struct {
	topicID        string
	subscriptionID string
}

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

	gccMux sync.Mutex

	subscriptions []subscriptionWithTopic
	client        *pubsub.Client

	connectionsBySubscription map[string]*internal.Connection
	connectionsByTopic        map[string]*internal.Connection

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

	if t.connectionsBySubscription == nil {
		t.connectionsBySubscription = make(map[string]*internal.Connection, 0)
	}

	if t.connectionsByTopic == nil {
		t.connectionsByTopic = make(map[string]*internal.Connection, 0)
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

func (t *Transport) getConnection(ctx context.Context, topic, subscription string) *internal.Connection {
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

func (t *Transport) getOrCreateConnection(ctx context.Context, topic, subscription string) *internal.Connection {
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
		Client:                  t.client,
		ProjectID:               t.projectID,
		TopicID:                 topic,
		SubscriptionID:          subscription,
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

// Send implements Transport.Send
func (t *Transport) Send(ctx context.Context, event cloudevents.Event) (context.Context, *cloudevents.Event, error) {
	// TODO populate response context properly.
	if ok := t.loadCodec(ctx); !ok {
		return ctx, nil, fmt.Errorf("unknown encoding set on transport: %d", t.Encoding)
	}

	topic := cecontext.TopicFrom(ctx)
	if topic == "" {
		topic = t.topicID
	}

	conn := t.getOrCreateConnection(ctx, topic, "")

	msg, err := t.codec.Encode(ctx, event)
	if err != nil {
		return ctx, nil, err
	}

	if m, ok := msg.(*Message); ok {
		respEvent, err := conn.Publish(ctx, &pubsub.Message{
			Attributes: m.Attributes,
			Data:       m.Data,
		})
		return ctx, respEvent, err
	}

	return ctx, nil, fmt.Errorf("failed to encode Event into a Message")
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

func (t *Transport) startSubscriber(ctx context.Context, sub subscriptionWithTopic, done func(error)) {
	logger := cecontext.LoggerFrom(ctx)
	logger.Infof("starting subscriber for Topic %q, Subscription %q", sub.topicID, sub.subscriptionID)
	conn := t.getOrCreateConnection(ctx, sub.topicID, sub.subscriptionID)

	logger.Info("conn is", conn)
	if conn == nil {
		err := fmt.Errorf("failed to find connection for Topic: %q, Subscription: %q", sub.topicID, sub.subscriptionID)
		done(err)
		return
	}
	// Ok, ready to start pulling.
	err := conn.Receive(ctx, func(ctx context.Context, m *pubsub.Message) {
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
	done(err)
}

// StartReceiver implements Transport.StartReceiver
// NOTE: This is a blocking call.
func (t *Transport) StartReceiver(ctx context.Context) error {
	// Load the codec.
	if ok := t.loadCodec(ctx); !ok {
		return fmt.Errorf("unknown encoding set on transport: %d", t.Encoding)
	}

	cctx, cancel := context.WithCancel(ctx)
	defer cancel()
	n := len(t.subscriptions)

	// Make the channels for quit and errors.
	quit := make(chan struct{}, n)
	errc := make(chan error, n)

	// Start up each subscription.
	for _, sub := range t.subscriptions {
		go t.startSubscriber(cctx, sub, func(err error) {
			if err != nil {
				errc <- err
			} else {
				quit <- struct{}{}
			}
		})
	}

	// Collect errors and done calls until we have n of them.
	errs := []string(nil)
	for success := 0; success < n; success++ {
		var err error
		select {
		case <-ctx.Done(): // Block for parent context to finish.
			success--
		case err = <-errc: // Collect errors
		case <-quit:
		}
		if cancel != nil {
			// Stop all other subscriptions.
			cancel()
			cancel = nil
		}
		if err != nil {
			errs = append(errs, err.Error())
		}
	}

	close(quit)
	close(errc)

	return errors.New(strings.Join(errs, "\n"))
}
