package cloudevents_test

import (
	ce "github.com/cloudevents/sdk-go/pkg/cloudevents"
	c "github.com/cloudevents/sdk-go/pkg/cloudevents/context"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/types"
	"github.com/gin-gonic/gin/json"
	"github.com/google/go-cmp/cmp"
	"net/url"
	"reflect"
	"testing"
	"time"
)

func TestContextAsV01(t *testing.T) {
	now := types.Timestamp{Time: time.Now()}

	testCases := map[string]struct {
		event ce.Event
		want  c.EventContextV01
	}{
		"empty, no conversion": {
			event: ce.Event{
				Context: c.EventContextV01{},
			},
			want: c.EventContextV01{
				CloudEventsVersion: "0.1",
			},
		},
		"min v01, no conversion": {
			event: ce.Event{
				Context: MinEventContextV01(),
			},
			want: MinEventContextV01(),
		},
		"full v01, no conversion": {
			event: ce.Event{
				Context: FullEventContextV01(now),
			},
			want: FullEventContextV01(now),
		},
		"min v02 -> v01": {
			event: ce.Event{
				Context: MinEventContextV02(),
			},
			want: MinEventContextV01(),
		},
		"full v02 -> v01": {
			event: ce.Event{
				Context: FullEventContextV02(now),
			},
			want: FullEventContextV01(now),
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			got := tc.event.Context.AsV01()

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("unexpected (-want, +got) = %v", diff)
			}
		})
	}
}

func TestContextAsV02(t *testing.T) {
	now := types.Timestamp{Time: time.Now()}

	testCases := map[string]struct {
		event ce.Event
		want  c.EventContextV02
	}{
		"empty, no conversion": {
			event: ce.Event{
				Context: c.EventContextV02{},
			},
			want: c.EventContextV02{
				SpecVersion: "0.2",
			},
		},
		"min v02, no conversion": {
			event: ce.Event{
				Context: MinEventContextV02(),
			},
			want: MinEventContextV02(),
		},
		"full v02, no conversion": {
			event: ce.Event{
				Context: FullEventContextV02(now),
			},
			want: FullEventContextV02(now),
		},
		"min v01 -> v02": {
			event: ce.Event{
				Context: MinEventContextV01(),
			},
			want: MinEventContextV02(),
		},
		"full v01 -> v2": {
			event: ce.Event{
				Context: FullEventContextV01(now),
			},
			want: FullEventContextV02(now),
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			got := tc.event.Context.AsV02()

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("unexpected (-want, +got) = %v", diff)
			}
		})
	}
}

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
		event ce.Event
		want  interface{}
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

			dataType := reflect.TypeOf(tc.want)

			t.Logf("got dataType: %s", dataType)

			got, _ := allocate(dataType)

			err := tc.event.DataAs(got)
			//got := data

			_ = err // TODO

			if dataType != nil {
				if diff := cmp.Diff(tc.want, got); diff != "" {
					t.Errorf("unexpected data (-want, +got) = %v", diff)
				}
			}
		})
	}
}

// Alocates a new instance of type t and returns:
// asPtr is of type t if t is a pointer type and of type &t otherwise (used for unmarshalling)
// asValue is a Value of type t pointing to the same data as asPtr
func allocate(t reflect.Type) (asPtr interface{}, asValue reflect.Value) {
	if t == nil {
		return nil, reflect.Value{}
	}
	if t.Kind() == reflect.Ptr {
		reflectPtr := reflect.New(t.Elem())
		asPtr = reflectPtr.Interface()
		asValue = reflectPtr
	} else {
		reflectPtr := reflect.New(t)
		asPtr = reflectPtr.Interface()
		asValue = reflectPtr.Elem()
	}
	return
}

func MinEventContextV01() c.EventContextV01 {
	sourceUrl, _ := url.Parse("http://example.com/source")
	source := &types.URLRef{URL: *sourceUrl}

	return c.EventContextV01{
		CloudEventsVersion: c.CloudEventsVersionV01,
		EventType:          "com.example.simple",
		Source:             *source,
		EventID:            "ABC-123",
	}
}

func MinEventContextV02() c.EventContextV02 {
	sourceUrl, _ := url.Parse("http://example.com/source")
	source := &types.URLRef{*sourceUrl}

	return c.EventContextV02{
		SpecVersion: c.CloudEventsVersionV02,
		Type:        "com.example.simple",
		Source:      *source,
		ID:          "ABC-123",
	}
}

func FullEventContextV01(now types.Timestamp) c.EventContextV01 {
	sourceUrl, _ := url.Parse("http://example.com/source")
	source := &types.URLRef{URL: *sourceUrl}

	schemaUrl, _ := url.Parse("http://example.com/schema")
	schema := &types.URLRef{URL: *schemaUrl}

	extensions := make(map[string]interface{})
	extensions["test"] = "extended"

	return c.EventContextV01{
		CloudEventsVersion: c.CloudEventsVersionV01,
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

func FullEventContextV02(now types.Timestamp) c.EventContextV02 {
	sourceUrl, _ := url.Parse("http://example.com/source")
	source := &types.URLRef{URL: *sourceUrl}

	schemaUrl, _ := url.Parse("http://example.com/schema")
	schema := &types.URLRef{URL: *schemaUrl}

	extensions := make(map[string]interface{})
	extensions["test"] = "extended"
	extensions["eventTypeVersion"] = "v1alpha1"

	return c.EventContextV02{
		SpecVersion: c.CloudEventsVersionV02,
		ID:          "ABC-123",
		Time:        &now,
		Type:        "com.example.simple",
		SchemaURL:   schema,
		ContentType: "application/json",
		Source:      *source,
		Extensions:  extensions,
	}
}
