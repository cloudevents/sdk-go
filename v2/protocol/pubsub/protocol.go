package pubsub

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"

	"cloud.google.com/go/pubsub"
	"github.com/cloudevents/sdk-go/v2/binding"
	cecontext "github.com/cloudevents/sdk-go/v2/context"
	"github.com/cloudevents/sdk-go/v2/protocol"
	"github.com/cloudevents/sdk-go/v2/protocol/pubsub/internal"
)

const (
	ProtocolName = "Pub/Sub"
)

type subscriptionWithTopic struct {
	topicID        string
	subscriptionID string
}

// Protocol acts as both a pubsub topic and a pubsub subscription .
type Protocol struct {
	transformers binding.TransformerFactories

	// PubSub

	// ReceiveSettings is used to configure Pubsub pull subscription.
	ReceiveSettings *pubsub.ReceiveSettings

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
		t.connectionsBySubscription = make(map[string]*internal.Connection, 0)
	}

	if t.connectionsByTopic == nil {
		t.connectionsByTopic = make(map[string]*internal.Connection, 0)
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
func (t *Protocol) Send(ctx context.Context, in binding.Message) error {
	var err error
	defer func() { _ = in.Finish(err) }()

	topic := cecontext.TopicFrom(ctx)
	if topic == "" {
		topic = t.topicID
	}

	conn := t.getOrCreateConnection(ctx, topic, "")

	msg := &pubsub.Message{}
	if err := WritePubSubMessage(ctx, in, msg, t.transformers); err != nil {
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

func (t *Protocol) getOrCreateConnection(ctx context.Context, topic, subscription string) *internal.Connection {
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

// Receive implements Receiver.Receive
func (t *Protocol) Receive(ctx context.Context) (binding.Message, error) {
	m, ok := <-t.incoming
	if !ok {
		return nil, io.EOF
	}

	msg := NewMessage(m.Data, m.Attributes)
	m.Ack()
	// TODO: when to do m.Nack()?
	return msg, nil
}

func (t *Protocol) startSubscriber(ctx context.Context, sub subscriptionWithTopic, done func(error)) {
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
		t.incoming <- *m
	})
	done(err)
}

func (t *Protocol) OpenInbound(ctx context.Context) error {
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
