/*
 Copyright 2026 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package kafka_franz

import (
	"bytes"
	"context"
	"errors"
	"strconv"
	"strings"
	"sync"

	"github.com/twmb/franz-go/pkg/kgo"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/binding/format"
	"github.com/cloudevents/sdk-go/v2/binding/spec"
)

const (
	prefix            = "ce_"
	contentTypeHeader = "content-type"
)

const (
	KafkaOffsetKey    = "kafkaoffset"
	KafkaPartitionKey = "kafkapartition"
	KafkaTopicKey     = "kafkatopic"
	KafkaMessageKey   = "kafkamessagekey"
)

var specs = spec.WithPrefix(prefix)

// Message represents a Kafka message.
// This message can be read several times safely.
type Message struct {
	record      *kgo.Record
	properties  map[string][]byte
	format      format.Format
	version     spec.Version
	contentType string
}

var (
	_ binding.Message               = (*Message)(nil)
	_ binding.MessageMetadataReader = (*Message)(nil)
)

// NewMessage returns a binding.Message that holds the provided kgo.Record.
// The returned binding.Message can be read several times safely.
func NewMessage(record *kgo.Record) *Message {
	if record == nil {
		panic("the kgo.Record must not be nil")
	}

	var contentType, contentVersion string
	properties := make(map[string][]byte, len(record.Headers)+4)
	for _, header := range record.Headers {
		key := strings.ToLower(header.Key)
		if key == contentTypeHeader {
			contentType = string(header.Value)
		}
		if key == specs.PrefixedSpecVersionName() {
			contentVersion = string(header.Value)
		}
		properties[key] = append([]byte(nil), header.Value...)
	}

	properties[prefix+KafkaOffsetKey] = []byte(strconv.FormatInt(record.Offset, 10))
	properties[prefix+KafkaPartitionKey] = []byte(strconv.FormatInt(int64(record.Partition), 10))
	properties[prefix+KafkaTopicKey] = []byte(record.Topic)
	if len(record.Key) > 0 {
		properties[prefix+KafkaMessageKey] = append([]byte(nil), record.Key...)
	}

	message := &Message{
		record:      record,
		properties:  properties,
		contentType: contentType,
	}
	if ft := format.Lookup(contentType); ft != nil {
		message.format = ft
	} else if v := specs.Version(contentVersion); v != nil {
		message.version = v
	}

	return message
}

func (m *Message) ReadEncoding() binding.Encoding {
	if m.version != nil {
		return binding.EncodingBinary
	}
	if m.format != nil {
		return binding.EncodingStructured
	}
	return binding.EncodingUnknown
}

func (m *Message) ReadStructured(ctx context.Context, encoder binding.StructuredWriter) error {
	if m.format == nil {
		return binding.ErrNotStructured
	}
	return encoder.SetStructuredEvent(ctx, m.format, bytes.NewReader(m.record.Value))
}

func (m *Message) ReadBinary(ctx context.Context, encoder binding.BinaryWriter) error {
	if m.version == nil {
		return binding.ErrNotBinary
	}

	var err error
	for key, value := range m.properties {
		switch {
		case strings.HasPrefix(key, prefix):
			attr := m.version.Attribute(key)
			if attr != nil {
				err = encoder.SetAttribute(attr, string(value))
			} else {
				err = encoder.SetExtension(strings.TrimPrefix(key, prefix), string(value))
			}
		case key == contentTypeHeader:
			err = encoder.SetAttribute(m.version.AttributeFromKind(spec.DataContentType), string(value))
		}
		if err != nil {
			return err
		}
	}

	if m.record.Value != nil {
		err = encoder.SetData(bytes.NewBuffer(m.record.Value))
	}
	return err
}

func (m *Message) Finish(error) error {
	return nil
}

func (m *Message) GetAttribute(k spec.Kind) (spec.Attribute, interface{}) {
	if m.version == nil {
		return nil, nil
	}
	attr := m.version.AttributeFromKind(k)
	if attr == nil {
		return nil, nil
	}
	return attr, string(m.properties[attr.PrefixedName()])
}

func (m *Message) GetExtension(name string) interface{} {
	value, ok := m.properties[prefix+name]
	if !ok {
		return nil
	}
	return string(value)
}

type receivedMessage struct {
	*Message
	finish     func(error) error
	finishOnce sync.Once
	finishErr  error
}

var _ binding.MessageWrapper = (*receivedMessage)(nil)

func (m *receivedMessage) Finish(err error) error {
	m.finishOnce.Do(func() {
		m.finishErr = m.Message.Finish(err)
		if m.finish != nil {
			m.finishErr = errors.Join(m.finishErr, m.finish(err))
		}
	})
	return m.finishErr
}

func (m *receivedMessage) GetWrappedMessage() binding.Message {
	return m.Message
}
