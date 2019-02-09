package http

import (
	"bytes"
	"fmt"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/canonical"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport"
	"io/ioutil"
	"net/http"
)

// type check that this transport message impl matches the contract
var _ transport.Sender = (*Sender)(nil)

type Sender struct {
	Encoding Encoding
	Client   *http.Client

	codec transport.Codec
}

func (s *Sender) Send(event canonical.Event, req *http.Request) (*http.Response, error) {
	if s.Client == nil {
		s.Client = &http.Client{}
	}

	if s.codec == nil {
		switch s.Encoding {
		case Default: // Move this to set default codec
			fallthrough
		case BinaryV01:
			fallthrough
		case StructuredV01:
			s.codec = &CodecV01{Encoding: s.Encoding}
		case BinaryV02:
			fallthrough
		case StructuredV02:
			fallthrough
		default:
			return nil, fmt.Errorf("unknown codec set on sender: %d", s.codec)
		}
	}

	msg, err := s.codec.Encode(event)
	if err != nil {
		return nil, err
	}

	// TODO: merge the incoming request with msg, for now just replace.
	if m, ok := msg.(*Message); ok {
		req.Header = m.Header
		req.Body = ioutil.NopCloser(bytes.NewBuffer(m.Body))
		req.ContentLength = int64(len(m.Body))
		return s.Client.Do(req)
	}

	return nil, fmt.Errorf("failed to encode Event into a Message")
}
