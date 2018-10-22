package v01_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	cloudevents "github.com/cloudevents/sdk-go"
	mocks "github.com/cloudevents/sdk-go/mocks"
	"github.com/cloudevents/sdk-go/v01"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestFromRequestNilRequest(t *testing.T) {
	principal := v01.NewDefaultHTTPMarshaller()

	event, err := principal.FromRequest(nil)

	assert.Nil(t, event)
	require.Error(t, err)
	assert.Equal(t, cloudevents.IllegalArgumentError("req"), err)
}

func TestFromRequestNoContentType(t *testing.T) {
	principal := v01.NewDefaultHTTPMarshaller()

	req := httptest.NewRequest("GET", "localhost:8080", nil)

	event, err := principal.FromRequest(req)
	assert.Nil(t, event)
	assert.Error(t, err)
}

func TestFromRequestInvalidContentType(t *testing.T) {
	principal := v01.NewDefaultHTTPMarshaller()

	req := httptest.NewRequest("GET", "localhost:8080", nil)
	req.Header.Set("Content-Type", "application/")

	event, err := principal.FromRequest(req)
	assert.Nil(t, event)
	assert.Error(t, err)
}

func TestFromRequestNoConverters(t *testing.T) {
	principal := v01.NewHTTPMarshaller()

	req := httptest.NewRequest("GET", "localhost:8080", nil)
	req.Header.Set("Content-Type", "application/cloudevents+json")

	event, err := principal.FromRequest(req)

	assert.Nil(t, event)
	require.Error(t, err)
	assert.Equal(t, cloudevents.ContentTypeNotSupportedError("application/cloudevents+json"), err)
}

func TestFromRequestWrongConverter(t *testing.T) {
	jsonConverter := &mocks.HTTPCloudEventConverter{}
	jsonConverter.On("CanRead", reflect.TypeOf(v01.Event{}), "application/json").Return(false)
	principal := v01.NewHTTPMarshaller(jsonConverter)

	req := httptest.NewRequest("GET", "localhost:8080", nil)
	req.Header.Set("Content-Type", "application/json")

	event, err := principal.FromRequest(req)

	assert.Nil(t, event)
	jsonConverter.AssertNumberOfCalls(t, "CanRead", 1)
	jsonConverter.AssertNotCalled(t, "Read")
	require.Error(t, err)
	assert.Equal(t, cloudevents.ContentTypeNotSupportedError("application/json"), err)
}

func TestFromRequestConverterError(t *testing.T) {
	jsonConverter := &mocks.HTTPCloudEventConverter{}
	jsonConverter.On("CanRead", reflect.TypeOf(v01.Event{}), "application/json").Return(true)
	jsonConverter.On("Read", reflect.TypeOf(v01.Event{}), mock.AnythingOfType(reflect.TypeOf((*http.Request)(nil)).String())).Return(nil, errors.New("read error"))
	principal := v01.NewHTTPMarshaller(jsonConverter)

	req := httptest.NewRequest("GET", "localhost:8080", nil)
	req.Header.Set("Content-Type", "application/json")

	event, err := principal.FromRequest(req)

	assert.Nil(t, event)
	jsonConverter.AssertNumberOfCalls(t, "CanRead", 1)
	jsonConverter.AssertNumberOfCalls(t, "Read", 1)
	require.Error(t, err)
	assert.Equal(t, errors.New("read error"), err)
}

func TestFromRequestSuccess(t *testing.T) {
	expected := &v01.Event{
		EventType:        "com.example.someevent",
		EventID:          "00001",
		EventTypeVersion: "0.1",
	}

	jsonConverter := &mocks.HTTPCloudEventConverter{}
	jsonConverter.On("CanRead", reflect.TypeOf(v01.Event{}), "application/json").Return(false)
	binaryConverter := &mocks.HTTPCloudEventConverter{}
	binaryConverter.On("CanRead", reflect.TypeOf(v01.Event{}), "application/json").Return(true)
	binaryConverter.On("Read", reflect.TypeOf(v01.Event{}), mock.AnythingOfType(reflect.TypeOf((*http.Request)(nil)).String())).Return(expected, nil)

	principal := v01.NewHTTPMarshaller(jsonConverter, binaryConverter)

	req := &http.Request{
		Header: map[string][]string{},
	}
	req.Header.Set("Content-Type", "application/json")

	actual, err := principal.FromRequest(req)

	require.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestToRequestNilRequest(t *testing.T) {
	principal := v01.NewHTTPMarshaller()

	event := &v01.Event{}
	err := principal.ToRequest(nil, event)

	assert.Error(t, err)
	assert.Equal(t, cloudevents.IllegalArgumentError("req"), err)
}

func TestToRequestNilEvent(t *testing.T) {
	principal := v01.NewHTTPMarshaller()

	req := httptest.NewRequest("GET", "localhost:8080", nil)
	err := principal.ToRequest(req, nil)

	assert.Error(t, err)
	assert.Equal(t, cloudevents.IllegalArgumentError("event"), err)
}

func TestToRequestDefaultContentType(t *testing.T) {
	jsonConverter := &mocks.HTTPCloudEventConverter{}
	jsonConverter.On("CanWrite", reflect.TypeOf(v01.Event{}), "application/cloudevents+json").Return(true)
	jsonConverter.On("Write", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	principal := v01.NewHTTPMarshaller(jsonConverter)

	req := httptest.NewRequest("GET", "localhost:8080", nil)
	event := &v01.Event{}
	err := principal.ToRequest(req, event)

	assert.NoError(t, err)
	jsonConverter.AssertCalled(t, "CanWrite", reflect.TypeOf(v01.Event{}), "application/cloudevents+json")
}

func TestToRequestInvalidContentType(t *testing.T) {
	principal := v01.NewDefaultHTTPMarshaller()

	req := httptest.NewRequest("GET", "localhost:8080", nil)
	event := &v01.Event{
		ContentType: "application/",
	}

	err := principal.ToRequest(req, event)

	assert.Error(t, err)
}

func TestToRequestNoConverters(t *testing.T) {
	principal := v01.NewHTTPMarshaller()

	req := httptest.NewRequest("GET", "localhost:8080", nil)
	event := &v01.Event{}
	err := principal.ToRequest(req, event)

	require.Error(t, err)
	assert.Equal(t, cloudevents.ContentTypeNotSupportedError("application/cloudevents+json"), err)
}

func TestToRequestWriteError(t *testing.T) {
	jsonConverter := &mocks.HTTPCloudEventConverter{}
	jsonConverter.On("CanWrite", reflect.TypeOf(v01.Event{}), "application/cloudevents+json").Return(true)
	jsonConverter.On("Write", reflect.TypeOf(v01.Event{}), mock.AnythingOfType(reflect.TypeOf((*http.Request)(nil)).String()), &v01.Event{}).Return(errors.New("write error"))

	principal := v01.NewHTTPMarshaller(jsonConverter)

	req := httptest.NewRequest("GET", "localhost:8080", nil)
	event := &v01.Event{}

	err := principal.ToRequest(req, event)

	assert.Error(t, err)
}

func TestToRequestWrongConverter(t *testing.T) {
	binaryConverter := &mocks.HTTPCloudEventConverter{}
	binaryConverter.On("CanWrite", reflect.TypeOf(v01.Event{}), "application/cloudevents+json").Return(false)

	principal := v01.NewHTTPMarshaller(binaryConverter)

	req := httptest.NewRequest("GET", "localhost:8080", nil)
	event := &v01.Event{}

	err := principal.ToRequest(req, event)

	require.Error(t, err)
	assert.Equal(t, cloudevents.ContentTypeNotSupportedError("application/cloudevents+json"), err)
}

func TestJSONCanReadCanWriteBothWrong(t *testing.T) {
	principal := v01.NewJSONHTTPCloudEventConverter()

	actual := principal.CanRead(reflect.TypeOf((*cloudevents.Event)(nil)), "application/json")
	assert.Equal(t, false, actual)

	actual = principal.CanWrite(reflect.TypeOf((*cloudevents.Event)(nil)), "application/json")
	assert.Equal(t, false, actual)
}
func TestJSONConverterCanReadCanWriteWrongType(t *testing.T) {
	principal := v01.NewJSONHTTPCloudEventConverter()

	actual := principal.CanRead(reflect.TypeOf((*cloudevents.Event)(nil)), "application/cloudevents+json")
	assert.Equal(t, false, actual)

	actual = principal.CanWrite(reflect.TypeOf((*cloudevents.Event)(nil)), "application/cloudevents+json")
	assert.Equal(t, false, actual)
}

func TestJSONConverterCanReadCanWriteWrongMediaType(t *testing.T) {
	principal := v01.NewJSONHTTPCloudEventConverter()

	actual := principal.CanRead(reflect.TypeOf(v01.Event{}), "application/json")
	assert.Equal(t, false, actual)

	actual = principal.CanWrite(reflect.TypeOf(v01.Event{}), "application/json")
	assert.Equal(t, false, actual)
}

func TestJSONConverterCanReadCanWriteSuccess(t *testing.T) {
	principal := v01.NewJSONHTTPCloudEventConverter()

	actual := principal.CanRead(reflect.TypeOf(v01.Event{}), "application/cloudevents+json")
	assert.Equal(t, true, actual)

	actual = principal.CanWrite(reflect.TypeOf(v01.Event{}), "application/cloudevents+json")
	assert.Equal(t, true, actual)
}

func TestJSONConverterReadError(t *testing.T) {
	principal := v01.NewJSONHTTPCloudEventConverter()

	req := httptest.NewRequest("GET", "locahost:8080", nil)
	actual, err := principal.Read(reflect.TypeOf(v01.Event{}), req)

	assert.Error(t, err)
	assert.Nil(t, actual)
}

func TestJSONConverterReadSuccess(t *testing.T) {
	principal := v01.NewJSONHTTPCloudEventConverter()

	body := bytes.NewBufferString("{\"eventType\":\"com.example.someevent\"}")
	req := httptest.NewRequest("GET", "localhost:8080", body)

	actual, err := principal.Read(reflect.TypeOf(v01.Event{}), req)

	require.NoError(t, err)

	expected := &v01.Event{
		EventType: "com.example.someevent",
	}

	assert.Equal(t, expected, actual)
}

func TestJSONConverterWriteError(t *testing.T) {
	principal := v01.NewJSONHTTPCloudEventConverter()

	req := httptest.NewRequest("GET", "localhost:8080", nil)
	event := &v01.Event{}

	err := principal.Write(reflect.TypeOf(v01.Event{}), req, event)

	assert.NoError(t, err)
}

func TestFromRequestJSONSuccess(t *testing.T) {
	factory := v01.NewDefaultHTTPMarshaller()

	myDate, err := time.Parse(time.RFC3339, "2018-04-05T03:56:24Z")
	if err != nil {
		t.Error(err.Error())
	}
	event := v01.Event{
		CloudEventsVersion: "0.1",
		EventType:          "com.example.someevent",
		EventTypeVersion:   "1.1",
		EventID:            "1234-1234-1234",
		Source:             "/mycontext/subcontext",
		SchemaURL:          "http://example.com",
		EventTime:          &myDate,
	}

	var buffer bytes.Buffer
	json.NewEncoder(&buffer).Encode(event)
	req := httptest.NewRequest("GET", "/", &buffer)
	req.Header = http.Header{}
	req.Header.Set("Content-Type", "application/cloudevents+json")

	actual, err := factory.FromRequest(req)
	require.NoError(t, err)

	expected := &v01.Event{
		CloudEventsVersion: "0.1",
		EventType:          "com.example.someevent",
		EventTypeVersion:   "1.1",
		EventID:            "1234-1234-1234",
		Source:             "/mycontext/subcontext",
		SchemaURL:          "http://example.com",
		EventTime:          &myDate,
	}

	assert.EqualValues(t, expected, actual)
}

func TestHTTPMarshallerToRequestJSONSuccess(t *testing.T) {
	factory := v01.NewDefaultHTTPMarshaller()

	event := v01.Event{
		EventType:        "com.example.someevent",
		EventTypeVersion: "1.1",
		EventID:          "1234-1234-1234",
		Source:           "/mycontext/subcontext",
	}
	event.Set("myint", 100)
	event.Set("myfloat", 100e+3)
	event.Set("mybool", true)
	event.Set("mystring", "string")

	actual, _ := http.NewRequest("GET", "localhost:8080", nil)
	err := factory.ToRequest(actual, &event)
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
	factory := v01.NewDefaultHTTPMarshaller()

	header := http.Header{}
	header.Set("Content-Type", "application/json")
	header.Set("CE-EventType", "com.example.someevent")
	header.Set("CE-Source", "/mycontext/subcontext")
	header.Set("CE-EventID", "1234-1234-1234")
	header.Set("CE-MyExtension", "myvalue")
	header.Set("CE-AnotherExtension", "anothervalue")
	header.Set("CE-EventTime", "2018-04-05T03:56:24Z")

	body := bytes.NewBufferString("{\"key1\":\"value1\", \"key2\":\"value2\"}")
	req := httptest.NewRequest("GET", "localhost:8080", ioutil.NopCloser(body))
	req.Header = header

	actual, err := factory.FromRequest(req)
	require.NoError(t, err)

	timestamp, err := time.Parse(time.RFC3339, "2018-04-05T03:56:24Z")
	expected := &v01.Event{
		ContentType: "application/json",
		EventType:   "com.example.someevent",
		Source:      "/mycontext/subcontext",
		EventID:     "1234-1234-1234",
		EventTime:   &timestamp,
		Data: map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
		},
	}

	expected.Set("myextension", "myvalue")
	expected.Set("anotherextension", "anothervalue")

	assert.EqualValues(t, expected, actual)
}

func TestHTTPMarshallerToRequestBinarySuccess(t *testing.T) {
	factory := v01.NewDefaultHTTPMarshaller()

	event := v01.Event{
		EventType:        "com.example.someevent",
		EventTypeVersion: "1.1",
		EventID:          "1234-1234-1234",
		Source:           "/mycontext/subcontext",
		ContentType:      "application/json",
		Data: map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
		},
	}

	event.Set("myfloat", 100e+3)
	event.Set("myint", 100)
	event.Set("mybool", true)
	event.Set("mystring", "string")

	actual, _ := http.NewRequest("GET", "localhost:8080", nil)
	err := factory.ToRequest(actual, &event)
	require.NoError(t, err)

	buffer := bytes.Buffer{}
	json.NewEncoder(&buffer).Encode(map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
	})
	expected, _ := http.NewRequest("GET", "localhost:8080", &buffer)
	expected.Header.Set("CE-EventID", "1234-1234-1234")
	expected.Header.Set("CE-EventType", "com.example.someevent")
	expected.Header.Set("CE-EventTypeVersion", "1.1")
	expected.Header.Set("CE-Source", "/mycontext/subcontext")
	expected.Header.Set("CE-Myfloat", "100000")
	expected.Header.Set("CE-Myint", "100")
	expected.Header.Set("CE-Mybool", "true")
	expected.Header.Set("CE-Mystring", "string")
	expected.Header.Set("Content-Type", "application/json")

	// Can't test function equality
	expected.GetBody = nil
	actual.GetBody = nil

	assert.Equal(t, expected, actual)
}
