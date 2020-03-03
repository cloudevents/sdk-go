package test

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"

	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/binding/spec"
	"github.com/cloudevents/sdk-go/pkg/event"
)

// MockBinaryMessage implements a binary-mode message as a simple struct.
type MockBinaryMessage struct {
	Metadata   map[spec.Attribute]interface{}
	Extensions map[string]interface{}
	Body       []byte
}

func (bm *MockBinaryMessage) Start(ctx context.Context) error {
	bm.Metadata = make(map[spec.Attribute]interface{})
	bm.Extensions = make(map[string]interface{})
	return nil
}

func (bm *MockBinaryMessage) SetAttribute(attribute spec.Attribute, value interface{}) error {
	bm.Metadata[attribute] = value
	return nil
}

func (bm *MockBinaryMessage) SetExtension(name string, value interface{}) error {
	bm.Extensions[name] = value
	return nil
}

func (bm *MockBinaryMessage) SetData(data io.Reader) (err error) {
	bm.Body, err = ioutil.ReadAll(data)
	return err
}

func (bm *MockBinaryMessage) End() error {
	return nil
}

var versions = spec.New()

func NewMockBinaryMessage(e event.Event) binding.Message {
	version, err := versions.Version(e.SpecVersion())
	if err != nil {
		panic(err)
	}

	m := MockBinaryMessage{
		Metadata:   make(map[spec.Attribute]interface{}),
		Extensions: make(map[string]interface{}),
	}

	for _, attribute := range version.Attributes() {
		val := attribute.Get(e.Context)
		if val != nil {
			m.Metadata[attribute] = val
		}
	}

	for k, v := range e.Extensions() {
		m.Extensions[k] = v
	}

	m.Body, err = e.DataBytes()
	if err != nil {
		panic(err)
	}

	return &m
}

func (bm *MockBinaryMessage) GetParent() binding.Message {
	return nil
}

func (bm *MockBinaryMessage) Structured(context.Context, binding.StructuredEncoder) error {
	return binding.ErrNotStructured
}

func (bm *MockBinaryMessage) Binary(ctx context.Context, b binding.BinaryEncoder) error {
	err := b.Start(ctx)
	if err != nil {
		return err
	}
	for k, v := range bm.Metadata {
		err = b.SetAttribute(k, v)
		if err != nil {
			return err
		}
	}
	for k, v := range bm.Extensions {
		err = b.SetExtension(k, v)
		if err != nil {
			return err
		}
	}
	if len(bm.Body) != 0 {
		err = b.SetData(bytes.NewReader(bm.Body))
		if err != nil {
			return err
		}
	}
	return b.End()
}

func (bm *MockBinaryMessage) Encoding() binding.Encoding {
	return binding.EncodingBinary
}

func (bm *MockBinaryMessage) Finish(error) error { return nil }

var _ binding.Message = (*MockBinaryMessage)(nil) // Test it conforms to the interface
var _ binding.BinaryEncoder = (*MockBinaryMessage)(nil)
