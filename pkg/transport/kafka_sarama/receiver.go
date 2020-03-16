package kafka_sarama

import (
	"context"
	"io"

	"github.com/Shopify/sarama"

	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/transport"
)

type msgErr struct {
	msg binding.Message
	err error
}

// Receiver which implements sarama.ConsumerGroupHandler
// After the first invocation of Receiver.Receive(), the sarama.ConsumerGroup is created and started.
type Receiver struct {
	incoming chan msgErr
}

// NewReceiver creates a Receiver which implements sarama.ConsumerGroupHandler
// The sarama.ConsumerGroup must be started invoking
// After the first invocation of Receiver.Receive(), the sarama.ConsumerGroup is created and started.
func NewReceiver(client sarama.Client, groupId string, topic string) *Receiver {
	return &Receiver{
		incoming: make(chan msgErr),
	}
}

func (r *Receiver) Setup(sess sarama.ConsumerGroupSession) error {
	return nil
}

func (r *Receiver) Cleanup(sarama.ConsumerGroupSession) error {
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
	msgErr, ok := <-r.incoming
	if !ok {
		return nil, io.EOF
	}
	return msgErr.msg, msgErr.err
}

func (r *Receiver) Close(ctx context.Context) error {
	close(r.incoming)
	return nil
}

var _ transport.Receiver = (*Receiver)(nil)
var _ transport.Closer = (*Receiver)(nil)

type Consumer struct {
	Receiver

	client              sarama.Client
	topic               string
	groupId             string
	saramaConsumerGroup sarama.ConsumerGroup
}

func NewConsumer(client sarama.Client, groupId string, topic string) *Consumer {
	return &Consumer{
		Receiver: Receiver{
			incoming: make(chan msgErr),
		},
		client:  client,
		topic:   topic,
		groupId: groupId,
	}
}

func (c *Consumer) OpenInbound(ctx context.Context) error {
	cg, err := sarama.NewConsumerGroupFromClient(c.groupId, c.client)
	if err != nil {
		return err
	}
	c.saramaConsumerGroup = cg

	if err := cg.Consume(ctx, []string{c.topic}, c); err != nil {
		return err
	}
	return nil
}

func (r *Consumer) Close(ctx context.Context) error {
	if r.saramaConsumerGroup != nil {
		return r.saramaConsumerGroup.Close()
	}
	return nil
}

var _ transport.Opener = (*Consumer)(nil)
