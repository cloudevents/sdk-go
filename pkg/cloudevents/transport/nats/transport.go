package nats

import (
	"context"
	"fmt"
	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/binding/spec"
	"github.com/cloudevents/sdk-go/pkg/binding/transcoder"
	bindings_nats "github.com/cloudevents/sdk-go/pkg/bindings/nats"
	"go.uber.org/zap"

	cecontext "github.com/cloudevents/sdk-go/pkg/cloudevents/context"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport"
	"github.com/nats-io/nats.go"
)

// Transport adheres to transport.Transport.
var _ transport.Transport = (*Transport)(nil)

const (
	// TransportName is the name of this transport.
	TransportName = "NATS"
)

// Transport acts as both a NATS client and a NATS handler.
type Transport struct {
	binding.BindingTransport

	Encoding    Encoding
	Conn        *nats.Conn
	ConnOptions []nats.Option
	NatsURL     string
	Subject     string

	sub *nats.Subscription

	// Converter is invoked if the incoming transport receives an undecodable
	// message.
	Converter transport.Converter
}

// New creates a new NATS transport.
func New(natsURL, subject string, opts ...Option) (*Transport, error) {
	t := &Transport{
		Subject:     subject,
		NatsURL:     natsURL,
		ConnOptions: []nats.Option{},
	}

	err := t.applyOptions(opts...)
	if err != nil {
		return nil, err
	}

	err = t.connect()
	if err != nil {
		return nil, err
	}

	t.BindingTransport.Sender, t.BindingTransport.SenderContextDecorators = t.applyEncoding()
	return t, nil
}

func (t *Transport) connect() error {
	var err error

	t.Conn, err = nats.Connect(t.NatsURL, t.ConnOptions...)

	return err
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
func (t *Transport) StartReceiver(ctx context.Context) (err error) {
	logger := cecontext.LoggerFrom(ctx)
	logger.Info("StartReceiver on ", t.Subject)

	if t.Conn == nil {
		return fmt.Errorf("no active nats connection")
	}
	if t.sub != nil {
		return fmt.Errorf("already subscribed")
	}

	// TODO: there could be more than one subscription. Might have to do a map
	// of subject to subscription.

	if t.Subject == "" {
		return fmt.Errorf("subject required for nats listen")
	}

	t.sub, err = t.Conn.SubscribeSync(t.Subject)
	if err != nil {
		return err
	}

	defer func() {
		if t.sub != nil && t.sub.IsValid() {
			err2 := t.sub.Unsubscribe()
			if err2 != nil {
				logger.Errorw("failed to unsubscribe transport", zap.Error(err2))
			}
			// TODO: We're logging the "Unsubscribe" error so should we be hiding deeper errors?
			if err != nil {
				err = err2 // Set the returned error if not already set.
			}
			t.sub = nil
		}
	}()

	t.BindingTransport.Receiver = bindings_nats.NewReceiver(t.sub)

	// this blocks until ctx is closed
	return t.BindingTransport.StartReceiver(ctx)
}

// HasTracePropagation implements Transport.HasTracePropagation
func (t *Transport) HasTracePropagation() bool {
	return false
}

func (t *Transport) applyEncoding() (binding.Sender, []func(context.Context) context.Context) {
	ctxDecorator := []func(context.Context) context.Context{binding.WithForceStructured}

	switch t.Encoding {
	case Default:
		fallthrough
	case StructuredV02:
		return bindings_nats.NewSender(
			t.Conn,
			t.Subject,
			bindings_nats.WithTranscoder(transcoder.Version(spec.V02)),
		), ctxDecorator
	case StructuredV03:
		return bindings_nats.NewSender(
			t.Conn,
			t.Subject,
			bindings_nats.WithTranscoder(transcoder.Version(spec.V03)),
		), ctxDecorator
	case StructuredV1:
		return bindings_nats.NewSender(
			t.Conn,
			t.Subject,
			bindings_nats.WithTranscoder(transcoder.Version(spec.V1)),
		), ctxDecorator
	}
	return bindings_nats.NewSender(t.Conn, t.Subject), []func(context.Context) context.Context{}
}
