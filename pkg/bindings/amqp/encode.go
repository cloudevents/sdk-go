package amqp

import (
	"pack.ag/amqp"

	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/binding/format"
	"github.com/cloudevents/sdk-go/pkg/binding/spec"
	ce "github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/types"
)

// NewStruct returns a new structured amqp.Message
func NewStruct(mediaType string, data []byte) *amqp.Message {
	return &amqp.Message{
		Data:       [][]byte{data},
		Properties: &amqp.MessageProperties{ContentType: mediaType},
	}
}

// NewBinary returns a new structured amqp.Message
func NewBinary(e ce.Event) (*amqp.Message, error) {
	data, err := e.DataBytes()
	if err != nil {
		return nil, err
	}
	version, err := specs.Version(e.SpecVersion())
	if err != nil {
		return nil, err
	}
	attrs := version.Attributes()
	ext := e.Extensions()
	var props map[string]interface{}
	if n := len(attrs) - 1 + len(ext); n > 0 { // Don't make a property map unless needed
		props = make(map[string]interface{}, n)
	}
	m := &amqp.Message{
		Data:                  [][]byte{data},
		Properties:            &amqp.MessageProperties{ContentType: e.DataContentType()},
		ApplicationProperties: props,
	}
	for _, a := range attrs { // Standard attributes
		if a.Kind() != spec.DataContentType { // Skip, encoded as Properties.ContentType
			if v := a.Get(e.Context); v != nil {
				m.ApplicationProperties[a.Name()] = v
			}
		}
	}
	for k, v := range ext { // Extension attributes
		v, err := types.Validate(v)
		if err != nil {
			return m, err
		}
		switch t := v.(type) {
		case types.URI: // Use string form of URLs.
			v = t.String()
		case types.URIRef:
			v = t.String()
		case types.URLRef:
			v = t.String()
		case types.Timestamp:
			v = t.Time
		case int32: // Use AMQP long for Integer as per CE spec.
			v = int64(t)
		}
		m.ApplicationProperties[prefix+k] = v
	}
	return m, nil
}

type BinaryEncoder struct{}

func (BinaryEncoder) Encode(e ce.Event) (binding.Message, error) {
	m, err := NewBinary(e)
	return Message{AMQP: m}, err
}

type StructEncoder struct{ Format format.Format }

func (enc StructEncoder) Encode(e ce.Event) (binding.Message, error) {
	b, err := enc.Format.Marshal(e)
	return Message{AMQP: NewStruct(enc.Format.MediaType(), b)}, err
}
