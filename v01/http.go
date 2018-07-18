package v01

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"mime"
	"net/http"
	"strings"
	"time"

	"github.com/dispatchframework/cloudevents-go-sdk"
)

// HTTPFormat type wraps supported modes of formatting CloudEvent as HTTP request.
// Currently, only binary mode and structured mode with JSON encoding are supported.
type HTTPFormat string

const (
	// FormatBinary corresponds to Binary mode in CloudEvents HTTP transport binding.
	// https://github.com/cloudevents/spec/blob/a12b6b618916c89bfa5595fc76732f07f89219b5/http-transport-binding.md#31-binary-content-mode
	FormatBinary HTTPFormat = "binary"
	// FormatJSON corresponds to Structured mode using JSON encoding.
	// https://github.com/cloudevents/spec/blob/a12b6b618916c89bfa5595fc76732f07f89219b5/http-transport-binding.md#32-structured-content-mode
	FormatJSON HTTPFormat = "json"
)

const (
	ceContentType = "application/cloudevents"
)

// FromHTTPRequest parses the http request and returns a CloudEvent.
// https://github.com/cloudevents/spec/blob/a12b6b618916c89bfa5595fc76732f07f89219b5/http-transport-binding.md
func (e *Event) FromHTTPRequest(req *http.Request) error {
	if req == nil {
		return errors.New("cannot process nil-request")
	}

	reqContentType := req.Header.Get("Content-Type")
	if strings.HasPrefix(reqContentType, ceContentType) {
		// CE content type should indicate structured mode according to spec
		return e.eventFromRequestStructured(req)
	}
	// binary mode

	e.Source = req.Header.Get(headerize(sourceKey))
	e.EventID = req.Header.Get(headerize(eventIDKey))
	e.EventType = req.Header.Get(headerize(eventTypeKey))

	if err := e.Validate(); err != nil {
		return fmt.Errorf("unable to parse event context from request headers: %s", err.Error())
	}

	e.ContentType = req.Header.Get("Content-Type")
	e.EventTypeVersion = req.Header.Get(headerize(eventTypeVersionKey))
	e.SchemaURL = req.Header.Get(headerize(schemaURLKey))
	timeString := req.Header.Get(headerize(eventTimeKey))
	if timeString != "" {

		t, err := time.Parse(time.RFC3339, timeString)
		if err != nil {
			return fmt.Errorf("error parsing the %s header: %s", headerize(eventTimeKey), err.Error())
		}
		e.EventTime = &t
	}

	// TODO: implement encoder/decoder registry

	if req.ContentLength == 0 {
		return nil
	}

	var err error
	e.Data, err = ioutil.ReadAll(req.Body)
	if err != nil {
		return fmt.Errorf("error reading request body: %s", err.Error())
	}
	return nil

}

func (e *Event) eventFromRequestStructured(req *http.Request) error {
	mimeType, _, err := mime.ParseMediaType(req.Header.Get("Content-Type"))
	if err != nil {
		return fmt.Errorf("error parsing request content type: %s", err.Error())
	}
	switch {
	case strings.HasSuffix(mimeType, "+json"):
		data, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return fmt.Errorf("error reading request body: %s", err.Error())
		}

		err = json.Unmarshal(data, e)
		if err != nil {
			return fmt.Errorf("error parsing request json: %s", err.Error())
		}
		return nil

	default:
		return cloudevents.ContentTypeNotSupportedError(mimeType)
	}
}

// ToHTTPRequest takes a pointer to existing http.Request struct and a binding format,
// and injects the event into to the struct using a format specified in CloudEvents HTTP transport binding.
// https://github.com/cloudevents/spec/blob/a12b6b618916c89bfa5595fc76732f07f89219b5/http-transport-binding.md
func (e *Event) ToHTTPRequest(req *http.Request, format HTTPFormat) error {
	switch format {
	case FormatBinary:
		req.Header.Set(headerize(eventTypeKey), e.EventType)
		req.Header.Set(headerize(eventIDKey), e.EventID)
		req.Header.Set(headerize(sourceKey), e.Source)

		if e.ContentType != "" {
			req.Header.Set("Content-Type", e.ContentType)
		}
		if e.EventTypeVersion != "" {
			req.Header.Set(headerize(eventTypeVersionKey), e.EventTypeVersion)
		}
		if e.SchemaURL != "" {
			req.Header.Set(headerize(schemaURLKey), e.SchemaURL)
		}
		if e.EventTime != nil {
			req.Header.Set(headerize(eventTimeKey), e.EventTime.Format(time.RFC3339))
		}

		// TODO: this is not bulletproof, should take care of nested keys
		for key, extensionVal := range e.Extensions {
			req.Header.Set("CE-X-"+strings.Title(key), fmt.Sprintf("%v", extensionVal))
		}

		data, err := marshalEventData(e.ContentType, e.Data)
		if err != nil {
			return fmt.Errorf("error marshaling event data: %s", err.Error())
		}
		req.Body = ioutil.NopCloser(bytes.NewReader(data))

		return nil

	case FormatJSON:
		data, err := e.MarshalJSON()
		if err != nil {
			return fmt.Errorf("error marshaling event to JSON: %s", err.Error())
		}
		req.Body = ioutil.NopCloser(bytes.NewBuffer(data))
		return nil
	default:
		return fmt.Errorf("format %s not implemented", format)
	}

}

func headerize(property string) string {
	return "CE-" + strings.Title(property)
}

func marshalEventData(encoding string, data interface{}) ([]byte, error) {
	var b []byte
	var err error

	switch {
	case isJSONEncoding(encoding):
		b, err = json.Marshal(data)

	default:
		return nil, fmt.Errorf("cannot encode content type %s", encoding)
	}

	if err != nil {
		return nil, err
	}
	return b, nil
}

func isJSONEncoding(encoding string) bool {
	return encoding == "application/json" || encoding == "text/json"
}
