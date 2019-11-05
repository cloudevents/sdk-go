package http_test

import (
	nethttp "net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/binding/format"
	"github.com/cloudevents/sdk-go/pkg/binding/test"
	"github.com/cloudevents/sdk-go/pkg/bindings/http"
	ce "github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testEvent = test.FullEvent()

	testHTTPBinary = &http.Message{
		Body: testEvent.Data.([]byte),
		Header: nethttp.Header{
			http.ContentType: {testEvent.DataContentType()},
			"Ce-Specversion": {"1.0"},
			"Ce-Type":        {testEvent.Type()},
			"Ce-Source":      {testEvent.Source()},
			"Ce-Id":          {testEvent.ID()},
			"Ce-Time":        {types.FormatTime(testEvent.Time())},
			"Ce-Dataschema":  {testEvent.DataSchema()},
			"Ce-Subject":     {testEvent.Subject()},

			"Ce-Exbool":   {types.FormatBool(true)},
			"Ce-Exint":    {types.FormatInteger(42)},
			"Ce-Exstring": {"exstring"},
			"Ce-Exbinary": {types.FormatBinary([]byte{0, 1, 2, 3})},
			"Ce-Extime":   {types.FormatTime(testEvent.Time())},
			"Ce-Exurl":    {testEvent.Source()},
		},
	}

	testHTTPStruct = &http.Message{
		Body:   test.MustJSON(exToStr(testEvent)),
		Header: nethttp.Header{http.ContentType: {format.JSON.MediaType()}},
	}
)

// Convert all extension attribute values to string form.
func exToStr(e ce.Event) ce.Event {
	for k, v := range e.Extensions() {
		s, _ := types.Format(v)
		e.SetExtension(k, s)
	}
	return e
}

func TestBinary(t *testing.T) {
	enc := http.BinaryEncoder{}
	// Encode
	m, err := enc.Encode(testEvent)
	assert.NoError(t, err)
	assert.Equal(t, testHTTPBinary, m.(*http.Message))
	// Decode
	e, err := m.Event()
	assert.NoError(t, err)
	assert.Equal(t, exToStr(testEvent), e)
	test.EachEvent(t, test.Events(), func(t *testing.T, in ce.Event) {
		assert.Equal(t, exToStr(in), test.EncodeDecode(t, in, enc))
	})
}

func TestStruct(t *testing.T) {
	enc := http.StructEncoder{Format: format.JSON}
	m, err := enc.Encode(testEvent)
	assert.NoError(t, err)
	assert.Equal(t, string(testHTTPStruct.Body), string(m.(*http.Message).Body))
	assert.Equal(t, testHTTPStruct, m.(*http.Message))
	// TODO(alanconway)
	t.Skip("JSON does not round-trip extension attributes")
	test.EachEvent(t, test.Events(), func(t *testing.T, in ce.Event) {
		assert.Equal(t, in, test.EncodeDecode(t, in, enc))
	})
}

func TestSendReceiveBinary(t *testing.T) {
	close, s, r := testSenderReceiver(t)
	defer close()
	test.EachEvent(t, test.Events(), func(t *testing.T, e ce.Event) {
		in := binding.EventMessage(exToStr(e))
		out := test.SendReceive(t, in, s, r)
		test.AssertMessageEqual(t, in, out)
	})
}

func TestSendReceiveStruct(t *testing.T) {
	close, s, r := testSenderReceiver(t)
	defer close()
	test.EachEvent(t, test.Events(), func(t *testing.T, e ce.Event) {
		in, err := binding.StructEncoder{Format: format.JSON}.Encode(e)
		require.NoError(t, err)
		out := test.SendReceive(t, in, s, r)
		test.AssertMessageEqual(t, in, out)
	})
}

func testSenderReceiver(t testing.TB) (func(), binding.Sender, binding.Receiver) {
	r := http.NewReceiver() // Parameters? Capacity, sync.
	srv := httptest.NewServer(r)
	u, err := url.Parse(srv.URL)
	require.NoError(t, err)
	s := http.NewSender(&nethttp.Client{}, u) // Capacity, sync etc.
	return func() { srv.Close() }, s, r
}

func BenchmarkSendReceive(b *testing.B) {
	c, s, r := testSenderReceiver(b)
	defer c() // Cleanup
	test.BenchmarkSendReceive(b, s, r)
}
