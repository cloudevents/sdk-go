package transport

import "github.com/cloudevents/sdk-go/pkg/cloudevents"

// Codec is the interface for transport codecs to convert between transport
// specific payloads and the Message interface.
type Codec interface {
	Encode(cloudevents.Event) (Message, error)
	Decode(Message) (*cloudevents.Event, error)

	BatchEncode([]cloudevents.Event) (Message, error)
	BatchDecode(Message) ([]cloudevents.Event, error)
}
