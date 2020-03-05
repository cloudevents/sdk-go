package test

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"

	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/binding/format"
	"github.com/cloudevents/sdk-go/pkg/event"
)

// MockStructuredMessage implements a structured-mode message as a simple struct.
// MockStructuredMessage implements both the binding.Message interface and the binding.StructuredEncoder
type MockStructuredMessage struct {
	Format format.Format
	Bytes  []byte
}

// Create a new MockStructuredMessage starting from an event.Event. Panics in case of error
func MustCreateMockStructuredMessage(e event.Event) binding.Message {
	testEventSerialized, err := format.JSON.Marshal(e)
	if err != nil {
		panic(err)
	}
	return &MockStructuredMessage{
		Bytes:  testEventSerialized,
		Format: format.JSON,
	}
}

func (s *MockStructuredMessage) Structured(ctx context.Context, b binding.StructuredEncoder) error {
	return b.SetStructuredEvent(ctx, s.Format, bytes.NewReader(s.Bytes))
}

func (s *MockStructuredMessage) Binary(context.Context, binding.BinaryEncoder) error {
	return binding.ErrNotBinary
}

func (bm *MockStructuredMessage) Encoding() binding.Encoding {
	return binding.EncodingStructured
}

func (s *MockStructuredMessage) Finish(error) error { return nil }

func (s *MockStructuredMessage) SetStructuredEvent(ctx context.Context, format format.Format, event io.Reader) (err error) {
	s.Format = format
	s.Bytes, err = ioutil.ReadAll(event)
	if err != nil {
		return
	}

	return nil
}

var _ binding.Message = (*MockStructuredMessage)(nil)
var _ binding.StructuredEncoder = (*MockStructuredMessage)(nil)
