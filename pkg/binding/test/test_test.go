package test_test

import (
	"context"
	"testing"

	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/binding/test"
	ce "github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/stretchr/testify/assert"
)

func TestEvent(t *testing.T) {
	assert := assert.New(t)

	e := test.FullEvent()
	assert.Equal("1.0", e.SpecVersion())
	assert.Equal("com.example.FullEvent", e.Type())
	assert.Equal(true, e.DataEncoded)
	var s string
	err := e.DataAs(&s)
	assert.NoError(err)
	assert.Equal("hello", s)

	e = test.MinEvent()
	assert.Equal("1.0", e.SpecVersion())
	assert.Equal("com.example.MinEvent", e.Type())
	assert.Nil(e.Data)
	assert.Empty(e.DataContentType())
}

func TestEncodeDecodeBinary(t *testing.T) {
	test.EachEvent(t, test.Events(), func(t *testing.T, in ce.Event) {
		assert.Equal(t, in, test.EncodeDecode(t, in, binding.BinaryEncoder{}))
	})
}

func TestEncodeDecodeStruct(t *testing.T) {
	test.EachEvent(t, test.Events(), func(t *testing.T, in ce.Event) {
		assert.Equal(t, in, test.EncodeDecode(t, in, binding.BinaryEncoder{}))
	})
}

type dummySR chan binding.Message

func (d dummySR) Send(ctx context.Context, m binding.Message) (err error) { d <- m; return nil }
func (d dummySR) Receive(ctx context.Context) (binding.Message, error)    { return <-d, nil }

func TestSendReceive(t *testing.T) {
	sr := make(dummySR)
	allIn := test.EncodeAll(t, test.Events(), binding.BinaryEncoder{})
	allOut := []binding.Message{}
	test.EachMessage(t, allIn, func(t *testing.T, in binding.Message) {
		out := test.SendReceive(t, in, sr, sr)
		assert.Equal(t, in, out)
		allOut = append(allOut, out)
	})
	assert.Equal(t, allIn, allOut)
}
