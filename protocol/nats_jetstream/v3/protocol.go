/*
 Copyright 2024 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package nats_jetstream

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/protocol"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

// Protocol is a reference implementation for using the CloudEvents binding
// integration. Protocol acts as both a NATS client and a NATS handler.
type Protocol struct {

	// nats connection
	conn     *nats.Conn
	url      string
	natsOpts []nats.Option

	// jetstream options
	jetStreamOpts []jetstream.JetStreamOpt
	jetStream     jetstream.JetStream

	// receiver
	incoming              chan msgErr
	subMtx                sync.Mutex
	internalClose         chan struct{}
	consumerConfig        *jetstream.ConsumerConfig
	orderedConsumerConfig *jetstream.OrderedConsumerConfig
	pullConsumeOpts       []jetstream.PullConsumeOpt
	jetstreamConsumer     jetstream.Consumer

	// sender
	publishOpts []jetstream.PublishOpt
	sendSubject string
}

// New creates a new NATS protocol.
func New(ctx context.Context, opts ...ProtocolOption) (*Protocol, error) {
	p := &Protocol{
		incoming:      make(chan msgErr),
		internalClose: make(chan struct{}, 1),
	}
	if err := p.applyOptions(opts...); err != nil {
		return nil, err
	}

	if err := validateOptions(p); err != nil {
		return nil, err
	}

	var errConnection error
	defer func() {
		// close connection if an error occurs
		if errConnection != nil {
			p.Close(ctx)
		}
	}()

	// if a URL was given create the nats connection
	if p.conn == nil {
		p.conn, errConnection = nats.Connect(p.url, p.natsOpts...)
		if errConnection != nil {
			return nil, errConnection
		}
	}

	if p.jetStream, errConnection = jetstream.New(p.conn, p.jetStreamOpts...); errConnection != nil {
		return nil, errConnection
	}

	return p, nil
}

// Send sends messages. Send implements Sender.Sender
func (p *Protocol) Send(ctx context.Context, in binding.Message, transformers ...binding.Transformer) (err error) {
	if p.sendSubject == "" {
		return newValidationError(fieldSendSubject, messageNoSendSubject)
	}

	defer func() {
		if err2 := in.Finish(err); err2 != nil {
			if err == nil {
				err = err2
			} else {
				err = fmt.Errorf("failed to call in.Finish() when error already occurred: %s: %w", err2.Error(), err)
			}
		}
	}()

	if _, err = p.jetStream.StreamNameBySubject(ctx, p.sendSubject); err != nil {
		return err
	}

	writer := new(bytes.Buffer)
	header, err := WriteMsg(ctx, in, writer, transformers...)
	if err != nil {
		return err
	}

	natsMsg := &nats.Msg{
		Subject: p.sendSubject,
		Data:    writer.Bytes(),
		Header:  header,
	}

	_, err = p.jetStream.PublishMsg(ctx, natsMsg, p.publishOpts...)

	return err
}

// OpenInbound implements Opener.OpenInbound
func (p *Protocol) OpenInbound(ctx context.Context) error {
	p.subMtx.Lock()
	defer p.subMtx.Unlock()

	var consumeContext jetstream.ConsumeContext
	var err error
	if err = p.createJetstreamConsumer(ctx); err != nil {
		return err
	}
	if consumeContext, err = p.jetstreamConsumer.Consume(p.MsgHandler, p.pullConsumeOpts...); err != nil {
		return err
	}

	// Wait until external or internal context done
	select {
	case <-ctx.Done():
	case <-p.internalClose:
	}

	// Finish to consume messages in the queue and close the subscription
	if consumeContext != nil {
		consumeContext.Drain()
	}
	return nil
}

// Receive implements Receiver.Receive.
func (p *Protocol) Receive(ctx context.Context) (binding.Message, error) {
	select {
	case msgErr, ok := <-p.incoming:
		if !ok {
			return nil, io.EOF
		}
		return msgErr.msg, nil
	case <-ctx.Done():
		return nil, io.EOF
	}
}

// MsgHandler implements nats.MsgHandler and publishes messages onto our internal incoming channel to be delivered
// via r.Receive(ctx)
func (p *Protocol) MsgHandler(msg jetstream.Msg) {
	p.incoming <- msgErr{msg: NewMessage(msg)}
}

// Close must be called after use to releases internal resources.
// If WithURL option was used, the NATS connection internally opened will be closed.
func (p *Protocol) Close(ctx context.Context) error {
	// Before closing, let's be sure OpenInbound completes
	// We send a signal to close and then we lock on subMtx in order
	// to wait OpenInbound to finish draining the queue
	p.internalClose <- struct{}{}
	p.subMtx.Lock()
	defer p.subMtx.Unlock()

	// if an URL was provided, then we must close the internally opened NATS connection
	// since the connection is not exposed.
	// If the connection was passed in, then leave the connection available.
	if p.url != "" && p.conn != nil {
		p.conn.Close()
	}

	close(p.internalClose)

	return nil
}

// applyOptions at the protocol layer should run before the sender and receiver are created.
// This allows the protocol to create a nats connection that can be shared for both the sender and receiver.
func (p *Protocol) applyOptions(opts ...ProtocolOption) error {
	for _, fn := range opts {
		if err := fn(p); err != nil {
			return err
		}
	}
	return nil
}

// createJetstreamConsumer creates a consumer based on the configured consumer config
func (p *Protocol) createJetstreamConsumer(ctx context.Context) error {
	var err error
	var stream string
	if stream, err = p.getStreamFromSubjects(ctx); err != nil {
		return err
	}
	var consumerErr error
	if p.consumerConfig != nil {
		p.jetstreamConsumer, consumerErr = p.jetStream.CreateOrUpdateConsumer(ctx, stream, *p.consumerConfig)
	} else if p.orderedConsumerConfig != nil {
		p.jetstreamConsumer, consumerErr = p.jetStream.OrderedConsumer(ctx, stream, *p.orderedConsumerConfig)
	} else {
		return newValidationError(fieldConsumerConfig, messageNoConsumerConfig)
	}
	return consumerErr
}

// getStreamFromSubjects finds the unique stream for the set of filter subjects
// If more than one stream is found, returns ErrMoreThanOneStream
func (p *Protocol) getStreamFromSubjects(ctx context.Context) (string, error) {
	var subjects []string
	if p.consumerConfig != nil && p.consumerConfig.FilterSubject != "" {
		subjects = []string{p.consumerConfig.FilterSubject}
	}
	if p.consumerConfig != nil && len(p.consumerConfig.FilterSubjects) > 0 {
		subjects = p.consumerConfig.FilterSubjects
	}
	if p.orderedConsumerConfig != nil && len(p.orderedConsumerConfig.FilterSubjects) > 0 {
		subjects = p.orderedConsumerConfig.FilterSubjects
	}
	if len(subjects) == 0 {
		return "", newValidationError(fieldFilterSubjects, messageNoFilterSubjects)
	}
	var finalStream string
	for i, subject := range subjects {
		currentStream, err := p.jetStream.StreamNameBySubject(ctx, subject)
		if err != nil {
			return "", err
		}
		if i == 0 {
			finalStream = currentStream
			continue
		}
		if finalStream != currentStream {
			return "", newValidationError(fieldFilterSubjects, messageMoreThanOneStream)
		}
	}
	return finalStream, nil
}

type msgErr struct {
	msg binding.Message
}

var _ protocol.Receiver = (*Protocol)(nil)
var _ protocol.Sender = (*Protocol)(nil)
var _ protocol.Opener = (*Protocol)(nil)
var _ protocol.Closer = (*Protocol)(nil)
