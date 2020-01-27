package event

import (
	"context"
	"io"

	"github.com/cloudevents/sdk-go/pkg/binding"
	ce "github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport"
)

// BindingTransport implements transport.Transport using a Sender and Receiver.
type BindingTransport struct {
	Sender   binding.Sender
	Receiver binding.Receiver
	handler  transport.Receiver
}

var _ transport.Transport = (*BindingTransport)(nil) // Conforms to the interface

func NewTransportAdapter(s binding.Sender, r binding.Receiver) *BindingTransport {
	return &BindingTransport{Sender: s, Receiver: r}
}

func (t *BindingTransport) Send(ctx context.Context, e ce.Event) (context.Context, *ce.Event, error) {
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

func (t *BindingTransport) handle(ctx context.Context, m binding.Message) (err error) {
	defer func() {
		if err2 := m.Finish(err); err == nil {
			err = err2
		}
	}()
	e, _, _, err := ToEvent(m)
	if err != nil {
		return err
	}
	return t.handler.Receive(ctx, e, nil)
}

func (t *BindingTransport) SetConverter(transport.Converter) {}
func (t *BindingTransport) HasConverter() bool               { return false }
