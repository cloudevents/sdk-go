/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package pubsub

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"cloud.google.com/go/pubsub/v2"
	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/binding/format"
	"github.com/cloudevents/sdk-go/v2/binding/spec"
	"github.com/cloudevents/sdk-go/v2/protocol"
)

const (
	prefix = "ce-"

	// contentType identifies the event format in structured mode and the data media type in binary mode.
	contentType = "content-type"
	// legacyContentType is retained for compatibility with older senders and consumers.
	legacyContentType = "Content-Type"
)

var specs = spec.WithPrefix(prefix)

// Message represents a Pub/Sub message.
// This message *can* be read several times safely
type Message struct {
	internal *pubsub.Message
	format   format.Format
	version  spec.Version
}

// NewMessage returns a binding.Message with data and attributes.
// This message *can* be read several times safely
func NewMessage(pm *pubsub.Message) *Message {
	var f format.Format = nil
	var version spec.Version = nil
	if pm.Attributes != nil {
		s, ok := pm.Attributes[contentType]
		if !ok {
			// Use the legacy key only when the binding's content-type is absent.
			s = pm.Attributes[legacyContentType]
		}
		if format.IsFormat(s) {
			f = format.Lookup(s)
		}
		if s := pm.Attributes[specs.PrefixedSpecVersionName()]; s != "" {
			version = specs.Version(s)
		}
	}

	return &Message{
		internal: pm,
		format:   f,
		version:  version,
	}
}

// Check if pubsub.Message implements binding.Message
var _ binding.Message = (*Message)(nil)
var _ binding.MessageMetadataReader = (*Message)(nil)

func (m *Message) ReadEncoding() binding.Encoding {
	// Structured messages may duplicate CloudEvents attributes in metadata.
	if m.format != nil {
		return binding.EncodingStructured
	}
	if m.version != nil {
		return binding.EncodingBinary
	}
	return binding.EncodingUnknown
}

func (m *Message) ReadStructured(ctx context.Context, encoder binding.StructuredWriter) error {
	if m.ReadEncoding() != binding.EncodingStructured {
		return binding.ErrNotStructured
	}
	return encoder.SetStructuredEvent(ctx, m.format, bytes.NewReader(m.internal.Data))
}

func (m *Message) ReadBinary(ctx context.Context, encoder binding.BinaryWriter) (err error) {
	if m.ReadEncoding() != binding.EncodingBinary {
		return binding.ErrNotBinary
	}
	if _, err = m.binaryDataContentType(m.version.AttributeFromKind(spec.DataContentType)); err != nil {
		return err
	}

	for k, v := range m.internal.Attributes {
		if strings.HasPrefix(k, prefix) {
			attr := m.version.Attribute(k)
			if attr != nil {
				err = encoder.SetAttribute(attr, string(v))
			} else {
				err = encoder.SetExtension(strings.TrimPrefix(k, prefix), string(v))
			}
		} else if k == contentType || k == legacyContentType {
			attr := m.version.AttributeFromKind(spec.DataContentType)
			// Let the prefixed attribute provide the value when present.
			// content-type was checked for agreement; Content-Type is only a fallback.
			if _, ok := m.internal.Attributes[prefix+attr.Name()]; ok {
				continue
			}
			if k == legacyContentType {
				if _, ok := m.internal.Attributes[contentType]; ok {
					continue
				}
			}
			err = encoder.SetAttribute(attr, string(v))
		}
		if err != nil {
			return err
		}
	}

	if m.internal.Data != nil {
		return encoder.SetData(bytes.NewBuffer(m.internal.Data))
	}

	return
}

func (m *Message) GetAttribute(k spec.Kind) (spec.Attribute, interface{}) {
	attr := m.version.AttributeFromKind(k)
	if attr != nil {
		if k == spec.DataContentType {
			value, err := m.binaryDataContentType(attr)
			if err != nil {
				// There is no unambiguous value; ReadBinary reports the conflict.
				return attr, nil
			}
			return attr, value
		}
		return attr, m.internal.Attributes[prefix+attr.Name()]
	}
	return nil, nil
}

func (m *Message) binaryDataContentType(attr spec.Attribute) (string, error) {
	prefixedName := prefix + attr.Name()
	prefixedValue, hasPrefixedValue := m.internal.Attributes[prefixedName]
	messageValue, hasMessageValue := m.internal.Attributes[contentType]
	if hasPrefixedValue && hasMessageValue && prefixedValue != messageValue {
		return "", fmt.Errorf("pubsub: conflicting %q and %q values: %q != %q", contentType, prefixedName, messageValue, prefixedValue)
	}
	if hasMessageValue {
		return messageValue, nil
	}
	if hasPrefixedValue {
		return prefixedValue, nil
	}
	return m.internal.Attributes[legacyContentType], nil
}

func (m *Message) GetExtension(name string) interface{} {
	return m.internal.Attributes[prefix+name]
}

// Finish marks the message to be forgotten and returns the provided error without modification.
// If err is nil or of type protocol.ResultACK the PubSub message will be acknowledged, otherwise nack-ed.
func (m *Message) Finish(err error) error {
	if protocol.IsACK(err) {
		m.internal.Ack()
	} else {
		m.internal.Nack()
	}
	return err
}
