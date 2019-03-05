package http

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	cecontext "github.com/cloudevents/sdk-go/pkg/cloudevents/context"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport"
)

type EncodingSelector func(e cloudevents.Event) Encoding

// type check that this transport message impl matches the contract
var _ transport.Transport = (*Transport)(nil)

// Transport acts as both a http client and a http handler.
type Transport struct {
	Encoding                   Encoding
	DefaultEncodingSelectionFn EncodingSelector

	// Sending
	Client *http.Client
	Req    *http.Request

	// Receiving
	Port     int    // default 8080
	Path     string // default "/"
	Receiver transport.Receiver

	codec transport.Codec
}

func (t *Transport) loadCodec() bool {
	if t.codec == nil {
		if t.DefaultEncodingSelectionFn != nil && t.Encoding != Default {
			log.Printf("[warn] Transport has a DefaultEncodingSelectionFn set but Encoding is not Default. DefaultEncodingSelectionFn will be ignored.")
		}
		t.codec = &Codec{
			Encoding:                   t.Encoding,
			DefaultEncodingSelectionFn: t.DefaultEncodingSelectionFn,
		}
	}
	return true
}

func (t *Transport) Send(ctx context.Context, event cloudevents.Event) (*cloudevents.Event, error) {
	if t.Client == nil {
		t.Client = &http.Client{}
	}

	var req http.Request
	if t.Req != nil {
		req.Method = t.Req.Method
		req.URL = t.Req.URL
	}

	// Override the default request with target from context.
	if target := cecontext.TargetFromContext(ctx); target != nil {
		req.URL = target
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
		return httpDo(ctx, &req, func(resp *http.Response, err error) (*cloudevents.Event, error) {
			if err != nil {
				return nil, err
			}
			defer resp.Body.Close()

			body, _ := ioutil.ReadAll(resp.Body)
			msg := &Message{
				Header: resp.Header,
				Body:   body,
			}

			var respEvent *cloudevents.Event
			if msg.CloudEventsVersion() != "" {
				if ok := t.loadCodec(); !ok {
					err := fmt.Errorf("unknown encoding set on transport: %d", t.Encoding)
					log.Printf("failed to load codec: %s", err)
				}
				if respEvent, err = t.codec.Decode(msg); err != nil {
					log.Printf("failed to decode message: %s %v", err, resp)
				}
			}

			if accepted(resp) {
				return respEvent, nil
			}
			return respEvent, fmt.Errorf("error sending cloudevent: %s", status(resp))
		})
	}

	return nil, fmt.Errorf("failed to encode Event into a Message")
}

type eventError struct {
	event *cloudevents.Event
	err   error
}

// TODO: finalize
type TransportContext struct {
	URI    string
	Host   string
	Method string
}

func httpDo(ctx context.Context, req *http.Request, f func(*http.Response, error) (*cloudevents.Event, error)) (*cloudevents.Event, error) {
	// Run the HTTP request in a goroutine and pass the response to f.
	c := make(chan eventError, 1)
	req = req.WithContext(ctx)
	go func() {
		event, err := f(http.DefaultClient.Do(req))

		if event != nil {
			event.TransportContext = &TransportContext{
				URI:    req.RequestURI,
				Host:   req.Host,
				Method: req.Method,
			}
		}

		c <- eventError{event: event, err: err}
	}()
	select {
	case <-ctx.Done():
		<-c // Wait for f to return.
		return nil, ctx.Err()
	case ee := <-c:
		return ee.event, ee.err
	}
}

// accepted is a helper method to understand if the response from the target
// accepted the CloudEvent.
func accepted(resp *http.Response) bool {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return true
	}
	return false
}

// status is a helper method to read the response of the target.
func status(resp *http.Response) string {
	status := resp.Status
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Sprintf("Status[%s] error reading response body: %v", status, err)
	}
	return fmt.Sprintf("Status[%s] %s", status, body)
}

func (t *Transport) invokeReceiver(event cloudevents.Event) (*Response, error) {
	if t.Receiver != nil {
		respEvent, err := t.Receiver.Receive(event)
		resp := Response{}
		if respEvent != nil && t.loadCodec() {
			m, err2 := t.codec.Encode(*respEvent)
			if err2 != nil {
				log.Printf("failed to encode response from receiver fn: %s", err2.Error())
			} else if msg, ok := m.(*Message); ok {
				resp.Header = msg.Header
				resp.Body = msg.Body
			}
		}
		if err != nil {
			resp.StatusCode = http.StatusBadRequest
		} else {
			resp.StatusCode = http.StatusAccepted
		}
		return &resp, err
	}
	return nil, nil
}

// ServeHTTP implements http.Handler
func (t *Transport) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("failed to handle request: %s %v", err, r)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Invalid request"}`))
		return
	}
	msg := &Message{
		Header: r.Header,
		Body:   body,
	}

	if ok := t.loadCodec(); !ok {
		err := fmt.Errorf("unknown encoding set on transport: %d", t.Encoding)
		log.Printf("failed to load codec: %s", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf(`{"error":%q}`, err.Error())))
		return
	}
	event, err := t.codec.Decode(msg)
	if err != nil {
		log.Printf("failed to decode message: %s %v", err, r)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf(`{"error":%q}`, err.Error())))
		return
	}

	resp, err := t.invokeReceiver(*event)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf(`{"error":%q}`, err.Error())))
		return
	}
	if resp != nil {
		if len(resp.Header) > 0 {
			for k, vs := range resp.Header {
				for _, v := range vs {
					w.Header().Set(k, v)
				}
			}
		}
		if resp.StatusCode >= 200 && resp.StatusCode < 600 {
			w.WriteHeader(resp.StatusCode)
		}
		if len(resp.Body) > 0 {
			w.Write(resp.Body)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (t *Transport) GetPort() int {
	if t.Port > 0 {
		return t.Port
	}
	return 8080 // default
}

func (t *Transport) GetPath() string {
	path := strings.TrimSpace(t.Path)
	if len(path) > 0 {
		return path
	}
	return "/" // default
}
