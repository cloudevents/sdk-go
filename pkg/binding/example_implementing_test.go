package binding_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"

	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/binding/format"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport"
)

// ExMessage is a json.RawMessage, a byte slice containing a JSON encoded event.
// It implements binding.Message
type ExMessage json.RawMessage

func NewMessage(m json.RawMessage) binding.Message {
	return ExMessage(m)
}

func (m ExMessage) Encoding() binding.Encoding {
	return binding.EncodingStructured
}

func (m ExMessage) Structured(ctx context.Context, b binding.StructuredEncoder) error {
	return b.SetStructuredEvent(ctx, format.JSON, bytes.NewReader(m))
}

func (m ExMessage) Binary(context.Context, binding.BinaryEncoder) error {
	return binding.ErrNotBinary
}

func (m ExMessage) Finish(error) error { return nil }

var _ binding.Message = (*ExMessage)(nil)

// ExSender sends by writing JSON encoded events to an io.Writer
// ExSender implements directly StructuredEncoder in order to perform the conversion
type ExSender struct {
	encoder      *json.Encoder
	transformers binding.TransformerFactories
}

func NewExSender(w io.Writer, factories ...binding.TransformerFactory) binding.Sender {
	return &ExSender{encoder: json.NewEncoder(w), transformers: factories}
}

func (s *ExSender) Send(ctx context.Context, m binding.Message) error {
	// Encode tries to perform the encoding directly, otherwise
	// It fallbacks to converting to event.Event and then convert back to the provided encoders
	_, err := binding.Encode(
		ctx,
		m,
		s,
		nil,
		s.transformers,
	)

	return err
}

func (s *ExSender) SetStructuredEvent(ctx context.Context, f format.Format, event io.Reader) error {
	if f == format.JSON {
		b, err := ioutil.ReadAll(event)
		if err != nil {
			return err
		}
		return s.encoder.Encode(json.RawMessage(b))
	} else {
		return binding.ErrNotStructured
	}
}

var _ binding.Sender = (*ExSender)(nil)
var _ binding.StructuredEncoder = (*ExSender)(nil)

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
	return binding.NewSendingTransport(NewExSender(w), NewExReceiver(r), []func(ctx context.Context) context.Context{})
}
