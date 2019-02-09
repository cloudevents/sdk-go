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
var _ transport.Sender = (*Transport)(nil)

type Transport struct {
	Encoding Encoding
	Client   *http.Client

	codec transport.Codec
}

func (s *Transport) Send(event canonical.Event, req *http.Request) (*http.Response, error) {
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
			s.codec = &CodecV02{Encoding: s.Encoding}
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

// TODO: next is the decode.
// ServeHTTP implements http.Handler
//func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
//	args := make([]reflect.Value, 0, 2)
//
//	if h.numIn > 0 {
//		dataPtr, dataArg := allocate(h.dataType)
//		eventContext, err := FromRequest(dataPtr, r)
//		if err != nil {
//			log.Printf("Failed to handle request %s; error %s", spew.Sdump(r), err)
//			w.WriteHeader(http.StatusBadRequest)
//			w.Write([]byte(`Invalid request`))
//			return
//		}
//
//		ctx := r.Context()
//		ctx = context.WithValue(ctx, contextKey, eventContext)
//		args = append(args, reflect.ValueOf(ctx))
//
//		if h.numIn == 2 {
//			args = append(args, dataArg)
//		}
//	}
//
//	res := h.fnValue.Call(args)
//	respondHTTP(res, h.fnValue, w)
//}
