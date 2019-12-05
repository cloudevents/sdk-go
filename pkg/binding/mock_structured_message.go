package binding

import (
	"bytes"

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/cloudevents/sdk-go/pkg/binding/format"
)

// MockStructuredMessage implements a structured-mode message as a simple struct.
type MockStructuredMessage struct {
	Format format.Format
	Bytes  []byte
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

func (s *MockStructuredMessage) Event(b EventMessageBuilder) error {
	e := cloudevents.Event{}
	err := s.Format.Unmarshal(s.Bytes, &e)
	if err != nil {
		return err
	}
	return b.Encode(e)
}

func (s *MockStructuredMessage) Structured(b StructuredMessageBuilder) error {
	return b.Event(s.Format, bytes.NewReader(s.Bytes))
}

func (s *MockStructuredMessage) Binary(BinaryMessageBuilder) error {
	return ErrNotBinary
}

func (s *MockStructuredMessage) Finish(error) error { return nil }

var _ Message = (*MockStructuredMessage)(nil) // Test it conforms to the interface
