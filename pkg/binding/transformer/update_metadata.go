package transformer

import (
	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/binding/spec"
	"github.com/cloudevents/sdk-go/pkg/event"
)

// Update cloudevents attribute (if present) using the provided function
func UpdateAttribute(attributeKind spec.Kind, updater func(interface{}) (interface{}, error)) binding.TransformerFactory {
	return updateAttributeTranscoderFactory{attributeKind: attributeKind, updater: updater}
}

// Update cloudevents extension (if present) using the provided function
func UpdateExtension(name string, updater func(interface{}) (interface{}, error)) binding.TransformerFactory {
	return updateExtensionTranscoderFactory{name: name, updater: updater}
}

type updateAttributeTranscoderFactory struct {
	attributeKind spec.Kind
	updater       func(interface{}) (interface{}, error)
}

func (a updateAttributeTranscoderFactory) StructuredTransformer(binding.StructuredEncoder) binding.StructuredEncoder {
	return nil
}

func (a updateAttributeTranscoderFactory) BinaryTransformer(encoder binding.BinaryEncoder) binding.BinaryEncoder {
	return &updateAttributeTransformer{
		BinaryEncoder: encoder,
		attributeKind: a.attributeKind,
		updater:       a.updater,
	}
}

func (a updateAttributeTranscoderFactory) EventTransformer() binding.EventTransformer {
	return func(event *event.Event) error {
		v, err := spec.VS.Version(event.SpecVersion())
		if err != nil {
			return err
		}
		if val := v.AttributeFromKind(a.attributeKind).Get(event.Context); val != nil {
			newVal, err := a.updater(val)
			if err != nil {
				return err
			}
			if newVal == nil {
				return v.AttributeFromKind(a.attributeKind).Delete(event.Context)
			} else {
				return v.AttributeFromKind(a.attributeKind).Set(event.Context, newVal)
			}
		}
		return nil
	}
}

type updateExtensionTranscoderFactory struct {
	name    string
	updater func(interface{}) (interface{}, error)
}

func (a updateExtensionTranscoderFactory) StructuredTransformer(binding.StructuredEncoder) binding.StructuredEncoder {
	return nil
}

func (a updateExtensionTranscoderFactory) BinaryTransformer(encoder binding.BinaryEncoder) binding.BinaryEncoder {
	return &updateExtensionTransformer{
		BinaryEncoder: encoder,
		name:          a.name,
		updater:       a.updater,
	}
}

func (a updateExtensionTranscoderFactory) EventTransformer() binding.EventTransformer {
	return func(event *event.Event) error {
		if val, ok := event.Extensions()[a.name]; ok {
			newVal, err := a.updater(val)
			if err != nil {
				return err
			}
			return event.Context.SetExtension(a.name, newVal)
		}
		return nil
	}
}

type updateAttributeTransformer struct {
	binding.BinaryEncoder
	attributeKind spec.Kind
	updater       func(interface{}) (interface{}, error)
}

func (b *updateAttributeTransformer) SetAttribute(attribute spec.Attribute, value interface{}) error {
	if attribute.Kind() == b.attributeKind {
		newVal, err := b.updater(value)
		if err != nil {
			return err
		}
		if newVal != nil {
			return b.BinaryEncoder.SetAttribute(attribute, newVal)
		}
		return nil
	}
	return b.BinaryEncoder.SetAttribute(attribute, value)
}

type updateExtensionTransformer struct {
	binding.BinaryEncoder
	name    string
	updater func(interface{}) (interface{}, error)
}

func (b *updateExtensionTransformer) SetExtension(name string, value interface{}) error {
	if name == b.name {
		newVal, err := b.updater(value)
		if err != nil {
			return err
		}
		if newVal != nil {
			return b.BinaryEncoder.SetExtension(name, newVal)
		}
		return nil
	}
	return b.BinaryEncoder.SetExtension(name, value)
}
