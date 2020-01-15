package binding_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/binding/test"
	"github.com/cloudevents/sdk-go/pkg/cloudevents"
)

type toEventTestCase struct {
	name         string
	isStructured bool
	isBinary     bool
	message      binding.Message
	want         cloudevents.Event
}

func TestToEvent(t *testing.T) {
	tests := []toEventTestCase{}

	for _, v := range test.Events() {
		tests = append(tests, []toEventTestCase{
			{
				name:         "From structured with payload/" + test.NameOf(v),
				isStructured: true,
				isBinary:     false,
				message:      binding.NewMockStructuredMessage(v),
				want:         v,
			},
			{
				name:         "From structured without payload/" + test.NameOf(v),
				isStructured: true,
				isBinary:     false,
				message:      binding.NewMockStructuredMessage(v),
				want:         v,
			},
			{
				name:         "From binary with payload/" + test.NameOf(v),
				isBinary:     true,
				isStructured: false,
				message:      binding.NewMockBinaryMessage(v),
				want:         v,
			},
			{
				name:         "From binary without payload/" + test.NameOf(v),
				isBinary:     true,
				isStructured: false,
				message:      binding.NewMockBinaryMessage(v),
				want:         v,
			},
			{
				name:         "From event with payload/" + test.NameOf(v),
				isBinary:     false,
				isStructured: false,
				message:      binding.EventMessage(v),
				want:         v,
			},
			{
				name:         "From event without payload/" + test.NameOf(v),
				isBinary:     false,
				isStructured: false,
				message:      binding.EventMessage(v),
				want:         v,
			},
		}...)
	}
	for _, tt := range tests {
		tt := tt // Don't use range variable in Run() scope
		t.Run(tt.name, func(t *testing.T) {
			got, isStructured, isBinary, err := binding.ToEvent(tt.message)
			require.NoError(t, err)
			require.Equal(t, tt.isStructured, isStructured)
			require.Equal(t, tt.isBinary, isBinary)
			test.AssertEventEquals(t, test.ExToStr(t, tt.want), test.ExToStr(t, got))
		})
	}
}
