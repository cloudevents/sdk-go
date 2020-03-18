package binding

import (
	"bytes"
	"context"
	"errors"
	"io"

	cloudevents "github.com/cloudevents/sdk-go/v1"
	"github.com/cloudevents/sdk-go/v1/binding/format"
	"github.com/cloudevents/sdk-go/v1/binding/spec"
	ce "github.com/cloudevents/sdk-go/v1/cloudevents"
	"github.com/cloudevents/sdk-go/v1/cloudevents/types"
)

var ErrCannotConvertToEvent = errors.New("cannot convert message to event")

// Translates a Message with a valid Structured or Binary representation to an Event
// The TransformerFactories **aren't invoked** during the transformation to event,
// but after the event instance is generated
func ToEvent(ctx context.Context, message Message, transformers ...TransformerFactory) (e ce.Event, encoding Encoding, err error) {
	e = cloudevents.NewEvent()

	messageEncoding := message.Encoding()
	if messageEncoding == EncodingEvent {
		for m := message; m != nil; {
			if em, ok := m.(EventMessage); ok {
				e = ce.Event(em)
				encoding = EncodingEvent
				err = TransformerFactories(transformers).EventTransformer()(&e)
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
	err = TransformerFactories(transformers).EventTransformer()(&e)
	return
}

type messageToEventBuilder struct {
	event *ce.Event
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
		case cloudevents.VersionV01:
			b.event.Context = b.event.Context.AsV01()
		case cloudevents.VersionV02:
			b.event.Context = b.event.Context.AsV02()
		case cloudevents.VersionV03:
			b.event.Context = b.event.Context.AsV03()
		case cloudevents.VersionV1:
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
