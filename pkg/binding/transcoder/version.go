package transcoder

import (
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

func (v versionTranscoderFactory) EventTransformer(encoder binding.EventEncoder) binding.EventEncoder {
	return eventVersionTransformer{version: v.version, delegate: encoder}
}

type binaryVersionTransformer struct {
	delegate binding.BinaryEncoder
	version  spec.Version
}

func (b binaryVersionTransformer) SetData(data binding.MessagePayloadReader) error {
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

type eventVersionTransformer struct {
	delegate binding.EventEncoder
	version  spec.Version
}

func (e eventVersionTransformer) SetEvent(event ce.Event) error {
	event.Context = e.version.Convert(event.Context)
	return e.delegate.SetEvent(event)
}
