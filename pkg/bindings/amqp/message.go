package amqp

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/cloudevents/sdk-go/pkg/binding/format"
	"github.com/cloudevents/sdk-go/pkg/binding/spec"
	ce "github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/types"
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
	version, err := findVersion(m)
	if err != nil {
		return ce.Event{}, err
	}
	c := version.NewContext()
	if m.AMQP.Properties != nil {
		if err := c.SetDataContentType(m.AMQP.Properties.ContentType); err != nil {
			return ce.Event{}, err
		}
	}
	for k, v := range m.AMQP.ApplicationProperties {
		if a := version.Attribute(k); a != nil { // A standard CE attribute
			if err := a.Set(c, v); err != nil {
				return ce.Event{}, err
			}
		} else if strings.HasPrefix(k, prefix) { // Extension attribute
			k = strings.TrimPrefix(k, prefix)
			// Ignore ill-formed attributes.
			if v, err := types.Validate(v); err == nil { // CE attribute value conversions.
				_ = c.SetExtension(k, v)
			}
		}
	}
	data := m.AMQP.GetData()
	if len(data) == 0 { // No data
		return ce.Event{Context: c}, nil
	}
	return ce.Event{Data: data, DataEncoded: true, Context: c}, nil
}

func findVersion(m Message) (spec.Version, error) {
	for _, sv := range specs.SpecVersionNames() {
		if v, ok := m.AMQP.ApplicationProperties[sv]; ok {
			if s, ok := v.(string); !ok {
				return nil, fmt.Errorf("%v property: want string, got %T", sv, v)
			} else if version, _ := specs.Version(s); version == nil {
				return nil, fmt.Errorf("not a valid version: %#v", s)
			} else {
				return version, nil
			}
		}
	}
	return nil, fmt.Errorf("no version property found")
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
