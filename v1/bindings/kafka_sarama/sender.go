package kafka_sarama

import (
	"context"

	"github.com/Shopify/sarama"

	"github.com/cloudevents/sdk-go/v1/binding"
)

type Sender struct {
	topic        string
	syncProducer sarama.SyncProducer

	transformerFactories binding.TransformerFactories
}

func (s *Sender) Send(ctx context.Context, m binding.Message) error {
	kafkaMessage := sarama.ProducerMessage{Topic: s.topic}

	if err := EncodeKafkaProducerMessage(ctx, m, &kafkaMessage, s.transformerFactories); err != nil {
		return err
	}

	_, _, err := s.syncProducer.SendMessage(&kafkaMessage)
	return err
}

func (s *Sender) Close(ctx context.Context) error {
	return s.syncProducer.Close()
}

func NewSender(client sarama.Client, topic string, options ...SenderOptionFunc) (*Sender, error) {
	producer, err := sarama.NewSyncProducerFromClient(client)
	if err != nil {
		return nil, err
	}

	s := &Sender{
		topic:                topic,
		syncProducer:         producer,
		transformerFactories: make(binding.TransformerFactories, 0),
	}
	for _, o := range options {
		o(s)
	}
	return s, nil
}
