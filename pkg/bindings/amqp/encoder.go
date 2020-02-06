package amqp

import (
	"io"
	"io/ioutil"
	"net/url"

	"pack.ag/amqp"

	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/binding/format"
	"github.com/cloudevents/sdk-go/pkg/binding/spec"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/types"
)

//TODO (slinkydeveloper) this is the public access to http encoders, document it
func EncodeAMQPMessage(m binding.Message, amqpMessage *amqp.Message, forceStructured bool, forceBinary bool, transformerFactories binding.TransformerFactories) error {
	//TODO remove this stuff with lifecycle of encoders?
	amqpMessage.Properties = &amqp.MessageProperties{}
	amqpMessage.ApplicationProperties = make(map[string]interface{})

	var structuredEncoder binding.StructuredEncoder
	if !forceBinary {
		structuredEncoder = (*amqpMessageEncoder)(amqpMessage)
	}

	var binaryEncoder binding.BinaryEncoder
	if !forceStructured {
		binaryEncoder = (*amqpMessageEncoder)(amqpMessage)
	}

	var preferredEventEncoding binding.Encoding
	if forceStructured {
		preferredEventEncoding = binding.EncodingStructured
	} else {
		preferredEventEncoding = binding.EncodingBinary
	}

	_, err := binding.Encode(
		m,
		structuredEncoder,
		binaryEncoder,
		transformerFactories,
		preferredEventEncoding,
	)
	return err
}

type amqpMessageEncoder amqp.Message

func (b *amqpMessageEncoder) SetStructuredEvent(format format.Format, event io.Reader) error {
	val, err := ioutil.ReadAll(event)
	if err != nil {
		return err
	}
	b.Data = [][]byte{val}
	b.Properties.ContentType = format.MediaType()
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
	v, err := types.Validate(value)
	if err != nil {
		return err
	}
	switch t := v.(type) {
	case *url.URL: // Use string form of URLs.
		v = t.String()
	case int32: // Use AMQP long for Integer as per CE spec.
		v = int64(t)
	}
	b.ApplicationProperties[prefix+name] = v
	return nil
}

var _ binding.BinaryEncoder = (*amqpMessageEncoder)(nil)     // Test it conforms to the interface
var _ binding.StructuredEncoder = (*amqpMessageEncoder)(nil) // Test it conforms to the interface
