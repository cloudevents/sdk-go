package cloudevents_test

import (
	"encoding/json"
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/types"
)

//type DataExample struct {
//	AnInt   int        `json:"a,omitempty"`
//	AString string     `json:"b,omitempty"`
//	AnArray []string   `json:"c,omitempty"`
//	ATime   *time.Time `json:"e,omitempty"`
//}

func TestMarshal(t *testing.T) {
	now := types.Timestamp{Time: time.Now().UTC()}
	sourceUrl, _ := url.Parse("http://example.com/source")
	source := &types.URLRef{URL: *sourceUrl}

	schemaUrl, _ := url.Parse("http://example.com/schema")
	schema := &types.URLRef{URL: *schemaUrl}

	testCases := map[string]struct {
		event   cloudevents.Event
		want    []byte
		wantErr *string
	}{
		"empty struct": {
			event:   cloudevents.Event{},
			wantErr: strptr("json: error calling MarshalJSON for type cloudevents.Event: every event conforming to the CloudEvents specification MUST include a context"),
		},
		"struct data v0.1": {
			event: cloudevents.Event{
				Context: cloudevents.EventContextV01{
					EventType:        "com.example.test",
					Source:           *source,
					SchemaURL:        schema,
					EventTypeVersion: strptr("version1"),
					EventID:          "ABC-123",
					EventTime:        &now,
					ContentType:      cloudevents.StringOfApplicationJSON(),
					Extensions: map[string]interface{}{
						"ex1": 42,
						"ex2": "testing",
						"ex3": map[string]interface{}{
							"an": "inner key",
						},
					},
				}.AsV01(),
				Data: DataExample{
					AnInt:   42,
					AString: "testing",
				},
			},
			want: toBytes(map[string]interface{}{
				"cloudEventsVersion": "0.1",
				"contentType":        "application/json",
				"data": map[string]interface{}{
					"a": 42,
					"b": "testing",
				},
				"eventID":          "ABC-123",
				"eventTime":        now.Format(time.RFC3339Nano),
				"eventType":        "com.example.test",
				"eventTypeVersion": "version1",
				"extensions": map[string]interface{}{
					"ex1": 42,
					"ex2": "testing",
					"ex3": map[string]interface{}{
						"an": "inner key",
					},
				},
				"schemaURL": "http://example.com/schema",
				"source":    "http://example.com/source",
			}),
		},
		"struct data v0.2": {
			event: cloudevents.Event{
				Context: cloudevents.EventContextV02{
					Type:        "com.example.test",
					Source:      *source,
					SchemaURL:   schema,
					ID:          "ABC-123",
					Time:        &now,
					ContentType: cloudevents.StringOfApplicationJSON(),
					Extensions: map[string]interface{}{
						"ex1": 42,
						"ex2": "testing",
						"ex3": map[string]interface{}{
							"an": "inner key",
						},
					},
				}.AsV02(),
				Data: DataExample{
					AnInt:   42,
					AString: "testing",
				},
			},
			want: toBytes(map[string]interface{}{
				"specversion": "0.2",
				"contenttype": "application/json",
				"data": map[string]interface{}{
					"a": 42,
					"b": "testing",
				},
				"id":   "ABC-123",
				"time": now.Format(time.RFC3339Nano),
				"type": "com.example.test",
				"ex1":  42,
				"ex2":  "testing",
				"ex3": map[string]interface{}{
					"an": "inner key",
				},
				"schemaurl": "http://example.com/schema",
				"source":    "http://example.com/source",
			}),
		},
		"struct data v0.3": {
			event: cloudevents.Event{
				Context: cloudevents.EventContextV03{
					Type:            "com.example.test",
					Source:          *source,
					SchemaURL:       schema,
					ID:              "ABC-123",
					Time:            &now,
					DataContentType: cloudevents.StringOfApplicationJSON(),
					Extensions: map[string]interface{}{
						"ex1": 42,
						"ex2": "testing",
						"ex3": map[string]interface{}{
							"an": "inner key",
						},
					},
				}.AsV03(),
				Data: DataExample{
					AnInt:   42,
					AString: "testing",
				},
			},
			want: toBytes(map[string]interface{}{
				"specversion":     "0.3",
				"datacontenttype": "application/json",
				"data": map[string]interface{}{
					"a": 42,
					"b": "testing",
				},
				"id":   "ABC-123",
				"time": now.Format(time.RFC3339Nano),
				"type": "com.example.test",
				"ex1":  42,
				"ex2":  "testing",
				"ex3": map[string]interface{}{
					"an": "inner key",
				},
				"schemaurl": "http://example.com/schema",
				"source":    "http://example.com/source",
			}),
		},
		"nil data v0.3": {
			event: cloudevents.Event{
				Context: cloudevents.EventContextV03{
					Type:            "com.example.test",
					Source:          *source,
					SchemaURL:       schema,
					ID:              "ABC-123",
					Time:            &now,
					DataContentType: cloudevents.StringOfApplicationJSON(),
					Extensions: map[string]interface{}{
						"ex1": 42,
						"ex2": "testing",
						"ex3": map[string]interface{}{
							"an": "inner key",
						},
					},
				}.AsV03(),
			},
			want: toBytes(map[string]interface{}{
				"specversion":     "0.3",
				"datacontenttype": "application/json",
				"id":              "ABC-123",
				"time":            now.Format(time.RFC3339Nano),
				"type":            "com.example.test",
				"ex1":             42,
				"ex2":             "testing",
				"ex3": map[string]interface{}{
					"an": "inner key",
				},
				"schemaurl": "http://example.com/schema",
				"source":    "http://example.com/source",
			}),
		},
		"string data v0.3": {
			event: cloudevents.Event{
				Context: cloudevents.EventContextV03{
					Type:            "com.example.test",
					Source:          *source,
					SchemaURL:       schema,
					ID:              "ABC-123",
					Time:            &now,
					DataContentType: cloudevents.StringOfApplicationJSON(),
					Extensions: map[string]interface{}{
						"ex1": 42,
						"ex2": "testing",
						"ex3": map[string]interface{}{
							"an": "inner key",
						},
					},
				}.AsV03(),
				Data: "This is a string.",
			},
			want: toBytes(map[string]interface{}{
				"specversion":     "0.3",
				"datacontenttype": "application/json",
				"data":            "This is a string.",
				"id":              "ABC-123",
				"time":            now.Format(time.RFC3339Nano),
				"type":            "com.example.test",
				"ex1":             42,
				"ex2":             "testing",
				"ex3": map[string]interface{}{
					"an": "inner key",
				},
				"schemaurl": "http://example.com/schema",
				"source":    "http://example.com/source",
			}),
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			gotBytes, err := json.Marshal(tc.event)

			if tc.wantErr != nil || err != nil {
				if diff := cmp.Diff(*tc.wantErr, err.Error()); diff != "" {
					t.Errorf("unexpected error (-want, +got) = %v", diff)
				}
				return
			}

			// so we can understand the diff, turn bytes to strings
			want := string(tc.want)
			got := string(gotBytes)

			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("unexpected event (-want, +got) = %v", diff)
			}
		})
	}
}

func TestUnmarshal(t *testing.T) {
	now := types.Timestamp{Time: time.Now().UTC()}
	sourceUrl, _ := url.Parse("http://example.com/source")
	source := &types.URLRef{URL: *sourceUrl}

	schemaUrl, _ := url.Parse("http://example.com/schema")
	schema := &types.URLRef{URL: *schemaUrl}

	testCases := map[string]struct {
		body    []byte
		want    *cloudevents.Event
		wantErr error
	}{
		"struct data v0.1": {
			body: toBytes(map[string]interface{}{
				"cloudEventsVersion": "0.1",
				"contentType":        "application/json",
				"data": map[string]interface{}{
					"a": 42,
					"b": "testing",
				},
				"eventID":          "ABC-123",
				"eventTime":        now.Format(time.RFC3339Nano),
				"eventType":        "com.example.test",
				"eventTypeVersion": "version1",
				"extensions": map[string]interface{}{
					"ex1": 42,
					"ex2": "testing",
					"ex3": map[string]interface{}{
						"an": "inner key",
					},
				},
				"schemaURL": "http://example.com/schema",
				"source":    "http://example.com/source",
			}),
			want: &cloudevents.Event{
				Context: cloudevents.EventContextV01{
					EventType:        "com.example.test",
					Source:           *source,
					SchemaURL:        schema,
					EventTypeVersion: strptr("version1"),
					EventID:          "ABC-123",
					EventTime:        &now,
					ContentType:      cloudevents.StringOfApplicationJSON(),
					Extensions: map[string]interface{}{
						"ex1": float64(42), // json auto-creates float64 from int.
						"ex2": "testing",
						"ex3": map[string]interface{}{
							"an": "inner key",
						},
					},
				}.AsV01(),
				Data: toBytes(DataExample{
					AnInt:   42,
					AString: "testing",
				}),
				DataEncoded: true,
			},
		},
		"struct data v0.2": {
			body: toBytes(map[string]interface{}{
				"specversion": "0.2",
				"contenttype": "application/json",
				"data": map[string]interface{}{
					"a": 42,
					"b": "testing",
				},
				"id":   "ABC-123",
				"time": now.Format(time.RFC3339Nano),
				"type": "com.example.test",
				"ex1":  42,
				"ex2":  "testing",
				"ex3": map[string]interface{}{
					"an": "inner key",
				},
				"schemaurl": "http://example.com/schema",
				"source":    "http://example.com/source",
			}),
			want: &cloudevents.Event{
				Context: cloudevents.EventContextV02{
					Type:        "com.example.test",
					Source:      *source,
					SchemaURL:   schema,
					ID:          "ABC-123",
					Time:        &now,
					ContentType: cloudevents.StringOfApplicationJSON(),
					Extensions: map[string]interface{}{
						"ex1": toRawMessage(float64(42)),
						"ex2": toRawMessage("testing"),
						"ex3": toRawMessage(map[string]interface{}{
							"an": "inner key",
						}),
					},
				}.AsV02(),
				Data: toBytes(DataExample{
					AnInt:   42,
					AString: "testing",
				}),
				DataEncoded: true,
			},
		},
		"struct data v0.3": {
			body: toBytes(map[string]interface{}{
				"specversion":     "0.3",
				"datacontenttype": "application/json",
				"data": map[string]interface{}{
					"a": 42,
					"b": "testing",
				},
				"id":   "ABC-123",
				"time": now.Format(time.RFC3339Nano),
				"type": "com.example.test",
				"ex1":  42,
				"ex2":  "testing",
				"ex3": map[string]interface{}{
					"an": "inner key",
				},
				"schemaurl": "http://example.com/schema",
				"source":    "http://example.com/source",
			}),
			want: &cloudevents.Event{
				Context: cloudevents.EventContextV03{
					Type:            "com.example.test",
					Source:          *source,
					SchemaURL:       schema,
					ID:              "ABC-123",
					Time:            &now,
					DataContentType: cloudevents.StringOfApplicationJSON(),
					Extensions: map[string]interface{}{
						"ex1": toRawMessage(float64(42)),
						"ex2": toRawMessage("testing"),
						"ex3": toRawMessage(map[string]interface{}{
							"an": "inner key",
						}),
					},
				}.AsV03(),
				Data: toBytes(DataExample{
					AnInt:   42,
					AString: "testing",
				}),
				DataEncoded: true,
			},
		},
		"string data v0.3": {
			body: toBytes(map[string]interface{}{
				"specversion":     "0.3",
				"datacontenttype": "application/json",
				"data":            "This is a string.",
				"id":              "ABC-123",
				"time":            now.Format(time.RFC3339Nano),
				"type":            "com.example.test",
				"ex1":             42,
				"ex2":             "testing",
				"ex3": map[string]interface{}{
					"an": "inner key",
				},
				"schemaurl": "http://example.com/schema",
				"source":    "http://example.com/source",
			}),
			want: &cloudevents.Event{
				Context: cloudevents.EventContextV03{
					Type:            "com.example.test",
					Source:          *source,
					SchemaURL:       schema,
					ID:              "ABC-123",
					Time:            &now,
					DataContentType: cloudevents.StringOfApplicationJSON(),
					Extensions: map[string]interface{}{
						"ex1": toRawMessage(float64(42)),
						"ex2": toRawMessage("testing"),
						"ex3": toRawMessage(map[string]interface{}{
							"an": "inner key",
						}),
					},
				}.AsV03(),
				Data:        toBytes("This is a string."),
				DataEncoded: true,
			},
		},
		"nil data v0.3": {
			body: toBytes(map[string]interface{}{
				"specversion":     "0.3",
				"datacontenttype": "application/json",
				"id":              "ABC-123",
				"time":            now.Format(time.RFC3339Nano),
				"type":            "com.example.test",
				"ex1":             42,
				"ex2":             "testing",
				"ex3": map[string]interface{}{
					"an": "inner key",
				},
				"schemaurl": "http://example.com/schema",
				"source":    "http://example.com/source",
			}),
			want: &cloudevents.Event{
				Context: cloudevents.EventContextV03{
					Type:            "com.example.test",
					Source:          *source,
					SchemaURL:       schema,
					ID:              "ABC-123",
					Time:            &now,
					DataContentType: cloudevents.StringOfApplicationJSON(),
					Extensions: map[string]interface{}{
						"ex1": toRawMessage(float64(42)),
						"ex2": toRawMessage("testing"),
						"ex3": toRawMessage(map[string]interface{}{
							"an": "inner key",
						}),
					},
				}.AsV03(),
			},
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			got := &cloudevents.Event{}
			err := json.Unmarshal(tc.body, got)

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

func TestUnmarshal_nilEvent(t *testing.T) {
	wantErr := "json: Unmarshal(nil)"

	err := json.Unmarshal(toBytes("{}"), nil)

	if err != nil {
		if diff := cmp.Diff(wantErr, err.Error()); diff != "" {
			t.Errorf("unexpected error (-want, +got) = %v", diff)
		}
		return
	}
}

func toBytes(body interface{}) []byte {
	b, err := json.Marshal(body)
	if err != nil {
		return []byte(fmt.Sprintf(`{"error":%q}`, err.Error()))
	}
	return b
}

func toRawMessage(body interface{}) json.RawMessage {
	return toBytes(body)
}
