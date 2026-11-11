/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package sqs

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/binding/format"
	"github.com/cloudevents/sdk-go/v2/binding/spec"
	cetypes "github.com/cloudevents/sdk-go/v2/types"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

func WriteMessageInput(ctx context.Context, m binding.Message, msgInput *sqs.SendMessageInput, transformers ...binding.Transformer) error {
	structuredWriter := (*sqsMessageWriter)(msgInput)
	binaryWriter := (*sqsMessageWriter)(msgInput)

	_, err := binding.Write(
		ctx,
		m,
		structuredWriter,
		binaryWriter,
		transformers...,
	)
	return err
}

type sqsMessageWriter sqs.SendMessageInput

// StructuredWriter  implements StructuredWriter.SetStructuredEvent
func (w *sqsMessageWriter) SetStructuredEvent(_ context.Context, _ format.Format, event io.Reader) error {
	val, err := io.ReadAll(event)
	if err != nil {
		return err
	}
	w.MessageBody = aws.String(string(val))
	w.MessageAttributes = make(map[string]types.MessageAttributeValue)
	return nil
}

// Start implements BinaryWriter.Start
func (b *sqsMessageWriter) Start(ctx context.Context) error {
	b.MessageAttributes = make(map[string]types.MessageAttributeValue)
	return nil
}

// End implements BinaryWriter.End
func (b *sqsMessageWriter) End(ctx context.Context) error {
	return nil
}

// SetData implements BinaryWriter.SetData
func (b *sqsMessageWriter) SetData(reader io.Reader) error {
	data, err := io.ReadAll(reader)
	if err != nil {
		return err
	}
	b.MessageBody = aws.String(string(data))
	return nil
}

// SetAttribute implements MessageMetadataWriter.SetAttribute
func (b *sqsMessageWriter) SetAttribute(attribute spec.Attribute, value interface{}) error {
	if attribute.Kind() == spec.DataContentType {
		if value == nil {
			return nil
		}
		s, err := cetypes.Format(value)
		if err != nil {
			return err
		}
		b.MessageAttributes[contentType] = stringMessageAttribute(s)
	} else {
		prefixedName := withPrefix(attribute.Name())
		if value == nil {
			delete(b.MessageAttributes, prefixedName)
			return nil
		}
		convertedValue := fmt.Sprint(value)
		if attribute.Kind().String() == spec.Time.String() {
			timeValue := value.(time.Time)
			convertedValue = timeValue.Format(time.RFC3339Nano)
		}
		b.MessageAttributes[prefixedName] = stringMessageAttribute(convertedValue)
	}
	return nil
}

// SetExtension implements MessageMetadataWriter.SetExtension
func (b *sqsMessageWriter) SetExtension(name string, value interface{}) error {
	prefixedName := withPrefix(name)
	convertedValue := fmt.Sprint(value)
	b.MessageAttributes[prefixedName] = stringMessageAttribute(convertedValue)
	return nil
}

var (
	_ binding.BinaryWriter     = (*sqsMessageWriter)(nil) // Test it conforms to the interface
	_ binding.StructuredWriter = (*sqsMessageWriter)(nil) // Test it conforms to the interface
)

func stringMessageAttribute(val string) types.MessageAttributeValue {
	return types.MessageAttributeValue{
		DataType:    aws.String("String"),
		StringValue: aws.String(val),
	}
}
