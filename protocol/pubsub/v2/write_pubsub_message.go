package pubsub

import (
	"bytes"
	"context"
	"io"

	"cloud.google.com/go/pubsub"
	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/binding/format"
	"github.com/cloudevents/sdk-go/v2/binding/spec"
	"github.com/cloudevents/sdk-go/v2/types"
)

// WritePubSubMessage fills the provided pubsubMessage with the message m.
// Using context you can tweak the encoding processing (more details on binding.Write documentation).
func WritePubSubMessage(ctx context.Context, m binding.Message, pubsubMessage *pubsub.Message, transformers ...binding.Transformer) error {
	structuredWriter := (*pubsubMessagePublisher)(pubsubMessage)
	binaryWriter := (*pubsubMessagePublisher)(pubsubMessage)

	_, err := binding.Write(
		ctx,
		m,
		structuredWriter,
		binaryWriter,
		transformers...,
	)
	return err
}

type pubsubMessagePublisher pubsub.Message

func (b *pubsubMessagePublisher) SetStructuredEvent(ctx context.Context, f format.Format, event io.Reader) error {
	var buf bytes.Buffer
	_, err := io.Copy(&buf, event)
	if err != nil {
		return err
	}
	b.Data = buf.Bytes()
	return nil
}

func (b *pubsubMessagePublisher) Start(ctx context.Context) error {
	b.Attributes = make(map[string]string)
	return nil
}

func (b *pubsubMessagePublisher) End(ctx context.Context) error {
	return nil
}

func (b *pubsubMessagePublisher) SetData(reader io.Reader) error {
	buf, ok := reader.(*bytes.Buffer)
	if !ok {
		buf = new(bytes.Buffer)
		_, err := io.Copy(buf, reader)
		if err != nil {
			return err
		}
	}
	b.Data = buf.Bytes()
	return nil
}

func (b *pubsubMessagePublisher) SetAttribute(attribute spec.Attribute, value interface{}) error {
	if attribute.Kind() == spec.DataContentType {
		if value == nil {
			delete(b.Attributes, contentType)
		}

		// Everything is a string here
		s, err := types.Format(value)
		if err != nil {
			return err
		}
		b.Attributes[contentType] = s
	} else {
		if value == nil {
			delete(b.Attributes, prefix+attribute.Name())
		}

		// Everything is a string here
		s, err := types.Format(value)
		if err != nil {
			return err
		}
		b.Attributes[prefix+attribute.Name()] = s
	}
	return nil
}

func (b *pubsubMessagePublisher) SetExtension(name string, value interface{}) error {
	if value == nil {
		delete(b.Attributes, prefix+name)
	}

	// Store extensions as string attrs as well
	s, err := types.Format(value)
	if err != nil {
		return err
	}
	b.Attributes[prefix+name] = s
	return nil
}

var _ binding.StructuredWriter = (*pubsubMessagePublisher)(nil) // Test it conforms to the interface
var _ binding.BinaryWriter = (*pubsubMessagePublisher)(nil)     // Test it conforms to the interface
