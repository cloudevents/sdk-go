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

// Transport acts as both a http client and a http handler.
type Transport struct {
	Encoding Encoding
	Client   *http.Client

	Receiver transport.Receiver

	codec transport.Codec
}

func (t *Transport) loadCodec() bool {
	if t.codec == nil {
		switch t.Encoding {
		case Default:
			t.codec = &Codec{}
		case BinaryV01:
			fallthrough
		case StructuredV01:
			t.codec = &CodecV01{Encoding: t.Encoding}
		case BinaryV02:
			fallthrough
		case StructuredV02:
			t.codec = &CodecV02{Encoding: t.Encoding}
		default:
			return false
		}
	}
	return true
}

func (t *Transport) Send(event cloudevents.Event, req *http.Request) (*http.Response, error) {
	if t.Client == nil {
		t.Client = &http.Client{}
	}

	if ok := t.loadCodec(); !ok {
		return nil, fmt.Errorf("unknown encoding set on transport: %d", t.Encoding)
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
		log.Printf("failed to handle request: %s %s", err, spew.Sdump(r))
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Invalid request"}`))
		return
	}
	msg := &Message{
		Header: r.Header,
		Body:   body,
	}
	_ = msg // TODO

	if ok := t.loadCodec(); !ok {
		err := fmt.Errorf("unknown encoding set on transport: %d", t.Encoding)
		log.Printf("failed to load codec: %s", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf(`{"error":%q}`, err.Error())))
		return
	}
	event, err := t.codec.Decode(msg)
	if err != nil {
		log.Printf("failed to decode message: %s %s", err, spew.Sdump(msg))
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Decoding Error"}`))
		return
	}

	if t.Receiver != nil {
		go t.Receiver.Receive(*event)
	}

	// TODO: respond correctly based on decode.
	w.WriteHeader(http.StatusNoContent)
}
