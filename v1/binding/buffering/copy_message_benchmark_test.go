package buffering

import (
	"context"
	"testing"

	"github.com/cloudevents/sdk-go/v1/binding/test"
)

var err error

func BenchmarkBufferMessageFromStructured(b *testing.B) {
	e := test.FullEvent()
	input := test.NewMockStructuredMessage(e)
	ctx := context.Background()
	for i := 0; i < b.N; i++ {
		outputMessage, _ := BufferMessage(ctx, input)
		err = outputMessage.Finish(nil)
	}
}

func BenchmarkBufferMessageFromBinary(b *testing.B) {
	e := test.FullEvent()
	input := test.NewMockBinaryMessage(e)
	ctx := context.Background()
	for i := 0; i < b.N; i++ {
		outputMessage, _ := BufferMessage(ctx, input)
		err = outputMessage.Finish(nil)
	}
}
