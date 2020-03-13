package amqp

import (
	"context"

	"pack.ag/amqp"

	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/binding/spec"
	"github.com/cloudevents/sdk-go/pkg/binding/transformer"
	cecontext "github.com/cloudevents/sdk-go/pkg/context"
	"github.com/cloudevents/sdk-go/pkg/transport"
)

type Protocol struct {
	connOpts         []amqp.ConnOption
	sessionOpts      []amqp.SessionOption
	senderLinkOpts   []amqp.LinkOption
	receiverLinkOpts []amqp.LinkOption

	// Encoding
	Encoding Encoding

	// AMQP
	Client  *amqp.Client
	Session *amqp.Session
	Node    string

	// Sender
	Sender                  transport.Sender
	SenderContextDecorators []func(context.Context) context.Context

	// Receiver
	Receiver transport.Receiver
}

// New creates a new amqp transport.
func New(server, queue string, opts ...Option) (*Protocol, error) {
	t := &Protocol{
		Node:             queue,
		connOpts:         []amqp.ConnOption(nil),
		sessionOpts:      []amqp.SessionOption(nil),
		senderLinkOpts:   []amqp.LinkOption(nil),
		receiverLinkOpts: []amqp.LinkOption(nil),
	}
	if err := t.applyOptions(opts...); err != nil {
		return nil, err
	}

	client, err := amqp.Dial(server, t.connOpts...)
	if err != nil {
		return nil, err
	}
	t.Client = client

	// Open a session
	session, err := client.NewSession(t.sessionOpts...)
	if err != nil {
		_ = client.Close()
		return nil, err
	}
	t.Session = session

	t.senderLinkOpts = append(t.senderLinkOpts, amqp.LinkTargetAddress(queue))

	// Create a sender
	sender, err := session.NewSender(t.senderLinkOpts...)
	if err != nil {
		_ = client.Close()
		_ = session.Close(context.Background())
		return nil, err
	}
	// TODO: in the future we might have more than one sender.
	t.Sender, t.SenderContextDecorators = t.applyEncoding(sender)
	return t, nil
}

func (t *Protocol) applyEncoding(amqpSender *amqp.Sender) (transport.Sender, []func(context.Context) context.Context) {
	switch t.Encoding {
	case BinaryV03:
		return NewSender(
			amqpSender,
			WithTransformer(transformer.Version(spec.V03)),
		), []func(context.Context) context.Context{binding.WithForceBinary}
	case BinaryV1:
		return NewSender(
			amqpSender,
			WithTransformer(transformer.Version(spec.V1)),
		), []func(context.Context) context.Context{binding.WithForceBinary}
	case StructuredV03:
		return NewSender(
			amqpSender,
			WithTransformer(transformer.Version(spec.V03)),
		), []func(context.Context) context.Context{binding.WithForceStructured}
	case StructuredV1:
		return NewSender(
			amqpSender,
			WithTransformer(transformer.Version(spec.V1)),
		), []func(context.Context) context.Context{binding.WithForceStructured}
	}
	return NewSender(amqpSender), []func(context.Context) context.Context{}
}

func (t *Protocol) applyOptions(opts ...Option) error {
	for _, fn := range opts {
		if err := fn(t); err != nil {
			return err
		}
	}
	return nil
}

// StartReceiver implements Protocol.StartReceiver
// NOTE: This is a blocking call.
func (t *Protocol) StartInbound(ctx context.Context) error {
	logger := cecontext.LoggerFrom(ctx)
	logger.Info("StartReceiver on ", t.Node)

	t.receiverLinkOpts = append(t.receiverLinkOpts, amqp.LinkSourceAddress(t.Node))
	receiver, err := t.Session.NewReceiver(t.receiverLinkOpts...)
	if err != nil {
		return err
	}
	t.Receiver = NewReceiver(receiver)
	return nil
}

// HasTracePropagation implements Protocol.HasTracePropagation
func (t *Protocol) HasTracePropagation() bool {
	return false
}

func (t *Protocol) Close() error {
	return t.Client.Close()
}

func (t *Protocol) Send(ctx context.Context, in binding.Message) error {
	return t.Sender.Send(ctx, in)
}

func (t *Protocol) Receive(ctx context.Context) (binding.Message, error) {
	return t.Receiver.Receive(ctx)
}
