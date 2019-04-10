package nats

import (
	"context"
	"fmt"
	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	context2 "github.com/cloudevents/sdk-go/pkg/cloudevents/context"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport"
	"github.com/nats-io/go-nats"
	"go.uber.org/zap"
)

// Transport adheres to transport.Transport.
var _ transport.Transport = (*Transport)(nil)

// Transport acts as both a http client and a http handler.
type Transport struct {
	Encoding Encoding
	Conn     *nats.Conn
	Subject  string

	sub *nats.Subscription

	Receiver transport.Receiver

	codec transport.Codec
}

// New creates a new NATS transport.
func New(natsServer, subject string, opts ...Option) (*Transport, error) {
	conn, err := nats.Connect(natsServer)
	if err != nil {
		return nil, err
	}
	t := &Transport{
		Conn:    conn,
		Subject: subject,
	}
	if err := t.applyOptions(opts...); err != nil {
		return nil, err
	}

	return t, nil
}

func (t *Transport) applyOptions(opts ...Option) error {
	for _, fn := range opts {
		if err := fn(t); err != nil {
			return err
		}
	}
	return nil
}

func (t *Transport) loadCodec() bool {
	if t.codec == nil {
		switch t.Encoding {
		case Default:
			t.codec = &Codec{}
		case StructuredV02:
			t.codec = &CodecV02{Encoding: t.Encoding}
		case StructuredV03:
			t.codec = &CodecV03{Encoding: t.Encoding}
		default:
			return false
		}
	}
	return true
}

// Send implements Transport.Send
func (t *Transport) Send(ctx context.Context, event cloudevents.Event) (*cloudevents.Event, error) {
	if ok := t.loadCodec(); !ok {
		return nil, fmt.Errorf("unknown encoding set on transport: %d", t.Encoding)
	}

	msg, err := t.codec.Encode(event)
	if err != nil {
		return nil, err
	}

	if m, ok := msg.(*Message); ok {
		return nil, t.Conn.Publish(t.Subject, m.Body)
	}

	return nil, fmt.Errorf("failed to encode Event into a Message")
}

// SetReceiver implements Transport.SetReceiver
func (t *Transport) SetReceiver(r transport.Receiver) {
	t.Receiver = r
}

// StartReceiver implements Transport.StartReceiver
// NOTE: This is a blocking call.
func (t *Transport) StartReceiver(ctx context.Context) error {
	if t.Conn == nil {
		return fmt.Errorf("no active nats connection")
	}
	if t.sub != nil {
		return fmt.Errorf("already subscribed")
	}
	if ok := t.loadCodec(); !ok {
		return fmt.Errorf("unknown encoding set on transport: %d", t.Encoding)
	}

	// TODO: there could be more than one subscription. Might have to do a map
	// of subject to subscription.

	if t.Subject == "" {
		return fmt.Errorf("subject required for nats listen")
	}

	var err error
	// Simple Async Subscriber
	t.sub, err = t.Conn.Subscribe(t.Subject, func(m *nats.Msg) {
		logger := context2.LoggerFrom(ctx)
		msg := &Message{
			Body: m.Data,
		}
		event, err := t.codec.Decode(msg)
		if err != nil {
			logger.Errorw("failed to decode message", zap.Error(err)) // TODO: create an error channel to pass this up
			return
		}
		// TODO: I do not know enough about NATS to implement reply.
		// For now, NATS does not support reply.
		if err := t.Receiver.Receive(context.TODO(), *event, nil); err != nil {
			logger.Warnw("nats receiver return err", zap.Error(err))
		}
	})
	defer func() {
		if t.sub != nil {
			t.sub.Unsubscribe() // TODO: create an error channel to pass this up
			t.sub = nil
		}
	}()
	<-ctx.Done()
	return err
}
