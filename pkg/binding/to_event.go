package binding

import (
	"io"
	"io/ioutil"

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/cloudevents/sdk-go/pkg/binding/format"
	"github.com/cloudevents/sdk-go/pkg/binding/spec"
	ce "github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/types"
)

// Translates a Message with a valid Structured, Binary or Event representation to an Event
// Returns:
// * event, true, false, nil if message was structured and correctly translated to Event
// * event, false, true, nil if message was binary and correctly translated to Event
// * event, false, false, nil if message was event and correctly translated to Event
// * nil, true, false, err if message was structured but error happened during translation
// * nil, false, true, err if message was binary but error happened during translation
// * nil, false, false, err if message was event but error happened during translation
// * nil, false, false, err in other cases
func ToEvent(message Message, factories ...TranscoderFactory) (ce.Event, bool, bool, error) {
	e := cloudevents.NewEvent()
	fs := TranscoderFactories(factories)
	builder := &MessageToEventBuilder{event: &e}
	structuredBuilder := fs.StructuredMessageTranscoder(builder)

	if structuredBuilder != nil {
		if err := message.Structured(structuredBuilder); err == nil {
			return e, true, false, nil
		} else if err != ErrNotStructured {
			return e, true, false, err
		}
	}

	binaryBuilder := fs.BinaryMessageTranscoder(builder)
	if binaryBuilder != nil {
		if err := message.Binary(binaryBuilder); err == nil {
			return e, false, true, nil
		} else if err != ErrNotBinary {
			return e, false, true, err
		}
	}

	eventBuilder := fs.EventMessageTranscoder(builder)
	return e, false, false, message.Event(eventBuilder)
}

type MessageToEventBuilder struct {
	event *ce.Event
}

func (b *MessageToEventBuilder) Encode(e ce.Event) error {
	b.event.Data = e.Data
	b.event.Context = e.Context.Clone()
	b.event.DataBinary = e.DataBinary
	b.event.DataEncoded = e.DataEncoded
	return nil
}

func (b *MessageToEventBuilder) Event(format format.Format, event io.Reader) error {
	//TODO(slinkydeveloper) can we do pooling for this allocation?
	val, err := ioutil.ReadAll(event)
	if err != nil {
		return err
	}
	return format.Unmarshal(val, b.event)
}

func (b *MessageToEventBuilder) Data(data io.Reader) error {
	//TODO(slinkydeveloper) can we do pooling for this allocation?
	val, err := ioutil.ReadAll(data)
	if err != nil {
		return err
	}
	if len(val) != 0 {
		return b.event.SetData(val)
	}
	return nil
}

func (b *MessageToEventBuilder) Set(attribute spec.Attribute, value interface{}) error {
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

func (b *MessageToEventBuilder) SetExtension(name string, value interface{}) error {
	value, err := types.Validate(value)
	if err != nil {
		return err
	}
	b.event.SetExtension(name, value)
	return nil
}
