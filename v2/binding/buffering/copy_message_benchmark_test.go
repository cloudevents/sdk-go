package buffering

import (
	"context"
	"testing"

	. "github.com/cloudevents/sdk-go/v2/binding/test"
	. "github.com/cloudevents/sdk-go/v2/test"
)

var err error

func BenchmarkBufferMessageFromStructured(b *testing.B) {
	e := FullEvent()
	input := MustCreateMockStructuredMessage(b, e)
	ctx := context.Background()
	for i := 0; i < b.N; i++ {
		outputMessage, _ := BufferMessage(ctx, input)
		err = outputMessage.Finish(nil)
	}
}

func BenchmarkBufferMessageFromBinary(b *testing.B) {
	e := FullEvent()
	input := MustCreateMockBinaryMessage(e)
	ctx := context.Background()
	for i := 0; i < b.N; i++ {
		outputMessage, _ := BufferMessage(ctx, input)
		err = outputMessage.Finish(nil)
	}
}
