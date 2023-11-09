/*
 Copyright 2023 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package grpc

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
	"google.golang.org/grpc"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/protocol"

	cecontext "github.com/cloudevents/sdk-go/v2/context"
)

// protocol for grpc
// define protocol for grpc

type Protocol struct {
	client          pb.CloudEventServiceClient
	publishOption   *PublishOption
	subscribeOption *SubscribeOption
	// receiver
	incoming chan *pb.CloudEvent
	// inOpen
	openerMutex sync.Mutex

	closeChan chan struct{}
}

var (
	_ protocol.Sender   = (*Protocol)(nil)
	_ protocol.Opener   = (*Protocol)(nil)
	_ protocol.Receiver = (*Protocol)(nil)
	_ protocol.Closer   = (*Protocol)(nil)
)

// new create grpc protocol
func NewProtocol(clientConn grpc.ClientConnInterface, opts ...Option) (*Protocol, error) {
	if clientConn == nil {
		return nil, fmt.Errorf("the client connection must not be nil")
	}

	p := &Protocol{
		client: pb.NewCloudEventServiceClient(clientConn),
		// subClient:
		incoming:  make(chan *pb.CloudEvent),
		closeChan: make(chan struct{}),
	}

	if err := p.applyOptions(opts...); err != nil {
		return nil, err
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

func (p *Protocol) Send(ctx context.Context, m binding.Message, transformers ...binding.Transformer) error {
	var err error
	defer m.Finish(err)

	msg := &pb.CloudEvent{}
	err = WritePubMessage(ctx, m, msg, transformers...)
	if err != nil {
		return err
	}

	_, err = p.client.Publish(ctx, msg)
	if err != nil {
		return err
	}
	return err
}

func (p *Protocol) OpenInbound(ctx context.Context) error {
	if p.subscribeOption == nil {
		return fmt.Errorf("the subscribe option must not be nil")
	}

	if p.subscribeOption.Topic == "" {
		return fmt.Errorf("the subscribe option topic must not be empty")
	}

	p.openerMutex.Lock()
	defer p.openerMutex.Unlock()

	logger := cecontext.LoggerFrom(ctx)

	logger.Infof("subscribing to topic: %v", p.subscribeOption.Topic)
	subClient, err := p.client.Subscribe(ctx, &pb.Subscription{
		Subscription: p.subscribeOption.Topic,
	})
	if err != nil {
		return err
	}

	go func() {
		for {
			msg, err := subClient.Recv()
			if err != nil {
				return
			}
			p.incoming <- msg
		}
	}()

	// Wait until external or internal context done
	select {
	case <-ctx.Done():
	case <-p.closeChan:
	}

	return nil
}

// Receive implements Receiver.Receive
func (p *Protocol) Receive(ctx context.Context) (binding.Message, error) {
	select {
	case m, ok := <-p.incoming:
		if !ok {
			return nil, io.EOF
		}
		msg := NewMessage(m)
		return msg, nil
	case <-ctx.Done():
		return nil, io.EOF
	}
}

func (p *Protocol) Close(ctx context.Context) error {
	close(p.closeChan)
	return nil
}
