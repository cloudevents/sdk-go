/*
 Copyright 2024 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package nats_jetstream

import (
	"bytes"
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/cloudevents/sdk-go/v2/binding/spec"
	bindingtest "github.com/cloudevents/sdk-go/v2/binding/test"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/test"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type jetStreamMsg struct {
	jetstream.Msg
	msg *nats.Msg
}

func (j *jetStreamMsg) Data() []byte         { return j.msg.Data }
func (j *jetStreamMsg) Headers() nats.Header { return j.msg.Header }

var (
	outBinaryMessage = bindingtest.MockBinaryMessage{
		Metadata:   map[spec.Attribute]interface{}{},
		Extensions: map[string]interface{}{},
	}
	outStructMessage = bindingtest.MockStructuredMessage{}

	testEvent     = test.FullEvent()
	binaryData, _ = json.Marshal(map[string]string{
		"ce-type":            testEvent.Type(),
		"ce-source":          testEvent.Source(),
		"ce-id":              testEvent.ID(),
		"ce-time":            test.Timestamp.String(),
		"ce-specversion":     "1.0",
		"ce-dataschema":      test.Schema.String(),
		"ce-datacontenttype": "text/json",
		"ce-subject":         "receiverTopic",
		"ce-exta":            "someext",
	})
	structuredReceiverMessage = &jetStreamMsg{
		msg: &nats.Msg{
			Subject: "hello",
			Data:    binaryData,
		},
	}
	binaryReceiverMessage = &jetStreamMsg{
		msg: &nats.Msg{
			Subject: "hello",
			Data:    testEvent.Data(),
			Header: nats.Header{
				"ce-type":            {testEvent.Type()},
				"ce-source":          {testEvent.Source()},
				"ce-id":              {testEvent.ID()},
				"ce-time":            {test.Timestamp.String()},
				"ce-specversion":     {"1.0"},
				"ce-dataschema":      {test.Schema.String()},
				"ce-datacontenttype": {"text/json"},
				"ce-subject":         {"receiverTopic"},
				"ce-exta":            {"someext"},
			},
		},
	}
)

func TestNewMessage(t *testing.T) {
	tests := []struct {
		name                    string
		receiverMessage         jetstream.Msg
		expectedEncoding        binding.Encoding
		expectedStructuredError error
		expectedBinaryError     error
	}{
		{
			name:                    "Structured encoding",
			receiverMessage:         structuredReceiverMessage,
			expectedEncoding:        binding.EncodingStructured,
			expectedStructuredError: nil,
			expectedBinaryError:     binding.ErrNotBinary,
		},
		{
			name:                    "Binary encoding",
			receiverMessage:         binaryReceiverMessage,
			expectedEncoding:        binding.EncodingBinary,
			expectedStructuredError: binding.ErrNotStructured,
			expectedBinaryError:     nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewMessage(tt.receiverMessage)
			if got == nil {
				t.Errorf("Error in NewMessage!")
			}
			err := got.ReadBinary(context.TODO(), &outBinaryMessage)
			if err != tt.expectedBinaryError {
				t.Errorf("ReadBinary err:%s", err.Error())
			}
			if tt.expectedEncoding == binding.EncodingBinary {
				if !bytes.Equal(outBinaryMessage.Body, tt.receiverMessage.Data()) {
					t.Fail()
				}
			}
			err = got.ReadStructured(context.TODO(), &outStructMessage)
			if err != tt.expectedStructuredError {
				t.Errorf("ReadStructured err:%s", err.Error())
			}
			if tt.expectedEncoding == binding.EncodingStructured {
				if !bytes.Equal(outStructMessage.Bytes, tt.receiverMessage.Data()) {
					t.Fail()
				}
			}
			if got.ReadEncoding() != tt.expectedEncoding {
				t.Errorf("ExpectedEncoding %s, while got %s", tt.expectedEncoding, got.ReadEncoding())
			}
		})
	}
}

func TestGetAttribute(t *testing.T) {
	specs = spec.WithPrefix(prefix)
	tests := []struct {
		name                   string
		receiverMessage        jetstream.Msg
		attributeKind          spec.Kind
		expectedAttribute      spec.Attribute
		expectedAttributeValue interface{}
	}{
		{
			name:                   "Binary encoding", // test only makes sense for binary
			receiverMessage:        binaryReceiverMessage,
			attributeKind:          spec.Type,
			expectedAttribute:      specs.Version(spec.V1.String()).AttributeFromKind(spec.Type),
			expectedAttributeValue: "com.example.FullEvent",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			message := NewMessage(tt.receiverMessage)
			if message == nil {
				t.Errorf("Error in NewMessage!")
			}
			gotAttribute, gotAttributeValue := message.GetAttribute(tt.attributeKind)
			if gotAttributeValue != tt.expectedAttributeValue {
				t.Errorf("ExpectedAttributeValue %s, while got %s", tt.expectedAttributeValue, gotAttributeValue)
			}
			if !reflect.DeepEqual(gotAttribute, tt.expectedAttribute) {
				t.Errorf("ExpectedAttribute %s, while got %s", tt.expectedAttribute, gotAttribute)
			}
		})
	}
}

func TestGetExtension(t *testing.T) {
	specs = spec.WithPrefix(prefix)
	tests := []struct {
		name                   string
		receiverMessage        jetstream.Msg
		extensionName          string
		expectedAttribute      spec.Attribute
		expectedExtensionValue interface{}
	}{
		{
			name:                   "Binary encoding", // test only makes sense for binary
			receiverMessage:        binaryReceiverMessage,
			extensionName:          "exta",
			expectedExtensionValue: "someext",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			message := NewMessage(tt.receiverMessage)
			if message == nil {
				t.Errorf("Error in NewMessage!")
			}
			gotExtensionValue := message.GetExtension(tt.extensionName)
			if gotExtensionValue != tt.expectedExtensionValue {
				t.Errorf("ExpectedExtensionValue %s, while got %s", tt.expectedExtensionValue, gotExtensionValue)
			}
		})
	}
}
