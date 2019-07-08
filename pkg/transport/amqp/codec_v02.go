package amqp

import (
	"fmt"

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/cloudevents/sdk-go/pkg/transport"
)

type CodecV02 struct {
	CodecStructured

	Encoding Encoding
}

var _ transport.Codec = (*CodecV02)(nil)

func (v CodecV02) Encode(e cloudevents.Event) (transport.Message, error) {
	switch v.Encoding {
	case Default:
		fallthrough
	case StructuredV02:
		return v.encodeStructured(e)
	default:
		return nil, fmt.Errorf("unknown encoding: %d", v.Encoding)
	}
}

func (v CodecV02) Decode(msg transport.Message) (*cloudevents.Event, error) {
	// only structured is supported as of v0.2
	switch v.inspectEncoding(msg) {
	case StructuredV02:
		return v.decodeStructured(cloudevents.VersionV02, msg)
	default:
		return nil, transport.NewErrMessageEncodingUnknown("v02", TransportName)
	}
}

func (v CodecV02) inspectEncoding(msg transport.Message) Encoding {
	version := msg.CloudEventsVersion()
	if version != cloudevents.VersionV02 {
		return Unknown
	}
	m, ok := msg.(*Message)
	if !ok {
		return Unknown
	}
	contentType := m.ContentType
	if contentType == cloudevents.ApplicationCloudEventsJSON {
		return StructuredV02
	}
	return BinaryV02
}
