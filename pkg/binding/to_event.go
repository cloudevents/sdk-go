package binding

import (
	"bytes"
	"context"
	"errors"
	"io"

	"github.com/cloudevents/sdk-go/pkg/binding/format"
	"github.com/cloudevents/sdk-go/pkg/binding/spec"
	"github.com/cloudevents/sdk-go/pkg/event"
	"github.com/cloudevents/sdk-go/pkg/types"
)

// Generic error when a conversion of a Message to an Event fails
var ErrCannotConvertToEvent = errors.New("cannot convert message to event")

// Translates a Message with a valid Structured or Binary representation to an Event
// The transformers aren't invoked during the transformation to event,
// but after the event instance is generated.
// This function returns the Event generated from the Message and the original encoding of the message or
// an error that points the conversion error
func ToEvent(ctx context.Context, message Message, transformers TransformerFactories) (e event.Event, encoding Encoding, err error) {
	e = event.New()

	messageEncoding := message.Encoding()
	if messageEncoding == EncodingEvent {
		for m := message; m != nil; {
			if em, ok := m.(EventMessage); ok {
				e = event.Event(em)
				encoding = EncodingEvent
				err = transformers.EventTransformer()(&e)
				return
			}
			if mw, ok := m.(MessageWrapper); ok {
				m = mw.GetWrappedMessage()
			} else {
				break
			}
		}
		err = ErrCannotConvertToEvent
		return
	}

	encoder := &messageToEventBuilder{event: &e}
	encoding, err = RunDirectEncoding(
		context.TODO(),
		message,
		encoder,
		encoder,
		[]TransformerFactory{},
	)
	if err != nil {
		return e, encoding, err
	}
	err = transformers.EventTransformer()(&e)
	return
}

type messageToEventBuilder struct {
	event *event.Event
}

var _ StructuredEncoder = (*messageToEventBuilder)(nil)
var _ BinaryEncoder = (*messageToEventBuilder)(nil)

func (b *messageToEventBuilder) SetStructuredEvent(ctx context.Context, format format.Format, event io.Reader) error {
	var buf bytes.Buffer
	_, err := io.Copy(&buf, event)
	if err != nil {
		return err
	}
	return format.Unmarshal(buf.Bytes(), b.event)
}

func (b *messageToEventBuilder) Start(ctx context.Context) error {
	return nil
}

func (b *messageToEventBuilder) End() error {
	return nil
}

func (b *messageToEventBuilder) SetData(data io.Reader) error {
	var buf bytes.Buffer
	w, err := io.Copy(&buf, data)
	if err != nil {
		return err
	}
	if w != 0 {
		return b.event.SetData(buf.Bytes())
	}
	return nil
}

func (b *messageToEventBuilder) SetAttribute(attribute spec.Attribute, value interface{}) error {
	// If spec version we need to change to right context struct
	if attribute.Kind() == spec.SpecVersion {
		str, err := types.ToString(value)
		if err != nil {
			return err
		}
		switch str {
		case event.CloudEventsVersionV03:
			b.event.Context = b.event.Context.AsV03()
		case event.CloudEventsVersionV1:
			b.event.Context = b.event.Context.AsV1()
		}
		return nil
	}
	return attribute.Set(b.event.Context, value)
}

func (b *messageToEventBuilder) SetExtension(name string, value interface{}) error {
	value, err := types.Validate(value)
	if err != nil {
		return err
	}
	b.event.SetExtension(name, value)
	return nil
}
