package amqp

import (
	"pack.ag/amqp"

	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/binding/format"
	"github.com/cloudevents/sdk-go/pkg/binding/spec"
	ce "github.com/cloudevents/sdk-go/pkg/cloudevents"
)

type eventToBinaryMessageBuilder struct {
	amqpMessage *amqp.Message
}

var _ binding.EventMessageBuilder = (*eventToBinaryMessageBuilder)(nil) // Test it conforms to the interface

func (b *eventToBinaryMessageBuilder) Encode(e ce.Event) error {
	version, err := specs.Version(e.SpecVersion())
	if err != nil {
		return err
	}
	attrs := version.Attributes()
	ext := e.Extensions()
	props := make(map[string]interface{}, len(attrs)-1+len(ext))

	b.amqpMessage.Properties = &amqp.MessageProperties{ContentType: e.DataContentType()}
	b.amqpMessage.ApplicationProperties = props

	for _, a := range attrs { // Standard attributes
		if a.Kind() != spec.DataContentType { // Skip, encoded as Properties.ContentType
			if v := a.Get(e.Context); v != nil {
				props[a.Name()] = v
			}
		}
	}

	for k, v := range ext { // Extension attributes
		v, err := safeAMQPPropertiesUnwrap(v)
		if err != nil {
			return err
		}
		props[prefix+k] = v
	}

	data, err := e.DataBytes()
	if err != nil {
		return err
	}
	b.amqpMessage.Data = [][]byte{data}
	return nil
}

type eventToStructuredMessageBuilder struct {
	format      format.Format
	amqpMessage *amqp.Message
}

var _ binding.EventMessageBuilder = (*eventToStructuredMessageBuilder)(nil) // Test it conforms to the interface

func (b *eventToStructuredMessageBuilder) Encode(event ce.Event) error {
	data, err := b.format.Marshal(event)
	if err != nil {
		return err
	}
	b.amqpMessage.Data = [][]byte{data}
	b.amqpMessage.Properties = &amqp.MessageProperties{ContentType: b.format.MediaType()}

	return nil
}
