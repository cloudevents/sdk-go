package v02_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
	"time"

	cloudevents "github.com/cloudevents/sdk-go"
	mocks "github.com/cloudevents/sdk-go/mocks"
	"github.com/cloudevents/sdk-go/v02"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestFromRequestNilRequest(t *testing.T) {
	principal := v02.NewDefaultHTTPMarshaller()

	var actual v02.Event
	err := principal.FromRequest(nil, &actual)

	assert.Zero(t, actual)
	require.Error(t, err)
	assert.Equal(t, cloudevents.IllegalArgumentError("req"), err)
}

func TestFromRequestNoContentType(t *testing.T) {
	principal := v02.NewDefaultHTTPMarshaller()

	req := httptest.NewRequest("GET", "localhost:8080", nil)

	var actual v02.Event
	err := principal.FromRequest(req, &actual)

	assert.Zero(t, actual)
	assert.Error(t, err)
}

func TestFromRequestInvalidContentType(t *testing.T) {
	principal := v02.NewDefaultHTTPMarshaller()

	req := httptest.NewRequest("GET", "localhost:8080", nil)
	req.Header.Set("Content-Type", "application/")

	var actual v02.Event
	err := principal.FromRequest(req, &actual)
	assert.Zero(t, actual)
	assert.Error(t, err)
}

func TestFromRequestNoConverters(t *testing.T) {
	principal := v02.NewHTTPMarshaller()

	req := httptest.NewRequest("GET", "localhost:8080", nil)
	req.Header.Set("Content-Type", "application/cloudevents+json")

	var actual v02.Event
	err := principal.FromRequest(req, &actual)

	assert.Zero(t, actual)
	require.Error(t, err)
	assert.Equal(t, cloudevents.ContentTypeNotSupportedError("application/cloudevents+json"), err)
}

func TestFromRequestWrongConverter(t *testing.T) {
	jsonConverter := &mocks.HTTPCloudEventConverter{}
	var actual v02.Event
	jsonConverter.On("CanRead", reflect.TypeOf(&actual), "application/json").Return(false)
	principal := v02.NewHTTPMarshaller(jsonConverter)

	req := httptest.NewRequest("GET", "localhost:8080", nil)
	req.Header.Set("Content-Type", "application/json")

	err := principal.FromRequest(req, &actual)

	assert.Zero(t, actual)
	jsonConverter.AssertNumberOfCalls(t, "CanRead", 1)
	jsonConverter.AssertNotCalled(t, "Read")
	require.Error(t, err)
	assert.Equal(t, cloudevents.ContentTypeNotSupportedError("application/json"), err)
}

func TestFromRequestConverterError(t *testing.T) {
	jsonConverter := &mocks.HTTPCloudEventConverter{}
	var actual v02.Event
	jsonConverter.On("CanRead", reflect.TypeOf(&actual), "application/json").Return(true)
	jsonConverter.On("Read", mock.AnythingOfType(reflect.TypeOf((*http.Request)(nil)).String()), &actual).Return(errors.New("read error"))
	principal := v02.NewHTTPMarshaller(jsonConverter)

	req := httptest.NewRequest("GET", "localhost:8080", nil)
	req.Header.Set("Content-Type", "application/json")

	err := principal.FromRequest(req, &actual)

	assert.Zero(t, actual)
	jsonConverter.AssertNumberOfCalls(t, "CanRead", 1)
	jsonConverter.AssertNumberOfCalls(t, "Read", 1)
	require.Error(t, err)
	assert.Equal(t, errors.New("read error"), err)
}

func TestFromRequestSuccess(t *testing.T) {
	builder := v02.NewCloudEventBuilder()
	expected, _ := builder.
		Type("com.example.someevent").
		ID("00001").
		Source(url.URL{
			Scheme: "http",
			Host:   "example.com",
			Path:   "/cloudevent",
		}).
		Build()

	var actual v02.Event

	jsonConverter := &mocks.HTTPCloudEventConverter{}
	jsonConverter.On("CanRead", reflect.TypeOf(&actual), "application/json").Return(false)
	binaryConverter := &mocks.HTTPCloudEventConverter{}
	binaryConverter.On("CanRead", reflect.TypeOf(&actual), "application/json").Return(true)
	binaryConverter.On("Read", mock.AnythingOfType(reflect.TypeOf((*http.Request)(nil)).String()), &actual).Run(func(args mock.Arguments) {
		arg := args.Get(1).(*v02.Event)
		*arg = expected
	}).Return(nil)

	principal := v02.NewHTTPMarshaller(jsonConverter, binaryConverter)

	req := &http.Request{
		Header: map[string][]string{},
	}
	req.Header.Set("Content-Type", "application/json")

	err := principal.FromRequest(req, &actual)

	binaryConverter.AssertNumberOfCalls(t, "CanRead", 1)
	binaryConverter.AssertNumberOfCalls(t, "Read", 1)
	require.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestToRequestNilRequest(t *testing.T) {
	principal := v02.NewHTTPMarshaller()

	event := &v02.Event{}
	err := principal.ToRequest(nil, event)

	assert.Error(t, err)
	assert.Equal(t, cloudevents.IllegalArgumentError("req"), err)
}

func TestToRequestNilEvent(t *testing.T) {
	principal := v02.NewHTTPMarshaller()

	req := httptest.NewRequest("GET", "localhost:8080", nil)
	err := principal.ToRequest(req, nil)

	assert.Error(t, err)
	assert.Equal(t, cloudevents.IllegalArgumentError("event"), err)
}

func TestToRequestDefaultContentType(t *testing.T) {
	jsonConverter := &mocks.HTTPCloudEventConverter{}
	event := &v02.Event{}
	jsonConverter.On("CanWrite", reflect.TypeOf(event), "application/cloudevents+json").Return(true)
	jsonConverter.On("Write", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	principal := v02.NewHTTPMarshaller(jsonConverter)

	req := httptest.NewRequest("GET", "localhost:8080", nil)
	err := principal.ToRequest(req, event)

	assert.NoError(t, err)
	jsonConverter.AssertCalled(t, "CanWrite", reflect.TypeOf(event), "application/cloudevents+json")
}

func TestToRequestInvalidContentType(t *testing.T) {
	principal := v02.NewDefaultHTTPMarshaller()

	req := httptest.NewRequest("GET", "localhost:8080", nil)
	builder := v02.NewCloudEventBuilder()
	event, _ := builder.
		ContentType("application/").
		ID("000001").
		Type("com.example.sample").
		Source(url.URL{
			Scheme: "http",
			Host:   "example.com",
			Path:   "/cloudevent",
		}).
		Build()

	err := principal.ToRequest(req, event)

	assert.Error(t, err)
}

func TestToRequestNoConverters(t *testing.T) {
	principal := v02.NewHTTPMarshaller()

	req := httptest.NewRequest("GET", "localhost:8080", nil)
	event := &v02.Event{}
	err := principal.ToRequest(req, event)

	require.Error(t, err)
	assert.Equal(t, cloudevents.ContentTypeNotSupportedError("application/cloudevents+json"), err)
}

func TestToRequestWriteError(t *testing.T) {
	jsonConverter := &mocks.HTTPCloudEventConverter{}
	event := v02.Event{}
	jsonConverter.On("CanWrite", reflect.TypeOf(event), "application/cloudevents+json").Return(true)
	jsonConverter.On("Write", reflect.TypeOf(event), mock.AnythingOfType(reflect.TypeOf((*http.Request)(nil)).String()), event).Return(errors.New("write error"))

	principal := v02.NewHTTPMarshaller(jsonConverter)

	req := httptest.NewRequest("GET", "localhost:8080", nil)

	err := principal.ToRequest(req, event)

	assert.Error(t, err)
}

func TestToRequestWrongConverter(t *testing.T) {
	binaryConverter := &mocks.HTTPCloudEventConverter{}
	event := &v02.Event{}
	binaryConverter.On("CanWrite", reflect.TypeOf(event), "application/cloudevents+json").Return(false)

	principal := v02.NewHTTPMarshaller(binaryConverter)

	req := httptest.NewRequest("GET", "localhost:8080", nil)

	err := principal.ToRequest(req, event)

	require.Error(t, err)
	assert.Equal(t, cloudevents.ContentTypeNotSupportedError("application/cloudevents+json"), err)
}

func TestJSONCanReadCanWriteBothWrong(t *testing.T) {
	principal := v02.NewJSONHTTPCloudEventConverter()

	actual := principal.CanRead(reflect.TypeOf((*cloudevents.Event)(nil)), "application/json")
	assert.Equal(t, false, actual)

	actual = principal.CanWrite(reflect.TypeOf((*cloudevents.Event)(nil)), "application/json")
	assert.Equal(t, false, actual)
}
func TestJSONConverterCanReadCanWriteWrongType(t *testing.T) {
	principal := v02.NewJSONHTTPCloudEventConverter()

	actual := principal.CanRead(reflect.TypeOf((*cloudevents.Event)(nil)), "application/cloudevents+json")
	assert.Equal(t, false, actual)

	actual = principal.CanWrite(reflect.TypeOf((*cloudevents.Event)(nil)), "application/cloudevents+json")
	assert.Equal(t, false, actual)
}

func TestJSONConverterCanReadCanWriteWrongMediaType(t *testing.T) {
	principal := v02.NewJSONHTTPCloudEventConverter()

	actual := principal.CanRead(reflect.TypeOf(v02.Event{}), "application/json")
	assert.Equal(t, false, actual)

	actual = principal.CanWrite(reflect.TypeOf(v02.Event{}), "application/json")
	assert.Equal(t, false, actual)
}

func TestJSONConverterCanReadCanWriteSuccess(t *testing.T) {
	principal := v02.NewJSONHTTPCloudEventConverter()

	var event v02.Event
	actual := principal.CanRead(reflect.TypeOf(&event), "application/cloudevents+json")
	assert.Equal(t, true, actual)

	actual = principal.CanWrite(reflect.TypeOf(&event), "application/cloudevents+json")
	assert.Equal(t, true, actual)
}

func TestJSONConverterReadError(t *testing.T) {
	principal := v02.NewJSONHTTPCloudEventConverter()

	req := httptest.NewRequest("GET", "locahost:8080", nil)
	var actual v02.Event
	err := principal.Read(req, &actual)

	assert.Error(t, err)
	assert.Zero(t, actual)
}

func TestJSONConverterReadSuccess(t *testing.T) {
	principal := v02.NewJSONHTTPCloudEventConverter()

	body := bytes.NewBufferString("{\"id\":\"1234-1234-1234\", \"source\": \"http://example.com/cloudevent\", \"specversion\": \"0.2\", \"type\":\"com.example.someevent\"}")
	req := httptest.NewRequest("GET", "localhost:8080", body)

	var actual v02.Event
	err := principal.Read(req, &actual)

	require.NoError(t, err)

	builder := v02.NewCloudEventBuilder()
	expected, err := builder.
		Type("com.example.someevent").
		ID("1234-1234-1234").
		Source(url.URL{
			Scheme: "http",
			Host:   "example.com",
			Path:   "/cloudevent",
		}).
		Build()

	assert.Equal(t, expected, actual)
}

func TestJSONConverterWriteError(t *testing.T) {
	principal := v02.NewJSONHTTPCloudEventConverter()

	req := httptest.NewRequest("GET", "localhost:8080", nil)
	event := &v02.Event{}

	err := principal.Write(req, event)

	assert.NoError(t, err)
}

func TestFromRequestJSONSuccess(t *testing.T) {
	factory := v02.NewDefaultHTTPMarshaller()

	myDate, err := time.Parse(time.RFC3339, "2018-04-05T03:56:24Z")
	if err != nil {
		t.Error(err.Error())
	}

	builder := v02.NewCloudEventBuilder()
	event, err := builder.
		SpecVersion("0.2").
		Type("com.example.someevent").
		ID("1234-1234-1234").
		Source(url.URL{
			Scheme: "http",
			Host:   "example.com",
			Path:   "/mycontext/subcontext",
		}).
		SchemaURL(url.URL{
			Scheme: "http",
			Host:   "example.com",
		}).
		Time(myDate).
		Build()

	var buffer bytes.Buffer
	json.NewEncoder(&buffer).Encode(event)
	req := httptest.NewRequest("GET", "/", &buffer)
	req.Header = http.Header{}
	req.Header.Set("Content-Type", "application/cloudevents+json")

	var actual v02.Event
	err = factory.FromRequest(req, &actual)
	require.NoError(t, err)

	builder = v02.NewCloudEventBuilder()
	expected, err := builder.
		SpecVersion("0.2").
		Type("com.example.someevent").
		ID("1234-1234-1234").
		Source(url.URL{
			Scheme: "http",
			Host:   "example.com",
			Path:   "/mycontext/subcontext",
		}).
		SchemaURL(url.URL{
			Scheme: "http",
			Host:   "example.com",
		}).
		Time(myDate).
		Build()

	assert.EqualValues(t, expected, actual)
}

func TestHTTPMarshallerToRequestJSONSuccess(t *testing.T) {
	factory := v02.NewDefaultHTTPMarshaller()

	builder := v02.NewCloudEventBuilder()
	event, err := builder.
		Type("com.example.someevent").
		ID("1234-1234-1234").
		Source(url.URL{
			Path: "/mycontext/subcontext",
		}).
		Build()

	event.Set("myint", 100)
	event.Set("myfloat", 100e+3)
	event.Set("mybool", true)
	event.Set("mystring", "string")

	actual, _ := http.NewRequest("GET", "localhost:8080", nil)
	err = factory.ToRequest(actual, event)
	require.NoError(t, err)

	buffer := bytes.Buffer{}
	json.NewEncoder(&buffer).Encode(event)
	expected, _ := http.NewRequest("GET", "localhost:8080", &buffer)
	expected.Header.Set("Content-Type", "application/cloudevents+json")

	// Can't test function equality
	expected.GetBody = nil
	actual.GetBody = nil

	assert.EqualValues(t, expected, actual)
}

func TestHTTPMarshallerFromRequestBinarySuccess(t *testing.T) {
	factory := v02.NewDefaultHTTPMarshaller()

	header := http.Header{}
	header.Set("ce-specversion", "0.2")
	header.Set("content-type", "application/json")
	header.Set("ce-type", "com.example.someevent")
	header.Set("ce-source", "/mycontext/subcontext")
	header.Set("ce-id", "1234-1234-1234")
	header.Set("ce-myextension", "myvalue")
	header.Set("ce-anotherextension", "anothervalue")
	header.Set("ce-time", "2018-04-05T03:56:24Z")

	body := bytes.NewBufferString("{\"key1\":\"value1\", \"key2\":\"value2\"}")
	req := httptest.NewRequest("GET", "localhost:8080", ioutil.NopCloser(body))
	req.Header = header

	var actual v02.Event
	err := factory.FromRequest(req, &actual)
	require.NoError(t, err)

	timestamp, err := time.Parse(time.RFC3339, "2018-04-05T03:56:24Z")
	builder := v02.NewCloudEventBuilder()
	expected, err := builder.
		ContentType("application/json").
		Type("com.example.someevent").
		Source(url.URL{
			Path: "/mycontext/subcontext",
		}).
		ID("1234-1234-1234").
		Time(timestamp).
		Data(map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
		}).
		Build()

	expected.Set("myextension", "myvalue")
	expected.Set("anotherextension", "anothervalue")

	assert.EqualValues(t, expected, actual)
}

func TestHTTPMarshallerToRequestBinarySuccess(t *testing.T) {
	factory := v02.NewDefaultHTTPMarshaller()

	builder := v02.NewCloudEventBuilder()
	event, err := builder.
		Type("com.example.someevent").
		ID("1234-1234-1234").
		Source(url.URL{
			Scheme: "http",
			Host:   "example.com",
			Path:   "/mycontext/subcontext",
		}).
		ContentType("application/json").
		Data(map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
		}).
		Build()

	event.Set("myfloat", 100e+3)
	event.Set("myint", 100)
	event.Set("mybool", true)
	event.Set("mystring", "string")

	actual, _ := http.NewRequest("GET", "localhost:8080", nil)
	err = factory.ToRequest(actual, &event)
	require.NoError(t, err)

	buffer := bytes.Buffer{}
	json.NewEncoder(&buffer).Encode(map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
	})
	expected, _ := http.NewRequest("GET", "localhost:8080", &buffer)
	expected.Header.Set("ce-specversion", "0.2")
	expected.Header.Set("ce-id", "1234-1234-1234")
	expected.Header.Set("ce-type", "com.example.someevent")
	expected.Header.Set("ce-source", "http://example.com/mycontext/subcontext")
	expected.Header.Set("ce-myfloat", "100000")
	expected.Header.Set("ce-myint", "100")
	expected.Header.Set("ce-mybool", "true")
	expected.Header.Set("ce-mystring", "string")
	expected.Header.Set("content-type", "application/json")

	// Can't test function equality
	expected.GetBody = nil
	actual.GetBody = nil

	assert.Equal(t, expected, actual)
}
