/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package amqp

import (
	"bytes"
	"context"
	"reflect"
	"strings"

	"github.com/Azure/go-amqp"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/binding/format"
	"github.com/cloudevents/sdk-go/v2/binding/spec"
)

const (
	prefix     = "cloudEvents:" // Name prefix for AMQP properties that hold CE attributes.
	amqpPrefix = "cloudEvents_" // AMQP prefix compatible with JMS 2.0
)

var (
	// Use the package path as AMQP error condition name
	condition = amqp.ErrorCondition(reflect.TypeOf(Message{}).PkgPath())
	specs     = []*spec.Versions{
		spec.WithPrefix(prefix),
		spec.WithPrefix(amqpPrefix),
	}
)

// Message implements binding.Message by wrapping an *amqp.Message.
// This message *can* be read several times safely
type Message struct {
	AMQP    *amqp.Message
	AMQPrcv *amqp.Receiver

	version spec.Version
	format  format.Format
}

// NewMessage wrap an *amqp.Message in a binding.Message.
// The returned message *can* be read several times safely
func NewMessage(message *amqp.Message, receiver *amqp.Receiver) *Message {
	var vn spec.Version
	var fmt format.Format
	if message.Properties != nil && message.Properties.ContentType != nil &&
		format.IsFormat(*message.Properties.ContentType) {
		fmt = format.Lookup(*message.Properties.ContentType)
	} else if sv := getSpecVersion(message); sv != nil {
		vn = sv
	}
	return &Message{AMQP: message, AMQPrcv: receiver, format: fmt, version: vn}
}

var (
	_ binding.Message               = (*Message)(nil)
	_ binding.MessageMetadataReader = (*Message)(nil)
)

func getSpecVersion(message *amqp.Message) spec.Version {
	specUsed := 1
	sv, foundSpecVersion := message.ApplicationProperties[specs[specUsed].PrefixedSpecVersionName()]
	if !foundSpecVersion {
		// if not found try with the other spec
		specUsed = 0
		sv, foundSpecVersion = message.ApplicationProperties[specs[specUsed].PrefixedSpecVersionName()]
	}

	if foundSpecVersion {
		if svs, ok := sv.(string); ok {
			return specs[specUsed].Version(svs)
		}
	}
	return nil
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
	if m.format != nil {
		return encoder.SetStructuredEvent(ctx, m.format, bytes.NewReader(m.AMQP.GetData()))
	}
	return binding.ErrNotStructured
}

// ReadBinary transfers the AMQP message into binary (encoder)
// it supports two type of prefix with ":" and "_"
func (m *Message) ReadBinary(ctx context.Context, encoder binding.BinaryWriter) error {
	if m.version == nil {
		return binding.ErrNotBinary
	}
	var err error

	if m.AMQP.Properties != nil && m.AMQP.Properties.ContentType != nil {
		err = encoder.SetAttribute(m.version.AttributeFromKind(spec.DataContentType), m.AMQP.Properties.ContentType)
		if err != nil {
			return err
		}
	}

	for k, v := range m.AMQP.ApplicationProperties {

		hasAMQPPrefix := strings.HasPrefix(k, amqpPrefix)
		hasRegularPrefix := strings.HasPrefix(k, prefix)

		// skip the properties if no prefix is there
		if !hasAMQPPrefix && !hasRegularPrefix {
			continue
		}

		// if the key is an attribue setup and continue
		attr := m.version.Attribute(k)
		if attr != nil {
			err = encoder.SetAttribute(attr, v)
			if err != nil {
				return err
			}
			continue
		}

		// if the key is an extension, find which prefix is used
		currentPrefix := amqpPrefix
		if hasRegularPrefix {
			currentPrefix = prefix
		}
		err = encoder.SetExtension(strings.ToLower(strings.TrimPrefix(k, currentPrefix)), v)
		if err != nil {
			return err
		}
	}

	data := m.AMQP.GetData()
	if len(data) != 0 { // Some data
		err = encoder.SetData(bytes.NewBuffer(data))
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *Message) GetAttribute(k spec.Kind) (spec.Attribute, interface{}) {
	attr := m.version.AttributeFromKind(k)
	if attr != nil {
		return attr, m.AMQP.ApplicationProperties[attr.PrefixedName()]
	}
	return nil, nil
}

func (m *Message) GetExtension(name string) interface{} {
	return m.AMQP.ApplicationProperties[amqpPrefix+name]
}

func (m *Message) Finish(err error) error {
	if err != nil {
		return m.AMQPrcv.RejectMessage(context.Background(), m.AMQP, &amqp.Error{
			Condition:   condition,
			Description: err.Error(),
		})
	}
	return m.AMQPrcv.AcceptMessage(context.Background(), m.AMQP)
}
