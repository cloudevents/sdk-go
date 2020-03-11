package buffering

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cloudevents/sdk-go/pkg/binding"
	. "github.com/cloudevents/sdk-go/pkg/binding/test"
	"github.com/cloudevents/sdk-go/pkg/event"
)

type copyMessageTestCase struct {
	name     string
	encoding binding.Encoding
	message  binding.Message
	event    event.Event
	want     event.Event
}

func TestCopyMessage(t *testing.T) {
	tests := []copyMessageTestCase{}

	for _, v := range Events() {
		tests = append(tests, []copyMessageTestCase{
			{
				name:     "From structured with payload/" + NameOf(v),
				encoding: binding.EncodingStructured,
				message:  MustCreateMockStructuredMessage(v),
				want:     v,
			},
			{
				name:     "From structured without payload/" + NameOf(v),
				encoding: binding.EncodingStructured,
				message:  MustCreateMockStructuredMessage(v),
				want:     v,
			},
			{
				name:     "From binary with payload/" + NameOf(v),
				encoding: binding.EncodingBinary,
				message:  MustCreateMockBinaryMessage(v),
				want:     v,
			},
			{
				name:     "From binary without payload/" + NameOf(v),
				encoding: binding.EncodingBinary,
				message:  MustCreateMockBinaryMessage(v),
				want:     v,
			},
			{
				name:     "From event with payload/" + NameOf(v),
				encoding: binding.EncodingEvent,
				event:    v,
				want:     v,
			},
			{
				name:     "From event without payload/" + NameOf(v),
				encoding: binding.EncodingEvent,
				event:    v,
				want:     v,
			},
		}...)
	}
	for _, tt := range tests {
		tt := tt // Don't use range variable in Run() scope

		var inputMessage binding.Message
		if tt.message != nil {
			inputMessage = tt.message
		} else {
			e := tt.event.Clone()
			inputMessage = binding.ToEventMessage(&e)
		}

		t.Run(fmt.Sprintf("CopyMessage: %s", tt.name), func(t *testing.T) {
			finished := false
			message := binding.WithFinish(inputMessage, func(err error) {
				finished = true
			})
			cpy, err := CopyMessage(context.Background(), message, nil)
			require.NoError(t, err)
			// The copy can be read any number of times
			for i := 0; i < 3; i++ {
				got, err := binding.ToEvent(context.Background(), cpy, nil)
				assert.NoError(t, err)
				require.Equal(t, tt.encoding, cpy.ReadEncoding())
				AssertEventEquals(t, ExToStr(t, tt.want), ExToStr(t, *got))
			}
			require.NoError(t, cpy.Finish(nil))
			require.Equal(t, false, finished)
		})
		t.Run(fmt.Sprintf("BufferMessage: %s", tt.name), func(t *testing.T) {
			finished := false
			message := binding.WithFinish(inputMessage, func(err error) {
				finished = true
			})
			cpy, err := BufferMessage(context.Background(), message, nil)
			require.NoError(t, err)
			// The copy can be read any number of times
			for i := 0; i < 3; i++ {
				got, err := binding.ToEvent(context.Background(), cpy, nil)
				assert.NoError(t, err)
				require.Equal(t, tt.encoding, cpy.ReadEncoding())
				AssertEventEquals(t, ExToStr(t, tt.want), ExToStr(t, *got))
			}
			require.NoError(t, cpy.Finish(nil))
			require.Equal(t, true, finished)
		})
	}
}
