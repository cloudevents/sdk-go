package text_test

import (
	"testing"

	"github.com/cloudevents/sdk-go/v2/event/datacodec/text"
	"github.com/stretchr/testify/assert"
)

func TestEncodeObserved(t *testing.T) {
	assert := assert.New(t)

	b, err := text.EncodeObserved(ctx, "")
	assert.NoError(err)
	assert.Empty(b)

	b, err = text.EncodeObserved(ctx, "hello😀")
	assert.NoError(err)
	assert.Equal("hello😀", string(b))

	_, err = text.EncodeObserved(ctx, []byte("x"))
	assert.EqualError(err, "text.Encode in: want string, got []uint8")
	_, err = text.EncodeObserved(ctx, nil)
	assert.EqualError(err, "text.Encode in: want string, got <nil>")
}

func TestDecodeObserved(t *testing.T) {
	assert := assert.New(t)
	var s string
	assert.NoError(text.DecodeObserved(ctx, []byte("hello"), &s))
	assert.Equal("hello", s)
	assert.NoError(text.DecodeObserved(ctx, []byte("bye"), &s))
	assert.Equal("bye", s)
	assert.NoError(text.DecodeObserved(ctx, []byte{}, &s))
	assert.Equal("", s)
	s = "xxx"
	assert.NoError(text.DecodeObserved(ctx, nil, &s))
	assert.Equal("", s)
}
