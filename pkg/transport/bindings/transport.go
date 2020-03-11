package bindings

import (
	"context"
	"errors"
	"io"

	"github.com/cloudevents/sdk-go/pkg/binding"
	bindings "github.com/cloudevents/sdk-go/pkg/transport"

	"go.uber.org/zap"

	cecontext "github.com/cloudevents/sdk-go/pkg/context"
	"github.com/cloudevents/sdk-go/pkg/event"
	"github.com/cloudevents/sdk-go/pkg/transport"
)

// BindingTransport adapts the transport.Transport interface to binding.Sender and binding.Receiver.
// This adapter doesn't support the request/response model
type BindingTransport struct {
	Sender    bindings.Sender
	Requester bindings.Requester
	Receiver  bindings.Receiver
	Responder bindings.Responder
	// SenderContextDecorators can be used to decorate the context passed to the Sender.Send() method
	SenderContextDecorators []func(context.Context) context.Context
	handler                 transport.Delivery
}

var _ transport.Transport = (*BindingTransport)(nil) // Conforms to the interface

func NewRequestingTransport(r bindings.Requester, rx bindings.Receiver, senderContextDecorators []func(context.Context) context.Context) *BindingTransport {
	if senderContextDecorators == nil {
		senderContextDecorators = []func(ctx context.Context) context.Context{}
	}
	return &BindingTransport{Requester: r, Receiver: rx, SenderContextDecorators: senderContextDecorators}
}

func NewSendingTransport(s bindings.Sender, rx bindings.Receiver, senderContextDecorators []func(context.Context) context.Context) *BindingTransport {
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
	return t.Sender.Send(ctx, (*binding.EventMessage)(&e))
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
	msg, err := t.Requester.Request(ctx, (*binding.EventMessage)(&e))
	defer func() {
		if err := msg.Finish(err); err != nil {
			cecontext.LoggerFrom(ctx).Warnw("failed calling message.Finish", zap.Error(err))
		}
	}()
	if err == nil {
		if rs, err := binding.ToEvent(ctx, msg, nil); err != nil {
			cecontext.LoggerFrom(ctx).Warnw("failed calling ToEvent", zap.Error(err), zap.Any("resp", msg))
		} else {
			resp = rs
		}
	}
	return resp, err
}

func (t *BindingTransport) SetDelivery(r transport.Delivery) {
	t.handler = r
}

// Legacy Transport StartReceiver is really start Responder.
func (t *BindingTransport) StartReceiver(ctx context.Context) error {
	var msg binding.Message
	var err error
	var respFn transport.ResponseFn
	for {
		if t.Responder != nil {
			msg, respFn, err = t.Responder.Respond(ctx)
		} else if t.Receiver != nil {
			msg, err = t.Receiver.Receive(ctx)
		} else {
			return errors.New("responder and receiver not set")
		}

		if err == io.EOF { // Normal close
			return nil
		} else if err != nil {
			return err
		}
		if err := t.handle(ctx, msg, respFn); err != nil {
			return err
		}
	}
}

func (t *BindingTransport) handle(ctx context.Context, m binding.Message, respFn transport.ResponseFn) (err error) {
	defer func() {
		if err2 := m.Finish(err); err2 == nil {
			err = err2
		}
	}()

	if t.handler == nil {
		return nil
	}

	e, err := binding.ToEvent(ctx, m, nil)
	if err != nil {
		return err
	}
	eventResp := event.EventResponse{}
	if err := t.handler.Delivery(ctx, *e, &eventResp); err != nil {
		return err
	}

	if eventResp.Event != nil {
		// TODO: this does not give control over the http response code at the moment.
		if respFn != nil {
			return respFn(ctx, (*binding.EventMessage)(eventResp.Event))
		}
	}

	return nil
}

func (t *BindingTransport) HasTracePropagation() bool { return false } // TODO
