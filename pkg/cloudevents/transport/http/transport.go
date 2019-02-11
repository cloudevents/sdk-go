package http

import (
	"bytes"
	"fmt"
	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport"
	"github.com/davecgh/go-spew/spew"
	"io/ioutil"
	"log"
	"net/http"
)

// type check that this transport message impl matches the contract
var _ transport.Sender = (*Transport)(nil)

type Transport struct {
	Encoding Encoding
	Client   *http.Client

	codec transport.Codec
}

func (t *Transport) Send(event cloudevents.Event, req *http.Request) (*http.Response, error) {
	if t.Client == nil {
		t.Client = &http.Client{}
	}

	if t.codec == nil {
		switch t.Encoding {
		case Default: // Move this to set default codec
			fallthrough
		case BinaryV01:
			fallthrough
		case StructuredV01:
			t.codec = &CodecV01{Encoding: t.Encoding}
		case BinaryV02:
			fallthrough
		case StructuredV02:
			t.codec = &CodecV02{Encoding: t.Encoding}
		default:
			return nil, fmt.Errorf("unknown codec set on sender: %d", t.codec)
		}
	}

	msg, err := t.codec.Encode(event)
	if err != nil {
		return nil, err
	}

	// TODO: merge the incoming request with msg, for now just replace.
	if m, ok := msg.(*Message); ok {
		req.Header = m.Header
		req.Body = ioutil.NopCloser(bytes.NewBuffer(m.Body))
		req.ContentLength = int64(len(m.Body))
		return t.Client.Do(req)
	}

	return nil, fmt.Errorf("failed to encode Event into a Message")
}

// ServeHTTP implements http.Handler
func (t *Transport) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Failed to handle request: %s %s", err, spew.Sdump(r))
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Invalid request"}`))
		return
	}
	msg := &Message{
		Header: r.Header,
		Body:   body,
	}
	_ = msg // TODO

	// TODO: respond correctly based on decode.
	w.WriteHeader(http.StatusNoContent)
}
