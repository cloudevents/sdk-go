package binding

import (
	"context"
	"io"

	ce "github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport"
)

// Transport implements transport.Transport using a Sender and Receiver.
type Transport struct {
	Sender   Sender
	Receiver Receiver
	handler  transport.Receiver
}

var _ transport.Transport = (*Transport)(nil) // Conforms to the interface

func NewTransport(s Sender, r Receiver) *Transport {
	return &Transport{Sender: s, Receiver: r}
}

func (t *Transport) Send(ctx context.Context, e ce.Event) (context.Context, *ce.Event, error) {
	return ctx, nil, t.Sender.Send(ctx, EventMessage(e))
}

func (t *Transport) SetReceiver(r transport.Receiver) { t.handler = r }

func (t *Transport) StartReceiver(ctx context.Context) error {
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

func (t *Transport) handle(ctx context.Context, m Message) (err error) {
	defer func() {
		if err2 := m.Finish(err); err == nil {
			err = err2
		}
	}()
	e, err := m.Event()
	if err != nil {
		return err
	}
	return t.handler.Receive(ctx, e, nil)
}

func (t *Transport) SetConverter(transport.Converter) {}
func (t *Transport) HasConverter() bool               { return false }
