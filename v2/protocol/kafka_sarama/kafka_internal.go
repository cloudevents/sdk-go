package kafka_sarama

import (
	"github.com/Shopify/sarama"
)

// kafkaInternal kafka internal of message
type kafkaInternal struct {
	topic           string
	session         sarama.ConsumerGroupSession
	consumerMessage *sarama.ConsumerMessage

	// used for nack, the nack if put the message to the end of the mq
	nackProducer sarama.SyncProducer
}

// Ack indicates successful processing of a Message passed to the Subscriber.Receive callback.
func (i *kafkaInternal) Ack() {
	if i.session == nil {
		return
	}
	i.session.MarkMessage(i.consumerMessage, "")
}

// Nack indicates that the client put the message to the end of the mq
func (i *kafkaInternal) Nack() error {
	if i.nackProducer == nil {
		return nil
	}
	var err error
	kafkaMessage := sarama.ProducerMessage{
		Topic: i.topic,
		Key:   sarama.ByteEncoder(i.consumerMessage.Key),
		Value: sarama.ByteEncoder(i.consumerMessage.Value),
	}
	for _, v := range i.consumerMessage.Headers {
		kafkaMessage.Headers = append(kafkaMessage.Headers, *v)
	}
	kafkaMessage.Timestamp = i.consumerMessage.Timestamp

	// put the message to the end of mq, and mark the message acked
	_, _, err = i.nackProducer.SendMessage(&kafkaMessage)
	if err == nil {
		i.session.MarkMessage(i.consumerMessage, "")
	}
	return err
}

// NewKafkaInternalFromConsumerMessage new kafka internal from consumer message used for unit test
func NewKafkaInternalFromConsumerMessage(cm *sarama.ConsumerMessage) *kafkaInternal {
	return &kafkaInternal{
		consumerMessage: cm,
	}
}
