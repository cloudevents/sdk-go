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

// Protocol is an AMQP 1.0 protocol binding for CloudEvents.
// It manages the AMQP connection, session, and sender/receiver links.
type Protocol struct {
	senderLinkOpts   *amqp.SenderOptions
	receiverLinkOpts *amqp.ReceiverOptions

	// AMQP connection and session
	Conn        *amqp.Conn
	Session     *amqp.Session
	ownedClient bool
	Node        string

	// Sender for publishing CloudEvents
	Sender                  *sender
	SenderContextDecorators []func(context.Context) context.Context

	// Receiver for consuming CloudEvents
	Receiver *receiver
}

// NewProtocolFromConn creates a new AMQP protocol from an existing connection and session.
func NewProtocolFromConn(conn *amqp.Conn, session *amqp.Session, queue string, opts ...Option) (*Protocol, error) {
	t := &Protocol{
		Node:             queue,
		senderLinkOpts:   &amqp.SenderOptions{},
		receiverLinkOpts: &amqp.ReceiverOptions{},
		Conn:             conn,
		Session:          session,
	}
	if err := t.applyOptions(opts...); err != nil {
		return nil, err
	}

	// Create a sender
	ctx := context.Background()
	amqpSender, err := session.NewSender(ctx, queue, t.senderLinkOpts)
	if err != nil {
		_ = conn.Close()
		_ = session.Close(ctx)
		return nil, err
	}
	t.Sender = NewSender(amqpSender).(*sender)
	t.SenderContextDecorators = []func(context.Context) context.Context{}

	// Create a receiver
	amqpReceiver, err := t.Session.NewReceiver(ctx, t.Node, t.receiverLinkOpts)
	if err != nil {
		return nil, err
	}
	t.Receiver = NewReceiver(amqpReceiver).(*receiver)
	return t, nil
}

// NewProtocol creates a new AMQP protocol.
func NewProtocol(server, queue string, connOption *amqp.ConnOptions, sessionOption *amqp.SessionOptions, opts ...Option) (*Protocol, error) {
	ctx := context.Background()
	conn, err := amqp.Dial(ctx, server, connOption)
	if err != nil {
		return nil, err
	}

	// Open a session
	session, err := conn.NewSession(ctx, sessionOption)
	if err != nil {
		_ = conn.Close()
		return nil, err
	}

	p, err := NewProtocolFromConn(conn, session, queue, opts...)
	if err != nil {
		return nil, err
	}

	p.ownedClient = true
	return p, nil
}

// NewSenderProtocolFromConn creates a send-only AMQP protocol from an existing connection.
func NewSenderProtocolFromConn(conn *amqp.Conn, session *amqp.Session, address string, opts ...Option) (*Protocol, error) {
	t := &Protocol{
		Node:             address,
		senderLinkOpts:   &amqp.SenderOptions{},
		receiverLinkOpts: &amqp.ReceiverOptions{},
		Conn:             conn,
		Session:          session,
	}
	if err := t.applyOptions(opts...); err != nil {
		return nil, err
	}

	// Create a sender
	ctx := context.Background()
	amqpSender, err := session.NewSender(ctx, address, t.senderLinkOpts)
	if err != nil {
		_ = conn.Close()
		_ = session.Close(ctx)
		return nil, err
	}
	t.Sender = NewSender(amqpSender).(*sender)
	t.SenderContextDecorators = []func(context.Context) context.Context{}

	return t, nil
}

// NewReceiverProtocolFromConn creates a receive-only AMQP protocol from an existing connection.
func NewReceiverProtocolFromConn(conn *amqp.Conn, session *amqp.Session, address string, opts ...Option) (*Protocol, error) {
	t := &Protocol{
		Node:             address,
		senderLinkOpts:   &amqp.SenderOptions{},
		receiverLinkOpts: &amqp.ReceiverOptions{},
		Conn:             conn,
		Session:          session,
	}
	if err := t.applyOptions(opts...); err != nil {
		return nil, err
	}

	ctx := context.Background()
	amqpReceiver, err := t.Session.NewReceiver(ctx, address, t.receiverLinkOpts)
	if err != nil {
		return nil, err
	}
	t.Receiver = NewReceiver(amqpReceiver).(*receiver)
	return t, nil
}

// NewSenderProtocol creates a send-only AMQP protocol.
func NewSenderProtocol(server, address string, connOption *amqp.ConnOptions, sessionOption *amqp.SessionOptions, opts ...Option) (*Protocol, error) {
	ctx := context.Background()
	conn, err := amqp.Dial(ctx, server, connOption)
	if err != nil {
		return nil, err
	}

	// Open a session
	session, err := conn.NewSession(ctx, sessionOption)
	if err != nil {
		_ = conn.Close()
		return nil, err
	}

	p, err := NewSenderProtocolFromConn(conn, session, address, opts...)
	if err != nil {
		return nil, err
	}

	p.ownedClient = true
	return p, nil
}

// NewReceiverProtocol creates a receive-only AMQP protocol.
func NewReceiverProtocol(server, address string, connOption *amqp.ConnOptions, sessionOption *amqp.SessionOptions, opts ...Option) (*Protocol, error) {
	ctx := context.Background()
	conn, err := amqp.Dial(ctx, server, connOption)
	if err != nil {
		return nil, err
	}

	// Open a session
	session, err := conn.NewSession(ctx, sessionOption)
	if err != nil {
		_ = conn.Close()
		return nil, err
	}

	p, err := NewReceiverProtocolFromConn(conn, session, address, opts...)

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
		// Closing the connection will close at cascade sender and receiver
		return t.Conn.Close()
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
