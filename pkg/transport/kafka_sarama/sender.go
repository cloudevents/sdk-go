package kafka_sarama

import (
	"context"

	"github.com/Shopify/sarama"

	"github.com/cloudevents/sdk-go/pkg/binding"
)

// Sender implements binding.Sender that sends messages to a specific topic using sarama.SyncProducer
type Sender struct {
	topic        string
	syncProducer sarama.SyncProducer

	transformers binding.TransformerFactories
}

// Returns a binding.Sender that sends messages to a specific topic using sarama.SyncProducer
func NewSender(client sarama.Client, topic string, options ...SenderOptionFunc) (*Sender, error) {
	producer, err := sarama.NewSyncProducerFromClient(client)
	if err != nil {
		return nil, err
	}

	s := &Sender{
		topic:        topic,
		syncProducer: producer,
		transformers: make(binding.TransformerFactories, 0),
	}
	for _, o := range options {
		o(s)
	}
	return s, nil
}

func (s *Sender) Send(ctx context.Context, m binding.Message) error {
	kafkaMessage := sarama.ProducerMessage{Topic: s.topic}

	if err := WriteProducerMessage(ctx, m, &kafkaMessage, s.transformers); err != nil {
		return err
	}

	_, _, err := s.syncProducer.SendMessage(&kafkaMessage)
	return err
}

func (s *Sender) Close(ctx context.Context) error {
	return s.syncProducer.Close()
}
