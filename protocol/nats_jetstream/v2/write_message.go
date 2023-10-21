/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package nats_jetstream

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/binding/format"
	"github.com/cloudevents/sdk-go/v2/binding/spec"
	"github.com/nats-io/nats.go"
)

// WriteMsg fills the provided writer with the bindings.Message m.
// Using context you can tweak the encoding processing (more details on binding.Write documentation).
// The nats.Header returned is not deep-copied. The header values should be deep-copied to an event object.
func WriteMsg(ctx context.Context, m binding.Message, writer io.ReaderFrom, transformers ...binding.Transformer) (nats.Header, error) {
	structuredWriter := &natsMessageWriter{writer}
	binaryWriter := &natsBinaryMessageWriter{ReaderFrom: writer}

	_, err := binding.Write(
		ctx,
		m,
		structuredWriter,
		binaryWriter,
		transformers...,
	)
	natsHeader := binaryWriter.header

	return natsHeader, err
}

type natsMessageWriter struct {
	io.ReaderFrom
}

// StructuredWriter  implements StructuredWriter.SetStructuredEvent
func (w *natsMessageWriter) SetStructuredEvent(_ context.Context, _ format.Format, event io.Reader) error {
	if _, err := w.ReadFrom(event); err != nil {
		return err
	}

	return nil
}

var _ binding.StructuredWriter = (*natsMessageWriter)(nil) // Test it conforms to the interface

type natsBinaryMessageWriter struct {
	io.ReaderFrom
	header nats.Header
}

// SetAttribute implements MessageMetadataWriter.SetAttribute
func (w *natsBinaryMessageWriter) SetAttribute(attribute spec.Attribute, value interface{}) error {
	prefixedName := withPrefix(attribute.Name())
	convertedValue := fmt.Sprint(value)
	switch attribute.Kind().String() {
	case spec.Time.String():
		timeValue := value.(time.Time)
		convertedValue = timeValue.Format(time.RFC3339Nano)
	}
	w.header.Set(prefixedName, convertedValue)
	return nil
}

// SetExtension implements MessageMetadataWriter.SetExtension
func (w *natsBinaryMessageWriter) SetExtension(name string, value interface{}) error {
	prefixedName := withPrefix(name)
	convertedValue := fmt.Sprint(value)
	w.header.Set(prefixedName, convertedValue)
	return nil
}

// Start implements BinaryWriter.Start
func (w *natsBinaryMessageWriter) Start(ctx context.Context) error {
	w.header = nats.Header{}
	return nil
}

// SetData implements BinaryWriter.SetData
func (w *natsBinaryMessageWriter) SetData(data io.Reader) error {
	if _, err := w.ReadFrom(data); err != nil {
		return err
	}

	return nil
}

// End implements BinaryWriter.End
func (w *natsBinaryMessageWriter) End(ctx context.Context) error {
	return nil
}
