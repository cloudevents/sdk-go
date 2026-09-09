/*
Copyright 2023 The CloudEvents Authors
SPDX-License-Identifier: Apache-2.0
*/

package sqs

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/binding/format"
	"github.com/cloudevents/sdk-go/v2/binding/spec"
	"github.com/cloudevents/sdk-go/v2/protocol"
)

const (
	prefix      = "ce-"
	contentType = "Content-Type"
)

var specs = spec.WithPrefix(prefix)

type SQSDeleteMessageAPI interface {
	DeleteMessage(ctx context.Context, params *sqs.DeleteMessageInput, optFns ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error)
}

// Message implements binding.Message by wrapping an SQS types.Message.
// This message *can* be read several times safely
type Message struct {
	Msg      *types.Message
	queueURL *string
	client   SQSDeleteMessageAPI

	version spec.Version
	format  format.Format
}

var (
	_ binding.Message               = (*Message)(nil)
	_ binding.MessageMetadataReader = (*Message)(nil)
)

// NewMessage wraps an SQS types.Message in a binding.Message.
// The returned message *can* be read several times safely
// The default encoding returned is EncodingStructured unless the SQS message contains a specversion
// header.
func NewMessage(msg *types.Message, sqsClient SQSDeleteMessageAPI, queueURL *string) *Message {
	v := getSpecVersion(msg)
	fmt := getFormat(msg)
	if v == nil && fmt == nil {
		fmt = format.JSON
	}
	return &Message{Msg: msg, version: v, format: fmt, client: sqsClient, queueURL: queueURL}
}

func getSpecVersion(message *types.Message) spec.Version {
	if sv, ok := message.MessageAttributes[specs.PrefixedSpecVersionName()]; ok {
		if sv.StringValue != nil {
			return specs.Version(aws.ToString(sv.StringValue))
		}
	}
	return nil
}

func getFormat(message *types.Message) format.Format {
	if sv, ok := message.MessageAttributes[contentType]; ok {
		if sv.StringValue != nil && format.IsFormat(aws.ToString(sv.StringValue)) {
			return format.Lookup(aws.ToString(sv.StringValue))
		}
	}
	return nil
}

// ReadEncoding return the type of the message Encoding.
func (m *Message) ReadEncoding() binding.Encoding {
	if m.version != nil {
		return binding.EncodingBinary
	}
	if m.format != nil {
		return binding.EncodingStructured
	}
	return binding.EncodingUnknown
}

// ReadStructured transfers a structured-mode event to a StructuredWriter.
func (m *Message) ReadStructured(ctx context.Context, encoder binding.StructuredWriter) error {
	if m.version != nil {
		return binding.ErrNotStructured
	}
	if m.format == nil {
		return binding.ErrNotStructured
	}
	data := []byte(*m.Msg.Body)
	return encoder.SetStructuredEvent(ctx, m.format, bytes.NewReader(data))
}

// ReadBinary transfers a binary-mode event to an BinaryWriter.
func (m *Message) ReadBinary(ctx context.Context, encoder binding.BinaryWriter) error {
	if m.format != nil {
		return binding.ErrNotBinary
	}
	var err error
	for k, attr := range m.Msg.MessageAttributes {
		v := aws.ToString(attr.StringValue)
		if strings.HasPrefix(k, prefix) {
			attr := m.version.Attribute(k)
			if attr != nil {
				err = encoder.SetAttribute(attr, v)
			} else {
				err = encoder.SetExtension(strings.ToLower(strings.TrimPrefix(k, prefix)), v)
			}
		} else if k == contentType {
			err = encoder.SetAttribute(m.version.AttributeFromKind(spec.DataContentType), v)
		}
		if err != nil {
			return err
		}
	}

	if m.Msg.Body != nil {
		err = encoder.SetData(bytes.NewBuffer([]byte(*m.Msg.Body)))
	}

	return err
}

// GetAttribute implements binding.MessageMetadataReader
func (m *Message) GetAttribute(k spec.Kind) (spec.Attribute, interface{}) {
	attr := m.version.AttributeFromKind(k)
	if attr == nil {
		return nil, nil
	}
	key := withPrefix(attr.Name())
	if msgAttr, ok := m.Msg.MessageAttributes[key]; ok && aws.ToString(msgAttr.StringValue) != "" {
		return attr, aws.ToString(msgAttr.StringValue)
	}
	return nil, nil
}

// GetExtension implements binding.MessageMetadataReader
func (m *Message) GetExtension(name string) interface{} {
	key := withPrefix(name)
	if attr, ok := m.Msg.MessageAttributes[key]; ok && attr.StringValue != nil {
		return aws.ToString(attr.StringValue)
	}
	return nil
}

// Finish implements binding.Message
// It deletes the message from SQS when the CloudEvent has been ACKed.
func (m *Message) Finish(err error) error {
	if protocol.IsACK(err) {
		// If the error is an ACK, we delete the message from SQS.
		ctx := context.Background()
		_, err = m.client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
			QueueUrl:      m.queueURL,
			ReceiptHandle: m.Msg.ReceiptHandle,
		})
		return err
	}
	return nil
}

// withPrefix prepends the prefix to the attribute name
func withPrefix(attributeName string) string {
	return fmt.Sprintf("%s%s", prefix, attributeName)
}
