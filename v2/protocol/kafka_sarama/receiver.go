package kafka_sarama

import (
	"context"
	"io"
	"sync"

	"github.com/Shopify/sarama"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/protocol"
)

type msgErr struct {
	msg binding.Message
	err error
}

// Receiver which implements sarama.ConsumerGroupHandler
// After the first invocation of Receiver.Receive(), the sarama.ConsumerGroup is created and started.
type Receiver struct {
	once     sync.Once
	incoming chan msgErr
}

// NewReceiver creates a Receiver which implements sarama.ConsumerGroupHandler
// The sarama.ConsumerGroup must be started invoking. If you need a Receiver which also manage the ConsumerGroup, use NewConsumer
// After the first invocation of Receiver.Receive(), the sarama.ConsumerGroup is created and started.
func NewReceiver() *Receiver {
	return &Receiver{
		incoming: make(chan msgErr),
	}
}

func (r *Receiver) Setup(sess sarama.ConsumerGroupSession) error {
	return nil
}

func (r *Receiver) Cleanup(sarama.ConsumerGroupSession) error {
	r.once.Do(func() {
		close(r.incoming)
	})
	return nil
}

func (r *Receiver) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		m := NewMessageFromConsumerMessage(message)

		r.incoming <- msgErr{
			msg: binding.WithFinish(m, func(err error) { session.MarkMessage(message, "") }),
		}
	}
	return nil
}

func (r *Receiver) Receive(ctx context.Context) (binding.Message, error) {
	select {
	case <-ctx.Done():
		return nil, io.EOF
	case msgErr, ok := <-r.incoming:
		if !ok {
			return nil, io.EOF
		}
		return msgErr.msg, msgErr.err
	}
}

var _ protocol.Receiver = (*Receiver)(nil)

type Consumer struct {
	Receiver

	client    sarama.Client
	ownClient bool

	topic   string
	groupId string

	cgMtx sync.Mutex
}

func NewConsumer(brokers []string, saramaConfig *sarama.Config, groupId string, topic string) (*Consumer, error) {
	client, err := sarama.NewClient(brokers, saramaConfig)
	if err != nil {
		return nil, err
	}

	consumer := NewConsumerFromClient(client, groupId, topic)
	consumer.ownClient = true

	return consumer, nil
}

func NewConsumerFromClient(client sarama.Client, groupId string, topic string) *Consumer {
	return &Consumer{
		Receiver: Receiver{
			incoming: make(chan msgErr),
		},
		client:    client,
		topic:     topic,
		groupId:   groupId,
		ownClient: false,
	}
}

func (c *Consumer) OpenInbound(ctx context.Context) error {
	c.cgMtx.Lock()
	defer c.cgMtx.Unlock()
	cg, err := sarama.NewConsumerGroupFromClient(c.groupId, c.client)
	if err != nil {
		return err
	}

	errCh := make(chan error)

	go func(errs chan error) {
		errs <- cg.Consume(context.Background(), []string{c.topic}, c)
	}(errCh)

	select {
	case <-ctx.Done():
		return cg.Close()
	case err = <-errCh:
		// We still need to close this thing
		err2 := cg.Close()
		if err == nil {
			err = err2
		}
		// Somebody else closed the client, so no problem here
		if err == sarama.ErrClosedClient || err == sarama.ErrClosedConsumerGroup {
			return nil
		}
		return err
	}
}

func (c *Consumer) Close(ctx context.Context) error {
	if c.ownClient {
		return c.client.Close()
	}
	return nil
}

var _ protocol.Opener = (*Consumer)(nil)
var _ protocol.Closer = (*Consumer)(nil)
