/*
Copyright 2023 The CloudEvents Authors
SPDX-License-Identifier: Apache-2.0
*/

package sqs

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/cloudevents/sdk-go/v2/event"
)

type mockGetObjectAPI func(ctx context.Context, params *sqs.DeleteMessageInput, optFns ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error)

func (m mockGetObjectAPI) DeleteMessage(ctx context.Context, params *sqs.DeleteMessageInput, optFns ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error) {
	return m(ctx, params, optFns...)
}

func TestReadStructured(t *testing.T) {
	tests := []struct {
		name      string
		client    func(t *testing.T) SQSDeleteMessageAPI
		msg       *types.Message
		queueName *string
		wantErr   error
	}{
		{
			name: "nil format",
			msg: &types.Message{
				Body: aws.String(""),
			},
			client: func(t *testing.T) SQSDeleteMessageAPI {
				return mockGetObjectAPI(func(ctx context.Context, params *sqs.DeleteMessageInput, optFns ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error) {
					return nil, nil
				})
			},
		},
		{
			name: "json format",
			msg: &types.Message{
				Body: aws.String(""),
				MessageAttributes: map[string]types.MessageAttributeValue{
					contentType: stringMessageAttribute(event.ApplicationCloudEventsJSON),
				},
			},
			client: func(t *testing.T) SQSDeleteMessageAPI {
				return mockGetObjectAPI(func(ctx context.Context, params *sqs.DeleteMessageInput, optFns ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error) {
					t.Helper()
					return nil, nil
				})
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			msg := NewMessage(tc.msg, tc.client(t), tc.queueName)
			err := msg.ReadStructured(context.Background(), msgToWriter(tc.msg))
			if err != tc.wantErr {
				t.Errorf("Error unexpected. got: %v, want: %v", err, tc.wantErr)
			}
		})
	}
}

func TestReadBinary(t *testing.T) {
	msg := &types.Message{
		Body: aws.String("{hello:world}"),
		MessageAttributes: map[string]types.MessageAttributeValue{
			"ce-specversion": stringMessageAttribute("1.0"),
			"ce-type":        stringMessageAttribute("binary.test"),
			"ce-source":      stringMessageAttribute("test-source"),
			"ce-id":          stringMessageAttribute("ABC-123"),
		},
		ReceiptHandle: aws.String("test-receipt-handle"),
	}

	client := func(t *testing.T) SQSDeleteMessageAPI {
		return mockGetObjectAPI(func(ctx context.Context, params *sqs.DeleteMessageInput, optFns ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error) {
			t.Helper()
			return nil, nil
		})
	}
	message := NewMessage(msg, client(t), aws.String("test-queue"))
	err := message.ReadBinary(context.Background(), msgToWriter(msg))
	if err != nil {
		t.Errorf("Error unexpected. got: %v", err)
	}
}

func msgToWriter(msg *types.Message) *sqsMessageWriter {
	return &sqsMessageWriter{
		MessageBody:       msg.Body,
		MessageAttributes: msg.MessageAttributes,
	}
}
