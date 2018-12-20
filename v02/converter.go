package v02

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"reflect"

	"github.com/cloudevents/sdk-go"
)

// HTTPMarshaller A struct representing the v02 version of the HTTPMarshaller
type HTTPMarshaller struct {
	converters []cloudevents.HTTPCloudEventConverter
}

// NewDefaultHTTPMarshaller creates a new v02 HTTPMarshaller prepopulated with the Binary and JSON
// CloudEvent converters
func NewDefaultHTTPMarshaller() cloudevents.HTTPMarshaller {
	return NewHTTPMarshaller(
		NewJSONHTTPCloudEventConverter(),
		NewBinaryHTTPCloudEventConverter())
}

// NewHTTPMarshaller creates a new HTTPMarshaller with the given HTTPCloudEventConverters
func NewHTTPMarshaller(converters ...cloudevents.HTTPCloudEventConverter) cloudevents.HTTPMarshaller {
	return &HTTPMarshaller{
		converters: converters,
	}
}

// FromRequest creates a new CloudEvent from an http Request
func (e HTTPMarshaller) FromRequest(req *http.Request, event interface{}) error {
	if req == nil {
		return cloudevents.IllegalArgumentError("req")
	}

	mimeType, _, err := mime.ParseMediaType(req.Header.Get("Content-Type"))
	if err != nil {
		return fmt.Errorf("error parsing request content type: %s", err.Error())
	}

	for _, v := range e.converters {
		if v.CanRead(reflect.TypeOf(event), mimeType) {
			return v.Read(req, event)
		}
	}
	return cloudevents.ContentTypeNotSupportedError(mimeType)
}

// ToRequest populates an http Request with the given CloudEvent
func (e HTTPMarshaller) ToRequest(req *http.Request, event interface{}) error {
	if req == nil {
		return cloudevents.IllegalArgumentError("req")
	}

	if event == nil || reflect.DeepEqual(event, reflect.Zero(reflect.TypeOf(event)).Interface()) {
		return cloudevents.IllegalArgumentError("event")
	}

	var mimeType string
	if reflect.TypeOf(event).Implements(reflect.TypeOf(new(cloudevents.HasContentType)).Elem()) {
		if contentTyped, ok := event.(cloudevents.HasContentType); ok {
			mimeType = contentTyped.GetContentType()
		}
	}

	if mimeType == "" {
		mimeType = "application/cloudevents+json"
	}

	mimeType, _, err := mime.ParseMediaType(mimeType)
	if err != nil {
		return fmt.Errorf("error parsing event content type: %s", err.Error())
	}

	for _, v := range e.converters {
		if v.CanWrite(reflect.TypeOf(event), mimeType) {
			err := v.Write(req, event)
			if err != nil {
				return err
			}

			return nil
		}
	}

	return cloudevents.ContentTypeNotSupportedError(mimeType)
}

// jsonhttpCloudEventConverter new converter for reading/writing CloudEvents to JSON
type jsonhttpCloudEventConverter struct {
	supportedMediaTypes      map[string]bool
	supportedMediaTypesSlice []string
}

// NewJSONHTTPCloudEventConverter creates a new JSONHTTPCloudEventConverter
func NewJSONHTTPCloudEventConverter() cloudevents.HTTPCloudEventConverter {
	mediaTypes := map[string]bool{
		"application/cloudevents+json": true,
	}

	return &jsonhttpCloudEventConverter{
		supportedMediaTypes: mediaTypes,
	}
}

// CanRead specifies if this converter can read the given mediaType into a given reflect.Type
func (j *jsonhttpCloudEventConverter) CanRead(t reflect.Type, mediaType string) bool {
	return t.Implements(reflect.TypeOf(new(json.Unmarshaler)).Elem()) && j.supportedMediaTypes[mediaType]
}

// CanWrite specifies if this converter can write the given Type into the given mediaType
func (j *jsonhttpCloudEventConverter) CanWrite(t reflect.Type, mediaType string) bool {
	return t.Implements(reflect.TypeOf(new(json.Marshaler)).Elem()) && j.supportedMediaTypes[mediaType]
}

func (j *jsonhttpCloudEventConverter) Read(req *http.Request, event interface{}) error {
	err := json.NewDecoder(req.Body).Decode(event)

	if err != nil {
		return fmt.Errorf("error parsing request: %s", err.Error())
	}

	return nil
}

func (j *jsonhttpCloudEventConverter) Write(req *http.Request, event interface{}) error {
	buffer := bytes.Buffer{}
	if err := json.NewEncoder(&buffer).Encode(event); err != nil {
		return err
	}

	req.Body = ioutil.NopCloser(&buffer)
	req.ContentLength = int64(buffer.Len())
	req.GetBody = func() (io.ReadCloser, error) {
		reader := bytes.NewReader(buffer.Bytes())
		return ioutil.NopCloser(reader), nil
	}

	req.Header.Set("Content-Type", "application/cloudevents+json")
	return nil
}

// BinaryHTTPCloudEventConverter a converter for reading/writing CloudEvents into the binary format
type binaryHTTPCloudEventConverter struct {
	supportedMediaTypes map[string]bool
}

// NewBinaryHTTPCloudEventConverter creates a new BinaryHTTPCloudEventConverter
func NewBinaryHTTPCloudEventConverter() cloudevents.HTTPCloudEventConverter {
	mediaTypes := map[string]bool{
		"application/json":         true,
		"application/xml":          true,
		"application/octet-stream": true,
	}

	return &binaryHTTPCloudEventConverter{
		supportedMediaTypes: mediaTypes,
	}
}

// CanRead specifies if this converter can read the given mediaType into a given reflect.Type
func (b *binaryHTTPCloudEventConverter) CanRead(t reflect.Type, mediaType string) bool {
	return t.Implements(reflect.TypeOf(new(cloudevents.BinaryUnmarshaler)).Elem()) && b.supportedMediaTypes[mediaType]
}

// CanWrite specifies if this converter can write the given Type into the given mediaType
func (b *binaryHTTPCloudEventConverter) CanWrite(t reflect.Type, mediaType string) bool {
	return t.Implements(reflect.TypeOf(new(cloudevents.BinaryMarshaler)).Elem()) && b.supportedMediaTypes[mediaType]
}

func (b *binaryHTTPCloudEventConverter) Read(req *http.Request, event interface{}) error {
	m, ok := reflect.ValueOf(event).Interface().(cloudevents.BinaryUnmarshaler)
	if !ok {
		return cloudevents.IllegalArgumentError("ev")
	}

	if err := m.UnmarshalBinary(req); err != nil {
		return err
	}

	return nil
}

func (b *binaryHTTPCloudEventConverter) Write(req *http.Request, event interface{}) error {
	m, ok := reflect.ValueOf(event).Interface().(cloudevents.BinaryMarshaler)
	if !ok {
		return cloudevents.IllegalArgumentError("ev")
	}

	if err := m.MarshalBinary(req); err != nil {
		return err
	}
	return nil
}
