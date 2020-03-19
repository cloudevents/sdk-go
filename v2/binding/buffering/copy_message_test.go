package buffering

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cloudevents/sdk-go/v2/binding"
	. "github.com/cloudevents/sdk-go/v2/binding/test"
	"github.com/cloudevents/sdk-go/v2/event"
)

type copyMessageTestCase struct {
	name     string
	encoding binding.Encoding
	message  binding.Message
	event    event.Event
	want     event.Event
}

func TestCopyMessage_success(t *testing.T) {
	EachEvent(t, Events(), func(t *testing.T, v event.Event) {
		tests := []copyMessageTestCase{
			{
				name:     "from structured",
				encoding: binding.EncodingStructured,
				message:  MustCreateMockStructuredMessage(v),
				want:     v,
			},
			{
				name:     "from binary",
				encoding: binding.EncodingBinary,
				message:  MustCreateMockBinaryMessage(v),
				want:     v,
			},
			{
				name:     "from event",
				encoding: binding.EncodingEvent,
				event:    v,
				want:     v,
			},
		}
		for _, tt := range tests {
			tt := tt // Don't use range variable in Run() scope

			var inputMessage binding.Message
			if tt.message != nil {
				inputMessage = tt.message
			} else {
				e := tt.event.Clone()
				inputMessage = binding.ToMessage(&e)
			}

			t.Run(fmt.Sprintf("CopyMessage %s", tt.name), func(t *testing.T) {
				finished := false
				message := binding.WithFinish(inputMessage, func(err error) {
					finished = true
				})
				cpy, err := CopyMessage(context.Background(), message)
				require.NoError(t, err)
				// The copy can be read any number of times
				for i := 0; i < 3; i++ {
					got, err := binding.ToEvent(context.Background(), cpy)
					assert.NoError(t, err)
					require.Equal(t, tt.encoding, cpy.ReadEncoding())
					AssertEventEquals(t, ExToStr(t, tt.want), ExToStr(t, *got))
				}
				require.NoError(t, cpy.Finish(nil))
				require.Equal(t, false, finished)
			})
			t.Run(fmt.Sprintf("BufferMessage %s", tt.name), func(t *testing.T) {
				finished := false
				message := binding.WithFinish(inputMessage, func(err error) {
					finished = true
				})
				cpy, err := BufferMessage(context.Background(), message)
				require.NoError(t, err)
				// The copy can be read any number of times
				for i := 0; i < 3; i++ {
					got, err := binding.ToEvent(context.Background(), cpy)
					assert.NoError(t, err)
					require.Equal(t, tt.encoding, cpy.ReadEncoding())
					AssertEventEquals(t, ExToStr(t, tt.want), ExToStr(t, *got))
				}
				require.NoError(t, cpy.Finish(nil))
				require.Equal(t, true, finished)
			})
		}
	})
}

func TestCopyMessage_unknown(t *testing.T) {
	cpy, err := BufferMessage(context.Background(), UnknownMessage)
	require.Nil(t, cpy)
	require.Equal(t, binding.ErrUnknownEncoding, err)
}

func TestCopyMessage_transformers_applied_once(t *testing.T) {
	EachEvent(t, Events(), func(t *testing.T, v event.Event) {
		tests := []copyMessageTestCase{
			{
				name:    "From structured",
				message: MustCreateMockStructuredMessage(v),
				want:    v,
			},
			{
				name:    "From binary",
				message: MustCreateMockBinaryMessage(v),
				want:    v,
			},
			{
				name:  "From event",
				event: v,
				want:  v,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name+" with structured Transformer", func(t *testing.T) {
				testCopyMessageWithTransformer(t, tt, NewMockTransformerFactory(false, false))
			})
			t.Run(tt.name+" with binary Transformer", func(t *testing.T) {
				testCopyMessageWithTransformer(t, tt, NewMockTransformerFactory(true, false))
			})
			t.Run(tt.name+" with event Transformer", func(t *testing.T) {
				testCopyMessageWithTransformer(t, tt, NewMockTransformerFactory(true, true))
			})
			t.Run(tt.name+" with mixed Transformers", func(t *testing.T) {
				var inputMessage binding.Message
				if tt.message != nil {
					inputMessage = tt.message
				} else {
					e := tt.event.Clone()
					inputMessage = binding.ToMessage(&e)
				}

				transformerBinary := NewMockTransformerFactory(true, false)
				transformerEvent := NewMockTransformerFactory(true, true)

				cpy, err := CopyMessage(context.Background(), inputMessage, transformerBinary, transformerEvent)
				require.NoError(t, err)
				require.NotNil(t, cpy)
				got, err := binding.ToEvent(context.Background(), cpy)
				assert.NoError(t, err)
				AssertEventEquals(t, ExToStr(t, tt.want), ExToStr(t, *got))

				AssertTransformerInvokedOneTime(t, transformerBinary)
				require.Equal(t, 1, transformerBinary.InvokedEventTransformer)
				AssertTransformerInvokedOneTime(t, transformerEvent)
			})
		}
	})
}

func testCopyMessageWithTransformer(t *testing.T, tt copyMessageTestCase, transformer *MockTransformerFactory) {
	var inputMessage binding.Message
	if tt.message != nil {
		inputMessage = tt.message
	} else {
		e := tt.event.Clone()
		inputMessage = binding.ToMessage(&e)
	}

	cpy, err := CopyMessage(context.Background(), inputMessage, transformer)
	require.NoError(t, err)
	require.NotNil(t, cpy)
	got, err := binding.ToEvent(context.Background(), cpy)
	assert.NoError(t, err)
	AssertEventEquals(t, ExToStr(t, tt.want), ExToStr(t, *got))

	AssertTransformerInvokedOneTime(t, transformer)
}
