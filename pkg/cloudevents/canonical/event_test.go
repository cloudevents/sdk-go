package canonical

import (
	"github.com/google/go-cmp/cmp"
	"net/url"
	"testing"
	"time"
)

func TestContextAsV01(t *testing.T) {
	now := time.Now()

	testCases := map[string]struct {
		event Event
		want  EventContextV01
	}{
		"empty, no conversion": {
			event: Event{
				Context: EventContextV01{},
			},
			want: EventContextV01{},
		},
		"min v01, no conversion": {
			event: Event{
				Context: MinEventContextV01(),
			},
			want: MinEventContextV01(),
		},
		"full v01, no conversion": {
			event: Event{
				Context: FullEventContextV01(&now),
			},
			want: FullEventContextV01(&now),
		},
		"min v02 -> v01": {
			event: Event{
				Context: MinEventContextV02(),
			},
			want: MinEventContextV01(),
		},
		"full v02 -> v01": {
			event: Event{
				Context: FullEventContextV02(&now),
			},
			want: FullEventContextV01(&now),
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			got := tc.event.Context.AsV01()

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("unexpected resource (-want, +got) = %v", diff)
			}
		})
	}
}

func TestContextAsV02(t *testing.T) {
	now := time.Now()

	testCases := map[string]struct {
		event Event
		want  EventContextV02
	}{
		"empty, no conversion": {
			event: Event{
				Context: EventContextV02{},
			},
			want: EventContextV02{},
		},
		"min v02, no conversion": {
			event: Event{
				Context: MinEventContextV02(),
			},
			want: MinEventContextV02(),
		},
		"full v02, no conversion": {
			event: Event{
				Context: FullEventContextV02(&now),
			},
			want: FullEventContextV02(&now),
		},
		"min v01 -> v02": {
			event: Event{
				Context: MinEventContextV01(),
			},
			want: MinEventContextV02(),
		},
		"full v01 -> v2": {
			event: Event{
				Context: FullEventContextV01(&now),
			},
			want: FullEventContextV02(&now),
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			got := tc.event.Context.AsV02()

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("unexpected resource (-want, +got) = %v", diff)
			}
		})
	}
}

func TestGetDataContentType(t *testing.T) {
	now := time.Now()

	testCases := map[string]struct {
		event Event
		want  string
	}{
		"min v01, blank": {
			event: Event{
				Context: MinEventContextV01(),
			},
			want: "",
		},
		"full v01, json": {
			event: Event{
				Context: FullEventContextV01(&now),
			},
			want: "application/json",
		},
		"min v02, blank": {
			event: Event{
				Context: MinEventContextV02(),
			},
			want: "",
		},
		"full v02, json": {
			event: Event{
				Context: FullEventContextV02(&now),
			},
			want: "application/json",
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			got := tc.event.Context.DataContentType()

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("unexpected resource (-want, +got) = %v", diff)
			}
		})
	}
}

func MinEventContextV01() EventContextV01 {
	source, _ := url.Parse("http://example.com/source")

	return EventContextV01{
		CloudEventsVersion: CloudEventsVersionV01,
		EventType:          "com.example.simple",
		Source:             *source,
		EventID:            "ABC-123",
	}
}

func MinEventContextV02() EventContextV02 {
	source, _ := url.Parse("http://example.com/source")

	return EventContextV02{
		SpecVersion: CloudEventsVersionV02,
		Type:        "com.example.simple",
		Source:      *source,
		ID:          "ABC-123",
	}
}

func FullEventContextV01(now *time.Time) EventContextV01 {
	source, _ := url.Parse("http://example.com/source")
	schema, _ := url.Parse("http://example.com/schema")

	extensions := make(map[string]interface{})
	extensions["test"] = "extended"

	return EventContextV01{
		CloudEventsVersion: CloudEventsVersionV01,
		EventID:            "ABC-123",
		EventTime:          now,
		EventType:          "com.example.simple",
		EventTypeVersion:   "v1alpha1",
		SchemaURL:          schema,
		ContentType:        "application/json",
		Source:             *source,
		Extensions:         extensions,
	}
}

func FullEventContextV02(now *time.Time) EventContextV02 {
	source, _ := url.Parse("http://example.com/source")
	schema, _ := url.Parse("http://example.com/schema")

	extensions := make(map[string]interface{})
	extensions["test"] = "extended"
	extensions["eventTypeVersion"] = "v1alpha1"

	return EventContextV02{
		SpecVersion: CloudEventsVersionV02,
		ID:          "ABC-123",
		Time:        now,
		Type:        "com.example.simple",
		SchemaURL:   schema,
		ContentType: "application/json",
		Source:      *source,
		Extensions:  extensions,
	}
}
