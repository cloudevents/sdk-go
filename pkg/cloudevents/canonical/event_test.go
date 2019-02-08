package canonical_test

import (
	c "github.com/cloudevents/sdk-go/pkg/cloudevents/canonical"
	"github.com/google/go-cmp/cmp"
	"net/url"
	"testing"
	"time"
)

func TestContextAsV01(t *testing.T) {
	now := time.Now()

	testCases := map[string]struct {
		event c.Event
		want  c.EventContextV01
	}{
		"empty, no conversion": {
			event: c.Event{
				Context: c.EventContextV01{},
			},
			want: c.EventContextV01{},
		},
		"min v01, no conversion": {
			event: c.Event{
				Context: MinEventContextV01(),
			},
			want: MinEventContextV01(),
		},
		"full v01, no conversion": {
			event: c.Event{
				Context: FullEventContextV01(&now),
			},
			want: FullEventContextV01(&now),
		},
		"min v02 -> v01": {
			event: c.Event{
				Context: MinEventContextV02(),
			},
			want: MinEventContextV01(),
		},
		"full v02 -> v01": {
			event: c.Event{
				Context: FullEventContextV02(&now),
			},
			want: FullEventContextV01(&now),
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
	now := time.Now()

	testCases := map[string]struct {
		event c.Event
		want  c.EventContextV02
	}{
		"empty, no conversion": {
			event: c.Event{
				Context: c.EventContextV02{},
			},
			want: c.EventContextV02{},
		},
		"min v02, no conversion": {
			event: c.Event{
				Context: MinEventContextV02(),
			},
			want: MinEventContextV02(),
		},
		"full v02, no conversion": {
			event: c.Event{
				Context: FullEventContextV02(&now),
			},
			want: FullEventContextV02(&now),
		},
		"min v01 -> v02": {
			event: c.Event{
				Context: MinEventContextV01(),
			},
			want: MinEventContextV02(),
		},
		"full v01 -> v2": {
			event: c.Event{
				Context: FullEventContextV01(&now),
			},
			want: FullEventContextV02(&now),
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
	now := time.Now()

	testCases := map[string]struct {
		event c.Event
		want  string
	}{
		"min v01, blank": {
			event: c.Event{
				Context: MinEventContextV01(),
			},
			want: "",
		},
		"full v01, json": {
			event: c.Event{
				Context: FullEventContextV01(&now),
			},
			want: "application/json",
		},
		"min v02, blank": {
			event: c.Event{
				Context: MinEventContextV02(),
			},
			want: "",
		},
		"full v02, json": {
			event: c.Event{
				Context: FullEventContextV02(&now),
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

func MinEventContextV01() c.EventContextV01 {
	source, _ := url.Parse("http://example.com/source")

	return c.EventContextV01{
		CloudEventsVersion: c.CloudEventsVersionV01,
		EventType:          "com.example.simple",
		Source:             *source,
		EventID:            "ABC-123",
	}
}

func MinEventContextV02() c.EventContextV02 {
	source, _ := url.Parse("http://example.com/source")

	return c.EventContextV02{
		SpecVersion: c.CloudEventsVersionV02,
		Type:        "com.example.simple",
		Source:      *source,
		ID:          "ABC-123",
	}
}

func FullEventContextV01(now *time.Time) c.EventContextV01 {
	source, _ := url.Parse("http://example.com/source")
	schema, _ := url.Parse("http://example.com/schema")

	extensions := make(map[string]interface{})
	extensions["test"] = "extended"

	return c.EventContextV01{
		CloudEventsVersion: c.CloudEventsVersionV01,
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

func FullEventContextV02(now *time.Time) c.EventContextV02 {
	source, _ := url.Parse("http://example.com/source")
	schema, _ := url.Parse("http://example.com/schema")

	extensions := make(map[string]interface{})
	extensions["test"] = "extended"
	extensions["eventTypeVersion"] = "v1alpha1"

	return c.EventContextV02{
		SpecVersion: c.CloudEventsVersionV02,
		ID:          "ABC-123",
		Time:        now,
		Type:        "com.example.simple",
		SchemaURL:   schema,
		ContentType: "application/json",
		Source:      *source,
		Extensions:  extensions,
	}
}
