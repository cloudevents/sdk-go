package nats_jetstream

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/cloudevents/sdk-go/v2/binding/spec"
	bindingtest "github.com/cloudevents/sdk-go/v2/binding/test"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/test"
	"github.com/nats-io/nats.go"
)

var (
	outBinaryMessage = bindingtest.MockBinaryMessage{
		Metadata:   map[spec.Attribute]interface{}{},
		Extensions: map[string]interface{}{},
	}
	outStructMessage = bindingtest.MockStructuredMessage{}

	testEvent     = test.FullEvent()
	binaryData, _ = json.Marshal(map[string]string{
		"ce_type":            testEvent.Type(),
		"ce_source":          testEvent.Source(),
		"ce_id":              testEvent.ID(),
		"ce_time":            test.Timestamp.String(),
		"ce_specversion":     "1.0",
		"ce_dataschema":      test.Schema.String(),
		"ce_datacontenttype": "text/json",
		"ce_subject":         "receiverTopic",
		"ce_exta":            "someext",
	})
	structuredConsumerMessage = &nats.Msg{
		Subject: "hello",
		Data:    binaryData,
	}
	binaryConsumerMessage = &nats.Msg{
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
	}
)

func TestNewMessage(t *testing.T) {
	tests := []struct {
		name                    string
		consumerMessage         *nats.Msg
		expectedEncoding        binding.Encoding
		expectedStructuredError error
		expectedBinaryError     error
	}{
		{
			name:                    "Structured encoding",
			consumerMessage:         structuredConsumerMessage,
			expectedEncoding:        binding.EncodingStructured,
			expectedStructuredError: nil,
			expectedBinaryError:     binding.ErrNotBinary,
		},
		{
			name:                    "Binary encoding",
			consumerMessage:         binaryConsumerMessage,
			expectedEncoding:        binding.EncodingBinary,
			expectedStructuredError: binding.ErrNotStructured,
			expectedBinaryError:     nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewMessage(tt.consumerMessage)
			if got == nil {
				t.Errorf("Error in NewMessage!")
			}
			err := got.ReadBinary(context.TODO(), &outBinaryMessage)
			if err != tt.expectedBinaryError {
				t.Errorf("ReadBinary err:%s", err.Error())
			}
			err = got.ReadStructured(context.TODO(), &outStructMessage)
			if err != tt.expectedStructuredError {
				t.Errorf("ReadStructured err:%s", err.Error())
			}
			if got.ReadEncoding() != tt.expectedEncoding {
				t.Errorf("ExpectedEncoding %s, while got %s", tt.expectedEncoding, got.ReadEncoding())
			}
		})
	}
}
