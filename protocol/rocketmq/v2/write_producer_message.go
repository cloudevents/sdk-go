package rocketmq

import (
	"bytes"
	"context"
	"io"

	"github.com/apache/rocketmq-client-go/v2/primitive"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/binding/format"
)

// WriteProducerMessage fills the provided message with the message m.
// Using context you can tweak the encoding processing (more details on binding.Write documentation).
func WriteProducerMessage(ctx context.Context, m binding.Message, message *primitive.Message, transformers ...binding.Transformer) error {
	writer := (*rocketmqMessageWriter)(message)

	_, err := binding.Write(
		ctx,
		m,
		writer,
		nil,
		transformers...,
	)
	return err
}

type rocketmqMessageWriter primitive.Message

func (w *rocketmqMessageWriter) SetStructuredEvent(ctx context.Context, format format.Format, event io.Reader) error {
	var buf bytes.Buffer
	_, err := io.Copy(&buf, event)
	if err != nil {
		return err
	}

	w.body = string(buf.Bytes())
	return nil
}

var _ binding.StructuredWriter = (*rocketmqMessageWriter)(nil) // Test it conforms to the interface
