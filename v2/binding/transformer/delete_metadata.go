package transformer

import (
	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/binding/spec"
)

// DeleteAttribute deletes a cloudevents attribute during the encoding process
func DeleteAttribute(attributeKind spec.Kind) binding.TransformerFunc {
	return func(reader binding.MessageMetadataReader, writer binding.MessageMetadataWriter) error {
		if attr, v := reader.GetAttribute(attributeKind); v != nil {
			return writer.SetAttribute(attr, nil)
		}
		return nil
	}
}

// DeleteExtension deletes a cloudevents extension during the encoding process
func DeleteExtension(name string) binding.TransformerFunc {
	return func(reader binding.MessageMetadataReader, writer binding.MessageMetadataWriter) error {
		return writer.SetExtension(name, nil)
	}
}
