package stan

import (
	"context"
	"errors"
	"fmt"
	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/nats-io/stan.go"
)

var ErrSubscriptionAlreadyOpen = errors.New("subscription already open")

// Protocol is a reference implementation for using the CloudEvents binding
// integration. Protocol acts as both a STAN client and a STAN handler.
type Protocol struct {
	Conn stan.Conn

	Consumer        *Consumer
	consumerOptions []ConsumerOption

	Sender        *Sender
	senderOptions []SenderOption

	connOwned bool // whether this protocol created the stan connection
}

// NewProtocol creates a new STAN protocol including managing the lifecycle of the connection
func NewProtocol(clusterID, clientID, sendSubject, receiveSubject string, stanOpts []stan.Option, opts ...ProtocolOption) (*Protocol, error) {
	conn, err := stan.Connect(clusterID, clientID, stanOpts...)
	if err != nil {
		return nil, err
	}

	p, err := NewProtocolFromConn(conn, sendSubject, receiveSubject, opts...)
	if err != nil {
		if err2 := conn.Close(); err2 != nil {
			return nil, fmt.Errorf("failed to close conn: %s, when recovering from err: %w", err2, err)
		}
		return nil, err
	}

	p.connOwned = true

	return p, nil
}

// NewProtocolFromConn creates a new STAN protocol but leaves managing the lifecycle of the connection up to the caller
func NewProtocolFromConn(conn stan.Conn, sendSubject, receiveSubject string, opts ...ProtocolOption) (*Protocol, error) {
	var err error
	p := &Protocol{
		Conn: conn,
	}

	if err := p.applyOptions(opts...); err != nil {
		return nil, err
	}

	if p.Consumer, err = NewConsumerFromConn(conn, receiveSubject, p.consumerOptions...); err != nil {
		return nil, err
	}

	if p.Sender, err = NewSenderFromConn(conn, sendSubject, p.senderOptions...); err != nil {
		return nil, err
	}

	return p, nil
}

func (p *Protocol) applyOptions(opts ...ProtocolOption) error {
	for _, fn := range opts {
		if err := fn(p); err != nil {
			return err
		}
	}
	return nil
}

// Send implements Sender.Send
func (p *Protocol) Send(ctx context.Context, in binding.Message) error {
	return p.Sender.Send(ctx, in)
}

// OpenInbound implements Opener.OpenInbound
func (p *Protocol) OpenInbound(ctx context.Context) error {
	return p.Consumer.OpenInbound(ctx)
}

// Receive implements Receiver.Receive
func (p *Protocol) Receive(ctx context.Context) (binding.Message, error) {
	return p.Consumer.Receive(ctx)
}

// Close implements Closer.Close
func (p *Protocol) Close(ctx context.Context) error {
	if err := p.Consumer.Close(ctx); err != nil {
		return err
	}

	if err := p.Sender.Close(ctx); err != nil {
		return err
	}

	if p.connOwned {
		return p.Conn.Close()
	}

	return nil
}
