/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package pubsub

import (
	"bytes"
	"context"
	"io"
	"strings"

	"cloud.google.com/go/pubsub/v2"
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
	if b.Attributes == nil {
		b.Attributes = make(map[string]string)
	}
	b.clearCloudEventAttributes()
	b.Attributes[contentType] = f.MediaType()
	b.Data = buf.Bytes()
	return nil
}

func (b *pubsubMessagePublisher) Start(ctx context.Context) error {
	// Attributes and data omitted by the next event must not survive reuse.
	b.clearCloudEventAttributes()
	b.Data = nil
	return nil
}

func (b *pubsubMessagePublisher) End(ctx context.Context) error {
	return nil
}

func (b *pubsubMessagePublisher) clearCloudEventAttributes() {
	// Preserve custom Pub/Sub attributes while removing metadata from the previous event.
	for name := range b.Attributes {
		if strings.HasPrefix(name, prefix) || name == contentType || name == legacyContentType {
			delete(b.Attributes, name)
		}
	}
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
	if value == nil {
		delete(b.Attributes, prefix+attribute.Name())
		if attribute.Kind() == spec.DataContentType {
			delete(b.Attributes, legacyContentType)
		}
		return nil
	}

	// Everything is a string here
	s, err := types.Format(value)
	if err != nil {
		return err
	}
	b.Attributes[prefix+attribute.Name()] = s
	if attribute.Kind() == spec.DataContentType {
		// Retain Content-Type for backward compatibility with existing consumers.
		// Use ce-datacontenttype when filtering binary-mode events by data content type.
		b.Attributes[legacyContentType] = s
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
