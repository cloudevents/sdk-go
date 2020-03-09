package transformer

import (
	"time"

	"github.com/google/uuid"

	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/binding/spec"
	"github.com/cloudevents/sdk-go/pkg/event"
	"github.com/cloudevents/sdk-go/pkg/types"
)

var (
	// Sets the cloudevents id attribute (if missing) to a UUID.New()
	AddUUID binding.TransformerFactory = addUUID{}
	// Sets the cloudevents time attribute (if missing) to time.Now()
	AddTimeNow binding.TransformerFactory = addTimeNow{}
)

type addUUID struct{}

func (a addUUID) StructuredTransformer(binding.StructuredWriter) binding.StructuredWriter {
	return nil
}

func (a addUUID) BinaryTransformer(encoder binding.BinaryWriter) binding.BinaryWriter {
	return &addUUIDTransformer{
		BinaryWriter: encoder,
		found:        false,
	}
}

func (a addUUID) EventTransformer() binding.EventTransformer {
	return func(event *event.Event) error {
		if event.Context.GetID() == "" {
			return event.Context.SetID(uuid.New().String())
		}
		return nil
	}
}

type addUUIDTransformer struct {
	binding.BinaryWriter
	version spec.Version
	found   bool
}

func (b *addUUIDTransformer) SetAttribute(attribute spec.Attribute, value interface{}) error {
	if attribute.Kind() == spec.ID {
		b.found = true
	}
	b.version = attribute.Version()
	return b.BinaryWriter.SetAttribute(attribute, value)
}

func (b *addUUIDTransformer) End() error {
	if !b.found {
		err := b.BinaryWriter.SetAttribute(b.version.AttributeFromKind(spec.ID), uuid.New().String())
		if err != nil {
			return err
		}
	}
	return b.BinaryWriter.End()
}

type addTimeNow struct{}

func (a addTimeNow) StructuredTransformer(binding.StructuredWriter) binding.StructuredWriter {
	return nil
}

func (a addTimeNow) BinaryTransformer(encoder binding.BinaryWriter) binding.BinaryWriter {
	return &addTimeNowTransformer{
		BinaryWriter: encoder,
		found:        false,
	}
}

func (a addTimeNow) EventTransformer() binding.EventTransformer {
	return func(event *event.Event) error {
		if event.Context.GetTime().IsZero() {
			return event.Context.SetTime(time.Now())
		}
		return nil
	}
}

type addTimeNowTransformer struct {
	binding.BinaryWriter
	version spec.Version
	found   bool
}

func (b *addTimeNowTransformer) SetAttribute(attribute spec.Attribute, value interface{}) error {
	if attribute.Kind() == spec.Time {
		b.found = true
	}
	b.version = attribute.Version()
	return b.BinaryWriter.SetAttribute(attribute, value)
}

func (b *addTimeNowTransformer) End() error {
	if !b.found {
		err := b.BinaryWriter.SetAttribute(b.version.AttributeFromKind(spec.Time), types.Timestamp{Time: time.Now()})
		if err != nil {
			return err
		}
	}
	return b.BinaryWriter.End()
}
