package binding

import (
	"context"

	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport"
)

// Transport implements the transport.Transport interface using a
// Sender and Receiver.
type Transport struct {
	Sender   Sender
	Receiver Receiver
	handler  transport.Receiver
}

var _ transport.Transport = (*Transport)(nil) // Test it conforms to the interface

func NewTransport(s Sender, r Receiver) *Transport {
	return &Transport{Sender: s, Receiver: r}
}

func (t *Transport) Send(ctx context.Context, e cloudevents.Event) (context.Context, *cloudevents.Event, error) {
	return ctx, nil, t.Sender.Send(ctx, EventMessage(e))
}

func (t *Transport) SetReceiver(r transport.Receiver) { t.handler = r }

func (t *Transport) StartReceiver(ctx context.Context) error {
	for {
		if m, err := t.Receiver.Receive(ctx); err != nil {
			return err
		} else if e, err := m.Event(); err != nil {
			m.Finish(err)
			return err
		} else if err := t.handler.Receive(ctx, e, nil); err != nil {
			m.Finish(err)
			return err
		} else {
			m.Finish(nil)
		}
	}
}

func (t *Transport) SetConverter(transport.Converter) {
	// TODO(alanconway) Can we separate Converter from the base transport interface?
}

func (t *Transport) HasConverter() bool {
	return false
}
