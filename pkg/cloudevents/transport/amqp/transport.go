package amqp

import (
	"context"

	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/binding/format"
	"github.com/cloudevents/sdk-go/pkg/binding/spec"
	bindings_amqp "github.com/cloudevents/sdk-go/pkg/bindings/amqp"
	cecontext "github.com/cloudevents/sdk-go/pkg/cloudevents/context"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport"
	"pack.ag/amqp"
)

// Transport adheres to transport.Transport.
var _ transport.Transport = (*Transport)(nil)

const (
	// TransportName is the name of this transport.
	TransportName = "AMQP"
)

type Transport struct {
	binding.Transport
	connOpts         []amqp.ConnOption
	sessionOpts      []amqp.SessionOption
	senderLinkOpts   []amqp.LinkOption
	receiverLinkOpts []amqp.LinkOption

	// Encoding
	Encoding Encoding

	// AMQP
	Client  *amqp.Client
	Session *amqp.Session
	Sender  *amqp.Sender
	Node    string

	// Receiver
	Receiver transport.Receiver
	// Converter is invoked if the incoming transport receives an undecodable
	// message.
	Converter transport.Converter
}

// New creates a new amqp transport.
func New(server, queue string, opts ...Option) (*Transport, error) {
	t := &Transport{
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
	t.Transport.Sender = t.applyEncoding(bindings_amqp.Sender{AMQP: sender})
	return t, nil
}

func (t *Transport) applyEncoding(s bindings_amqp.Sender) binding.Sender {
	switch t.Encoding {
	case BinaryV02:
		return binding.VersionSender(s, spec.V02)
	case BinaryV03:
		return binding.VersionSender(s, spec.V03)
	case BinaryV1:
		return binding.VersionSender(s, spec.V1)
	case StructuredV02:
		return binding.StructSender(binding.VersionSender(s, spec.V02), format.JSON)
	case StructuredV03:
		return binding.StructSender(binding.VersionSender(s, spec.V03), format.JSON)
	case StructuredV1:
		return binding.StructSender(binding.VersionSender(s, spec.V1), format.JSON)
	}
	return s
}

func (t *Transport) applyOptions(opts ...Option) error {
	for _, fn := range opts {
		if err := fn(t); err != nil {
			return err
		}
	}
	return nil
}

// SetConverter implements Transport.SetConverter
func (t *Transport) SetConverter(c transport.Converter) {
	t.Converter = c
}

// HasConverter implements Transport.HasConverter
func (t *Transport) HasConverter() bool {
	return t.Converter != nil
}

// StartReceiver implements Transport.StartReceiver
// NOTE: This is a blocking call.
func (t *Transport) StartReceiver(ctx context.Context) error {
	logger := cecontext.LoggerFrom(ctx)
	logger.Info("StartReceiver on ", t.Node)

	t.receiverLinkOpts = append(t.receiverLinkOpts, amqp.LinkSourceAddress(t.Node))
	receiver, err := t.Session.NewReceiver(t.receiverLinkOpts...)
	if err != nil {
		return err
	}
	t.Transport.Receiver = bindings_amqp.Receiver{AMQP: receiver}
	return t.Transport.StartReceiver(ctx)
}

func (t *Transport) Close() error {
	return t.Client.Close()
}
