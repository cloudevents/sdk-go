package cloudevents_test

import (
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

			got := tc.event.Context.GetDataMediaType()

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
				Context: FullEventContextV01(now),
				Data:    []byte(`{"a":"apple","b":"banana"}`),
			},
			want: &map[string]string{
				"a": "apple",
				"b": "banana",
			},
		},
		"json complex empty": {
			event: ce.Event{
				Context: FullEventContextV01(now),
				Data:    []byte(`{}`),
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
					return j
				}(),
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
				Context: ce.EventContextV01{},
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
				Context: ce.EventContextV02{},
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
				Context: ce.EventContextV03{},
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
  eventTypeVersion: v1alpha1
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
  id: ABC-123
  time: %s
  schemaurl: http://example.com/schema
  datacontenttype: application/json
Extensions,
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

func strptr(s string) *string {
	return &s
}

func MinEventContextV01() ce.EventContextV01 {
	sourceUrl, _ := url.Parse("http://example.com/source")
	source := &types.URLRef{URL: *sourceUrl}

	return ce.EventContextV01{
		EventType: "com.example.simple",
		Source:    *source,
		EventID:   "ABC-123",
	}.AsV01()
}

func MinEventContextV02() ce.EventContextV02 {
	sourceUrl, _ := url.Parse("http://example.com/source")
	source := &types.URLRef{URL: *sourceUrl}

	return ce.EventContextV02{
		Type:   "com.example.simple",
		Source: *source,
		ID:     "ABC-123",
	}.AsV02()
}

func MinEventContextV03() ce.EventContextV03 {
	sourceUrl, _ := url.Parse("http://example.com/source")
	source := &types.URLRef{URL: *sourceUrl}

	return ce.EventContextV03{
		Type:   "com.example.simple",
		Source: *source,
		ID:     "ABC-123",
	}.AsV03()
}

func FullEventContextV01(now types.Timestamp) ce.EventContextV01 {
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
	}.AsV01()
	eventContextV01.Extension("test", "extended")
	return eventContextV01
}

func FullEventContextV02(now types.Timestamp) ce.EventContextV02 {
	sourceUrl, _ := url.Parse("http://example.com/source")
	source := &types.URLRef{URL: *sourceUrl}

	schemaUrl, _ := url.Parse("http://example.com/schema")
	schema := &types.URLRef{URL: *schemaUrl}

	extensions := make(map[string]interface{})
	extensions["test"] = "extended"

	eventContextV02 := ce.EventContextV02{
		ID:          "ABC-123",
		Time:        &now,
		Type:        "com.example.simple",
		SchemaURL:   schema,
		ContentType: ce.StringOfApplicationJSON(),
		Source:      *source,
		Extensions:  extensions,
	}.AsV02()
	eventContextV02.Extension("eventTypeVersion", "v1alpha1")
	return eventContextV02
}

func FullEventContextV03(now types.Timestamp) ce.EventContextV03 {
	sourceUrl, _ := url.Parse("http://example.com/source")
	source := &types.URLRef{URL: *sourceUrl}

	schemaUrl, _ := url.Parse("http://example.com/schema")
	schema := &types.URLRef{URL: *schemaUrl}

	eventContextV03 := ce.EventContextV03{
		ID:              "ABC-123",
		Time:            &now,
		Type:            "com.example.simple",
		SchemaURL:       schema,
		DataContentType: ce.StringOfApplicationJSON(),
		Source:          *source,
	}.AsV03()
	eventContextV03.Extension("test", "extended")
	eventContextV03.Extension("evenTypeVersion", "v1alpha1")
	return eventContextV03
}
