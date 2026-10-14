/*
 Copyright 2026 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package kafka_franz

import (
	"bytes"
	"context"
	"io"

	"github.com/twmb/franz-go/pkg/kgo"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/binding/format"
	"github.com/cloudevents/sdk-go/v2/binding/spec"
	"github.com/cloudevents/sdk-go/v2/types"
)

type kafkaRecordWriter kgo.Record

var (
	_ binding.StructuredWriter = (*kafkaRecordWriter)(nil)
	_ binding.BinaryWriter     = (*kafkaRecordWriter)(nil)
)

// WriteProducerMessage fills the provided record with the message in.
func WriteProducerMessage(ctx context.Context, in binding.Message, record *kgo.Record, transformers ...binding.Transformer) error {
	writer := (*kafkaRecordWriter)(record)
	_, err := binding.Write(ctx, in, writer, writer, transformers...)
	return err
}

func (w *kafkaRecordWriter) SetStructuredEvent(ctx context.Context, f format.Format, event io.Reader) error {
	w.Headers = []kgo.RecordHeader{{
		Key:   contentTypeHeader,
		Value: []byte(f.MediaType()),
	}}

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, event); err != nil {
		return err
	}

	w.Value = buf.Bytes()
	return nil
}

func (w *kafkaRecordWriter) Start(context.Context) error {
	w.Headers = []kgo.RecordHeader{}
	return nil
}

func (w *kafkaRecordWriter) End(context.Context) error {
	return nil
}

func (w *kafkaRecordWriter) SetData(reader io.Reader) error {
	buf, ok := reader.(*bytes.Buffer)
	if !ok {
		buf = new(bytes.Buffer)
		if _, err := io.Copy(buf, reader); err != nil {
			return err
		}
	}
	w.Value = buf.Bytes()
	return nil
}

func (w *kafkaRecordWriter) SetAttribute(attribute spec.Attribute, value interface{}) error {
	if attribute.Kind() == spec.DataContentType {
		if value == nil {
			w.removeHeader(contentTypeHeader)
			return nil
		}
		return w.addHeader(contentTypeHeader, value)
	}

	key := prefix + attribute.Name()
	if value == nil {
		w.removeHeader(key)
		return nil
	}
	return w.addHeader(key, value)
}

func (w *kafkaRecordWriter) SetExtension(name string, value interface{}) error {
	key := prefix + name
	if value == nil {
		w.removeHeader(key)
		return nil
	}
	return w.addHeader(key, value)
}

func (w *kafkaRecordWriter) removeHeader(key string) {
	for i, header := range w.Headers {
		if header.Key == key {
			w.Headers = append(w.Headers[:i], w.Headers[i+1:]...)
			return
		}
	}
}

func (w *kafkaRecordWriter) addHeader(key string, value interface{}) error {
	s, err := types.Format(value)
	if err != nil {
		return err
	}
	w.Headers = append(w.Headers, kgo.RecordHeader{Key: key, Value: []byte(s)})
	return nil
}
