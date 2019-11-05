package amqp

import (
	"errors"
	"reflect"

	"github.com/cloudevents/sdk-go/pkg/binding/format"
	"github.com/cloudevents/sdk-go/pkg/binding/spec"
	ce "github.com/cloudevents/sdk-go/pkg/cloudevents"
	"pack.ag/amqp"
)

const prefix = "cloudEvents:" // Name prefix for AMQP properties that hold CE attributes.

var (
	// Use the package path as AMQP error condition name
	condition = amqp.ErrorCondition(reflect.TypeOf(Message{}).PkgPath())
	specs     = spec.WithPrefix(prefix)
)

// Message implements binding.Message by wrapping an *amqp.Message.
type Message struct{ AMQP *amqp.Message }

func (m Message) Structured() (string, []byte) {
	if format.IsFormat(m.AMQP.Properties.ContentType) {
		return m.AMQP.Properties.ContentType, m.AMQP.GetData()
	}
	return "", nil
}

func (m Message) Event() (e ce.Event, err error) {
	if f, b := m.Structured(); f != "" {
		err = format.Unmarshal(f, b, &e)
		return e, err
	}
	if len(m.AMQP.ApplicationProperties) == 0 {
		return e, errors.New("AMQP CloudEvents message has no application properties")
	}
	version, err := specs.FindVersion(func(k string) string {
		s, _ := m.AMQP.ApplicationProperties[k].(string)
		return s
	})
	if err != nil {
		return e, err
	}
	c := version.NewContext()
	if m.AMQP.Properties != nil && m.AMQP.Properties.ContentType != "" {
		if err := c.SetDataContentType(m.AMQP.Properties.ContentType); err != nil {
			return e, err
		}
	}
	for k, v := range m.AMQP.ApplicationProperties {
		if err := version.SetAttribute(c, k, v); err != nil {
			return e, err
		}
	}
	data := m.AMQP.GetData()
	if len(data) == 0 { // No data
		return ce.Event{Context: c}, nil
	}
	return ce.Event{Data: data, DataEncoded: true, Context: c}, nil
}

func (m Message) Finish(err error) error {
	if err != nil {
		return m.AMQP.Reject(&amqp.Error{
			Condition:   condition,
			Description: err.Error(),
		})
	}
	return m.AMQP.Accept()
}
