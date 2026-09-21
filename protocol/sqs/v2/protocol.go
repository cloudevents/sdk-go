/*
Copyright 2023 The CloudEvents Authors
SPDX-License-Identifier: Apache-2.0
*/

package sqs

import (
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/cloudevents/sdk-go/v2/binding"
	cecontext "github.com/cloudevents/sdk-go/v2/context"
	"github.com/cloudevents/sdk-go/v2/protocol"
)

const (
	defaultVisibilityTimeout = 30
	defaultMaxMessages       = 10
	defaultWaitTimeSeconds   = 20
)

type Protocol struct {
	client            *sqs.Client
	incoming          chan *Message
	queueURL          *string
	maxMessages       int32
	waitTimeSeconds   int32
	visibilityTimeout int32
}

// New creates a new SQS protocol.
func New(queueName string, opts ...Option) (*Protocol, error) {
	p := &Protocol{
		incoming:          make(chan *Message),
		queueURL:          aws.String(queueName),
		visibilityTimeout: defaultVisibilityTimeout,
		maxMessages:       defaultMaxMessages,
		waitTimeSeconds:   defaultWaitTimeSeconds,
	}
	if err := p.applyOptions(opts...); err != nil {
		return nil, err
	}
	if p.client == nil {
		return nil, fmt.Errorf("sqs client is nil")
	}
	return p, nil
}

func (p *Protocol) applyOptions(opts ...Option) error {
	for _, fn := range opts {
		if err := fn(p); err != nil {
			return err
		}
	}
	return nil
}

// Send sends messages. Send implements Sender.Sender
func (p *Protocol) Send(ctx context.Context, in binding.Message, transformers ...binding.Transformer) error {
	var err error
	defer func() { _ = in.Finish(err) }()
	msgInput := sqs.SendMessageInput{
		QueueUrl: p.queueURL,
	}
	err = WriteMessageInput(ctx, in, &msgInput, transformers...)
	if err != nil {
		return err
	}
	_, err = p.client.SendMessage(ctx, &msgInput)
	return err
}

// OpenInbound implements Opener.OpenInbound
func (p *Protocol) OpenInbound(ctx context.Context) error {
	logger := cecontext.LoggerFrom(ctx)
	logger.Infof("Starting SQS Message polling for %s", aws.ToString(p.queueURL))
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			res, err := p.getMessages(ctx)
			if err != nil {
				continue
			}
			for _, message := range res.Messages {
				p.incoming <- NewMessage(&message, p.client, p.queueURL)
			}
		}
	}
}

func (p *Protocol) getMessages(ctx context.Context) (*sqs.ReceiveMessageOutput, error) {
	return p.client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:            p.queueURL,
		MaxNumberOfMessages: p.maxMessages,
		WaitTimeSeconds:     p.waitTimeSeconds,
		VisibilityTimeout:   p.visibilityTimeout,
		MessageSystemAttributeNames: []types.MessageSystemAttributeName{
			types.MessageSystemAttributeNameAWSTraceHeader,
		},
		MessageAttributeNames: []string{
			string(types.QueueAttributeNameAll),
		},
	})
}

// Receive implements Receiver.Receive.
func (p *Protocol) Receive(ctx context.Context) (binding.Message, error) {
	select {
	case msg, ok := <-p.incoming:
		if !ok {
			return nil, io.EOF
		}
		return msg, nil
	case <-ctx.Done():
		return nil, io.EOF
	}
}

var (
	_ protocol.Receiver = (*Protocol)(nil)
	_ protocol.Sender   = (*Protocol)(nil)
	_ protocol.Opener   = (*Protocol)(nil)
)
