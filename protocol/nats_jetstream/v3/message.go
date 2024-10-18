/*
 Copyright 2024 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package nats_jetstream

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/nats-io/nats.go/jetstream"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/binding/format"
	"github.com/cloudevents/sdk-go/v2/binding/spec"
	"github.com/cloudevents/sdk-go/v2/protocol"
)

const (
	// see https://github.com/cloudevents/spec/blob/main/cloudevents/bindings/nats-protocol-binding.md
	prefix            = "ce-"
	contentTypeHeader = "content-type"
)

var (
	specs = spec.WithPrefix(prefix)

	// ErrNoVersion returned when no version header is found in the protocol header.
	ErrNoVersion = errors.New("message does not contain version header")
)

// Message implements binding.Message by wrapping an jetstream.Msg.
// This message *can* be read several times safely
type Message struct {
	Msg      jetstream.Msg
	encoding binding.Encoding
}

// NewMessage wraps an *nats.Msg in a binding.Message.
// The returned message *can* be read several times safely
// The default encoding returned is EncodingStructured unless the NATS message contains a specversion header.
func NewMessage(msg jetstream.Msg) *Message {
	encoding := binding.EncodingStructured
	if msg.Headers() != nil {
		if msg.Headers().Get(specs.PrefixedSpecVersionName()) != "" {
			encoding = binding.EncodingBinary
		}
	}
	return &Message{Msg: msg, encoding: encoding}
}

var _ binding.Message = (*Message)(nil)

// ReadEncoding return the type of the message Encoding.
func (m *Message) ReadEncoding() binding.Encoding {
	return m.encoding
}

// ReadStructured transfers a structured-mode event to a StructuredWriter.
func (m *Message) ReadStructured(ctx context.Context, encoder binding.StructuredWriter) error {
	if m.encoding != binding.EncodingStructured {
		return binding.ErrNotStructured
	}
	return encoder.SetStructuredEvent(ctx, format.JSON, bytes.NewReader(m.Msg.Data()))
}

// ReadBinary transfers a binary-mode event to an BinaryWriter.
func (m *Message) ReadBinary(ctx context.Context, encoder binding.BinaryWriter) error {
	if m.encoding != binding.EncodingBinary {
		return binding.ErrNotBinary
	}

	version := m.GetVersion()
	if version == nil {
		return ErrNoVersion
	}

	var err error
	for k, v := range m.Msg.Headers() {
		headerValue := v[0]
		if strings.HasPrefix(k, prefix) {
			attr := version.Attribute(k)
			if attr != nil {
				err = encoder.SetAttribute(attr, headerValue)
			} else {
				err = encoder.SetExtension(strings.TrimPrefix(k, prefix), headerValue)
			}
		} else if k == contentTypeHeader {
			err = encoder.SetAttribute(version.AttributeFromKind(spec.DataContentType), headerValue)
		}
		if err != nil {
			return err
		}
	}

	if m.Msg.Data() != nil {
		err = encoder.SetData(bytes.NewBuffer(m.Msg.Data()))
	}

	return err
}

// Finish *must* be called when message from a Receiver can be forgotten by the receiver.
func (m *Message) Finish(err error) error {
	// Ack and Nak first checks to see if the message has been acknowleged
	// and if Ack/Nak was done, it immediately returns an error without applying any logic to the message on the server.
	// Nak will only be sent if the error given is explictly a NACK error(protocol.ResultNACK).
	// AckPolicy effects if an explict Ack/Nak is needed.
	// AckExplicit: The default policy. Each individual message must be acknowledged.
	// 		Recommended for most reliability and functionality.
	// AckNone: No acknowledgment needed; the server assumes acknowledgment on delivery.
	// AckAll: Acknowledge only the last message received in a series; all previous messages are automatically acknowledged.
	// 		Will acknowledge all pending messages for all subscribers for Pull Consumer.
	// see: github.com/nats-io/nats.go/jetstream/ConsumerConfig.AckPolicy
	if m.Msg == nil {
		return nil
	}
	if protocol.IsNACK(err) {
		if err = m.Msg.Nak(); err != jetstream.ErrMsgAlreadyAckd {
			return err
		}
	}
	if protocol.IsACK(err) {
		if err = m.Msg.Ack(); err != jetstream.ErrMsgAlreadyAckd {
			return err
		}
	}

	// In the case that we receive an unknown error, the intent of whether the message should Ack/Nak is unknown.
	// When this happens, the ack/nak behavior will be based on the consumer configuration.  There are several options such as:
	// AckPolicy, AckWait, MaxDeliver, MaxAckPending
	// that determine how messages would be redelivered by the server.
	// [consumers configuration]: https://docs.nats.io/nats-concepts/jetstream/consumers#configuration
	return nil
}

// GetAttribute implements binding.MessageMetadataReader
func (m *Message) GetAttribute(attributeKind spec.Kind) (spec.Attribute, interface{}) {
	key := withPrefix(attributeKind.String())
	if m.Msg.Headers() != nil {
		version := m.GetVersion()
		headerValue := m.Msg.Headers().Get(key)
		if headerValue != "" {
			return version.Attribute(key), headerValue
		}
		return version.Attribute(key), nil
	}
	// if the headers are nil, the version is also nil.  Therefore return nil.
	return nil, nil
}

// GetExtension implements binding.MessageMetadataReader
func (m *Message) GetExtension(name string) interface{} {
	key := withPrefix(name)
	if m.Msg.Headers() != nil {
		headerValue := m.Msg.Headers().Get(key)
		if headerValue != "" {
			return headerValue
		}
	}
	return nil
}

// GetVersion looks for specVersion header and returns a Version object
func (m *Message) GetVersion() spec.Version {
	if m.Msg.Headers() == nil {
		return nil
	}
	versionValue := m.Msg.Headers().Get(specs.PrefixedSpecVersionName())
	if versionValue == "" {
		return nil
	}
	return specs.Version(versionValue)
}

// withPrefix prepends the prefix to the attribute name
func withPrefix(attributeName string) string {
	return fmt.Sprintf("%s%s", prefix, attributeName)
}
