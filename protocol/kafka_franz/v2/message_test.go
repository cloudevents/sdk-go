/*
 Copyright 2026 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package kafka_franz

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/twmb/franz-go/pkg/kgo"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/binding/format"
	"github.com/cloudevents/sdk-go/v2/test"
)

var (
	ctx       = context.Background()
	testEvent = test.FullEvent()

	structuredConsumerRecord = &kgo.Record{
		Topic:     "test-topic",
		Partition: 0,
		Offset:    10,
		Value: func() []byte {
			b, _ := format.JSON.Marshal(&testEvent)
			return b
		}(),
		Headers: []kgo.RecordHeader{{
			Key:   contentTypeHeader,
			Value: []byte(cloudevents.ApplicationCloudEventsJSON),
		}},
	}

	binaryConsumerRecord = &kgo.Record{
		Topic:     "test-topic",
		Partition: 0,
		Offset:    10,
		Value:     []byte("hello world!"),
		Headers: mapToRecordHeaders(map[string]string{
			"ce_type":            testEvent.Type(),
			"ce_source":          testEvent.Source(),
			"ce_id":              testEvent.ID(),
			"ce_time":            test.Timestamp.String(),
			"ce_specversion":     "1.0",
			"ce_dataschema":      test.Schema.String(),
			"ce_datacontenttype": "text/json",
			"ce_subject":         "receiverTopic",
			"exta":               "someext",
		}),
	}
)

func TestNewMessage(t *testing.T) {
	tests := []struct {
		name             string
		record           *kgo.Record
		expectedEncoding binding.Encoding
	}{
		{
			name:             "structured encoding",
			record:           structuredConsumerRecord,
			expectedEncoding: binding.EncodingStructured,
		},
		{
			name:             "binary encoding",
			record:           binaryConsumerRecord,
			expectedEncoding: binding.EncodingBinary,
		},
		{
			name: "unknown encoding",
			record: &kgo.Record{
				Topic:     "test-topic",
				Partition: 0,
				Offset:    10,
				Value:     []byte("{}"),
				Headers: []kgo.RecordHeader{{
					Key:   contentTypeHeader,
					Value: []byte("application/json"),
				}},
			},
			expectedEncoding: binding.EncodingUnknown,
		},
		{
			name: "binary encoding with empty value",
			record: &kgo.Record{
				Topic:     "test-topic",
				Partition: 0,
				Offset:    10,
				Headers: mapToRecordHeaders(map[string]string{
					"ce_type":            testEvent.Type(),
					"ce_source":          testEvent.Source(),
					"ce_id":              testEvent.ID(),
					"ce_time":            test.Timestamp.String(),
					"ce_specversion":     "1.0",
					"ce_dataschema":      test.Schema.String(),
					"ce_datacontenttype": "text/json",
					"ce_subject":         "receiverTopic",
				}),
			},
			expectedEncoding: binding.EncodingBinary,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := NewMessage(tt.record)
			require.Equal(t, tt.expectedEncoding, msg.ReadEncoding())

			var err error
			switch tt.expectedEncoding {
			case binding.EncodingStructured:
				err = msg.ReadStructured(ctx, (*kafkaRecordWriter)(tt.record))
			case binding.EncodingBinary:
				err = msg.ReadBinary(ctx, (*kafkaRecordWriter)(tt.record))
			}
			require.NoError(t, err)
		})
	}
}

func mapToRecordHeaders(m map[string]string) []kgo.RecordHeader {
	res := make([]kgo.RecordHeader, 0, len(m))
	for k, v := range m {
		res = append(res, kgo.RecordHeader{Key: k, Value: []byte(v)})
	}
	return res
}
