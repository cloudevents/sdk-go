package cloudevents_test

import (
	ce "github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/types"
	"github.com/gin-gonic/gin/json"
	"github.com/google/go-cmp/cmp"
	"net/url"
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
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			got := tc.event.Context.DataContentType()

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

func MinEventContextV01() ce.EventContextV01 {
	sourceUrl, _ := url.Parse("http://example.com/source")
	source := &types.URLRef{URL: *sourceUrl}

	return ce.EventContextV01{
		CloudEventsVersion: ce.CloudEventsVersionV01,
		EventType:          "com.example.simple",
		Source:             *source,
		EventID:            "ABC-123",
	}
}

func MinEventContextV02() ce.EventContextV02 {
	sourceUrl, _ := url.Parse("http://example.com/source")
	source := &types.URLRef{*sourceUrl}

	return ce.EventContextV02{
		SpecVersion: ce.CloudEventsVersionV02,
		Type:        "com.example.simple",
		Source:      *source,
		ID:          "ABC-123",
	}
}

func FullEventContextV01(now types.Timestamp) ce.EventContextV01 {
	sourceUrl, _ := url.Parse("http://example.com/source")
	source := &types.URLRef{URL: *sourceUrl}

	schemaUrl, _ := url.Parse("http://example.com/schema")
	schema := &types.URLRef{URL: *schemaUrl}

	extensions := make(map[string]interface{})
	extensions["test"] = "extended"

	return ce.EventContextV01{
		CloudEventsVersion: ce.CloudEventsVersionV01,
		EventID:            "ABC-123",
		EventTime:          &now,
		EventType:          "com.example.simple",
		EventTypeVersion:   "v1alpha1",
		SchemaURL:          schema,
		ContentType:        "application/json",
		Source:             *source,
		Extensions:         extensions,
	}
}

func FullEventContextV02(now types.Timestamp) ce.EventContextV02 {
	sourceUrl, _ := url.Parse("http://example.com/source")
	source := &types.URLRef{URL: *sourceUrl}

	schemaUrl, _ := url.Parse("http://example.com/schema")
	schema := &types.URLRef{URL: *schemaUrl}

	extensions := make(map[string]interface{})
	extensions["test"] = "extended"
	extensions["eventTypeVersion"] = "v1alpha1"

	return ce.EventContextV02{
		SpecVersion: ce.CloudEventsVersionV02,
		ID:          "ABC-123",
		Time:        &now,
		Type:        "com.example.simple",
		SchemaURL:   schema,
		ContentType: "application/json",
		Source:      *source,
		Extensions:  extensions,
	}
}
