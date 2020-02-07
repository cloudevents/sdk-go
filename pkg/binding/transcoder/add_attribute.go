package transcoder

import (
	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/binding/spec"
)

// TODO(slinkydeveloper) docs
// Attribute is added only if missing!
func AddAttribute(attributeKind spec.Kind, value interface{}) binding.TransformerFactory {
	return setAttributeTranscoderFactory{attributeKind: attributeKind, value: value}
}

type setAttributeTranscoderFactory struct {
	attributeKind spec.Kind
	value         interface{}
}

func (a setAttributeTranscoderFactory) StructuredTransformer(binding.StructuredEncoder) binding.StructuredEncoder {
	return nil
}

func (a setAttributeTranscoderFactory) BinaryTransformer(encoder binding.BinaryEncoder) binding.BinaryEncoder {
	return &setAttributeTransformer{
		BinaryEncoder: encoder,
		attributeKind: a.attributeKind,
		value:         a.value,
		found:         false,
	}
}

func (a setAttributeTranscoderFactory) EventTransformer() binding.EventTransformer {
	return func(event *cloudevents.Event) error {
		v, err := spec.VS.Version(event.SpecVersion())
		if err != nil {
			return err
		}
		if v.AttributeFromKind(a.attributeKind).Get(event.Context) == nil {
			return v.AttributeFromKind(a.attributeKind).Set(event.Context, a.value)
		}
		return nil
	}
}

type setAttributeTransformer struct {
	binding.BinaryEncoder
	attributeKind spec.Kind
	value         interface{}
	version       spec.Version
	found         bool
}

func (b *setAttributeTransformer) SetAttribute(attribute spec.Attribute, value interface{}) error {
	if attribute.Kind() == b.attributeKind {
		b.found = true
	}
	b.version = attribute.Version()
	return b.BinaryEncoder.SetAttribute(attribute, value)
}

func (b *setAttributeTransformer) End() error {
	if !b.found {
		return b.BinaryEncoder.SetAttribute(b.version.AttributeFromKind(b.attributeKind), b.value)
	}
	return nil
}
