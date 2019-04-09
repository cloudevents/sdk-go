package client

import (
	"context"
	"errors"
	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/google/go-cmp/cmp"
	"testing"
)

func TestReceiverFnValidTypes(t *testing.T) {
	for name, fn := range map[string]interface{}{
		"no in, no out":                           func() {},
		"no in, error out":                        func() error { return nil },
		"ctx in, no out":                          func(context.Context) {},
		"ctx, Event in, no out":                   func(context.Context, cloudevents.Event) {},
		"ctx, EventResponse in, no out":           func(context.Context, *cloudevents.EventResponse) {},
		"ctx, Event, EventResponse in, no out":    func(context.Context, cloudevents.Event, *cloudevents.EventResponse) {},
		"ctx in, error out":                       func(context.Context) error { return nil },
		"ctx, Event in, error out":                func(context.Context, cloudevents.Event) error { return nil },
		"ctx, EventResponse in, error out":        func(context.Context, *cloudevents.EventResponse) error { return nil },
		"ctx, Event, EventResponse in, error out": func(context.Context, cloudevents.Event, *cloudevents.EventResponse) error { return nil },
		"Event in, no out":                        func(cloudevents.Event) {},
		"EventResponse in, no out":                func(*cloudevents.EventResponse) {},
		"Event, EventResponse in, no out":         func(cloudevents.Event, *cloudevents.EventResponse) {},
		"Event in, error out":                     func(cloudevents.Event) error { return nil },
		"EventResponse in, error out":             func(*cloudevents.EventResponse) error { return nil },
		"Event, EventResponse in, error out":      func(cloudevents.Event, *cloudevents.EventResponse) error { return nil },
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := receiver(fn); err != nil {
				t.Errorf("%q failed: %v", name, err)
			}
		})
	}
}

func TestReceiverFnInvalidTypes(t *testing.T) {
	for name, fn := range map[string]interface{}{
		"wrong type in":                func(string) {},
		"wrong type out":               func() string { return "" },
		"extra in":                     func(context.Context, cloudevents.Event, *cloudevents.EventResponse, map[string]string) {},
		"extra out":                    func(context.Context, *cloudevents.EventResponse) (error, int) { return nil, 0 },
		"context dup EventResponse in": func(context.Context, *cloudevents.EventResponse, *cloudevents.EventResponse) {},
		"dup EventResponse in":         func(*cloudevents.EventResponse, *cloudevents.EventResponse) {},
		"context dup Event in":         func(context.Context, cloudevents.Event, cloudevents.Event) {},
		"dup Event in":                 func(cloudevents.Event, cloudevents.Event) {},
		"wrong order, context3 in":     func(*cloudevents.EventResponse, *cloudevents.EventResponse, context.Context) {},
		"wrong order, event in":        func(context.Context, *cloudevents.EventResponse, cloudevents.Event) {},
		"wrong order, resp in":         func(*cloudevents.EventResponse, cloudevents.Event) {},
		"wrong order, context2 in":     func(*cloudevents.EventResponse, context.Context) {},
		"Event as ptr in":              func(*cloudevents.Event) {},
		"EventResponse as non-ptr in":  func(cloudevents.EventResponse) {},
		"extra Event in":               func(cloudevents.Event, *cloudevents.EventResponse, cloudevents.Event) {},
		"not a function":               map[string]string(nil),
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := receiver(fn); err == nil {
				t.Errorf("%q failed to catch the issue", name)
			}
		})
	}
}

func TestReceiverFnInvoke_1(t *testing.T) {
	wantErr := errors.New("UNIT TEST")
	key := struct{}{}
	wantCtx := context.WithValue(context.TODO(), key, "UNIT TEST")
	wantEvent := cloudevents.Event{
		Context: &cloudevents.EventContextV02{
			ID: "UNIT TEST",
		},
	}
	wantResp := &cloudevents.EventResponse{Reason: "UNIT TEST"}

	fn, err := receiver(func(ctx context.Context, event cloudevents.Event, resp *cloudevents.EventResponse) error {
		if diff := cmp.Diff(wantCtx.Value(key), ctx.Value(key)); diff != "" {
			t.Errorf("unexpected context (-want, +got) = %v", diff)
		}

		if diff := cmp.Diff(wantEvent, event); diff != "" {
			t.Errorf("unexpected event (-want, +got) = %v", diff)
		}

		if diff := cmp.Diff(wantResp, resp); diff != "" {
			t.Errorf("unexpected response (-want, +got) = %v", diff)
		}
		return wantErr
	})
	if err != nil {
		t.Errorf("unexpected error, wanted nil got = %v", err)
	}

	err = fn.invoke(wantCtx, wantEvent, wantResp)

	if diff := cmp.Diff(wantErr.Error(), err.Error()); diff != "" {
		t.Errorf("unexpected error (-want, +got) = %v", diff)
	}
}

func TestReceiverFnInvoke_2(t *testing.T) {
	wantErr := errors.New("UNIT TEST")
	key := struct{}{}
	ctx := context.WithValue(context.TODO(), key, "UNIT TEST")
	wantEvent := cloudevents.Event{
		Context: &cloudevents.EventContextV02{
			ID: "UNIT TEST",
		},
	}
	wantResp := &cloudevents.EventResponse{Reason: "UNIT TEST"}

	fn, err := receiver(func(event cloudevents.Event, resp *cloudevents.EventResponse) error {
		if diff := cmp.Diff(wantEvent, event); diff != "" {
			t.Errorf("unexpected event (-want, +got) = %v", diff)
		}

		if diff := cmp.Diff(wantResp, resp); diff != "" {
			t.Errorf("unexpected response (-want, +got) = %v", diff)
		}
		return wantErr
	})
	if err != nil {
		t.Errorf("unexpected error, wanted nil got = %v", err)
	}

	err = fn.invoke(ctx, wantEvent, wantResp)

	if diff := cmp.Diff(wantErr.Error(), err.Error()); diff != "" {
		t.Errorf("unexpected error (-want, +got) = %v", diff)
	}
}

func TestReceiverFnInvoke_3(t *testing.T) {
	key := struct{}{}
	ctx := context.WithValue(context.TODO(), key, "UNIT TEST")
	wantEvent := cloudevents.Event{
		Context: &cloudevents.EventContextV02{
			ID: "UNIT TEST",
		},
	}
	wantResp := &cloudevents.EventResponse{Reason: "UNIT TEST"}

	fn, err := receiver(func(event cloudevents.Event, resp *cloudevents.EventResponse) {
		if diff := cmp.Diff(wantEvent, event); diff != "" {
			t.Errorf("unexpected event (-want, +got) = %v", diff)
		}

		if diff := cmp.Diff(wantResp, resp); diff != "" {
			t.Errorf("unexpected response (-want, +got) = %v", diff)
		}
	})
	if err != nil {
		t.Errorf("unexpected error, wanted nil got = %v", err)
	}

	err = fn.invoke(ctx, wantEvent, wantResp)

	if err != nil {
		t.Errorf("unexpected error, want nil got got = %v", err.Error())
	}
}

func TestReceiverFnInvoke_4(t *testing.T) {
	wantErr := errors.New("UNIT TEST")
	key := struct{}{}
	ctx := context.WithValue(context.TODO(), key, "UNIT TEST")
	event := cloudevents.Event{
		Context: &cloudevents.EventContextV02{
			ID: "UNIT TEST",
		},
	}
	wantResp := &cloudevents.EventResponse{Reason: "UNIT TEST"}

	fn, err := receiver(func(resp *cloudevents.EventResponse) error {
		if diff := cmp.Diff(wantResp, resp); diff != "" {
			t.Errorf("unexpected response (-want, +got) = %v", diff)
		}
		return wantErr
	})
	if err != nil {
		t.Errorf("unexpected error, wanted nil got = %v", err)
	}

	err = fn.invoke(ctx, event, wantResp)

	if diff := cmp.Diff(wantErr.Error(), err.Error()); diff != "" {
		t.Errorf("unexpected error (-want, +got) = %v", diff)
	}
}

func TestReceiverFnInvoke_5(t *testing.T) {
	wantErr := errors.New("UNIT TEST")
	key := struct{}{}
	ctx := context.WithValue(context.TODO(), key, "UNIT TEST")
	event := cloudevents.Event{
		Context: &cloudevents.EventContextV02{
			ID: "UNIT TEST",
		},
	}
	resp := &cloudevents.EventResponse{Reason: "UNIT TEST"}

	fn, err := receiver(func() error {
		return wantErr
	})
	if err != nil {
		t.Errorf("unexpected error, wanted nil got = %v", err)
	}

	err = fn.invoke(ctx, event, resp)

	if diff := cmp.Diff(wantErr.Error(), err.Error()); diff != "" {
		t.Errorf("unexpected error (-want, +got) = %v", diff)
	}
}

func TestReceiverFnInvoke_6(t *testing.T) {
	key := struct{}{}
	ctx := context.WithValue(context.TODO(), key, "UNIT TEST")
	event := cloudevents.Event{
		Context: &cloudevents.EventContextV02{
			ID: "UNIT TEST",
		},
	}
	resp := &cloudevents.EventResponse{Reason: "UNIT TEST"}

	fn, err := receiver(func() {})
	if err != nil {
		t.Errorf("unexpected error, wanted nil got = %v", err)
	}

	err = fn.invoke(ctx, event, resp)

	if err != nil {
		t.Errorf("unexpected error, want nil got got = %v", err.Error())
	}
}
