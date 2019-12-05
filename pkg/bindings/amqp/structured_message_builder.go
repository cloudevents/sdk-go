package amqp

import (
	"io"
	"io/ioutil"

	"pack.ag/amqp"

	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/binding/format"
)

type structuredMessageBuilder struct {
	amqpMessage *amqp.Message
}

var _ binding.StructuredMessageBuilder = (*structuredMessageBuilder)(nil) // Test it conforms to the interface

func (b *structuredMessageBuilder) Event(format format.Format, event io.Reader) error {
	val, err := ioutil.ReadAll(event)
	if err != nil {
		return err
	}
	b.amqpMessage.Data = [][]byte{val}
	b.amqpMessage.Properties = &amqp.MessageProperties{ContentType: format.MediaType()}
	return nil
}
