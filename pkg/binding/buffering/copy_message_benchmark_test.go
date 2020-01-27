package buffering

import (
	"testing"

	"github.com/cloudevents/sdk-go/pkg/binding/test"
)

var err error

func BenchmarkBufferMessageFromStructured(b *testing.B) {
	e := test.FullEvent()
	input := test.NewMockStructuredMessage(e)
	for i := 0; i < b.N; i++ {
		outputMessage, _ := BufferMessage(input)
		err = outputMessage.Finish(nil)
	}
}

func BenchmarkBufferMessageFromBinary(b *testing.B) {
	e := test.FullEvent()
	input := test.NewMockBinaryMessage(e)
	for i := 0; i < b.N; i++ {
		outputMessage, _ := BufferMessage(input)
		err = outputMessage.Finish(nil)
	}
}
