package pubsub_test

import (
	"context"
	"encoding/json"
	"net/url"
	"testing"
	"time"

	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport/pubsub"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/types"
	"github.com/google/go-cmp/cmp"
)

func TestCodecV03_Encode(t *testing.T) {
	now := types.Timestamp{Time: time.Now().UTC()}
	sourceUrl, _ := url.Parse("http://example.com/source")
	source := &types.URLRef{URL: *sourceUrl}

	schemaUrl, _ := url.Parse("http://example.com/schema")
	schema := &types.URLRef{URL: *schemaUrl}

	testCases := map[string]struct {
		codec   pubsub.CodecV03
		event   cloudevents.Event
		want    *pubsub.Message
		wantErr error
	}{
		"simple v0.3 default": {
			codec: pubsub.CodecV03{},
			event: cloudevents.Event{
				Context: cloudevents.EventContextV03{
					Type:   "com.example.test",
					Source: *source,
					ID:     "ABC-123",
				}.AsV03(),
			},
			want: &pubsub.Message{
				Attributes: map[string]string{
					"Content-Type": cloudevents.ApplicationCloudEventsJSON,
				},
				Data: func() []byte {
					body := map[string]interface{}{
						"datacontenttype": "application/json",
						"specversion":     "0.3",
						"id":              "ABC-123",
						"type":            "com.example.test",
						"source":          "http://example.com/source",
					}
					return toBytes(body)
				}(),
			},
		},
		"full v0.3 default": {
			codec: pubsub.CodecV03{},
			event: cloudevents.Event{
				Context: cloudevents.EventContextV03{
					ID:              "ABC-123",
					Time:            &now,
					Type:            "com.example.test",
					SchemaURL:       schema,
					DataContentType: cloudevents.StringOfApplicationJSON(),
					Source:          *source,
					Extensions: map[string]interface{}{
						"test": "extended",
					},
				}.AsV03(),
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
						"specversion":     "0.3",
						"datacontenttype": "application/json",
						"data": map[string]interface{}{
							"hello": "world",
						},
						"id":        "ABC-123",
						"time":      now,
						"type":      "com.example.test",
						"test":      "extended",
						"schemaurl": "http://example.com/schema",
						"source":    "http://example.com/source",
					}
					return toBytes(body)
				}(),
			},
		},
		"simple v0.3 structured": {
			codec: pubsub.CodecV03{Encoding: pubsub.StructuredV03},
			event: cloudevents.Event{
				Context: cloudevents.EventContextV03{
					Type:   "com.example.test",
					Source: *source,
					ID:     "ABC-123",
				}.AsV03(),
			},
			want: &pubsub.Message{
				Attributes: map[string]string{
					"Content-Type": cloudevents.ApplicationCloudEventsJSON,
				},
				Data: func() []byte {
					body := map[string]interface{}{
						"datacontenttype": "application/json",
						"specversion":     "0.3",
						"id":              "ABC-123",
						"type":            "com.example.test",
						"source":          "http://example.com/source",
					}
					return toBytes(body)
				}(),
			},
		},
		"full v0.3 structured": {
			codec: pubsub.CodecV03{Encoding: pubsub.StructuredV03},
			event: cloudevents.Event{
				Context: cloudevents.EventContextV03{
					ID:              "ABC-123",
					Time:            &now,
					Type:            "com.example.test",
					SchemaURL:       schema,
					DataContentType: cloudevents.StringOfApplicationJSON(),
					Source:          *source,
					Extensions: map[string]interface{}{
						"test": "extended",
					},
				}.AsV03(),
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
						"specversion":     "0.3",
						"datacontenttype": "application/json",
						"data": map[string]interface{}{
							"hello": "world",
						},
						"id":        "ABC-123",
						"time":      now,
						"type":      "com.example.test",
						"test":      "extended",
						"schemaurl": "http://example.com/schema",
						"source":    "http://example.com/source",
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

// TODO: figure out extensions for v0.3

func TestCodecV03_Decode(t *testing.T) {
	now := types.Timestamp{Time: time.Now()}
	sourceUrl, _ := url.Parse("http://example.com/source")
	source := &types.URLRef{URL: *sourceUrl}

	schemaUrl, _ := url.Parse("http://example.com/schema")
	schema := &types.URLRef{URL: *schemaUrl}

	testCases := map[string]struct {
		codec   pubsub.CodecV03
		msg     *pubsub.Message
		want    *cloudevents.Event
		wantErr error
	}{
		"simple v0.3 structured": {
			codec: pubsub.CodecV03{},
			msg: &pubsub.Message{
				Attributes: map[string]string{
					"Content-Type": cloudevents.ApplicationCloudEventsJSON,
				},
				Data: toBytes(map[string]interface{}{
					"specversion": "0.3",
					"id":          "ABC-123",
					"type":        "com.example.test",
					"source":      "http://example.com/source",
				}),
			},
			want: &cloudevents.Event{
				Context: &cloudevents.EventContextV03{
					SpecVersion: cloudevents.CloudEventsVersionV03,
					Type:        "com.example.test",
					Source:      *source,
					ID:          "ABC-123",
				},
			},
		},
		"full v0.3 structured": {
			codec: pubsub.CodecV03{},
			msg: &pubsub.Message{
				Attributes: map[string]string{
					"Content-Type": cloudevents.ApplicationCloudEventsJSON,
				},
				Data: toBytes(map[string]interface{}{
					"specversion":     "0.3",
					"datacontenttype": "application/json",
					"data": map[string]interface{}{
						"hello": "world",
					},
					"id":        "ABC-123",
					"time":      now,
					"type":      "com.example.test",
					"test":      "extended",
					"schemaurl": "http://example.com/schema",
					"source":    "http://example.com/source",
				}),
			},
			want: &cloudevents.Event{
				Context: &cloudevents.EventContextV03{
					SpecVersion:     cloudevents.CloudEventsVersionV03,
					ID:              "ABC-123",
					Time:            &now,
					Type:            "com.example.test",
					SchemaURL:       schema,
					DataContentType: cloudevents.StringOfApplicationJSON(),
					Source:          *source,
					Extensions: map[string]interface{}{
						"test": json.RawMessage(`"extended"`),
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
