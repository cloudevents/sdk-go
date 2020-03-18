package client

import (
	"context"
	"testing"
	"time"

	"github.com/cloudevents/sdk-go/v1/cloudevents"
	"github.com/google/go-cmp/cmp"
)

var versions = []string{"0.1", "0.2", "0.3", "1.0"}

func TestDefaultIDToUUIDIfNotSet_empty(t *testing.T) {
	for _, tc := range versions {
		t.Run(tc, func(t *testing.T) {
			got := DefaultIDToUUIDIfNotSet(context.TODO(), cloudevents.New(tc))

			if got.Context != nil && got.ID() == "" {
				t.Errorf("failed to generate an id for event")
			}
		})
	}
}

func TestDefaultIDToUUIDIfNotSet_set(t *testing.T) {
	for _, tc := range versions {
		t.Run(tc, func(t *testing.T) {
			event := cloudevents.New(tc)
			event.SetID("abc-123")

			got := DefaultIDToUUIDIfNotSet(context.TODO(), event)

			if got.ID() != "abc-123" {
				t.Errorf("id was defaulted when already set")
			}
		})
	}
}

func TestDefaultIDToUUIDIfNotSet_nil(t *testing.T) {
	got := DefaultIDToUUIDIfNotSet(context.TODO(), cloudevents.Event{})

	if got.Context != nil && got.ID() == "" {
		t.Errorf("failed to generate time for nil context event")
	}
}

func TestDefaultIDToUUIDIfNotSetImmutable(t *testing.T) {
	event := cloudevents.Event{
		Context: &cloudevents.EventContextV01{},
	}

	got := DefaultIDToUUIDIfNotSet(context.TODO(), event)

	want := "0.1"

	if diff := cmp.Diff(want, got.SpecVersion()); diff != "" {
		t.Errorf("unexpected (-want, +got) = %v", diff)
	}

	if event.Context.AsV01().EventID != "" {
		t.Errorf("modified the original event")
	}

	if got.Context.AsV01().EventID == "" {
		t.Errorf("failed to generate an id for event")
	}
}

func TestDefaultTimeToNowIfNotSet_empty(t *testing.T) {
	for _, tc := range versions {
		t.Run(tc, func(t *testing.T) {
			got := DefaultTimeToNowIfNotSet(context.TODO(), cloudevents.New(tc))

			if got.Time().IsZero() {
				t.Errorf("failed to generate time for event")
			}
		})
	}
}

func TestDefaultTimeToNowIfNotSet_set(t *testing.T) {
	for _, tc := range versions {
		t.Run(tc, func(t *testing.T) {
			event := cloudevents.New(tc)
			now := time.Now()

			event.SetTime(now)

			got := DefaultTimeToNowIfNotSet(context.TODO(), event)

			if !got.Time().Equal(now) {
				t.Errorf("time was defaulted when already set")
			}
		})
	}
}

func TestDefaultTimeToNowIfNotSet_nil(t *testing.T) {
	got := DefaultTimeToNowIfNotSet(context.TODO(), cloudevents.Event{})

	if got.Context != nil && got.Time().IsZero() {
		t.Errorf("failed to generate time for nil context event")
	}
}

func TestDefaultTimeToNowIfNotSetImmutable(t *testing.T) {
	event := cloudevents.Event{
		Context: &cloudevents.EventContextV01{},
	}

	got := DefaultTimeToNowIfNotSet(context.TODO(), event)

	want := "0.1"

	if diff := cmp.Diff(want, got.SpecVersion()); diff != "" {
		t.Errorf("unexpected (-want, +got) = %v", diff)
	}

	if event.Context.AsV01().EventTime != nil {
		t.Errorf("modified the original event")
	}

	if got.Context.AsV01().EventTime.IsZero() {
		t.Errorf("failed to generate a time for event")
	}
}

func TestNewDefaultDataContentTypeIfNotSet_empty(t *testing.T) {
	ct := "a/b"
	for _, tc := range versions {
		t.Run(tc, func(t *testing.T) {
			fn := NewDefaultDataContentTypeIfNotSet(ct)
			got := fn(context.TODO(), cloudevents.New(tc))

			if got.DataContentType() != ct {
				t.Errorf("failed to default data content type for event")
			}
		})
	}
}

func TestNewDefaultDataContentTypeIfNotSet_set(t *testing.T) {
	ct := "a/b"
	for _, tc := range versions {
		t.Run(tc, func(t *testing.T) {
			event := cloudevents.New(tc)
			event.SetDataContentType(ct)

			fn := NewDefaultDataContentTypeIfNotSet("b/c")
			got := fn(context.TODO(), event)

			if got.DataContentType() != ct {
				t.Errorf("failed to preserve data content type for event")
			}
		})
	}
}
