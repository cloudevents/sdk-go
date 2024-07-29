/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package nats_jetstream

import (
	"context"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/protocol"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

// Protocol is a reference implementation for using the CloudEvents binding
// integration. Protocol acts as both a NATS client and a NATS handler.
type Protocol struct {
	Conn *nats.Conn

	Consumer        *Consumer
	consumerOptions []ConsumerOption

	Sender        *Sender
	senderOptions []SenderOption

	connOwned bool // whether this protocol created the nats connection
}

// NewProtocol creates a new NATS protocol.
func NewProtocol(url, stream, sendSubject, receiveSubject string, natsOpts []nats.Option, jsOps []nats.JSOpt, subOpts []nats.SubOpt, opts ...ProtocolOption) (*Protocol, error) {
	conn, err := nats.Connect(url, natsOpts...)
	if err != nil {
		return nil, err
	}

	p, err := NewProtocolFromConn(conn, stream, sendSubject, receiveSubject, jsOps, subOpts, opts...)
	if err != nil {
		conn.Close()
		return nil, err
	}

	p.connOwned = true

	return p, nil
}

// NewProtocolV2 creates a new NATS protocol.
func NewProtocolV2(ctx context.Context, url, stream, sendSubject string, natsOpts []nats.Option, jsOpts []jetstream.JetStreamOpt, opts ...ProtocolOption) (*Protocol, error) {
	conn, err := nats.Connect(url, natsOpts...)
	if err != nil {
		return nil, err
	}

	p, err := NewProtocolFromConnV2(ctx, conn, stream, sendSubject, jsOpts, opts...)
	if err != nil {
		conn.Close()
		return nil, err
	}

	p.connOwned = true

	return p, nil
}

func NewProtocolFromConn(conn *nats.Conn, stream, sendSubject, receiveSubject string, jsOpts []nats.JSOpt, subOpts []nats.SubOpt, opts ...ProtocolOption) (*Protocol, error) {
	var err error
	p := &Protocol{
		Conn: conn,
	}

	if err := p.applyOptions(opts...); err != nil {
		return nil, err
	}

	if p.Consumer, err = NewConsumerFromConn(conn, stream, receiveSubject, jsOpts, subOpts, p.consumerOptions...); err != nil {
		return nil, err
	}

	if p.Sender, err = NewSenderFromConn(conn, stream, sendSubject, jsOpts, p.senderOptions...); err != nil {
		return nil, err
	}

	return p, nil
}

func NewProtocolFromConnV2(ctx context.Context, conn *nats.Conn, stream, sendSubject string, jsOpts []jetstream.JetStreamOpt, opts ...ProtocolOption) (*Protocol, error) {
	var err error
	var js jetstream.JetStream
	p := &Protocol{
		Conn: conn,
	}

	if err := p.applyOptions(opts...); err != nil {
		return nil, err
	}

	if js, err = jetstream.New(conn, jsOpts...); err != nil {
		return nil, err
	}
	streamConfig := jetstream.StreamConfig{Name: stream, Subjects: []string{sendSubject}}
	if _, err := js.CreateOrUpdateStream(ctx, streamConfig); err != nil {
		return nil, err
	}

	if p.Consumer, err = NewConsumerFromConnV2(ctx, conn, jsOpts, p.consumerOptions...); err != nil {
		return nil, err
	}

	if p.Sender, err = NewSenderFromConnV2(ctx, conn, sendSubject, jsOpts, p.senderOptions...); err != nil {
		return nil, err
	}

	return p, nil
}

// Send implements Sender.Send
func (p *Protocol) Send(ctx context.Context, in binding.Message, transformers ...binding.Transformer) error {
	return p.Sender.Send(ctx, in, transformers...)
}

func (p *Protocol) OpenInbound(ctx context.Context) error {
	return p.Consumer.OpenInbound(ctx)
}

// Receive implements Receiver.Receive
func (p *Protocol) Receive(ctx context.Context) (binding.Message, error) {
	return p.Consumer.Receive(ctx)
}

// Close implements Closer.Close
func (p *Protocol) Close(ctx context.Context) error {
	if p.connOwned {
		defer p.Conn.Close()
	}

	if err := p.Consumer.Close(ctx); err != nil {
		return err
	}

	if err := p.Sender.Close(ctx); err != nil {
		return err
	}

	return nil
}

func (p *Protocol) applyOptions(opts ...ProtocolOption) error {
	for _, fn := range opts {
		if err := fn(p); err != nil {
			return err
		}
	}
	return nil
}

var _ protocol.Receiver = (*Protocol)(nil)
var _ protocol.Sender = (*Protocol)(nil)
var _ protocol.Opener = (*Protocol)(nil)
var _ protocol.Closer = (*Protocol)(nil)
