/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package kafka_segmentio

import (
	"context"
	"errors"
	"sync"

	"github.com/segmentio/kafka-go"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/protocol"
)

const (
	defaultGroupID = "cloudevents-sdk-go"
)

type Protocol struct {
	// Kafka
	writerConfig kafka.WriterConfig
	readerConfig kafka.ReaderConfig

	// Sender
	Sender *Sender

	// Sender options
	SenderContextDecorators []func(context.Context) context.Context
	senderTopic             string

	// Consumer
	Consumer    *Consumer
	consumerMux sync.Mutex

	// Consumer options
	receiverTopic   string
	receiverGroupID string
}

// NewProtocol creates a new kafka transport.
func NewProtocol(brokers []string, writerConfig kafka.WriterConfig, readerConfig kafka.ReaderConfig, sendToTopic string, receiveFromTopic string, opts ...ProtocolOptionFunc) (*Protocol, error) {
	p, err := NewProtocolFromClient(writerConfig, readerConfig, sendToTopic, receiveFromTopic, opts...)
	if err != nil {
		return nil, err
	}
	return p, nil
}

// NewProtocolFromClient creates a new kafka transport starting from a kafka writerConfigs
func NewProtocolFromClient(writerConfig kafka.WriterConfig, readerConfig kafka.ReaderConfig, sendToTopic string, receiveFromTopic string, opts ...ProtocolOptionFunc) (*Protocol, error) {
	p := &Protocol{
		writerConfig:            writerConfig,
		readerConfig:            readerConfig,
		SenderContextDecorators: make([]func(context.Context) context.Context, 0),
		receiverGroupID:         readerConfig.GroupID,
		senderTopic:             sendToTopic,
		receiverTopic:           receiveFromTopic,
	}

	var err error
	if err = p.applyOptions(opts...); err != nil {
		return nil, err
	}

	if p.senderTopic == "" {
		return nil, errors.New("you didn't specify the topic to send to")
	}
	p.Sender, err = NewSenderFromClient(p.writerConfig, p.senderTopic)
	if err != nil {
		return nil, err
	}

	if p.receiverTopic == "" {
		return nil, errors.New("you didn't specify the topic to receive from")
	}
	reader := kafka.NewReader(p.readerConfig)
	p.Consumer = NewConsumerFromReader(reader, p.receiverGroupID, p.receiverTopic)

	return p, nil
}

func (p *Protocol) applyOptions(opts ...ProtocolOptionFunc) error {
	for _, fn := range opts {
		fn(p)
	}
	return nil
}

func (p *Protocol) Send(ctx context.Context, in binding.Message, transformers ...binding.Transformer) error {
	for _, f := range p.SenderContextDecorators {
		ctx = f(ctx)
	}
	return p.Sender.Send(ctx, in, transformers...)
}

func (p *Protocol) Close(ctx context.Context) error {

	if err := p.Consumer.Close(ctx); err != nil {
		return err
	}
	return p.Sender.Close(ctx)
}

// Kafka protocol implements Sender, Receiver
var _ protocol.Sender = (*Protocol)(nil)
var _ protocol.Closer = (*Protocol)(nil)
