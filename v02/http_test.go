package v02_test

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/cloudevents/sdk-go/v02"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTTPMarshallerFromRequestBinaryBase64Success(t *testing.T) {
	factory := v02.NewDefaultHTTPMarshaller()

	header := http.Header{}
	header.Set("content-type", "application/octet-stream")
	header.Set("ce-type", "com.example.someevent")
	header.Set("ce-source", "http://example.com/mycontext")
	header.Set("ce-id", "1234-1234-1234")
	header.Set("ce-myextension", "myvalue")
	header.Set("ce-anotherextension", "anothervalue")
	header.Set("ce-time", "2018-04-05T17:31:00Z")
	header.Set("ce-specversion", "0.2")

	body := bytes.NewBufferString("This is a byte array of data.")
	req := httptest.NewRequest("GET", "localhost:8080", ioutil.NopCloser(body))
	req.Header = header

	var actual v02.Event
	err := factory.FromRequest(req, &actual)
	require.NoError(t, err)

	timestamp, err := time.Parse(time.RFC3339, "2018-04-05T17:31:00Z")
	builder := v02.NewCloudEventBuilder()
	expected, _ := builder.
		SpecVersion("0.2").
		ContentType("application/octet-stream").
		Type("com.example.someevent").
		Source(url.URL{
			Scheme: "http",
			Host:   "example.com",
			Path:   "/mycontext",
		}).
		ID("1234-1234-1234").
		Time(timestamp).
		Data([]byte("This is a byte array of data.")).
		Build()

	expected.Set("myextension", "myvalue")
	expected.Set("anotherextension", "anothervalue")

	assert.EqualValues(t, expected, actual)
}

func TestHTTPMarshallerToRequestBinaryBase64Success(t *testing.T) {
	factory := v02.NewDefaultHTTPMarshaller()

	builder := v02.NewCloudEventBuilder()
	event, _ := builder.
		SpecVersion("0.2").
		Type("com.example.someevent").
		ID("1234-1234-1234").
		Source(url.URL{
			Scheme: "http",
			Host:   "example.com",
			Path:   "/mycontext",
		}).
		ContentType("application/octet-stream").
		Data([]byte("This is a byte array of data")).
		Build()

	event.Set("myfloat", 100e+3)
	event.Set("myint", 100)
	event.Set("mybool", true)
	event.Set("mystring", "string")

	actual, _ := http.NewRequest("GET", "localhost:8080", nil)
	err := factory.ToRequest(actual, &event)
	require.NoError(t, err)

	buffer := bytes.NewBufferString("This is a byte array of data")

	expected, _ := http.NewRequest("GET", "localhost:8080", buffer)
	expected.Header.Set("ce-specversion", "0.2")
	expected.Header.Set("ce-id", "1234-1234-1234")
	expected.Header.Set("ce-type", "com.example.someevent")
	expected.Header.Set("ce-source", "http://example.com/mycontext")
	expected.Header.Set("ce-myfloat", "100000")
	expected.Header.Set("ce-myint", "100")
	expected.Header.Set("ce-mybool", "true")
	expected.Header.Set("ce-mystring", "string")
	expected.Header.Set("content-type", "application/octet-stream")

	// Can't test function equality
	expected.GetBody = nil
	actual.GetBody = nil

	assert.EqualValues(t, expected, actual)
}
