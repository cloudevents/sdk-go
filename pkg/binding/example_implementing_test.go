package binding_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"

	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/binding/format"
	ce "github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport"
)

// ExMessage is a json.RawMessage, a byte slice containing a JSON encoded event.
// It implements binding.MockStructuredMessage
//
// Note: a good binding implementation should provide an easy way to convert
// between the Message implementation and the "native" message format.
// In this case it's as simple as:
//
//    native = ExMessage(impl)
//    impl = json.RawMessage(native)
//
// For example in a HTTP binding it should be easy to convert between
// the HTTP binding.Message implementation and net/http.Request and
// Response types.  There are no interfaces for this conversion as it
// requires the use of unknown types.
type ExMessage json.RawMessage

func (m ExMessage) Structured(b binding.StructuredEncoder) error {
	return b.SetStructuredEvent(format.JSON, &m)
}

func (m ExMessage) Binary(binding.BinaryEncoder) error {
	return binding.ErrNotBinary
}

func (m ExMessage) Event(b binding.EventEncoder) error {
	e := ce.Event{}
	err := json.Unmarshal(m, &e)
	if err != nil {
		return err
	}
	return b.SetEvent(e)
}

func (m *ExMessage) IsEmpty() bool {
	return m == nil
}

func (m *ExMessage) Bytes() []byte {
	return *m
}

func (m *ExMessage) Reader() io.Reader {
	return bytes.NewReader(m.Bytes())
}

func (m ExMessage) Finish(error) error { return nil }

var _ binding.Message = (*ExMessage)(nil)
var _ binding.MessagePayloadReader = (*ExMessage)(nil)

// ExSender sends by writing JSON encoded events to an io.Writer
// ExSender supports transcoding
// ExSender implements directly StructuredEncoder & EventEncoder
type ExSender struct {
	encoder      *json.Encoder
	transformers binding.TransformerFactories
}

func NewExSender(w io.Writer, factories ...binding.TransformerFactory) binding.Sender {
	return &ExSender{encoder: json.NewEncoder(w), transformers: factories}
}

func (s *ExSender) Send(ctx context.Context, m binding.Message) error {
	// Invoke m.Finish to notify the receiver that message was processed
	defer func() { _ = m.Finish(nil) }()

	// Translate tries the various encodings, starting with provided root encoder factories.
	// If a sender doesn't support a specific encoding, a null root encoder factory could be provided.
	_, _, err := binding.Translate(
		m,
		func() binding.StructuredEncoder {
			return s
		},
		nil,
		func() binding.EventEncoder {
			return s
		},
		s.transformers)

	return err
}

func (s *ExSender) SetStructuredEvent(f format.Format, event binding.MessagePayloadReader) error {
	if f == format.JSON {
		return s.encoder.Encode(event.Bytes())
	} else {
		return binding.ErrNotStructured
	}
}

func (s *ExSender) SetEvent(event ce.Event) error {
	return s.encoder.Encode(&event)
}

func (s *ExSender) Close(context.Context) error { return nil }

var _ binding.Sender = (*ExSender)(nil)
var _ binding.StructuredEncoder = (*ExSender)(nil)
var _ binding.EventEncoder = (*ExSender)(nil)

// ExReceiver receives by reading JSON encoded events from an io.Reader
type ExReceiver struct{ decoder *json.Decoder }

func NewExReceiver(r io.Reader) binding.Receiver { return &ExReceiver{json.NewDecoder(r)} }

func (r *ExReceiver) Receive(context.Context) (binding.Message, error) {
	var rm json.RawMessage
	err := r.decoder.Decode(&rm) // This is just a byte copy.
	return ExMessage(rm), err
}
func (r *ExReceiver) Close(context.Context) error { return nil }

// NewExTransport returns a transport.Transport which is implemented by
// an ExSender and an ExReceiver
func NewExTransport(r io.Reader, w io.Writer) transport.Transport {
	return binding.NewTransport(NewExSender(w), NewExReceiver(r))
}

// Example of implementing a transport including a simple message type,
// and a transport sender and receiver.
func Example_implementing() {}
