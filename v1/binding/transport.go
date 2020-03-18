package binding

import (
	"context"
	"io"

	ce "github.com/cloudevents/sdk-go/v1/cloudevents"
	"github.com/cloudevents/sdk-go/v1/cloudevents/transport"
)

// BindingTransport implements transport.Transport using a Sender and Receiver.
type BindingTransport struct {
	Sender                  Sender
	Receiver                Receiver
	SenderContextDecorators []func(context.Context) context.Context
	handler                 transport.Receiver
}

var _ transport.Transport = (*BindingTransport)(nil) // Conforms to the interface

func NewTransportAdapter(s Sender, r Receiver, senderContextDecorators []func(context.Context) context.Context) *BindingTransport {
	if senderContextDecorators == nil {
		senderContextDecorators = []func(ctx context.Context) context.Context{}
	}
	return &BindingTransport{Sender: s, Receiver: r, SenderContextDecorators: senderContextDecorators}
}

func (t *BindingTransport) Send(ctx context.Context, e ce.Event) (context.Context, *ce.Event, error) {
	for _, f := range t.SenderContextDecorators {
		ctx = f(ctx)
	}
	return ctx, nil, t.Sender.Send(ctx, EventMessage(e))
}

func (t *BindingTransport) SetReceiver(r transport.Receiver) { t.handler = r }

func (t *BindingTransport) StartReceiver(ctx context.Context) error {
	for {
		m, err := t.Receiver.Receive(ctx)
		if err == io.EOF { // Normal close
			return nil
		} else if err != nil {
			return err
		}
		if err := t.handle(ctx, m); err != nil {
			return err
		}
	}
}

func (t *BindingTransport) handle(ctx context.Context, m Message) (err error) {
	defer func() {
		if err2 := m.Finish(err); err == nil {
			err = err2
		}
	}()
	e, _, err := ToEvent(ctx, m)
	if err != nil {
		return err
	}
	return t.handler.Receive(ctx, e, nil)
}

func (t *BindingTransport) SetConverter(transport.Converter) {}
func (t *BindingTransport) HasConverter() bool               { return false }
func (t *BindingTransport) HasTracePropagation() bool        { return false }
