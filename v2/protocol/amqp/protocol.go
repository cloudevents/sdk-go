package amqp

import (
	"context"

	"pack.ag/amqp"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/protocol"
)

type Protocol struct {
	connOpts         []amqp.ConnOption
	sessionOpts      []amqp.SessionOption
	senderLinkOpts   []amqp.LinkOption
	receiverLinkOpts []amqp.LinkOption

	// AMQP
	Client  *amqp.Client
	Session *amqp.Session
	Node    string

	// Sender
	Sender                  *sender
	SenderContextDecorators []func(context.Context) context.Context

	// Receiver
	Receiver *receiver
}

// NewProtocol creates a new amqp transport.
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

	return NewProtocolFromClient(client, session, queue, opts...)
}

func (t *Protocol) applyOptions(opts ...Option) error {
	for _, fn := range opts {
		if err := fn(t); err != nil {
			return err
		}
	}
	return nil
}

func (t *Protocol) Close(ctx context.Context) error {
	if t.Sender != nil {
		if err := t.Sender.Close(ctx); err != nil {
			return err
		}
	}

	if t.Receiver != nil {
		if err := t.Receiver.Close(ctx); err != nil {
			return err
		}
	}

	return t.Client.Close()
}

func (t *Protocol) Send(ctx context.Context, in binding.Message) error {
	return t.Sender.Send(ctx, in)
}

func (t *Protocol) Receive(ctx context.Context) (binding.Message, error) {
	return t.Receiver.Receive(ctx)
}

var _ protocol.Sender = (*Protocol)(nil)
var _ protocol.Receiver = (*Protocol)(nil)
var _ protocol.Closer = (*Protocol)(nil)
