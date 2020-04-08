package transformer

import (
	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/binding/spec"
)

// AddAttribute adds a cloudevents attribute (if missing) during the encoding process
func AddAttribute(attributeKind spec.Kind, value interface{}) binding.TransformerFunc {
	return func(reader binding.MessageMetadataReader, writer binding.MessageMetadataWriter) error {
		// Get the spec version
		attr, val := reader.GetAttribute(attributeKind)
		if attr == nil || val != nil {
			return nil
		}
		return writer.SetAttribute(attr.Version().AttributeFromKind(attributeKind), value)
	}
}

// AddExtension adds a cloudevents extension (if missing) during the encoding process
func AddExtension(name string, value interface{}) binding.TransformerFunc {
	return func(reader binding.MessageMetadataReader, writer binding.MessageMetadataWriter) error {
		return writer.SetExtension(name, value)
	}
}
