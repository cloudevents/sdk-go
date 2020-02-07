package transcoder

import (
	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/binding/spec"
)

// TODO(slinkydeveloper) docs
func DeleteAttribute(attributeKind spec.Kind) binding.TransformerFactory {
	return deleteAttributeTranscoderFactory{attributeKind: attributeKind}
}

type deleteAttributeTranscoderFactory struct {
	attributeKind spec.Kind
}

func (a deleteAttributeTranscoderFactory) StructuredTransformer(binding.StructuredEncoder) binding.StructuredEncoder {
	return nil
}

func (a deleteAttributeTranscoderFactory) BinaryTransformer(encoder binding.BinaryEncoder) binding.BinaryEncoder {
	return &deleteAttributeTransformer{
		BinaryEncoder: encoder,
		attributeKind: a.attributeKind,
	}
}

func (a deleteAttributeTranscoderFactory) EventTransformer() binding.EventTransformer {
	return func(event *cloudevents.Event) error {
		v, err := spec.VS.Version(event.SpecVersion())
		if err != nil {
			return err
		}
		if v.AttributeFromKind(a.attributeKind).Get(event.Context) != nil {
			return v.AttributeFromKind(a.attributeKind).Delete(event.Context)
		}
		return nil
	}
}

type deleteAttributeTransformer struct {
	binding.BinaryEncoder
	attributeKind spec.Kind
}

func (b *deleteAttributeTransformer) SetAttribute(attribute spec.Attribute, value interface{}) error {
	if attribute.Kind() == b.attributeKind {
		return nil
	}
	return b.BinaryEncoder.SetAttribute(attribute, value)
}
