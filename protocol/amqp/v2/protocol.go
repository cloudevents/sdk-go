/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package amqp

import (
	"context"

	"github.com/Azure/go-amqp"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/protocol"
)

type Protocol struct {
	connOpts         []amqp.ConnOption
	sessionOpts      []amqp.SessionOption
	senderLinkOpts   []amqp.LinkOption
	receiverLinkOpts []amqp.LinkOption

	// AMQP
	Client      *amqp.Client
	Session     *amqp.Session
	ownedClient bool
	Node        string

	// Sender
	Sender                  *sender
	SenderContextDecorators []func(context.Context) context.Context

	// Receiver
	Receiver *receiver
}

// NewProtocolFromClient creates a new amqp transport.
func NewProtocolFromClient(client *amqp.Client, session *amqp.Session, queue string, opts ...Option) (*Protocol, error) {
	t := &Protocol{
		Node:             queue,
		senderLinkOpts:   []amqp.LinkOption(nil),
		receiverLinkOpts: []amqp.LinkOption(nil),
		Client:           client,
		Session:          session,
	}
	if err := t.applyOptions(opts...); err != nil {
		return nil, err
	}

	t.senderLinkOpts = append(t.senderLinkOpts, amqp.LinkTargetAddress(queue))

	// Create a sender
	amqpSender, err := session.NewSender(t.senderLinkOpts...)
	if err != nil {
		_ = client.Close()
		_ = session.Close(context.Background())
		return nil, err
	}
	t.Sender = NewSender(amqpSender).(*sender)
	t.SenderContextDecorators = []func(context.Context) context.Context{}

	t.receiverLinkOpts = append(t.receiverLinkOpts, amqp.LinkSourceAddress(t.Node))
	amqpReceiver, err := t.Session.NewReceiver(t.receiverLinkOpts...)
	if err != nil {
		return nil, err
	}
	t.Receiver = NewReceiver(amqpReceiver).(*receiver)
	return t, nil
}

// NewProtocol creates a new amqp transport.
func NewProtocol(server, queue string, connOption []amqp.ConnOption, sessionOption []amqp.SessionOption, opts ...Option) (*Protocol, error) {
	client, err := amqp.Dial(server, connOption...)
	if err != nil {
		return nil, err
	}

	// Open a session
	session, err := client.NewSession(sessionOption...)
	if err != nil {
		_ = client.Close()
		return nil, err
	}

	p, err := NewProtocolFromClient(client, session, queue, opts...)
	if err != nil {
		return nil, err
	}

	p.ownedClient = true
	return p, nil
}

// NewSenderProtocolFromClient creates a new amqp sender transport.
func NewSenderProtocolFromClient(client *amqp.Client, session *amqp.Session, address string, opts ...Option) (*Protocol, error) {
	t := &Protocol{
		Node:             address,
		senderLinkOpts:   []amqp.LinkOption(nil),
		receiverLinkOpts: []amqp.LinkOption(nil),
		Client:           client,
		Session:          session,
	}
	if err := t.applyOptions(opts...); err != nil {
		return nil, err
	}
	t.senderLinkOpts = append(t.senderLinkOpts, amqp.LinkTargetAddress(address))
	// Create a sender
	amqpSender, err := session.NewSender(t.senderLinkOpts...)
	if err != nil {
		_ = client.Close()
		_ = session.Close(context.Background())
		return nil, err
	}
	t.Sender = NewSender(amqpSender).(*sender)
	t.SenderContextDecorators = []func(context.Context) context.Context{}

	return t, nil
}

// NewReceiverProtocolFromClient creates a new receiver amqp transport.
func NewReceiverProtocolFromClient(client *amqp.Client, session *amqp.Session, address string, opts ...Option) (*Protocol, error) {
	t := &Protocol{
		Node:             address,
		senderLinkOpts:   []amqp.LinkOption(nil),
		receiverLinkOpts: []amqp.LinkOption(nil),
		Client:           client,
		Session:          session,
	}
	if err := t.applyOptions(opts...); err != nil {
		return nil, err
	}

	t.Node = address
	t.receiverLinkOpts = append(t.receiverLinkOpts, amqp.LinkSourceAddress(address))
	amqpReceiver, err := t.Session.NewReceiver(t.receiverLinkOpts...)
	if err != nil {
		return nil, err
	}
	t.Receiver = NewReceiver(amqpReceiver).(*receiver)
	return t, nil
}

// NewSenderProtocol creates a new sender amqp transport.
func NewSenderProtocol(server, address string, connOption []amqp.ConnOption, sessionOption []amqp.SessionOption, opts ...Option) (*Protocol, error) {
	client, err := amqp.Dial(server, connOption...)
	if err != nil {
		return nil, err
	}

	// Open a session
	session, err := client.NewSession(sessionOption...)
	if err != nil {
		_ = client.Close()
		return nil, err
	}

	p, err := NewSenderProtocolFromClient(client, session, address, opts...)
	if err != nil {
		return nil, err
	}

	p.ownedClient = true
	return p, nil
}

// NewReceiverProtocol creates a new receiver amqp transport.
func NewReceiverProtocol(server, address string, connOption []amqp.ConnOption, sessionOption []amqp.SessionOption, opts ...Option) (*Protocol, error) {
	client, err := amqp.Dial(server, connOption...)
	if err != nil {
		return nil, err
	}

	// Open a session
	session, err := client.NewSession(sessionOption...)
	if err != nil {
		_ = client.Close()
		return nil, err
	}

	p, err := NewReceiverProtocolFromClient(client, session, address, opts...)

	if err != nil {
		return nil, err
	}

	p.ownedClient = true
	return p, nil
}

func (t *Protocol) applyOptions(opts ...Option) error {
	for _, fn := range opts {
		if err := fn(t); err != nil {
			return err
		}
	}
	return nil
}

func (t *Protocol) Close(ctx context.Context) (err error) {
	if t.ownedClient {
		// Closing the client will close at cascade sender and receiver
		return t.Client.Close()
	} else {
		if t.Sender != nil {
			if err = t.Sender.amqp.Close(ctx); err != nil {
				return
			}
		}

		if t.Receiver != nil {
			if err = t.Receiver.amqp.Close(ctx); err != nil {
				return err
			}
		}
		return
	}
}

func (t *Protocol) Send(ctx context.Context, in binding.Message, transformers ...binding.Transformer) error {
	return t.Sender.Send(ctx, in, transformers...)
}

func (t *Protocol) Receive(ctx context.Context) (binding.Message, error) {
	return t.Receiver.Receive(ctx)
}

var _ protocol.Sender = (*Protocol)(nil)
var _ protocol.Receiver = (*Protocol)(nil)
var _ protocol.Closer = (*Protocol)(nil)
