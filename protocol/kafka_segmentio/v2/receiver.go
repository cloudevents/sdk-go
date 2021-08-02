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
	"github.com/cloudevents/sdk-go/v2/protocol"
)

type Consumer struct {
	reader    *kafka.Reader
	ownReader bool

	topic   string
	groupID string
}

func NewConsumer(brokers []string, readerConfig kafka.ReaderConfig, topic string) (*Consumer, error) {
	client := kafka.NewReader(readerConfig)

	consumer := NewConsumerFromReader(client, readerConfig.GroupID, topic)
	consumer.ownReader = true

	return consumer, nil
}

func NewConsumerFromReader(reader *kafka.Reader, groupId string, topic string) *Consumer {
	return &Consumer{
		reader:    reader,
		topic:     topic,
		groupID:   groupId,
		ownReader: false,
	}
}

func (c *Consumer) Receive(ctx context.Context) (binding.Message, error) {
	msg, err := c.reader.FetchMessage(ctx)
	fmt.Print("cloudevents kafka_segmentio msg received: %+v\n", msg)
	return NewMessageFromConsumerMessage(&msg), err
}

func (c *Consumer) Close(ctx context.Context) error {
	if c.ownReader {
		return c.reader.Close()
	}
	return nil
}

var _ protocol.Receiver = (*Consumer)(nil)

var _ protocol.Closer = (*Consumer)(nil)
