/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package amqp

import (
	"context"
	"testing"

	"github.com/Azure/go-amqp"
	"github.com/cloudevents/sdk-go/v2/binding"
	bindingtest "github.com/cloudevents/sdk-go/v2/binding/test"
	"github.com/cloudevents/sdk-go/v2/event"
	. "github.com/cloudevents/sdk-go/v2/test"
	"github.com/stretchr/testify/require"
)

var (
	testEvent                           = FullEvent()
	applicationPropertiesWithUnderscore = map[string]interface{}{
		"cloudEvents_type":            testEvent.Type(),
		"cloudEvents_source":          testEvent.Source(),
		"cloudEvents_id":              testEvent.ID(),
		"cloudEvents_time":            Timestamp.String(),
		"cloudEvents_specversion":     "1.0",
		"cloudEvents_dataschema":      Schema.String(),
		"cloudEvents_datacontenttype": "text/json",
		"cloudEvents_subject":         "receiverTopic",
		"cloudEvents_exta":            "someext",
	}
	sampleApplicationPropertiesWithDots = map[string]interface{}{
		"cloudEvents:type":            testEvent.Type(),
		"cloudEvents:source":          testEvent.Source(),
		"cloudEvents:id":              testEvent.ID(),
		"cloudEvents:time":            Timestamp.String(),
		"cloudEvents:specversion":     "1.0",
		"cloudEvents:dataschema":      Schema.String(),
		"cloudEvents:datacontenttype": "text/json",
		"cloudEvents:subject":         "receiverTopic",
		"cloudEvents:exta":            "someext",
	}
)

func TestNewMessage_success(t *testing.T) {
	tests := []struct {
		name     string
		encoding binding.Encoding
	}{
		{
			name:     "Structured encoding",
			encoding: binding.EncodingStructured,
		},
		{
			name:     "Binary encoding",
			encoding: binding.EncodingBinary,
		},
	}
	for _, tt := range tests {
		EachEvent(t, Events(), func(t *testing.T, eventIn event.Event) {
			t.Run(tt.name, func(t *testing.T) {
				eventIn = eventIn.Clone()

				ctx := context.TODO()
				if tt.encoding == binding.EncodingStructured {
					ctx = binding.WithForceStructured(ctx)
				} else if tt.encoding == binding.EncodingBinary {
					ctx = binding.WithForceBinary(ctx)
				}

				message := amqp.Message{}
				require.NoError(t, WriteMessage(ctx, binding.ToMessage(&eventIn), &message))

				rcv := amqp.Receiver{}

				got := NewMessage(&message, &rcv)
				require.Equal(t, tt.encoding, got.ReadEncoding())
			})
		})
	}
}

func TestNewMessage_message_unknown(t *testing.T) {
	message := amqp.NewMessage([]byte("hello-world"))
	rcv := amqp.Receiver{}

	got := NewMessage(message, &rcv)
	require.Equal(t, binding.EncodingUnknown, got.ReadEncoding())
}

func TestMessage_ReadBinary(t *testing.T) {
	ctx := context.Background()
	outBinaryMessage := bindingtest.MockBinaryMessage{}
	outBinaryMessage.Start(ctx)
	type fields struct {
		message  *amqp.Message
		receiver *amqp.Receiver
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Should read an AMQP message with regular prefix :",
			fields: fields{
				message: &amqp.Message{
					ApplicationProperties: applicationPropertiesWithUnderscore,
				},
				receiver: &amqp.Receiver{},
			},
			args: args{
				ctx: ctx,
			},
			wantErr: false,
		},
		{
			name: "Should read an AMQP message with regular prefix :",
			fields: fields{
				message: &amqp.Message{
					ApplicationProperties: sampleApplicationPropertiesWithDots,
				},
				receiver: &amqp.Receiver{},
			},
			args: args{
				ctx: ctx,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMessage(tt.fields.message, tt.fields.receiver)
			err := m.ReadBinary(tt.args.ctx, &outBinaryMessage)
			if (err != nil) != tt.wantErr {
				t.Errorf("Message.ReadBinary() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
