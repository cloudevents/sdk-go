package pubsub_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/cloudevents/sdk-go/v1/cloudevents"
	"github.com/cloudevents/sdk-go/v1/cloudevents/transport/pubsub"
	"github.com/cloudevents/sdk-go/v1/cloudevents/types"
	"github.com/google/go-cmp/cmp"
)

const (
	prefix = "ce-"
)

func strptr(s string) *string {
	return &s
}

func TestCodecEncode(t *testing.T) {
	sourceUrl, _ := url.Parse("http://example.com/source")
	source := &types.URLRef{URL: *sourceUrl}

	testCases := map[string]struct {
		codec   *pubsub.Codec
		event   *cloudevents.Event
		want    *pubsub.Message
		wantErr error
	}{
		"simple v03 structured": {
			codec: &pubsub.Codec{Encoding: pubsub.StructuredV03},
			event: &cloudevents.Event{
				Context: cloudevents.EventContextV03{
					Type:    "com.example.test",
					Source:  *source,
					ID:      "ABC-123",
					Subject: strptr("a-subject"),
				}.AsV03(),
			},
			want: &pubsub.Message{
				Attributes: map[string]string{
					"Content-Type": cloudevents.ApplicationCloudEventsJSON,
				},
				Data: func() []byte {
					body := map[string]interface{}{
						"specversion": "0.3",
						"id":          "ABC-123",
						"type":        "com.example.test",
						"source":      "http://example.com/source",
						"subject":     "a-subject",
					}
					return toBytes(body)
				}(),
			},
		},
		"simple v03 binary": {
			codec: &pubsub.Codec{Encoding: pubsub.BinaryV03},
			event: &cloudevents.Event{
				Context: &cloudevents.EventContextV03{
					Type:    "com.example.test",
					Source:  *source,
					ID:      "ABC-123",
					Subject: strptr("a-subject"),
				},
			},
			want: &pubsub.Message{
				Attributes: map[string]string{
					"ce-specversion":     "0.3",
					"ce-id":              "ABC-123",
					"ce-type":            "com.example.test",
					"ce-source":          "http://example.com/source",
					"ce-subject":         "a-subject",
					"ce-datacontenttype": "application/json",
				},
			},
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			got, err := tc.codec.Encode(context.TODO(), *tc.event)

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
						t.Errorf("unexpected (-want, +got) = %v", diff)
						return
					}
				}
				t.Errorf("unexpected message (-want, +got) = %v", diff)
			}
		})
	}
}

func TestCodecDecode(t *testing.T) {
	sourceUrl, _ := url.Parse("http://example.com/source")
	source := &types.URLRef{URL: *sourceUrl}

	testCases := map[string]struct {
		codec   *pubsub.Codec
		msg     *pubsub.Message
		want    *cloudevents.Event
		wantErr error
	}{
		"simple v3 structured": {
			codec: &pubsub.Codec{Encoding: pubsub.StructuredV03},
			msg: &pubsub.Message{
				Attributes: map[string]string{
					"Content-Type": cloudevents.ApplicationCloudEventsJSON,
				},
				Data: func() []byte {
					body := map[string]interface{}{
						"specversion": "0.3",
						"id":          "ABC-123",
						"type":        "com.example.test",
						"source":      "http://example.com/source",
						"subject":     "a-subject",
					}
					return toBytes(body)
				}(),
			},
			want: &cloudevents.Event{
				Context: &cloudevents.EventContextV03{
					SpecVersion: cloudevents.CloudEventsVersionV03,
					Type:        "com.example.test",
					Source:      *source,
					ID:          "ABC-123",
					Subject:     strptr("a-subject"),
				},
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

type DataExample struct {
	AnInt   int                       `json:"a,omitempty" xml:"a,omitempty"`
	AString string                    `json:"b,omitempty" xml:"b,omitempty"`
	AnArray []string                  `json:"c,omitempty" xml:"c,omitempty"`
	AMap    map[string]map[string]int `json:"d,omitempty" xml:"d,omitempty"`
	ATime   *time.Time                `json:"e,omitempty" xml:"e,omitempty"`
}

func TestCodecAsMiddleware(t *testing.T) {
	sourceUrl, _ := url.Parse("http://example.com/source")
	source := &types.URLRef{URL: *sourceUrl}

	for _, encoding := range []pubsub.Encoding{pubsub.StructuredV03} {

		testCases := map[string]struct {
			codec   *pubsub.Codec
			event   *cloudevents.Event
			want    *cloudevents.Event
			wantErr error
		}{
			"simple data": {
				codec: &pubsub.Codec{Encoding: encoding},
				event: &cloudevents.Event{
					Context: cloudevents.EventContextV03{
						Type:   "com.example.test",
						Source: *source,
						ID:     "ABC-123",
					}.AsV03(),
					Data: map[string]string{
						"a": "apple",
						"b": "banana",
					},
				},
				want: &cloudevents.Event{
					Context: &cloudevents.EventContextV03{
						SpecVersion: cloudevents.CloudEventsVersionV03,
						Type:        "com.example.test",
						Source:      *source,
						ID:          "ABC-123",
					},
					Data: map[string]interface{}{
						"a": "apple",
						"b": "banana",
					},
					DataEncoded: true,
				},
			},
			"struct data": {
				codec: &pubsub.Codec{Encoding: encoding},
				event: &cloudevents.Event{
					Context: cloudevents.EventContextV03{
						Type:   "com.example.test",
						Source: *source,
						ID:     "ABC-123",
					}.AsV03(),
					Data: DataExample{
						AnInt:   42,
						AString: "testing",
					},
				},
				want: &cloudevents.Event{
					Context: &cloudevents.EventContextV03{
						SpecVersion: cloudevents.CloudEventsVersionV03,
						Type:        "com.example.test",
						Source:      *source,
						ID:          "ABC-123",
					},
					Data: &DataExample{
						AnInt:   42,
						AString: "testing",
					},
					DataEncoded: true,
				},
			},
		}
		for n, tc := range testCases {
			n = fmt.Sprintf("%s, %s", encoding, n)
			t.Run(n, func(t *testing.T) {

				msg1, err := tc.codec.Encode(context.TODO(), *tc.event)
				if err != nil {
					if diff := cmp.Diff(tc.wantErr, err); diff != "" {
						t.Errorf("unexpected error (-want, +got) = %v", diff)
					}
					return
				}

				midEvent, err := tc.codec.Decode(context.TODO(), msg1)
				if err != nil {
					if diff := cmp.Diff(tc.wantErr, err); diff != "" {
						t.Errorf("unexpected error (-want, +got) = %v", diff)
					}
					return
				}

				msg2, err := tc.codec.Encode(context.TODO(), *midEvent)
				if err != nil {
					if diff := cmp.Diff(tc.wantErr, err); diff != "" {
						t.Errorf("unexpected error (-want, +got) = %v", diff)
					}
					return
				}

				got, err := tc.codec.Decode(context.TODO(), msg2)
				if err != nil {
					if diff := cmp.Diff(tc.wantErr, err); diff != "" {
						t.Errorf("unexpected error (-want, +got) = %v", diff)
					}
					return
				}

				if tc.event.Data != nil {
					// We have to be pretty tricky in the test to get the correct type.
					data, _ := types.Allocate(tc.want.Data)
					err := got.DataAs(&data)
					if err != nil {
						if diff := cmp.Diff(tc.wantErr, err); diff != "" {
							t.Errorf("unexpected error (-want, +got) = %v", diff)
						}
						return
					}
					got.Data = data
				}

				ctxv3 := got.Context.AsV03()
				got.Context = ctxv3

				if diff := cmp.Diff(tc.want, got); diff != "" {
					t.Errorf("unexpected event (-want, +got) = %v", diff)
				}
			})
		}
	}
}

func toBytes(body map[string]interface{}) []byte {
	b, err := json.Marshal(body)
	if err != nil {
		return []byte(fmt.Sprintf(`{"error":%q}`, err.Error()))
	}
	return b
}
