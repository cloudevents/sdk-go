package transcoder

import (
	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/binding/spec"
)

func AddOrUpdateAttribute(attribute spec.Kind, defaultValue interface{}, updater func(interface{}) (interface{}, error)) []binding.TransformerFactory {
	return []binding.TransformerFactory{
		UpdateAttribute(attribute, updater),
		AddAttribute(attribute, defaultValue),
	}
}

func AddOrUpdateExtension(name string, defaultValue interface{}, updater func(interface{}) (interface{}, error)) []binding.TransformerFactory {
	return []binding.TransformerFactory{
		UpdateExtension(name, updater),
		AddExtension(name, defaultValue),
	}
}
