package amqp

import (
	"io"
	"io/ioutil"
	"net/url"

	"pack.ag/amqp"

	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/binding/spec"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/types"
)

type binaryMessageBuilder struct {
	amqpMessage *amqp.Message
}

var _ binding.BinaryMessageBuilder = (*binaryMessageBuilder)(nil) // Test it conforms to the interface

func newBinaryMessageBuilder(amqpMessage *amqp.Message) *binaryMessageBuilder {
	amqpMessage.ApplicationProperties = make(map[string]interface{})
	return &binaryMessageBuilder{amqpMessage: amqpMessage}
}

func (b *binaryMessageBuilder) Data(reader io.Reader) error {
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}
	b.amqpMessage.Data = [][]byte{data}
	return nil
}

func (b *binaryMessageBuilder) Set(attribute spec.Attribute, value interface{}) error {
	if attribute.Kind() == spec.DataContentType {
		s, err := types.Format(value)
		if err != nil {
			return err
		}
		b.amqpMessage.Properties = &amqp.MessageProperties{ContentType: s}
	} else {
		v, err := safeAMQPPropertiesUnwrap(value)
		if err != nil {
			return err
		}
		b.amqpMessage.ApplicationProperties[prefix+attribute.Name()] = v
	}
	return nil
}

func (b *binaryMessageBuilder) SetExtension(name string, value interface{}) error {
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
	b.amqpMessage.ApplicationProperties[prefix+name] = v
	return nil
}
