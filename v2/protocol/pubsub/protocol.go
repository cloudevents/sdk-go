package pubsub

import (
	"context"
	"sync"

	"cloud.google.com/go/pubsub"
	"github.com/cloudevents/sdk-go/v2/binding"
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

	// Receiver
	Receiver protocol.Receiver
}

// New creates a new pubsub transport.
func New(ctx context.Context, opts ...Option) (*Protocol, error) {
	t := &Protocol{}
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
	msg := &PubSubMessage{}
	if err := WritePubSubMessage(ctx, in, msg, t.transformers); err != nil {
		return err
	}
	// TODO: Implement this.
	return nil
}

// Receive implements Receiver.Receive
func (t *Protocol) Receive(ctx context.Context) (binding.Message, error) {
	// TODO: Implement this.
	return nil, nil
}

// Close implements Closer.Close
func (t *Protocol) Close(ctx context.Context) error {
	// TODO: Implement this.
	return nil
}
