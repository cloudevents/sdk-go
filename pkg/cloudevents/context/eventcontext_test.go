package context_test

import (
	ce "github.com/cloudevents/sdk-go/pkg/cloudevents"
	c "github.com/cloudevents/sdk-go/pkg/cloudevents/context"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/types"
	"github.com/google/go-cmp/cmp"
	"net/url"
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
