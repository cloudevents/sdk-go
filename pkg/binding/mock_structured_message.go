package binding

import (
	"bytes"
	"io"

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/cloudevents/sdk-go/pkg/binding/format"
)

// MockStructuredMessage implements a structured-mode message as a simple struct.
type MockStructuredMessage struct {
	format format.Format
	body   []byte
}

func NewMockStructuredMessage(e cloudevents.Event) *MockStructuredMessage {
	testEventSerialized, err := format.JSON.Marshal(e)
	if err != nil {
		panic(err)
	}
	return &MockStructuredMessage{
		body:   testEventSerialized,
		format: format.JSON,
	}
}

func (s *MockStructuredMessage) Event(b EventEncoder) error {
	e := cloudevents.Event{}
	err := s.format.Unmarshal(s.body, &e)
	if err != nil {
		return err
	}
	return b.SetEvent(e)
}

func (s *MockStructuredMessage) Structured(b StructuredEncoder) error {
	return b.SetStructuredEvent(s.format, s)
}

func (s *MockStructuredMessage) Binary(BinaryEncoder) error {
	return ErrNotBinary
}

func (s *MockStructuredMessage) IsEmpty() bool {
	return s.body == nil
}

func (s *MockStructuredMessage) Bytes() []byte {
	return s.body
}

func (s *MockStructuredMessage) Reader() io.Reader {
	return bytes.NewReader(s.body)
}

func (s *MockStructuredMessage) Finish(error) error { return nil }

var _ Message = (*MockStructuredMessage)(nil) // Test it conforms to the interface
