/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package nats_jetstream

import (
	"context"
	"io"
	"strings"
	"sync"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"

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

// MsgHandlerV2 implements jetstream.MessageHandler and publishes messages onto our internal incoming channel to be delivered
// via r.Receive(ctx)
func (r *Receiver) MsgHandlerV2(msg jetstream.Msg) {
	// *nats.Msg does not implement jetstream.Msg.
	// jetstream.MessageHandler.Ack() returns error where nats.Msg.Ack() returns nothing
	natsMessage := &nats.Msg{
		Subject: msg.Subject(),
		Reply:   msg.Reply(),
		Header:  msg.Headers(),
		Data:    msg.Data(),
	}
	r.incoming <- msgErr{msg: NewMessage(natsMessage)}
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

	// v1 options
	Jsm        nats.JetStreamContext
	Subject    string
	Subscriber Subscriber
	SubOpt     []nats.SubOpt

	// v2 options
	JetStream             jetstream.JetStream
	ConsumerConfig        *jetstream.ConsumerConfig
	OrderedConsumerConfig *jetstream.OrderedConsumerConfig
	PullConsumeOpt        []jetstream.PullConsumeOpt
	JetstreamConsumer     jetstream.Consumer

	Conn          *nats.Conn
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

// NewConsumerV2 consumes from filterSubjects given in ConsumerConfig or OrderedConsumerConfig
// embedded in ConsumerOption.
// See: WithConsumerConfig(...) and WithOrderedConsumerConfig(...)
func NewConsumerV2(ctx context.Context, url string, natsOpts []nats.Option, jsOpts []jetstream.JetStreamOpt, opts ...ConsumerOption) (*Consumer, error) {
	conn, err := nats.Connect(url, natsOpts...)
	if err != nil {
		return nil, err
	}

	c, err := NewConsumerFromConnV2(ctx, conn, jsOpts, opts...)
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

	subjectMatch := stream + ".*"
	if strings.Count(strings.TrimPrefix(subject, stream), ".") > 1 {
		// More than one "." in the remainder of subject, use ".>" to match
		subjectMatch = stream + ".>"
	}
	if !strings.HasPrefix(subject, stream) {
		// Use an empty subject parameter in conjunction with
		// nats.ConsumerFilterSubjects
		subjectMatch = ""
	}

	if streamInfo == nil || err != nil && err.Error() == "stream not found" {
		_, err = jsm.AddStream(&nats.StreamConfig{
			Name:     stream,
			Subjects: []string{subjectMatch},
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

// NewConsumerFromConnV2 consumes from filterSubjects given in ConsumerConfig or OrderedConsumerConfig
// embedded in ConsumerOption. See: WithConsumerConfig(...) and WithOrderedConsumerConfig(...)
func NewConsumerFromConnV2(ctx context.Context, conn *nats.Conn, jsOpts []jetstream.JetStreamOpt, opts ...ConsumerOption) (*Consumer, error) {
	var js jetstream.JetStream
	var err error
	if js, err = jetstream.New(conn, jsOpts...); err != nil {
		return nil, err
	}

	c := &Consumer{
		Receiver:      *NewReceiver(),
		Conn:          conn,
		JetStream:     js,
		internalClose: make(chan struct{}, 1),
	}

	if err = c.applyOptions(opts...); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Consumer) OpenInbound(ctx context.Context) error {
	c.subMtx.Lock()
	defer c.subMtx.Unlock()

	var sub *nats.Subscription
	var consumeContext jetstream.ConsumeContext
	var err error
	version := c.getVersion()
	switch version {
	case 0:
		return ErrNoJetstream
	case 1:
		if sub, err = c.Subscriber.Subscribe(c.Jsm, c.Subject, c.MsgHandler, c.SubOpt...); err != nil {
			return err
		}
	case 2:
		if err = c.createConsumer(ctx); err != nil {
			return err
		}
		if consumeContext, err = c.JetstreamConsumer.Consume(c.MsgHandlerV2, c.PullConsumeOpt...); err != nil {
			return err
		}
	}

	// Wait until external or internal context done
	select {
	case <-ctx.Done():
	case <-c.internalClose:
	}

	// Finish to consume messages in the queue and close the subscription
	if consumeContext != nil {
		consumeContext.Drain()
		return nil
	}
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

// createConsumer creates a consumer based on the configured consumer config
func (c *Consumer) createConsumer(ctx context.Context) error {
	var err error
	var stream string
	if stream, err = c.getStreamFromSubjects(ctx); err != nil {
		return err
	}
	consumerType := c.getConsumerType()
	switch consumerType {
	case ConsumerType_Unknown:
		return ErrNoConsumerConfig
	case ConsumerType_Ordinary:
		c.JetstreamConsumer, err = c.JetStream.CreateOrUpdateConsumer(ctx, stream, *c.ConsumerConfig)
	case ConsumerType_Ordered:
		c.JetstreamConsumer, err = c.JetStream.OrderedConsumer(ctx, stream, *c.OrderedConsumerConfig)
	}
	return err
}

// getStreamFromSubjects finds the unique stream for the set of filter subjects
// If more than one stream is found, returns ErrMoreThanOneStream
func (c *Consumer) getStreamFromSubjects(ctx context.Context) (string, error) {
	var subjects []string
	if c.ConsumerConfig != nil && c.ConsumerConfig.FilterSubject != "" {
		subjects = []string{c.ConsumerConfig.FilterSubject}
	}
	if c.ConsumerConfig != nil && len(c.ConsumerConfig.FilterSubjects) > 0 {
		subjects = c.ConsumerConfig.FilterSubjects
	}
	if c.OrderedConsumerConfig != nil && len(c.OrderedConsumerConfig.FilterSubjects) > 0 {
		subjects = c.OrderedConsumerConfig.FilterSubjects
	}
	var finalStream string
	for i, subject := range subjects {
		currentStream, err := c.JetStream.StreamNameBySubject(ctx, subject)
		if err != nil {
			return "", err
		}
		if i == 0 {
			finalStream = currentStream
			continue
		}
		if finalStream != currentStream {
			return "", ErrMoreThanOneStream
		}
	}
	return finalStream, nil
}

// getVersion returns whether the consumer uses the older jsm or newer jetstream package
// 0 - None, 1 - jsm, 2 - jetstream
func (c *Consumer) getVersion() int {
	if c.Jsm != nil {
		return 1
	}
	if c.JetStream != nil {
		return 2
	}
	return 0
}

// getConsumerType returns consumer type based on what has been configured
func (c *Consumer) getConsumerType() ConsumerType {
	if c.ConsumerConfig != nil {
		return ConsumerType_Ordinary
	}
	if c.OrderedConsumerConfig != nil {
		return ConsumerType_Ordered
	}
	return ConsumerType_Unknown
}

var _ protocol.Opener = (*Consumer)(nil)
var _ protocol.Receiver = (*Consumer)(nil)
var _ protocol.Closer = (*Consumer)(nil)
