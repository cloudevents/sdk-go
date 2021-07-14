package nats_jetstream

import (
	"context"
	"encoding/json"
	bindingtest "github.com/cloudevents/sdk-go/v2/binding/test"
	"testing"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/test"
	"github.com/nats-io/nats.go"
)

var (
	outBinaryMessage = bindingtest.MockBinaryMessage{}
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
)

func TestNewMessage(t *testing.T) {
	tests := []struct {
		name             string
		consumerMessage  *nats.Msg
		expectedEncoding binding.Encoding
	}{
		{
			name:             "Structured encoding",
			consumerMessage:  structuredConsumerMessage,
			expectedEncoding: binding.EncodingStructured,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewMessage(tt.consumerMessage)
			if got == nil {
				t.Errorf("Error in NewMessage!")
			}
			err := got.ReadBinary(context.TODO(), &outBinaryMessage)
			if err == nil {
				t.Errorf("Response in ReadBinary should err")
			}
			err = got.ReadStructured(context.TODO(), &outStructMessage)
			if err != nil {
				t.Errorf("ReadStructured err:%s", err.Error())
			}
			if got.ReadEncoding() != tt.expectedEncoding {
				t.Errorf("ExpectedEncoding %s, while got %s", tt.expectedEncoding, got.ReadEncoding())
			}

		})
	}
}
