package nats_test

import (
	"encoding/json"
	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport/nats"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/types"
	"github.com/google/go-cmp/cmp"
	"net/url"
	"testing"
	"time"
)

func TestCodecV03_Encode(t *testing.T) {
	now := types.Timestamp{Time: time.Now().UTC()}
	sourceUrl, _ := url.Parse("http://example.com/source")
	source := &types.URLRef{URL: *sourceUrl}

	schemaUrl, _ := url.Parse("http://example.com/schema")
	schema := &types.URLRef{URL: *schemaUrl}

	testCases := map[string]struct {
		codec   nats.CodecV03
		event   cloudevents.Event
		want    *nats.Message
		wantErr error
	}{
		"simple v0.3 default": {
			codec: nats.CodecV03{},
			event: cloudevents.Event{
				Context: &cloudevents.EventContextV03{
					Type:   "com.example.test",
					Source: *source,
					ID:     "ABC-123",
				},
			},
			want: &nats.Message{
				Body: func() []byte {
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
			codec: nats.CodecV03{},
			event: cloudevents.Event{
				Context: &cloudevents.EventContextV03{
					ID:              "ABC-123",
					Time:            &now,
					Type:            "com.example.test",
					SchemaURL:       schema,
					DataContentType: cloudevents.StringOfApplicationJSON(),
					Source:          *source,
					Extensions: map[string]interface{}{
						"test": "extended",
					},
				},
				Data: map[string]interface{}{
					"hello": "world",
				},
			},
			want: &nats.Message{
				Body: func() []byte {
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
			codec: nats.CodecV03{Encoding: nats.StructuredV03},
			event: cloudevents.Event{
				Context: &cloudevents.EventContextV03{
					Type:   "com.example.test",
					Source: *source,
					ID:     "ABC-123",
				},
			},
			want: &nats.Message{
				Body: func() []byte {
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
			codec: nats.CodecV03{Encoding: nats.StructuredV03},
			event: cloudevents.Event{
				Context: &cloudevents.EventContextV03{
					ID:              "ABC-123",
					Time:            &now,
					Type:            "com.example.test",
					SchemaURL:       schema,
					DataContentType: cloudevents.StringOfApplicationJSON(),
					Source:          *source,
					Extensions: map[string]interface{}{
						"test": "extended",
					},
				},
				Data: map[string]interface{}{
					"hello": "world",
				},
			},
			want: &nats.Message{
				Body: func() []byte {
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

			got, err := tc.codec.Encode(tc.event)

			if tc.wantErr != nil || err != nil {
				if diff := cmp.Diff(tc.wantErr, err); diff != "" {
					t.Errorf("unexpected error (-want, +got) = %v", diff)
				}
				return
			}

			if diff := cmp.Diff(tc.want, got); diff != "" {

				if msg, ok := got.(*nats.Message); ok {
					// It is hard to read the byte dump
					want := string(tc.want.Body)
					got := string(msg.Body)
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
		codec   nats.CodecV03
		msg     *nats.Message
		want    *cloudevents.Event
		wantErr error
	}{
		"simple v0.3 structured": {
			codec: nats.CodecV03{},
			msg: &nats.Message{
				Body: toBytes(map[string]interface{}{
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
				DataEncoded: true,
			},
		},
		"full v0.3 structured": {
			codec: nats.CodecV03{},
			msg: &nats.Message{
				Body: toBytes(map[string]interface{}{
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

			got, err := tc.codec.Decode(tc.msg)

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
