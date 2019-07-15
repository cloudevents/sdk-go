package x_test

import (
	"context"
	"encoding/json"
	"io"

	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport/x"
)

// ExMessage is a json.RawMessage, which is just a byte slice
// containing a JSON encoded event.
type ExMessage struct{ json.RawMessage }

func (m ExMessage) Structured() (string, []byte) {
	return cloudevents.ApplicationCloudEventsJSON, []byte(m.RawMessage)
}

func (m ExMessage) Event() (e cloudevents.Event, err error) {
	err = json.Unmarshal(m.RawMessage, &e)
	return e, err
}

func (m ExMessage) Finish(error) {}

// ExSender sends by writing JSON encoded events to an io.Writer
type ExSender struct{ *json.Encoder }

func NewExSender(w io.Writer) ExSender { return ExSender{json.NewEncoder(w)} }

func (s ExSender) Send(ctx context.Context, m x.Message) error {
	if t, b := m.Structured(); t != "" {
		// Fast case: if the Message is already structured JSON we can
		// send it directly, no need to decode and re-encode. Encoding a
		// json.RawMessage to a json.Encoder() is just a byte-buffer copy.
		return s.Encode(json.RawMessage(b))
	} else {
		// Some other message encoding. Decode as a generic cloudevents.Event
		// and then re-encode as JSON
		if e, err := m.Event(); err != nil {
			return err
		} else if err := s.Encode(e); err != nil {
			return err
		}
		return nil
	}
}

// ExReceiver receives by reading JSON encoded events from an io.Reader
type ExReceiver struct{ *json.Decoder }

func NewExReceiver(r io.Reader) ExReceiver { return ExReceiver{json.NewDecoder(r)} }

func (sr ExReceiver) Receive(context.Context) (x.Message, error) {
	var m ExMessage
	err := sr.Decode(&m) // This is just a byte copy since m is a json.RawMessage
	return m, err
}

// NewExTransport returns a transport.Transport which is implemented by
// an ExSender and an ExReceiver
func NewExTransport(r io.Reader, w io.Writer) transport.Transport {
	return x.NewTransport(NewExSender(w), NewExReceiver(r))
}

// Example of implementing a transport including a simple message type,
// and a transport sender and receiver.
func Example_implementing() {}
