package pubsub_test

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/cloudevents/sdk-go/v1/cloudevents"
	"github.com/cloudevents/sdk-go/v1/cloudevents/transport/pubsub"
	"github.com/cloudevents/sdk-go/v1/cloudevents/types"
)

func TestCodecV1_Encode(t *testing.T) {
	now := types.Timestamp{Time: time.Now().UTC()}
	sourceUrl, _ := url.Parse("http://example.com/source")
	source := &types.URIRef{URL: *sourceUrl}

	schemaUrl, _ := url.Parse("http://example.com/schema")
	schema := &types.URI{URL: *schemaUrl}

	testCases := map[string]struct {
		codec   pubsub.CodecV1
		event   cloudevents.Event
		want    *pubsub.Message
		wantErr error
	}{
		"simple v1.0 default": {
			codec: pubsub.CodecV1{},
			event: cloudevents.Event{
				Context: cloudevents.EventContextV1{
					Type:   "com.example.test",
					Source: *source,
					ID:     "ABC-123",
				}.AsV1(),
			},
			want: &pubsub.Message{
				Attributes: map[string]string{
					"Content-Type": cloudevents.ApplicationCloudEventsJSON,
				},
				Data: func() []byte {
					body := map[string]interface{}{
						"specversion": "1.0",
						"id":          "ABC-123",
						"type":        "com.example.test",
						"source":      "http://example.com/source",
					}
					return toBytes(body)
				}(),
			},
		},
		"full v1.0 default": {
			codec: pubsub.CodecV1{},
			event: cloudevents.Event{
				Context: cloudevents.EventContextV1{
					ID:              "ABC-123",
					Time:            &now,
					Type:            "com.example.test",
					DataSchema:      schema,
					DataContentType: cloudevents.StringOfApplicationJSON(),
					Source:          *source,
					Extensions: map[string]interface{}{
						"test":       "extended",
						"generation": "1579743478182200",
					},
				}.AsV1(),
				Data: map[string]interface{}{
					"hello": "world",
				},
			},
			want: &pubsub.Message{
				Attributes: map[string]string{
					"Content-Type": cloudevents.ApplicationCloudEventsJSON,
				},
				Data: func() []byte {
					body := map[string]interface{}{
						"specversion":     "1.0",
						"datacontenttype": "application/json",
						"data": map[string]interface{}{
							"hello": "world",
						},
						"id":         "ABC-123",
						"time":       now,
						"type":       "com.example.test",
						"test":       "extended",
						"generation": "1579743478182200",
						"dataschema": "http://example.com/schema",
						"source":     "http://example.com/source",
					}
					return toBytes(body)
				}(),
			},
		},
		"full v1.0 binary": {
			codec: pubsub.CodecV1{
				DefaultEncoding: pubsub.BinaryV1,
			},
			event: cloudevents.Event{
				Context: cloudevents.EventContextV1{
					ID:         "ABC-123",
					Time:       &now,
					Type:       "com.example.test",
					DataSchema: schema,
					Source:     *source,
					Extensions: map[string]interface{}{
						"test":       "extended",
						"generation": "1579743478182200",
					},
				}.AsV1(),
				Data: map[string]interface{}{
					"hello": "world",
				},
			},
			want: &pubsub.Message{
				Attributes: map[string]string{
					prefix + "specversion":     "1.0",
					prefix + "id":              "ABC-123",
					prefix + "datacontenttype": cloudevents.ApplicationJSON,
					prefix + "time":            now.String(),
					prefix + "type":            "com.example.test",
					prefix + "dataschema":      "http://example.com/schema",
					prefix + "source":          "http://example.com/source",
					prefix + "test":            "extended",
					prefix + "generation":      "1579743478182200",
				},
				Data: func() []byte {
					data := map[string]interface{}{
						"hello": "world",
					}
					return toBytes(data)
				}(),
			},
		},
		"simple v1.0 structured": {
			codec: pubsub.CodecV1{DefaultEncoding: pubsub.StructuredV1},
			event: cloudevents.Event{
				Context: cloudevents.EventContextV1{
					Type:   "com.example.test",
					Source: *source,
					ID:     "ABC-123",
				}.AsV1(),
			},
			want: &pubsub.Message{
				Attributes: map[string]string{
					"Content-Type": cloudevents.ApplicationCloudEventsJSON,
				},
				Data: func() []byte {
					body := map[string]interface{}{
						"specversion": "1.0",
						"id":          "ABC-123",
						"type":        "com.example.test",
						"source":      "http://example.com/source",
					}
					return toBytes(body)
				}(),
			},
		},
		"full v1.0 structured": {
			codec: pubsub.CodecV1{DefaultEncoding: pubsub.StructuredV1},
			event: cloudevents.Event{
				Context: cloudevents.EventContextV1{
					ID:              "ABC-123",
					Time:            &now,
					Type:            "com.example.test",
					DataSchema:      schema,
					DataContentType: cloudevents.StringOfApplicationJSON(),
					Source:          *source,
					Extensions: map[string]interface{}{
						"test": "extended",
					},
				}.AsV1(),
				Data: map[string]interface{}{
					"hello": "world",
				},
			},
			want: &pubsub.Message{
				Attributes: map[string]string{
					"Content-Type": cloudevents.ApplicationCloudEventsJSON,
				},
				Data: func() []byte {
					body := map[string]interface{}{
						"specversion":     "1.0",
						"datacontenttype": "application/json",
						"data": map[string]interface{}{
							"hello": "world",
						},
						"id":         "ABC-123",
						"time":       now,
						"type":       "com.example.test",
						"test":       "extended",
						"dataschema": "http://example.com/schema",
						"source":     "http://example.com/source",
					}
					return toBytes(body)
				}(),
			},
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			got, err := tc.codec.Encode(context.TODO(), tc.event)

			if tc.wantErr != nil || err != nil {
				if diff := cmp.Diff(tc.wantErr, err); diff != "" {
					t.Errorf("unexpected error (-want, +got) = %v", diff)
				}
				return
			}

			if diff := cmp.Diff(tc.want, got); diff != "" {

				if msg, ok := got.(*pubsub.Message); ok {
					// It is hard to read the byte dump
					want := string(tc.want.Data)
					got := string(msg.Data)
					if diff := cmp.Diff(want, got); diff != "" {
						t.Errorf("unexpected message body (-want, +got) = %v", diff)
						return
					}
				}

				t.Errorf("unexpected message (-want, +got) = %v", diff)
			}
		})
	}
}

func TestCodecV1_Decode(t *testing.T) {
	now := types.Timestamp{Time: time.Now()}
	sourceUrl, _ := url.Parse("http://example.com/source")
	source := &types.URIRef{URL: *sourceUrl}

	schemaUrl, _ := url.Parse("http://example.com/schema")
	schema := &types.URI{URL: *schemaUrl}

	testCases := map[string]struct {
		codec   pubsub.CodecV1
		msg     *pubsub.Message
		want    *cloudevents.Event
		wantErr error
	}{
		"simple v1.0 structured": {
			codec: pubsub.CodecV1{},
			msg: &pubsub.Message{
				Attributes: map[string]string{
					"Content-Type": cloudevents.ApplicationCloudEventsJSON,
				},
				Data: toBytes(map[string]interface{}{
					"specversion": "1.0",
					"id":          "ABC-123",
					"type":        "com.example.test",
					"source":      "http://example.com/source",
				}),
			},
			want: &cloudevents.Event{
				Context: &cloudevents.EventContextV1{
					SpecVersion: cloudevents.CloudEventsVersionV1,
					Type:        "com.example.test",
					Source:      *source,
					ID:          "ABC-123",
				},
			},
		},
		"full v1.0 structured": {
			codec: pubsub.CodecV1{},
			msg: &pubsub.Message{
				Attributes: map[string]string{
					"Content-Type": cloudevents.ApplicationCloudEventsJSON,
				},
				Data: toBytes(map[string]interface{}{
					"specversion":     "1.0",
					"datacontenttype": "application/json",
					"data": map[string]interface{}{
						"hello": "world",
					},
					"id":         "ABC-123",
					"time":       now,
					"type":       "com.example.test",
					"test":       "extended",
					"generation": "1579743478182200",
					"dataschema": "http://example.com/schema",
					"source":     "http://example.com/source",
				}),
			},
			want: &cloudevents.Event{
				Context: &cloudevents.EventContextV1{
					SpecVersion:     cloudevents.CloudEventsVersionV1,
					ID:              "ABC-123",
					Time:            &now,
					Type:            "com.example.test",
					DataSchema:      schema,
					DataContentType: cloudevents.StringOfApplicationJSON(),
					Source:          *source,
					Extensions: map[string]interface{}{
						"test":       "extended",
						"generation": "1579743478182200",
					},
				},
				Data: toBytes(map[string]interface{}{
					"hello": "world",
				}),
				DataEncoded: true,
			},
		},
		"full v1.0 binary": {
			codec: pubsub.CodecV1{},
			msg: &pubsub.Message{
				Attributes: map[string]string{
					prefix + "specversion": "1.0",
					prefix + "id":          "ABC-123",
					prefix + "time":        now.String(),
					prefix + "type":        "com.example.test",
					prefix + "dataschema":  "http://example.com/schema",
					prefix + "source":      "http://example.com/source",
					prefix + "test":        "extended",
					prefix + "generation":  "1579743478182200",
				},
				Data: toBytes(map[string]interface{}{
					"hello": "world",
				}),
			},
			want: &cloudevents.Event{
				Context: &cloudevents.EventContextV1{
					SpecVersion: cloudevents.CloudEventsVersionV1,
					ID:          "ABC-123",
					Time:        &now,
					Type:        "com.example.test",
					DataSchema:  schema,
					Source:      *source,
					Extensions: map[string]interface{}{
						"test":       "extended",
						"generation": "1579743478182200",
					},
				},
				Data: toBytes(map[string]interface{}{
					"hello": "world",
				}),
				DataEncoded: true,
			},
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			got, err := tc.codec.Decode(context.TODO(), tc.msg)

			if tc.wantErr != nil || err != nil {
				if diff := cmp.Diff(tc.wantErr, err); diff != "" {
					t.Errorf("unexpected error (-want, +got) = %v", diff)
				}
				return
			}

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("unexpected event (-want, +got) = %v", diff)
			}
		})
	}
}
