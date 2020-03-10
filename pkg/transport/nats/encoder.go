package nats

import (
	"context"
	"io"
	"io/ioutil"

	"github.com/nats-io/nats.go"

	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/binding/format"
)

// Fill the provided amqpMessage with the message m.
// Using context you can tweak the encoding processing (more details on binding.Write documentation).
func WriteNATSMessage(ctx context.Context, m binding.Message, natsMessage *nats.Msg, transformers binding.TransformerFactories) error {
	structuredWriter := (*natsMessageWriter)(natsMessage)

	_, err := binding.Write(
		ctx,
		m,
		structuredWriter,
		nil,
		transformers,
	)
	return err
}

type natsMessageWriter nats.Msg

func (b *natsMessageWriter) SetStructuredEvent(ctx context.Context, f format.Format, event io.Reader) error {
	val, err := ioutil.ReadAll(event)
	if err != nil {
		return err
	}
	b.Data = val
	return nil
}

func (b *natsMessageWriter) Start(ctx context.Context) error {
	return nil
}

func (b *natsMessageWriter) End() error {
	return nil
}

var _ binding.StructuredWriter = (*natsMessageWriter)(nil) // Test it conforms to the interface
