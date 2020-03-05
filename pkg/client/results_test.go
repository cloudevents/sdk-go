package client

import (
	"context"
	"github.com/google/go-cmp/cmp"
	"testing"

	"github.com/cloudevents/sdk-go/pkg/event"
)

func TestResultsFnValidTypes(t *testing.T) {
	for name, fn := range map[string]interface{}{
		"none":       func() {},
		"ctx":        func(context.Context) {},
		"ctx, event": func(context.Context, *event.Event) {},
		"event":      func(*event.Event) {},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := parseResultsFn(fn); err != nil {
				t.Errorf("%q failed: %v", name, err)
			}
		})
	}
}

func TestResultsFnInvalidTypes(t *testing.T) {
	for name, fn := range map[string]interface{}{
		"wrong type in":        func(string) {},
		"wrong type out":       func() string { return "" },
		"context dup Event in": func(context.Context, *event.Event, *event.Event) {},
		"dup Event in":         func(*event.Event, *event.Event) {},
		"Event non pointer in": func(event.Event) {},
		"not a function":       map[string]string(nil),
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := parseResultsFn(fn); err == nil {
				t.Errorf("%q failed to catch the issue", name)
			}
		})
	}
}

func TestResultsFnInvoke_1(t *testing.T) {
	key := struct{}{}
	wantCtx := context.WithValue(context.TODO(), key, "UNIT TEST")
	wantEvent := event.New()
	wantEvent.SetID("UNIT TEST")

	fn, err := parseResultsFn(func(ctx context.Context, e *event.Event) {
		if diff := cmp.Diff(wantCtx.Value(key), ctx.Value(key)); diff != "" {
			t.Errorf("unexpected context (-want, +got) = %v", diff)
		}

		if diff := cmp.Diff(&wantEvent, e); diff != "" {
			t.Errorf("unexpected event (-want, +got) = %v", diff)
		}
	})
	if err != nil {
		t.Errorf("unexpected error, wanted nil got = %v", err)
	}

	fn.invoke(wantCtx, &wantEvent)
}

func TestResultsFnInvoke_2(t *testing.T) {
	key := struct{}{}
	wantCtx := context.WithValue(context.TODO(), key, "UNIT TEST")

	fn, err := parseResultsFn(func(ctx context.Context) {
		if diff := cmp.Diff(wantCtx.Value(key), ctx.Value(key)); diff != "" {
			t.Errorf("unexpected context (-want, +got) = %v", diff)
		}
	})
	if err != nil {
		t.Errorf("unexpected error, wanted nil got = %v", err)
	}

	fn.invoke(wantCtx, &event.Event{})
}

func TestResultsFnInvoke_3(t *testing.T) {
	wantEvent := event.New()
	wantEvent.SetID("UNIT TEST")

	fn, err := parseResultsFn(func(e *event.Event) {
		if diff := cmp.Diff(&wantEvent, e); diff != "" {
			t.Errorf("unexpected event (-want, +got) = %v", diff)
		}
	})
	if err != nil {
		t.Errorf("unexpected error, wanted nil got = %v", err)
	}

	fn.invoke(context.Background(), &wantEvent)
}
