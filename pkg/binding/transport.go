package binding

import (
	"context"
	"errors"
	"io"

	"go.uber.org/zap"

	cecontext "github.com/cloudevents/sdk-go/pkg/cloudevents/context"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport"
	"github.com/cloudevents/sdk-go/pkg/event"
)

// BindingTransport adapts the transport.Transport interface to binding.Sender and binding.Receiver.
// This adapter doesn't support the request/response model
type BindingTransport struct {
	Sender    Sender
	Requester Requester
	Receiver  Receiver
	// SenderContextDecorators can be used to decorate the context passed to the Sender.Send() method
	SenderContextDecorators []func(context.Context) context.Context
	handler                 transport.Receiver
	converter               transport.Converter
}

var _ transport.Transport = (*BindingTransport)(nil) // Conforms to the interface

func NewRequestingTransport(r Requester, rx Receiver, senderContextDecorators []func(context.Context) context.Context) *BindingTransport {
	if senderContextDecorators == nil {
		senderContextDecorators = []func(ctx context.Context) context.Context{}
	}
	return &BindingTransport{Requester: r, Receiver: rx, SenderContextDecorators: senderContextDecorators}
}

func NewSendingTransport(s Sender, rx Receiver, senderContextDecorators []func(context.Context) context.Context) *BindingTransport {
	if senderContextDecorators == nil {
		senderContextDecorators = []func(ctx context.Context) context.Context{}
	}
	return &BindingTransport{Sender: s, Receiver: rx, SenderContextDecorators: senderContextDecorators}
}

func (t *BindingTransport) Send(ctx context.Context, e event.Event) error {
	if t.Sender == nil {
		return errors.New("sender not set")
	}
	for _, f := range t.SenderContextDecorators {
		ctx = f(ctx)
	}
	return t.Sender.Send(ctx, EventMessage(e))
}

func (t *BindingTransport) Request(ctx context.Context, e event.Event) (*event.Event, error) {
	if t.Requester == nil {
		return nil, errors.New("requester not set")
	}
	for _, f := range t.SenderContextDecorators {
		ctx = f(ctx)
	}

	// If provided a requester, use it to do request/response.
	var resp *event.Event
	msg, err := t.Requester.Request(ctx, EventMessage(e))
	defer func() {
		if err := msg.Finish(err); err != nil {
			cecontext.LoggerFrom(ctx).Warnw("failed calling message.Finish", zap.Error(err))
		}
	}()
	if err == nil {
		if rs, err := ToEvent(ctx, msg, nil); err != nil {
			cecontext.LoggerFrom(ctx).Warnw("failed calling ToEvent", zap.Error(err), zap.Any("resp", msg))
		} else {
			resp = &rs
		}
	}
	return resp, err
}

func (t *BindingTransport) SetReceiver(r transport.Receiver) {
	t.handler = r
}

func (t *BindingTransport) StartReceiver(ctx context.Context) error {
	for {
		msg, err := t.Receiver.Receive(ctx)
		if err == io.EOF { // Normal close
			return nil
		} else if err != nil {
			return err
		}
		if err := t.handle(ctx, msg); err != nil {
			return err
		}
	}
}

func (t *BindingTransport) handle(ctx context.Context, m Message) (err error) {
	defer func() {
		if err2 := m.Finish(err); err2 == nil {
			err = err2
		}
	}()

	if t.handler == nil {
		return
	}

	e, err := ToEvent(ctx, m, nil)
	if err != nil {
		return err
	}
	eventResp := event.EventResponse{}
	if err := t.handler.Receive(ctx, e, &eventResp); err != nil {
		return err
	}

	if eventResp.Event != nil {
		// TODO: this does not give control over the http response code at the moment.
		if rs, ok := m.(ResponseMessage); ok {
			rs.Response(ctx, EventMessage(*eventResp.Event))
		}
	}

	return nil
}

func (t *BindingTransport) SetConverter(c transport.Converter) {
	t.converter = c // TODO: use converter.
}

func (t *BindingTransport) HasConverter() bool {
	return t.converter != nil
}

func (t *BindingTransport) HasTracePropagation() bool { return false } // TODO
