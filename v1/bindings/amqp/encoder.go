package amqp

import (
	"context"
	"io"
	"io/ioutil"

	"pack.ag/amqp"

	"github.com/cloudevents/sdk-go/v1/binding"
	"github.com/cloudevents/sdk-go/v1/binding/format"
	"github.com/cloudevents/sdk-go/v1/binding/spec"
	"github.com/cloudevents/sdk-go/v1/cloudevents/types"
)

// Fill the provided amqpMessage with the message m.
// Using context you can tweak the encoding processing (more details on binding.Translate documentation).
func EncodeAMQPMessage(ctx context.Context, m binding.Message, amqpMessage *amqp.Message, transformerFactories ...binding.TransformerFactory) error {
	structuredEncoder := (*amqpMessageEncoder)(amqpMessage)
	binaryEncoder := (*amqpMessageEncoder)(amqpMessage)

	_, err := binding.Encode(
		ctx,
		m,
		structuredEncoder,
		binaryEncoder,
		transformerFactories,
	)
	return err
}

type amqpMessageEncoder amqp.Message

func (b *amqpMessageEncoder) SetStructuredEvent(ctx context.Context, format format.Format, event io.Reader) error {
	val, err := ioutil.ReadAll(event)
	if err != nil {
		return err
	}
	b.Data = [][]byte{val}
	b.Properties = &amqp.MessageProperties{ContentType: format.MediaType()}
	return nil
}

func (b *amqpMessageEncoder) Start(ctx context.Context) error {
	b.Properties = &amqp.MessageProperties{}
	b.ApplicationProperties = make(map[string]interface{})
	return nil
}

func (b *amqpMessageEncoder) End() error {
	return nil
}

func (b *amqpMessageEncoder) SetData(reader io.Reader) error {
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}
	b.Data = [][]byte{data}
	return nil
}

func (b *amqpMessageEncoder) SetAttribute(attribute spec.Attribute, value interface{}) error {
	if attribute.Kind() == spec.DataContentType {
		s, err := types.Format(value)
		if err != nil {
			return err
		}
		b.Properties.ContentType = s
	} else {
		v, err := safeAMQPPropertiesUnwrap(value)
		if err != nil {
			return err
		}
		b.ApplicationProperties[prefix+attribute.Name()] = v
	}
	return nil
}

func (b *amqpMessageEncoder) SetExtension(name string, value interface{}) error {
	v, err := safeAMQPPropertiesUnwrap(value)
	if err != nil {
		return err
	}
	b.ApplicationProperties[prefix+name] = v
	return nil
}

var _ binding.BinaryEncoder = (*amqpMessageEncoder)(nil)     // Test it conforms to the interface
var _ binding.StructuredEncoder = (*amqpMessageEncoder)(nil) // Test it conforms to the interface
