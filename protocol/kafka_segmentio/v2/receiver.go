/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package kafka_segmentio

import (
	"context"
	"sync"

	"github.com/segmentio/kafka-go"

	"github.com/cloudevents/sdk-go/v2/binding"
)

type msgErr struct {
	msg binding.Message
	err error
}

type Consumer struct {
	reader    *kafka.Reader
	ownReader bool

	topic   string
	groupId string

	cgMtx sync.Mutex
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
		groupId:   groupId,
		ownReader: false,
	}
}

func (c *Consumer) Close(ctx context.Context) error {
	if c.ownReader {
		return c.reader.Close()
	}
	return nil
}
