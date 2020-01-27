package buffering

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/binding/event"
	"github.com/cloudevents/sdk-go/pkg/binding/test"
	"github.com/cloudevents/sdk-go/pkg/cloudevents"
)

type copyMessageTestCase struct {
	name         string
	isStructured bool
	isBinary     bool
	message      binding.Message
	want         cloudevents.Event
}

func TestCopyMessage(t *testing.T) {
	tests := []copyMessageTestCase{}

	for _, v := range test.Events() {
		tests = append(tests, []copyMessageTestCase{
			{
				name:         "From structured with payload/" + test.NameOf(v),
				isStructured: true,
				isBinary:     false,
				message:      test.NewMockStructuredMessage(v),
				want:         v,
			},
			{
				name:         "From structured without payload/" + test.NameOf(v),
				isStructured: true,
				isBinary:     false,
				message:      test.NewMockStructuredMessage(v),
				want:         v,
			},
			{
				name:         "From binary with payload/" + test.NameOf(v),
				isBinary:     true,
				isStructured: false,
				message:      test.NewMockBinaryMessage(v),
				want:         v,
			},
			{
				name:         "From binary without payload/" + test.NameOf(v),
				isBinary:     true,
				isStructured: false,
				message:      test.NewMockBinaryMessage(v),
				want:         v,
			},
			{
				name:         "From event with payload/" + test.NameOf(v),
				isBinary:     true,
				isStructured: false,
				message:      event.EventMessage(v),
				want:         v,
			},
			{
				name:         "From event without payload/" + test.NameOf(v),
				isBinary:     true,
				isStructured: false,
				message:      event.EventMessage(v),
				want:         v,
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
			cpy, err := CopyMessage(message)
			require.NoError(t, err)
			// The copy can be read any number of times
			for i := 0; i < 3; i++ {
				got, isStructured, isBinary, err := event.ToEvent(cpy)
				assert.NoError(t, err)
				require.Equal(t, tt.isStructured, isStructured)
				require.Equal(t, tt.isBinary, isBinary)
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
			cpy, err := BufferMessage(message)
			require.NoError(t, err)
			// The copy can be read any number of times
			for i := 0; i < 3; i++ {
				got, isStructured, isBinary, err := event.ToEvent(cpy)
				assert.NoError(t, err)
				require.Equal(t, tt.isStructured, isStructured)
				require.Equal(t, tt.isBinary, isBinary)
				test.AssertEventEquals(t, test.ExToStr(t, tt.want), test.ExToStr(t, got))
			}
			require.NoError(t, cpy.Finish(nil))
			require.Equal(t, true, finished)
		})
	}
}
