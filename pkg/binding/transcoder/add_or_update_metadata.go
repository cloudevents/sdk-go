package transcoder

import (
	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/binding/spec"
)

func SetAttribute(attribute spec.Kind, defaultValue interface{}, updater func(interface{}) (interface{}, error)) []binding.TransformerFactory {
	return []binding.TransformerFactory{
		UpdateAttribute(attribute, updater),
		AddAttribute(attribute, defaultValue),
	}
}

func SetExtension(name string, defaultValue interface{}, updater func(interface{}) (interface{}, error)) []binding.TransformerFactory {
	return []binding.TransformerFactory{
		UpdateExtension(name, updater),
		AddExtension(name, defaultValue),
	}
}
