package test

import (
	"bytes"

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/binding/format"
)

// MockStructuredMessage implements a structured-mode message as a simple struct.
type MockStructuredMessage struct {
	Format format.Format
	Bytes  []byte
}

func (s *MockStructuredMessage) GetParent() binding.Message {
	return nil
}

func NewMockStructuredMessage(e cloudevents.Event) *MockStructuredMessage {
	testEventSerialized, err := format.JSON.Marshal(e)
	if err != nil {
		panic(err)
	}
	return &MockStructuredMessage{
		Bytes:  testEventSerialized,
		Format: format.JSON,
	}
}

func (s *MockStructuredMessage) Structured(b binding.StructuredEncoder) error {
	return b.SetStructuredEvent(s.Format, bytes.NewReader(s.Bytes))
}

func (s *MockStructuredMessage) Binary(binding.BinaryEncoder) error {
	return binding.ErrNotBinary
}

func (bm *MockStructuredMessage) Encoding() binding.Encoding {
	return binding.EncodingStructured
}

func (s *MockStructuredMessage) Finish(error) error { return nil }

var _ binding.Message = (*MockStructuredMessage)(nil) // Test it conforms to the interface
