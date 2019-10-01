package nats_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport/nats"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/types"
	"github.com/google/go-cmp/cmp"
)

func TestCodecEncode(t *testing.T) {
	sourceUrl, _ := url.Parse("http://example.com/source")
	source := &types.URLRef{URL: *sourceUrl}

	testCases := map[string]struct {
		codec   nats.Codec
		event   cloudevents.Event
		want    *nats.Message
		wantErr error
	}{
		"simple v02 structured binary": {
			codec: nats.Codec{Encoding: nats.StructuredV02},
			event: cloudevents.Event{
				Context: cloudevents.EventContextV02{
					Type:   "com.example.test",
					Source: *source,
					ID:     "ABC-123",
				}.AsV02(),
			},
			want: &nats.Message{
				Body: func() []byte {
					body := map[string]interface{}{
						"specversion": "0.2",
						"id":          "ABC-123",
						"type":        "com.example.test",
						"source":      "http://example.com/source",
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
				if msg, ok := got.(*nats.Message); ok {
					// It is hard to read the byte dump
					want := string(tc.want.Body)
					got := string(msg.Body)
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
		codec   nats.Codec
		msg     *nats.Message
		want    *cloudevents.Event
		wantErr error
	}{
		"simple v2 structured": {
			codec: nats.Codec{Encoding: nats.StructuredV02},
			msg: &nats.Message{
				Body: func() []byte {
					body := map[string]interface{}{
						"specversion": "0.2",
						"id":          "ABC-123",
						"type":        "com.example.test",
						"source":      "http://example.com/source",
					}
					return toBytes(body)
				}(),
			},
			want: &cloudevents.Event{
				Context: &cloudevents.EventContextV02{
					SpecVersion: cloudevents.CloudEventsVersionV02,
					Type:        "com.example.test",
					Source:      *source,
					ID:          "ABC-123",
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

func TestCodecRoundTrip(t *testing.T) {
	sourceUrl, _ := url.Parse("http://example.com/source")
	source := &types.URLRef{URL: *sourceUrl}

	for _, encoding := range []nats.Encoding{nats.StructuredV02} {

		testCases := map[string]struct {
			codec   nats.Codec
			event   cloudevents.Event
			want    cloudevents.Event
			wantErr error
		}{
			"simple data": {
				codec: nats.Codec{Encoding: encoding},
				event: cloudevents.Event{
					Context: cloudevents.EventContextV01{
						EventType: "com.example.test",
						Source:    *source,
						EventID:   "ABC-123",
					}.AsV02(),
					Data: map[string]string{
						"a": "apple",
						"b": "banana",
					},
				},
				want: cloudevents.Event{
					Context: &cloudevents.EventContextV01{
						CloudEventsVersion: cloudevents.CloudEventsVersionV01,
						EventType:          "com.example.test",
						Source:             *source,
						EventID:            "ABC-123",
					},
					Data: map[string]interface{}{
						"a": "apple",
						"b": "banana",
					},
					DataEncoded: true,
				},
			},
			"struct data": {
				codec: nats.Codec{Encoding: encoding},
				event: cloudevents.Event{
					Context: cloudevents.EventContextV01{
						EventType: "com.example.test",
						Source:    *source,
						EventID:   "ABC-123",
					}.AsV02(),
					Data: DataExample{
						AnInt:   42,
						AString: "testing",
					},
				},
				want: cloudevents.Event{
					Context: &cloudevents.EventContextV01{
						CloudEventsVersion: cloudevents.CloudEventsVersionV01,
						EventType:          "com.example.test",
						Source:             *source,
						EventID:            "ABC-123",
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

				msg, err := tc.codec.Encode(context.TODO(), tc.event)
				if err != nil {
					if diff := cmp.Diff(tc.wantErr, err); diff != "" {
						t.Errorf("unexpected error (-want, +got) = %v", diff)
					}
					return
				}

				got, err := tc.codec.Decode(context.TODO(), msg)
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

				if tc.wantErr != nil {
					if diff := cmp.Diff(tc.wantErr, err); diff != "" {
						t.Errorf("unexpected error (-want, +got) = %v", diff)
					}
					return
				}

				// fix the context back to v01 to test.
				ctxv1 := got.Context.AsV01()
				got.Context = ctxv1

				if diff := cmp.Diff(tc.want, *got); diff != "" {
					t.Errorf("unexpected event (-want, +got) = %v", diff)
				}
			})
		}

	}
}

func TestCodecAsMiddleware(t *testing.T) {
	sourceUrl, _ := url.Parse("http://example.com/source")
	source := &types.URLRef{URL: *sourceUrl}

	for _, encoding := range []nats.Encoding{nats.StructuredV02} {

		testCases := map[string]struct {
			codec   nats.Codec
			event   cloudevents.Event
			want    cloudevents.Event
			wantErr error
		}{
			"simple data": {
				codec: nats.Codec{Encoding: encoding},
				event: cloudevents.Event{
					Context: cloudevents.EventContextV01{
						EventType: "com.example.test",
						Source:    *source,
						EventID:   "ABC-123",
					}.AsV02(),
					Data: map[string]string{
						"a": "apple",
						"b": "banana",
					},
				},
				want: cloudevents.Event{
					Context: &cloudevents.EventContextV01{
						CloudEventsVersion: cloudevents.CloudEventsVersionV01,
						EventType:          "com.example.test",
						Source:             *source,
						EventID:            "ABC-123",
					},
					Data: map[string]interface{}{
						"a": "apple",
						"b": "banana",
					},
					DataEncoded: true,
				},
			},
			"struct data": {
				codec: nats.Codec{Encoding: encoding},
				event: cloudevents.Event{
					Context: cloudevents.EventContextV01{
						EventType: "com.example.test",
						Source:    *source,
						EventID:   "ABC-123",
					}.AsV02(),
					Data: DataExample{
						AnInt:   42,
						AString: "testing",
					},
				},
				want: cloudevents.Event{
					Context: &cloudevents.EventContextV01{
						CloudEventsVersion: cloudevents.CloudEventsVersionV01,
						EventType:          "com.example.test",
						Source:             *source,
						EventID:            "ABC-123",
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

				msg1, err := tc.codec.Encode(context.TODO(), tc.event)
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

				if tc.wantErr != nil {
					if diff := cmp.Diff(tc.wantErr, err); diff != "" {
						t.Errorf("unexpected error (-want, +got) = %v", diff)
					}
					return
				}

				// fix the context back to v1 to test.
				ctxv1 := got.Context.AsV01()
				got.Context = ctxv1

				if diff := cmp.Diff(tc.want, *got); diff != "" {
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
