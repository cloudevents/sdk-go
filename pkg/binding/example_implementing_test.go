package binding_test

import (
	"context"
	"encoding/json"
	"io"

	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	ce "github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport"
)

// ExMessage is a json.RawMessage, a byte slice containing a JSON encoded event.
// It implements binding.StructMessage
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

func (m ExMessage) Structured() (string, []byte) {
	return cloudevents.ApplicationCloudEventsJSON, []byte(m)
}

func (m ExMessage) Event() (e cloudevents.Event, err error) {
	err = json.Unmarshal(json.RawMessage(m), &e)
	return e, err
}

func (m ExMessage) Finish(error) error { return nil }

// ExSender sends by writing JSON encoded events to an io.Writer
type ExSender struct{ encoder *json.Encoder }

func NewExSender(w io.Writer) ExSender { return ExSender{json.NewEncoder(w)} }

func (s ExSender) Send(ctx context.Context, m binding.Message) error {
	if f, b := m.Structured(); f == ce.ApplicationCloudEventsJSON {
		// Fast case: Message is already structured JSON.
		return s.encoder.Encode(json.RawMessage(b))
	}
	// Some other message encoding. Decode as generic Event and re-encode.
	if e, err := m.Event(); err != nil {
		return err
	} else if err := s.encoder.Encode(&e); err != nil {
		return err
	}
	return nil
}
func (s ExSender) Close(context.Context) error { return nil }

// ExReceiver receives by reading JSON encoded events from an io.Reader
type ExReceiver struct{ decoder *json.Decoder }

func NewExReceiver(r io.Reader) ExReceiver { return ExReceiver{json.NewDecoder(r)} }

func (r ExReceiver) Receive(context.Context) (binding.Message, error) {
	var rm json.RawMessage
	err := r.decoder.Decode(&rm) // This is just a byte copy.
	return ExMessage(rm), err
}
func (r ExReceiver) Close(context.Context) error { return nil }

// NewExTransport returns a transport.Transport which is implemented by
// an ExSender and an ExReceiver
func NewExTransport(r io.Reader, w io.Writer) transport.Transport {
	return binding.NewTransport(NewExSender(w), NewExReceiver(r))
}

// Example of implementing a transport including a simple message type,
// and a transport sender and receiver.
func Example_implementing() {}
