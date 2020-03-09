package buffering

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/binding/test"
	"github.com/cloudevents/sdk-go/pkg/event"
)

type copyMessageTestCase struct {
	name     string
	encoding binding.Encoding
	message  binding.Message
	want     event.Event
}

func TestCopyMessage(t *testing.T) {
	tests := []copyMessageTestCase{}

	for _, v := range test.Events() {
		tests = append(tests, []copyMessageTestCase{
			{
				name:     "From structured with payload/" + test.NameOf(v),
				encoding: binding.EncodingStructured,
				message:  test.MustCreateMockStructuredMessage(v),
				want:     v,
			},
			{
				name:     "From structured without payload/" + test.NameOf(v),
				encoding: binding.EncodingStructured,
				message:  test.MustCreateMockStructuredMessage(v),
				want:     v,
			},
			{
				name:     "From binary with payload/" + test.NameOf(v),
				encoding: binding.EncodingBinary,
				message:  test.MustCreateMockBinaryMessage(v),
				want:     v,
			},
			{
				name:     "From binary without payload/" + test.NameOf(v),
				encoding: binding.EncodingBinary,
				message:  test.MustCreateMockBinaryMessage(v),
				want:     v,
			},
			{
				name:     "From event with payload/" + test.NameOf(v),
				encoding: binding.EncodingEvent,
				message:  binding.EventMessage(v),
				want:     v,
			},
			{
				name:     "From event without payload/" + test.NameOf(v),
				encoding: binding.EncodingEvent,
				message:  binding.EventMessage(v),
				want:     v,
			},
		}...)
	}
	for _, tt := range tests {
		tt := tt // Don't use range variable in Run() scope
		t.Run(fmt.Sprintf("CopyMessage: %s", tt.name), func(t *testing.T) {
			finished := false
			message := binding.WithFinish(tt.message, func(err error) {
				finished = true
			})
			cpy, err := CopyMessage(context.Background(), message, nil)
			require.NoError(t, err)
			// The copy can be read any number of times
			for i := 0; i < 3; i++ {
				got, err := binding.ToEvent(context.Background(), cpy, nil)
				assert.NoError(t, err)
				require.Equal(t, tt.encoding, cpy.ReadEncoding())
				test.AssertEventEquals(t, test.ExToStr(t, tt.want), test.ExToStr(t, got))
			}
			require.NoError(t, cpy.Finish(nil))
			require.Equal(t, false, finished)
		})
		t.Run(fmt.Sprintf("BufferMessage: %s", tt.name), func(t *testing.T) {
			finished := false
			message := binding.WithFinish(tt.message, func(err error) {
				finished = true
			})
			cpy, err := BufferMessage(context.Background(), message, nil)
			require.NoError(t, err)
			// The copy can be read any number of times
			for i := 0; i < 3; i++ {
				got, err := binding.ToEvent(context.Background(), cpy, nil)
				assert.NoError(t, err)
				require.Equal(t, tt.encoding, cpy.ReadEncoding())
				test.AssertEventEquals(t, test.ExToStr(t, tt.want), test.ExToStr(t, got))
			}
			require.NoError(t, cpy.Finish(nil))
			require.Equal(t, true, finished)
		})
	}
}
