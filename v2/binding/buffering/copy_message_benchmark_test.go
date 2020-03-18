package buffering

import (
	"context"
	"testing"

	"github.com/cloudevents/sdk-go/v2/binding/test"
)

var err error

func BenchmarkBufferMessageFromStructured(b *testing.B) {
	e := test.FullEvent()
	input := test.MustCreateMockStructuredMessage(e)
	ctx := context.Background()
	for i := 0; i < b.N; i++ {
		outputMessage, _ := BufferMessage(ctx, input, nil)
		err = outputMessage.Finish(nil)
	}
}

func BenchmarkBufferMessageFromBinary(b *testing.B) {
	e := test.FullEvent()
	input := test.MustCreateMockBinaryMessage(e)
	ctx := context.Background()
	for i := 0; i < b.N; i++ {
		outputMessage, _ := BufferMessage(ctx, input, nil)
		err = outputMessage.Finish(nil)
	}
}
