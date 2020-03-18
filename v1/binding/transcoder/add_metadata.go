package transcoder

import (
	cloudevents "github.com/cloudevents/sdk-go/v1"
	"github.com/cloudevents/sdk-go/v1/binding"
	"github.com/cloudevents/sdk-go/v1/binding/spec"
)

// Add cloudevents attribute (if missing) during the encoding process
func AddAttribute(attributeKind spec.Kind, value interface{}) binding.TransformerFactory {
	return setAttributeTranscoderFactory{attributeKind: attributeKind, value: value}
}

// Add cloudevents extension (if missing) during the encoding process
func AddExtension(name string, value interface{}) binding.TransformerFactory {
	return setExtensionTranscoderFactory{name: name, value: value}
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

type setExtensionTranscoderFactory struct {
	name  string
	value interface{}
}

func (a setExtensionTranscoderFactory) StructuredTransformer(binding.StructuredEncoder) binding.StructuredEncoder {
	return nil
}

func (a setExtensionTranscoderFactory) BinaryTransformer(encoder binding.BinaryEncoder) binding.BinaryEncoder {
	return &setExtensionTransformer{
		BinaryEncoder: encoder,
		name:          a.name,
		value:         a.value,
		found:         false,
	}
}

func (a setExtensionTranscoderFactory) EventTransformer() binding.EventTransformer {
	return func(event *cloudevents.Event) error {
		if _, ok := event.Extensions()[a.name]; !ok {
			return event.Context.SetExtension(a.name, a.value)
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

type setExtensionTransformer struct {
	binding.BinaryEncoder
	name  string
	value interface{}
	found bool
}

func (b *setExtensionTransformer) SetExtension(name string, value interface{}) error {
	if name == b.name {
		b.found = true
	}
	return b.BinaryEncoder.SetExtension(name, value)
}

func (b *setExtensionTransformer) End() error {
	if !b.found {
		return b.BinaryEncoder.SetExtension(b.name, b.value)
	}
	return nil
}
