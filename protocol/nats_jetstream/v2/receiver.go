/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package nats_jetstream

import (
	"context"
	"io"
	"sync"

	"github.com/nats-io/nats.go"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/protocol"
)

type msgErr struct {
	msg binding.Message
}

type Receiver struct {
	incoming chan msgErr
}

// NewReceiver creates a new protocol.Receiver responsible for receiving messages.
func NewReceiver() *Receiver {
	return &Receiver{
		incoming: make(chan msgErr),
	}
}

// MsgHandler implements nats.MsgHandler and publishes messages onto our internal incoming channel to be delivered
// via r.Receive(ctx)
func (r *Receiver) MsgHandler(msg *nats.Msg) {
	r.incoming <- msgErr{msg: NewMessage(msg)}
}

// Receive implements Receiver.Receive.
func (r *Receiver) Receive(ctx context.Context) (binding.Message, error) {
	select {
	case msgErr, ok := <-r.incoming:
		if !ok {
			return nil, io.EOF
		}
		return msgErr.msg, nil
	case <-ctx.Done():
		return nil, io.EOF
	}
}

type Consumer struct {
	Receiver

	Conn       *nats.Conn
	Jsm        nats.JetStreamContext
	Subject    string
	Subscriber Subscriber
	SubOpt     []nats.SubOpt

	subMtx        sync.Mutex
	internalClose chan struct{}
	connOwned     bool
}

func NewConsumer(url, stream, subject string, natsOpts []nats.Option, jsmOpts []nats.JSOpt, subOpts []nats.SubOpt, opts ...ConsumerOption) (*Consumer, error) {
	conn, err := nats.Connect(url, natsOpts...)
	if err != nil {
		return nil, err
	}

	c, err := NewConsumerFromConn(conn, stream, subject, jsmOpts, subOpts, opts...)
	if err != nil {
		conn.Close()
		return nil, err
	}

	c.connOwned = true

	return c, err
}

func NewConsumerFromConn(conn *nats.Conn, stream, subject string, jsmOpts []nats.JSOpt, subOpts []nats.SubOpt, opts ...ConsumerOption) (*Consumer, error) {
	jsm, err := conn.JetStream(jsmOpts...)
	if err != nil {
		return nil, err
	}

	streamInfo, err := jsm.StreamInfo(stream, jsmOpts...)

	if streamInfo == nil || err != nil && err.Error() == "stream not found" {
		_, err = jsm.AddStream(&nats.StreamConfig{
			Name:     stream,
			Subjects: []string{stream + ".*"},
		})
		if err != nil {
			return nil, err
		}
	}

	c := &Consumer{
		Receiver:      *NewReceiver(),
		Conn:          conn,
		Jsm:           jsm,
		Subject:       subject,
		SubOpt:        subOpts,
		Subscriber:    &RegularSubscriber{},
		internalClose: make(chan struct{}, 1),
	}

	err = c.applyOptions(opts...)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Consumer) OpenInbound(ctx context.Context) error {
	c.subMtx.Lock()
	defer c.subMtx.Unlock()

	// Subscribe
	sub, err := c.Subscriber.Subscribe(c.Jsm, c.Subject, c.MsgHandler, c.SubOpt...)
	if err != nil {
		return err
	}

	// Wait until external or internal context done
	select {
	case <-ctx.Done():
	case <-c.internalClose:
	}

	// Finish to consume messages in the queue and close the subscription
	return sub.Drain()
}

func (c *Consumer) Close(ctx context.Context) error {
	// Before closing, let's be sure OpenInbound completes
	// We send a signal to close and then we lock on subMtx in order
	// to wait OpenInbound to finish draining the queue
	c.internalClose <- struct{}{}
	c.subMtx.Lock()
	defer c.subMtx.Unlock()

	if c.connOwned {
		c.Conn.Close()
	}

	close(c.internalClose)

	return nil
}

func (c *Consumer) applyOptions(opts ...ConsumerOption) error {
	for _, fn := range opts {
		if err := fn(c); err != nil {
			return err
		}
	}
	return nil
}

var _ protocol.Opener = (*Consumer)(nil)
var _ protocol.Receiver = (*Consumer)(nil)
var _ protocol.Closer = (*Consumer)(nil)
