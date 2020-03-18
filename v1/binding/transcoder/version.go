package transcoder

import (
	"github.com/cloudevents/sdk-go/v1/binding"
	"github.com/cloudevents/sdk-go/v1/binding/spec"
	ce "github.com/cloudevents/sdk-go/v1/cloudevents"
)

// Returns a TransformerFactory that converts the event context version to the specified one.
func Version(version spec.Version) binding.TransformerFactory {
	return versionTranscoderFactory{version: version}
}

type versionTranscoderFactory struct {
	version spec.Version
}

func (v versionTranscoderFactory) StructuredTransformer(binding.StructuredEncoder) binding.StructuredEncoder {
	return nil // Not supported, must fallback to EventTransformer!
}

func (v versionTranscoderFactory) BinaryTransformer(encoder binding.BinaryEncoder) binding.BinaryEncoder {
	return binaryVersionTransformer{BinaryEncoder: encoder, version: v.version}
}

func (v versionTranscoderFactory) EventTransformer() binding.EventTransformer {
	return func(e *ce.Event) error {
		e.Context = v.version.Convert(e.Context)
		return nil
	}
}

type binaryVersionTransformer struct {
	binding.BinaryEncoder
	version spec.Version
}

func (b binaryVersionTransformer) SetAttribute(attribute spec.Attribute, value interface{}) error {
	if attribute.Kind() == spec.SpecVersion {
		return b.BinaryEncoder.SetAttribute(b.version.AttributeFromKind(spec.SpecVersion), b.version.String())
	}
	attributeInDifferentVersion := b.version.AttributeFromKind(attribute.Kind())
	return b.BinaryEncoder.SetAttribute(attributeInDifferentVersion, value)
}
