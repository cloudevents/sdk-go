/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package kafka_segmentio

import (
	"context"
	"fmt"

	"github.com/segmentio/kafka-go"
	"github.com/cloudevents/sdk-go/v2/binding"
)

// Sender implements binding.Sender that sends messages to a specific receiverTopic using kafka.Writer
type Sender struct {
	receiverTopic string
	writer        *kafka.Writer
}

// NewSender returns a binding.Sender that sends messages to a specific receiverTopic using kafka.Writer
func NewSender(brokers []string, writerConfig kafka.WriterConfig, receiverTopic string, options ...SenderOptionFunc) (*Sender, error) {
	writer := kafka.NewWriter(writerConfig)
	return makeSender(writer, receiverTopic, options...), nil
}

// NewSenderFromConfig returns a binding.Sender that sends messages to a specific receiverTopic using kafka.Writer
func NewSenderFromConfig(writerConfig kafka.WriterConfig, receiverTopic string, options ...SenderOptionFunc) (*Sender, error) {
	writer := kafka.NewWriter(writerConfig)
	return makeSender(writer, receiverTopic, options...), nil
}

// NewSenderFromWriter returns a binding.Sender that sends messages to a specific topic using kafka.Writer
func NewSenderFromWriter(receiverTopic string, writer *kafka.Writer, options ...SenderOptionFunc) (*Sender, error) {
	return makeSender(writer, receiverTopic, options...), nil
}

func makeSender(writer *kafka.Writer, topic string, options ...SenderOptionFunc) *Sender {
	s := &Sender{
		receiverTopic: topic,
		writer:        writer,
	}
	for _, o := range options {
		o(s)
	}
	return s
}

func (s *Sender) Send(ctx context.Context, m binding.Message, transformers ...binding.Transformer) error {
	var err error
	defer m.Finish(err)

	kafkaMessage := kafka.Message{Topic: s.receiverTopic}

	if k := ctx.Value(withMessageKey{}); k != nil {
		kafkaMessage.Key = []byte(fmt.Sprintf("%v", k.(interface{})))
	}

	if err = WriteProducerMessage(ctx, m, &kafkaMessage, transformers...); err != nil {
		return err
	}

	err = s.writer.WriteMessages(ctx, kafkaMessage)
	return err
}

func (s *Sender) Close(ctx context.Context) error {
	// If the Sender was built with NewSenderFromClient, this Close will close only the producer,
	// otherwise it will close the whole client
	return s.writer.Close()
}

type withMessageKey struct{}

// WithMessageKey allows to set the key used when sending the producer message
func WithMessageKey(ctx context.Context, key []byte) context.Context {
	return context.WithValue(ctx, withMessageKey{}, key)
}
