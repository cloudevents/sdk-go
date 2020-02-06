package transcoder

import (
	"io"

	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/binding/spec"
	ce "github.com/cloudevents/sdk-go/pkg/cloudevents"
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
	return binaryVersionTransformer{version: v.version, delegate: encoder}
}

func (v versionTranscoderFactory) EventTransformer() binding.EventTransformer {
	return func(e *ce.Event) error {
		e.Context = v.version.Convert(e.Context)
		return nil
	}
}

type binaryVersionTransformer struct {
	delegate binding.BinaryEncoder
	version  spec.Version
}

func (b binaryVersionTransformer) SetData(data io.Reader) error {
	return b.delegate.SetData(data)
}

func (b binaryVersionTransformer) SetAttribute(attribute spec.Attribute, value interface{}) error {
	if attribute.Kind() == spec.SpecVersion {
		return b.delegate.SetAttribute(b.version.AttributeFromKind(spec.SpecVersion), b.version.String())
	}
	attributeInDifferentVersion := b.version.AttributeFromKind(attribute.Kind())
	return b.delegate.SetAttribute(attributeInDifferentVersion, value)
}

func (b binaryVersionTransformer) SetExtension(name string, value interface{}) error {
	return b.delegate.SetExtension(name, value)
}
