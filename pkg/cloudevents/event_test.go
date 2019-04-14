package cloudevents_test

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	ce "github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/types"
	"github.com/google/go-cmp/cmp"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestGetDataContentType(t *testing.T) {
	now := types.Timestamp{Time: time.Now()}

	testCases := map[string]struct {
		event ce.Event
		want  string
	}{
		"min v01, blank": {
			event: ce.Event{
				Context: MinEventContextV01(),
			},
			want: "",
		},
		"full v01, json": {
			event: ce.Event{
				Context: FullEventContextV01(now),
			},
			want: "application/json",
		},
		"min v02, blank": {
			event: ce.Event{
				Context: MinEventContextV02(),
			},
			want: "",
		},
		"full v02, json": {
			event: ce.Event{
				Context: FullEventContextV02(now),
			},
			want: "application/json",
		},
		"min v03, blank": {
			event: ce.Event{
				Context: MinEventContextV03(),
			},
			want: "",
		},
		"full v03, json": {
			event: ce.Event{
				Context: FullEventContextV03(now),
			},
			want: "application/json",
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			got, _ := tc.event.Context.GetDataMediaType()

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("unexpected (-want, +got) = %v", diff)
			}
		})
	}
}

func TestSource(t *testing.T) {
	now := types.Timestamp{Time: time.Now()}

	source := "http://example.com/source"

	testCases := map[string]struct {
		event ce.Event
		want  string
	}{
		"min v01": {
			event: ce.Event{
				Context: MinEventContextV01(),
			},
			want: source,
		},
		"full v01": {
			event: ce.Event{
				Context: FullEventContextV01(now),
			},
			want: source,
		},
		"min v02": {
			event: ce.Event{
				Context: MinEventContextV02(),
			},
			want: source,
		},
		"full v02": {
			event: ce.Event{
				Context: FullEventContextV02(now),
			},
			want: source,
		},
		"min v03": {
			event: ce.Event{
				Context: MinEventContextV03(),
			},
			want: source,
		},
		"full v03": {
			event: ce.Event{
				Context: FullEventContextV03(now),
			},
			want: source,
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			got := tc.event.Source()

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("unexpected (-want, +got) = %v", diff)
			}
		})
	}
}

func TestSchemaURL(t *testing.T) {
	now := types.Timestamp{Time: time.Now()}

	schema := "http://example.com/schema"

	testCases := map[string]struct {
		event ce.Event
		want  string
	}{
		"min v01, empty schema": {
			event: ce.Event{
				Context: MinEventContextV01(),
			},
			want: "",
		},
		"full v01, schema": {
			event: ce.Event{
				Context: FullEventContextV01(now),
			},
			want: schema,
		},
		"min v02, empty schema": {
			event: ce.Event{
				Context: MinEventContextV02(),
			},
			want: "",
		},
		"full v02, schema": {
			event: ce.Event{
				Context: FullEventContextV02(now),
			},
			want: schema,
		},
		"min v03, empty schema": {
			event: ce.Event{
				Context: MinEventContextV03(),
			},
			want: "",
		},
		"full v03, schema": {
			event: ce.Event{
				Context: FullEventContextV03(now),
			},
			want: schema,
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			got := tc.event.SchemaURL()

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("unexpected (-want, +got) = %v", diff)
			}
		})
	}
}

type DataExample struct {
	AnInt   int                       `json:"a,omitempty"`
	AString string                    `json:"b,omitempty"`
	AnArray []string                  `json:"c,omitempty"`
	AMap    map[string]map[string]int `json:"d,omitempty"`
	ATime   *time.Time                `json:"e,omitempty"`
}

func TestDataAs(t *testing.T) {
	now := types.Timestamp{Time: time.Now()}

	testCases := map[string]struct {
		event   ce.Event
		want    interface{}
		wantErr error
	}{
		"empty": {
			event: ce.Event{
				Context: MinEventContextV01(),
			},
			want: nil,
		},
		"json simple": {
			event: ce.Event{
				Context:     FullEventContextV01(now),
				Data:        []byte(`eyJhIjoiYXBwbGUiLCJiIjoiYmFuYW5hIn0K`),
				DataEncoded: true,
			},
			want: &map[string]string{
				"a": "apple",
				"b": "banana",
			},
		},
		"json complex empty": {
			event: ce.Event{
				Context:     FullEventContextV01(now),
				Data:        []byte(`e30K`),
				DataEncoded: true,
			},
			want: &DataExample{},
		},
		"json complex filled": {
			event: ce.Event{
				Context: FullEventContextV01(now),
				Data: func() []byte {
					data := &DataExample{
						AnInt: 42,
						AMap: map[string]map[string]int{
							"a": {"1": 1, "2": 2, "3": 3},
							"z": {"3": 3, "2": 2, "1": 1},
						},
						AString: "Hello, World!",
						ATime:   &now.Time,
						AnArray: []string{"Anne", "Bob", "Chad"},
					}
					j, err := json.Marshal(data)
					if err != nil {
						t.Errorf("failed to marshal test data: %s", err.Error())
					}
					buf := make([]byte, base64.StdEncoding.EncodedLen(len(j)))
					base64.StdEncoding.Encode(buf, j)
					return buf
				}(),
				DataEncoded: true,
			},
			want: &DataExample{
				AnInt: 42,
				AMap: map[string]map[string]int{
					"a": {"1": 1, "2": 2, "3": 3},
					"z": {"3": 3, "2": 2, "1": 1},
				},
				AString: "Hello, World!",
				ATime:   &now.Time,
				AnArray: []string{"Anne", "Bob", "Chad"},
			},
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			got, _ := types.Allocate(tc.want)
			err := tc.event.DataAs(got)

			if tc.wantErr != nil || err != nil {
				if diff := cmp.Diff(tc.wantErr, err); diff != "" {
					t.Errorf("unexpected error (-want, +got) = %v", diff)
				}
				return
			}

			if tc.want != nil {
				if diff := cmp.Diff(tc.want, got); diff != "" {
					t.Errorf("unexpected data (-want, +got) = %v", diff)
				}
			}
		})
	}
}

func TestValidate(t *testing.T) {
	now := types.Timestamp{Time: time.Now()}

	testCases := map[string]struct {
		event ce.Event
		want  []string
	}{
		"min valid v0.1": {
			event: ce.Event{
				Context: MinEventContextV01(),
			},
		},
		"min valid v0.2": {
			event: ce.Event{
				Context: MinEventContextV02(),
			},
		},
		"min valid v0.3": {
			event: ce.Event{
				Context: MinEventContextV03(),
			},
		},
		"json valid, v0.1": {
			event: ce.Event{
				Context: FullEventContextV01(now),
				Data:    []byte(`{"a":"apple","b":"banana"}`),
			},
		},
		"json valid, v0.2": {
			event: ce.Event{
				Context: FullEventContextV02(now),
				Data:    []byte(`{"a":"apple","b":"banana"}`),
			},
		},
		"json valid, v0.3": {
			event: ce.Event{
				Context: FullEventContextV03(now),
				Data:    []byte(`{"a":"apple","b":"banana"}`),
			},
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			got := tc.event.Validate()
			var gotErr string
			if got != nil {
				gotErr = got.Error()

				if len(tc.want) == 0 {
					t.Errorf("unexpected no error, got %q", gotErr)
				}
			}

			for _, want := range tc.want {
				if !strings.Contains(gotErr, want) {
					t.Errorf("unexpected error, expected to contain %q, got: %q ", want, gotErr)
				}
			}
		})
	}
}

func TestString(t *testing.T) {
	now := types.Timestamp{Time: time.Now()}

	testCases := map[string]struct {
		event ce.Event
		want  string
	}{
		"empty v0.1": {
			event: ce.Event{
				Context: &ce.EventContextV01{},
			},
			want: `Validation: invalid
Validation Error: 
eventType: MUST be a non-empty string
cloudEventsVersion: MUST be a non-empty string
source: REQUIRED
eventID: MUST be a non-empty string
Context Attributes,
  cloudEventsVersion: 
  eventType: 
  source: 
  eventID: 
`,
		},
		"empty v0.2": {
			event: ce.Event{
				Context: &ce.EventContextV02{},
			},
			want: `Validation: invalid
Validation Error: 
type: MUST be a non-empty string
specversion: MUST be a non-empty string
source: REQUIRED
id: MUST be a non-empty string
Context Attributes,
  specversion: 
  type: 
  source: 
  id: 
`,
		},
		"empty v0.3": {
			event: ce.Event{
				Context: &ce.EventContextV03{},
			},
			want: `Validation: invalid
Validation Error: 
type: MUST be a non-empty string
specversion: MUST be a non-empty string
source: REQUIRED
id: MUST be a non-empty string
Context Attributes,
  specversion: 
  type: 
  source: 
  id: 
`,
		},
		"min v0.1": {
			event: ce.Event{
				Context: MinEventContextV01(),
			},
			want: `Validation: valid
Context Attributes,
  cloudEventsVersion: 0.1
  eventType: com.example.simple
  source: http://example.com/source
  eventID: ABC-123
`,
		},
		"min v0.2": {
			event: ce.Event{
				Context: MinEventContextV02(),
			},
			want: `Validation: valid
Context Attributes,
  specversion: 0.2
  type: com.example.simple
  source: http://example.com/source
  id: ABC-123
`,
		},
		"min v0.3": {
			event: ce.Event{
				Context: MinEventContextV03(),
			},
			want: `Validation: valid
Context Attributes,
  specversion: 0.3
  type: com.example.simple
  source: http://example.com/source
  id: ABC-123
`,
		},
		"json simple, v0.1": {
			event: ce.Event{
				Context: FullEventContextV01(now),
				Data:    []byte(`{"a":"apple","b":"banana"}`),
			},
			want: fmt.Sprintf(`Validation: valid
Context Attributes,
  cloudEventsVersion: 0.1
  eventType: com.example.simple
  eventTypeVersion: v1alpha1
  source: http://example.com/source
  eventID: ABC-123
  eventTime: %s
  schemaURL: http://example.com/schema
  contentType: application/json
Extensions,
  another-test: 1
  datacontentencoding: base64
  subject: topic
  test: extended
Data,
  {
    "a": "apple",
    "b": "banana"
  }
`, now.String()),
		},
		"json simple, v0.2": {
			event: ce.Event{
				Context: FullEventContextV02(now),
				Data:    []byte(`{"a":"apple","b":"banana"}`),
			},
			want: fmt.Sprintf(`Validation: valid
Context Attributes,
  specversion: 0.2
  type: com.example.simple
  source: http://example.com/source
  id: ABC-123
  time: %s
  schemaurl: http://example.com/schema
  contenttype: application/json
Extensions,
  another-test: 1
  datacontentencoding: base64
  eventTypeVersion: v1alpha1
  subject: topic
  test: extended
Data,
  {
    "a": "apple",
    "b": "banana"
  }
`, now.String()),
		},
		"json simple, v0.3": {
			event: ce.Event{
				Context: FullEventContextV03(now),
				Data:    []byte(`{"a":"apple","b":"banana"}`),
			},
			want: fmt.Sprintf(`Validation: valid
Context Attributes,
  specversion: 0.3
  type: com.example.simple
  source: http://example.com/source
  subject: topic
  id: ABC-123
  time: %s
  schemaurl: http://example.com/schema
  datacontenttype: application/json
  datacontentencoding: base64
Extensions,
  another-test: 1
  eventTypeVersion: v1alpha1
  test: extended
Data,
  {
    "a": "apple",
    "b": "banana"
  }
`, now.String()),
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			got := tc.event.String()

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Log(got)
				t.Errorf("unexpected string (-want, +got) = %v", diff)
			}
		})
	}
}

func TestExtensionAs(t *testing.T) {
	now := types.Timestamp{Time: time.Now()}

	testCases := map[string]struct {
		event        ce.Event
		extension    string
		want         string
		wantError    bool
		wantErrorMsg string
	}{
		"min v01, no extension": {
			event: ce.Event{
				Context: MinEventContextV01(),
			},
			extension:    "test",
			wantError:    true,
			wantErrorMsg: `extension "test" does not exist`,
		},
		"full v01, test extension": {
			event: ce.Event{
				Context: FullEventContextV01(now),
			},
			extension: "test",
			want:      "extended",
		},
		"full v01, another-test extension invalid type": {
			event: ce.Event{
				Context: FullEventContextV01(now),
			},
			extension:    "another-test",
			wantError:    true,
			wantErrorMsg: `invalid type for extension "another-test"`,
		},
		"min v02, no extension": {
			event: ce.Event{
				Context: MinEventContextV02(),
			},
			extension:    "test",
			wantError:    true,
			wantErrorMsg: `extension "test" does not exist`,
		},
		"full v02, test extension": {
			event: ce.Event{
				Context: FullEventContextV02(now),
			},
			extension: "test",
			want:      "extended",
		},
		"full v02, another-test extension invalid type": {
			event: ce.Event{
				Context: FullEventContextV02(now),
			},
			extension:    "another-test",
			wantError:    true,
			wantErrorMsg: `invalid type for extension "another-test"`,
		},
		"min v03, no extension": {
			event: ce.Event{
				Context: MinEventContextV03(),
			},
			extension:    "test",
			wantError:    true,
			wantErrorMsg: `extension "test" does not exist`,
		},
		"full v03, test extension": {
			event: ce.Event{
				Context: FullEventContextV03(now),
			},
			extension: "test",
			want:      "extended",
		},
		"full v03, another-test extension invalid type": {
			event: ce.Event{
				Context: FullEventContextV03(now),
			},
			extension:    "another-test",
			wantError:    true,
			wantErrorMsg: `invalid type for extension "another-test"`,
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			var got string
			err := tc.event.Context.ExtensionAs(tc.extension, &got)

			if tc.wantError {
				if err == nil {
					t.Errorf("expected error %q, got nil", tc.wantErrorMsg)
				} else {
					if diff := cmp.Diff(tc.wantErrorMsg, err.Error()); diff != "" {
						t.Errorf("unexpected (-want, +got) = %v", diff)
					}
				}
			} else {
				if diff := cmp.Diff(tc.want, got); diff != "" {
					t.Errorf("unexpected (-want, +got) = %v", diff)
				}
			}
		})
	}
}

func strptr(s string) *string {
	return &s
}

func MinEventContextV01() *ce.EventContextV01 {
	sourceUrl, _ := url.Parse("http://example.com/source")
	source := &types.URLRef{URL: *sourceUrl}

	return ce.EventContextV01{
		EventType: "com.example.simple",
		Source:    *source,
		EventID:   "ABC-123",
	}.AsV01()
}

func MinEventContextV02() *ce.EventContextV02 {
	sourceUrl, _ := url.Parse("http://example.com/source")
	source := &types.URLRef{URL: *sourceUrl}

	return ce.EventContextV02{
		Type:   "com.example.simple",
		Source: *source,
		ID:     "ABC-123",
	}.AsV02()
}

func MinEventContextV03() *ce.EventContextV03 {
	sourceUrl, _ := url.Parse("http://example.com/source")
	source := &types.URLRef{URL: *sourceUrl}

	return ce.EventContextV03{
		Type:   "com.example.simple",
		Source: *source,
		ID:     "ABC-123",
	}.AsV03()
}

func FullEventContextV01(now types.Timestamp) *ce.EventContextV01 {
	sourceUrl, _ := url.Parse("http://example.com/source")
	source := &types.URLRef{URL: *sourceUrl}

	schemaUrl, _ := url.Parse("http://example.com/schema")
	schema := &types.URLRef{URL: *schemaUrl}

	eventContextV01 := ce.EventContextV01{
		EventID:          "ABC-123",
		EventTime:        &now,
		EventType:        "com.example.simple",
		EventTypeVersion: strptr("v1alpha1"),
		SchemaURL:        schema,
		ContentType:      ce.StringOfApplicationJSON(),
		Source:           *source,
	}
	eventContextV01.SetExtension(ce.SubjectKey, "topic")
	eventContextV01.SetExtension(ce.DataContentEncodingKey, ce.Base64)
	eventContextV01.SetExtension("test", "extended")
	eventContextV01.SetExtension("another-test", 1)
	return eventContextV01.AsV01()
}

func FullEventContextV02(now types.Timestamp) *ce.EventContextV02 {
	sourceUrl, _ := url.Parse("http://example.com/source")
	source := &types.URLRef{URL: *sourceUrl}

	schemaUrl, _ := url.Parse("http://example.com/schema")
	schema := &types.URLRef{URL: *schemaUrl}

	extensions := make(map[string]interface{})
	extensions["test"] = "extended"
	extensions["another-test"] = 1

	eventContextV02 := ce.EventContextV02{
		ID:          "ABC-123",
		Time:        &now,
		Type:        "com.example.simple",
		SchemaURL:   schema,
		ContentType: ce.StringOfApplicationJSON(),
		Source:      *source,
		Extensions:  extensions,
	}
	eventContextV02.SetExtension(ce.SubjectKey, "topic")
	eventContextV02.SetExtension(ce.DataContentEncodingKey, ce.Base64)
	eventContextV02.SetExtension(ce.EventTypeVersionKey, "v1alpha1")
	return eventContextV02.AsV02()
}

func FullEventContextV03(now types.Timestamp) *ce.EventContextV03 {
	sourceUrl, _ := url.Parse("http://example.com/source")
	source := &types.URLRef{URL: *sourceUrl}

	schemaUrl, _ := url.Parse("http://example.com/schema")
	schema := &types.URLRef{URL: *schemaUrl}

	eventContextV03 := ce.EventContextV03{
		ID:                  "ABC-123",
		Time:                &now,
		Type:                "com.example.simple",
		SchemaURL:           schema,
		DataContentType:     ce.StringOfApplicationJSON(),
		DataContentEncoding: ce.StringOfBase64(),
		Source:              *source,
		Subject:             strptr("topic"),
	}
	eventContextV03.SetExtension("test", "extended")
	eventContextV03.SetExtension("another-test", 1)
	eventContextV03.SetExtension(ce.EventTypeVersionKey, "v1alpha1")
	return eventContextV03.AsV03()
}
