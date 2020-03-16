package kafka_sarama

import (
	"context"
	"io"

	"github.com/Shopify/sarama"

	"github.com/cloudevents/sdk-go/pkg/binding"
)

type msgErr struct {
	msg binding.Message
	err error
}

// Receiver which implements sarama.ConsumerGroupHandler
// After the first invocation of Receiver.Receive(), the sarama.ConsumerGroup is created and started.
type Receiver struct {
	incoming chan msgErr

	client              sarama.Client
	topic               string
	groupId             string
	saramaConsumerGroup sarama.ConsumerGroup
}

// NewReceiver creates a Receiver which implements sarama.ConsumerGroupHandler
// After the first invocation of Receiver.Receive(), the sarama.ConsumerGroup is created and started.
func NewReceiver(client sarama.Client, groupId string, topic string) *Receiver {
	return &Receiver{
		incoming: make(chan msgErr),
		client:   client,
		groupId:  groupId,
		topic:    topic,
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
	// Consumer Group not started!
	if r.saramaConsumerGroup == nil {
		cg, err := sarama.NewConsumerGroupFromClient(r.groupId, r.client)
		if err != nil {
			return nil, err
		}
		r.saramaConsumerGroup = cg

		go func() {
			if err := cg.Consume(ctx, []string{r.topic}, r); err != nil {
				r.incoming <- msgErr{err: err}
			}
		}()
	}

	msgErr, ok := <-r.incoming
	if !ok {
		return nil, io.EOF
	}
	return msgErr.msg, msgErr.err
}

func (r *Receiver) Close(ctx context.Context) error {
	if r.saramaConsumerGroup != nil {
		return r.saramaConsumerGroup.Close()
	}
	return nil
}
