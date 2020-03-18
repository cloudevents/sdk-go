package client

import (
	"context"
	"testing"
	"time"

	"github.com/cloudevents/sdk-go/v2/event"
)

var versions = []string{"0.3", "1.0"}

func TestDefaultIDToUUIDIfNotSet_empty(t *testing.T) {
	for _, tc := range versions {
		t.Run(tc, func(t *testing.T) {
			got := DefaultIDToUUIDIfNotSet(context.TODO(), event.New(tc))

			if got.Context != nil && got.ID() == "" {
				t.Errorf("failed to generate an id for event")
			}
		})
	}
}

func TestDefaultIDToUUIDIfNotSet_set(t *testing.T) {
	for _, tc := range versions {
		t.Run(tc, func(t *testing.T) {
			event := event.New(tc)
			event.SetID("abc-123")

			got := DefaultIDToUUIDIfNotSet(context.TODO(), event)

			if got.ID() != "abc-123" {
				t.Errorf("id was defaulted when already set")
			}
		})
	}
}

func TestDefaultIDToUUIDIfNotSet_nil(t *testing.T) {
	got := DefaultIDToUUIDIfNotSet(context.TODO(), event.Event{})

	if got.Context != nil && got.ID() == "" {
		t.Errorf("failed to generate time for nil context event")
	}
}

func TestDefaultIDToUUIDIfNotSetImmutable(t *testing.T) {
	e := event.New()

	got := DefaultIDToUUIDIfNotSet(context.TODO(), e)

	if e.ID() != "" {
		t.Errorf("modified the original event")
	}

	if got.ID() == "" {
		t.Errorf("failed to generate an id for event")
	}
}

func TestDefaultTimeToNowIfNotSet_empty(t *testing.T) {
	for _, tc := range versions {
		t.Run(tc, func(t *testing.T) {
			got := DefaultTimeToNowIfNotSet(context.TODO(), event.New(tc))

			if got.Time().IsZero() {
				t.Errorf("failed to generate time for event")
			}
		})
	}
}

func TestDefaultTimeToNowIfNotSet_set(t *testing.T) {
	for _, tc := range versions {
		t.Run(tc, func(t *testing.T) {
			event := event.New(tc)
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
	got := DefaultTimeToNowIfNotSet(context.TODO(), event.Event{})

	if got.Context != nil && got.Time().IsZero() {
		t.Errorf("failed to generate time for nil context event")
	}
}

func TestDefaultTimeToNowIfNotSetImmutable(t *testing.T) {
	e := event.New()

	got := DefaultTimeToNowIfNotSet(context.TODO(), e)

	if !e.Time().IsZero() {
		t.Errorf("modified the original event")
	}

	if got.Time().IsZero() {
		t.Errorf("failed to generate a time for event")
	}
}

func TestNewDefaultDataContentTypeIfNotSet_empty(t *testing.T) {
	ct := "a/b"
	for _, tc := range versions {
		t.Run(tc, func(t *testing.T) {
			fn := NewDefaultDataContentTypeIfNotSet(ct)
			got := fn(context.TODO(), event.New(tc))

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
			event := event.New(tc)
			event.SetDataContentType(ct)

			fn := NewDefaultDataContentTypeIfNotSet("b/c")
			got := fn(context.TODO(), event)

			if got.DataContentType() != ct {
				t.Errorf("failed to preserve data content type for event")
			}
		})
	}
}
