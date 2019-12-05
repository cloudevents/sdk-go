package transcoder

import (
	"io"

	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/binding/spec"
	ce "github.com/cloudevents/sdk-go/pkg/cloudevents"
)

type versionTranscoderFactory struct {
	version spec.Version
}

func (v versionTranscoderFactory) StructuredMessageTranscoder(builder binding.StructuredMessageBuilder) binding.StructuredMessageBuilder {
	return nil // Not supported!
}

func (v versionTranscoderFactory) BinaryMessageTranscoder(builder binding.BinaryMessageBuilder) binding.BinaryMessageBuilder {
	return binaryVersionTranscoder{version: v.version, delegate: builder}
}

func (v versionTranscoderFactory) EventMessageTranscoder(builder binding.EventMessageBuilder) binding.EventMessageBuilder {
	return eventVersionTranscoder{version: v.version, delegate: builder}
}

func Version(version spec.Version) binding.TranscoderFactory {
	return versionTranscoderFactory{version: version}
}

type binaryVersionTranscoder struct {
	delegate binding.BinaryMessageBuilder
	version  spec.Version
}

func (b binaryVersionTranscoder) Data(data io.Reader) error {
	return b.delegate.Data(data)
}

func (b binaryVersionTranscoder) Set(attribute spec.Attribute, value interface{}) error {
	if attribute.Kind() == spec.SpecVersion {
		return b.delegate.Set(b.version.AttributeFromKind(spec.SpecVersion), b.version.String())
	}
	attributeInDifferentVersion := b.version.AttributeFromKind(attribute.Kind())
	return b.delegate.Set(attributeInDifferentVersion, value)
}

func (b binaryVersionTranscoder) SetExtension(name string, value interface{}) error {
	return b.delegate.SetExtension(name, value)
}

type eventVersionTranscoder struct {
	delegate binding.EventMessageBuilder
	version  spec.Version
}

func (e eventVersionTranscoder) Encode(event ce.Event) error {
	event.Context = e.version.Convert(event.Context)
	return e.delegate.Encode(event)
}
