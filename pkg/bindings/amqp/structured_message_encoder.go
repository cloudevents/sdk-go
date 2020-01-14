package amqp

import (
	"pack.ag/amqp"

	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/binding/format"
)

type structuredMessageEncoder struct {
	amqpMessage *amqp.Message
}

var _ binding.StructuredEncoder = (*structuredMessageEncoder)(nil) // Test it conforms to the interface

func (b *structuredMessageEncoder) SetStructuredEvent(format format.Format, event binding.MessagePayloadReader) error {
	b.amqpMessage.Data = [][]byte{event.Bytes()}
	b.amqpMessage.Properties = &amqp.MessageProperties{ContentType: format.MediaType()}
	return nil
}
