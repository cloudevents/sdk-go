package v02_test

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cloudevents/sdk-go/v02"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTTPMarshallerFromRequestBinaryBase64Success(t *testing.T) {
	factory := v02.NewDefaultHTTPMarshaller()

	header := http.Header{}
	header.Set("Content-Type", "application/octet-stream")
	header.Set("CE-Type", "com.example.someevent")
	header.Set("CE-Source", "/mycontext")
	header.Set("CE-ID", "1234-1234-1234")
	header.Set("CE-MyExtension", "myvalue")
	header.Set("CE-AnotherExtension", "anothervalue")
	header.Set("CE-Time", "2018-04-05T17:31:00Z")

	body := bytes.NewBufferString("This is a byte array of data.")
	req := httptest.NewRequest("GET", "localhost:8080", ioutil.NopCloser(body))
	req.Header = header

	actual, err := factory.FromRequest(req)
	require.NoError(t, err)

	timestamp, err := time.Parse(time.RFC3339, "2018-04-05T17:31:00Z")
	expected := &v02.Event{
		ContentType: "application/octet-stream",
		Type:        "com.example.someevent",
		Source:      "/mycontext",
		ID:          "1234-1234-1234",
		Time:        &timestamp,
		Data:        []byte("This is a byte array of data."),
	}

	expected.Set("myextension", "myvalue")
	expected.Set("anotherextension", "anothervalue")

	assert.EqualValues(t, expected, actual)
}

func TestHTTPMarshallerToRequestBinaryBase64Success(t *testing.T) {
	factory := v02.NewDefaultHTTPMarshaller()

	event := v02.Event{
		SpecVersion: "0.1",
		Type:        "com.example.someevent",
		ID:          "1234-1234-1234",
		Source:      "/mycontext",
		ContentType: "application/octet-stream",
		Data:        []byte("This is a byte array of data"),
	}

	event.Set("myfloat", 100e+3)
	event.Set("myint", 100)
	event.Set("mybool", true)
	event.Set("mystring", "string")

	actual, _ := http.NewRequest("GET", "localhost:8080", nil)
	err := factory.ToRequest(actual, &event)
	require.NoError(t, err)

	buffer := bytes.NewBufferString("This is a byte array of data")

	expected, _ := http.NewRequest("GET", "localhost:8080", buffer)
	expected.Header.Set("CE-SpecVersion", "0.1")
	expected.Header.Set("CE-ID", "1234-1234-1234")
	expected.Header.Set("CE-Type", "com.example.someevent")
	expected.Header.Set("CE-Source", "/mycontext")
	expected.Header.Set("CE-Myfloat", "100000")
	expected.Header.Set("CE-Myint", "100")
	expected.Header.Set("CE-Mybool", "true")
	expected.Header.Set("CE-Mystring", "string")
	expected.Header.Set("Content-Type", "application/octet-stream")

	// Can't test function equality
	expected.GetBody = nil
	actual.GetBody = nil

	assert.EqualValues(t, expected, actual)
}
