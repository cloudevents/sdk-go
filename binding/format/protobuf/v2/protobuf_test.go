package format_test

import (
	"net/url"
	"testing"
	stdtime "time"

	"github.com/stretchr/testify/require"

	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/cloudevents/sdk-go/v2/types"

	format "github.com/cloudevents/sdk-go/binding/format/protobuf/v2"
)

func TestProtobufFormatWithoutProtobufCodec(t *testing.T) {
	require := require.New(t)
	const test = "test"
	e := event.New()
	e.SetID(test)
	e.SetTime(stdtime.Date(2021, 1, 1, 1, 1, 1, 1, stdtime.UTC))
	e.SetExtension(test, test)
	e.SetExtension("int", 1)
	e.SetExtension("bool", true)
	e.SetExtension("URI", &url.URL{
		Host: "test-uri",
	})
	e.SetExtension("URIRef", types.URIRef{URL: url.URL{
		Host: "test-uriref",
	}})
	e.SetExtension("bytes", []byte(test))
	e.SetExtension("timestamp", stdtime.Date(2021, 2, 1, 1, 1, 1, 1, stdtime.UTC))
	e.SetSubject(test)
	e.SetSource(test)
	e.SetType(test)
	e.SetDataSchema(test)
	require.NoError(e.SetData(event.ApplicationJSON, "foo"))

	b, err := format.Protobuf.Marshal(&e)
	require.NoError(err)
	var e2 event.Event
	require.NoError(format.Protobuf.Unmarshal(b, &e2))
	require.Equal(e, e2)
}
