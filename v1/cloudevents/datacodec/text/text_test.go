package text_test

import (
	"context"
	"testing"

	"github.com/cloudevents/sdk-go/v1/cloudevents/datacodec/text"
	"github.com/stretchr/testify/assert"
)

var ctx = context.Background()

func TestEncode(t *testing.T) {
	assert := assert.New(t)

	b, err := text.Encode(ctx, "")
	assert.NoError(err)
	assert.Empty(b)

	b, err = text.Encode(ctx, "hello😀")
	assert.NoError(err)
	assert.Equal("hello😀", string(b))

	_, err = text.Encode(ctx, []byte("x"))
	assert.EqualError(err, "text.Encode in: want string, got []uint8")
	_, err = text.Encode(ctx, nil)
	assert.EqualError(err, "text.Encode in: want string, got <nil>")
}

func TestDecode(t *testing.T) {
	assert := assert.New(t)
	var s string
	assert.NoError(text.Decode(ctx, "hello", &s))
	assert.Equal("hello", s)
	assert.NoError(text.Decode(ctx, []byte("bye"), &s))
	assert.Equal("bye", s)
	assert.NoError(text.Decode(ctx, []byte{}, &s))
	assert.Equal("", s)
	s = "xxx"
	assert.NoError(text.Decode(ctx, nil, &s))
	assert.Equal("", s)

	assert.EqualError(text.Decode(ctx, 123, &s), "text.Decode in: want []byte or string, got int")
	assert.EqualError(text.Decode(ctx, "", nil), "text.Decode out: want *string, got <nil>")
	assert.EqualError(text.Decode(ctx, "", 1), "text.Decode out: want *string, got int")
}
