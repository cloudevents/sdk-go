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
	// AMQP
	Client      *amqp.Conn
	Session     *amqp.Session
	ownedClient bool
	Node        string

	connOpts     *amqp.ConnOptions
	sessionOpts  *amqp.SessionOptions
	senderOpts   *amqp.SenderOptions
	receiverOpts *amqp.ReceiverOptions
	sendOpts     *amqp.SendOptions
	receiveOpts  *amqp.ReceiveOptions

	// Sender
	Sender                  *sender
	SenderContextDecorators []func(context.Context) context.Context

	// Receiver
	Receiver *receiver
}

// NewProtocolFromClient creates a new amqp transport.
func NewProtocolFromClient(
	ctx context.Context,
	client *amqp.Conn,
	session *amqp.Session,
	queue string,
	opts ...Option,
) (*Protocol, error) {
	t := &Protocol{
		Node:    queue,
		Client:  client,
		Session: session,
	}
	if err := t.applyOptions(opts...); err != nil {
		return nil, err
	}

	// Create a sender
	amqpSender, err := session.NewSender(ctx, queue, t.senderOpts)
	if err != nil {
		_ = client.Close()
		_ = session.Close(context.Background())
		return nil, err
	}
	t.Sender = NewSender(amqpSender, WithSendOptions(t.sendOpts)).(*sender)
	t.SenderContextDecorators = []func(context.Context) context.Context{}

	amqpReceiver, err := t.Session.NewReceiver(ctx, t.Node, t.receiverOpts)
	if err != nil {
		return nil, err
	}
	t.Receiver = NewReceiver(amqpReceiver, WithReceiveOptions(t.receiveOpts)).(*receiver)
	return t, nil
}

// NewProtocol creates a new amqp transport.
func NewProtocol(
	ctx context.Context,
	server, queue string,
	connOptions amqp.ConnOptions,
	sessionOptions amqp.SessionOptions,
	opts ...Option,
) (*Protocol, error) {
	client, err := amqp.Dial(ctx, server, &connOptions)
	if err != nil {
		return nil, err
	}

	// Open a session
	session, err := client.NewSession(ctx, &sessionOptions)
	if err != nil {
		_ = client.Close()
		return nil, err
	}

	p, err := NewProtocolFromClient(ctx, client, session, queue, opts...)
	if err != nil {
		return nil, err
	}

	p.ownedClient = true
	return p, nil
}

// NewSenderProtocolFromClient creates a new amqp sender transport.
func NewSenderProtocolFromClient(
	ctx context.Context,
	client *amqp.Conn,
	session *amqp.Session,
	address string,
	opts ...Option,
) (*Protocol, error) {
	t := &Protocol{
		Node:    address,
		Client:  client,
		Session: session,
	}
	if err := t.applyOptions(opts...); err != nil {
		return nil, err
	}

	// Create a sender
	amqpSender, err := session.NewSender(ctx, address, t.senderOpts)
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
func NewReceiverProtocolFromClient(
	ctx context.Context,
	client *amqp.Conn,
	session *amqp.Session,
	address string,
	opts ...Option,
) (*Protocol, error) {
	t := &Protocol{
		Node:    address,
		Client:  client,
		Session: session,
	}
	if err := t.applyOptions(opts...); err != nil {
		return nil, err
	}

	t.Node = address
	amqpReceiver, err := t.Session.NewReceiver(ctx, address, t.receiverOpts)
	if err != nil {
		return nil, err
	}
	t.Receiver = NewReceiver(amqpReceiver, WithReceiveOptions(t.receiveOpts)).(*receiver)
	return t, nil
}

// NewSenderProtocol creates a new sender amqp transport.
func NewSenderProtocol(ctx context.Context, server, address string, connOptions amqp.ConnOptions, sessionOptions amqp.SessionOptions, opts ...Option) (*Protocol, error) {
	client, err := amqp.Dial(ctx, server, &connOptions)
	if err != nil {
		return nil, err
	}

	// Open a session
	session, err := client.NewSession(ctx, &sessionOptions)
	if err != nil {
		_ = client.Close()
		return nil, err
	}

	p, err := NewSenderProtocolFromClient(ctx, client, session, address, opts...)
	if err != nil {
		return nil, err
	}

	p.ownedClient = true
	return p, nil
}

// NewReceiverProtocol creates a new receiver amqp transport.
func NewReceiverProtocol(ctx context.Context, server, address string, connOptions amqp.ConnOptions, sessionOptions amqp.SessionOptions, opts ...Option) (*Protocol, error) {
	client, err := amqp.Dial(ctx, server, &connOptions)
	if err != nil {
		return nil, err
	}

	// Open a session
	session, err := client.NewSession(ctx, &sessionOptions)
	if err != nil {
		_ = client.Close()
		return nil, err
	}

	p, err := NewReceiverProtocolFromClient(ctx, client, session, address, opts...)

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
