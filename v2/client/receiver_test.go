package client

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/cloudevents/sdk-go/v2/protocol"

	"github.com/google/go-cmp/cmp"
)

func TestReceiverFnValidTypes(t *testing.T) {
	for name, fn := range map[string]interface{}{
		"no in, no out": func() {},

		"ctx in, no out":       func(context.Context) {},
		"Event in, no out":     func(event.Event) {},
		"ctx+Event in, no out": func(context.Context, event.Event) {},

		"no in, error out":       func() error { return nil },
		"no in, Result out":      func() protocol.Result { return nil },
		"no in, Event+error out": func() (*event.Event, error) { return nil, nil },

		"ctx in, error out":       func(context.Context) error { return nil },
		"Event in, error out":     func(event.Event) error { return nil },
		"ctx+Event in, error out": func(context.Context, event.Event) error { return nil },

		"ctx in, Event out":       func(context.Context) *event.Event { return nil },
		"Event in, Event out":     func(event.Event) *event.Event { return nil },
		"ctx+Event in, Event out": func(context.Context, event.Event) *event.Event { return nil },

		"ctx in, Event+error out":       func(context.Context) (*event.Event, error) { return nil, nil },
		"Event in, Event+error out":     func(event.Event) (*event.Event, error) { return nil, nil },
		"ctx+Event in, Event+error out": func(context.Context, event.Event) (*event.Event, error) { return nil, nil },

		"ctx in, Event+Result out":       func(context.Context) (*event.Event, protocol.Result) { return nil, nil },
		"Event in, Event+Result out":     func(event.Event) (*event.Event, protocol.Result) { return nil, nil },
		"ctx+Event in, Event+Result out": func(context.Context, event.Event) (*event.Event, protocol.Result) { return nil, nil },

		"input contravariance; may accept supertype": func(event.EventReader) {},
		"output covariance; may return subtype":      func() *myErr { return nil },
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
		"wrong type in":            func(string) {},
		"wrong type out":           func() string { return "" },
		"extra in":                 func(context.Context, event.Event, map[string]string) {},
		"extra out":                func(context.Context) (int, error) { return 0, nil },
		"dup error out":            func(context.Context) (protocol.Result, error) { return nil, nil },
		"context dup Event out":    func(context.Context) (*event.Event, *event.Event) { return nil, nil },
		"context dup Event in":     func(context.Context, event.Event, event.Event) {},
		"dup Event in":             func(event.Event, event.Event) {},
		"wrong order, context3 in": func(*event.Event, event.Event, context.Context) {},
		"wrong order, event in":    func(context.Context, *event.Event, event.Event) {},
		"wrong order, resp in":     func(*event.Event, event.Event) {},
		"wrong order, context2 in": func(*event.Event, context.Context) {},
		"Event as ptr in":          func(*event.Event) {},
		"Event as non-ptr out":     func() event.Event { return event.Event{} },
		"extra Event in":           func(event.Event, event.Event) {},
		"not a function":           map[string]string(nil),

		"input covariance; must not accept subtype": func(*myCtx) {},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := receiver(fn); err == nil {
				t.Errorf("%q failed to catch the issue", name)
			}
		})
	}
}

func TestReceiverFnInvoke_1(t *testing.T) {
	key := struct{}{}
	wantCtx := context.WithValue(context.TODO(), key, "UNIT TEST")
	wantEvent := event.Event{
		Context: &event.EventContextV1{ID: "UNIT TEST"},
	}
	wantResp := &event.Event{
		Context: &event.EventContextV1{ID: "UNIT TEST"},
	}
	wantResult := errors.New("UNIT TEST")

	fn, err := receiver(func(ctx context.Context, event event.Event) (*event.Event, protocol.Result) {
		if diff := cmp.Diff(wantCtx.Value(key), ctx.Value(key)); diff != "" {
			t.Errorf("unexpected context (-want, +got) = %v", diff)
		}

		if diff := cmp.Diff(wantEvent, event); diff != "" {
			t.Errorf("unexpected event (-want, +got) = %v", diff)
		}

		return wantResp, wantResult
	})
	if err != nil {
		t.Errorf("unexpected error, wanted nil got = %v", err)
	}

	resp, result := fn.invoke(wantCtx, &wantEvent)

	if diff := cmp.Diff(wantResp, resp); diff != "" {
		t.Errorf("unexpected response (-want, +got) = %v", diff)
	}

	if diff := cmp.Diff(wantResult.Error(), result.Error()); diff != "" {
		t.Errorf("unexpected error (-want, +got) = %v", diff)
	}
}

func TestReceiverFnInvoke_2(t *testing.T) {
	key := struct{}{}
	ctx := context.WithValue(context.TODO(), key, "UNIT TEST")
	wantEvent := event.Event{
		Context: &event.EventContextV1{
			ID: "UNIT TEST",
		},
	}
	wantResp := &event.Event{
		Context: &event.EventContextV1{ID: "UNIT TEST"},
	}
	wantResult := errors.New("UNIT TEST")

	fn, err := receiver(func(event event.Event) (*event.Event, protocol.Result) {
		if diff := cmp.Diff(wantEvent, event); diff != "" {
			t.Errorf("unexpected event (-want, +got) = %v", diff)
		}
		return wantResp, wantResult
	})
	if err != nil {
		t.Errorf("unexpected error, wanted nil got = %v", err)
	}

	resp, result := fn.invoke(ctx, &wantEvent)

	if diff := cmp.Diff(wantResp, resp); diff != "" {
		t.Errorf("unexpected response (-want, +got) = %v", diff)
	}

	if diff := cmp.Diff(wantResult.Error(), result.Error()); diff != "" {
		t.Errorf("unexpected error (-want, +got) = %v", diff)
	}
}

func TestReceiverFnInvoke_3(t *testing.T) {
	key := struct{}{}
	ctx := context.WithValue(context.TODO(), key, "UNIT TEST")
	wantEvent := event.Event{
		Context: &event.EventContextV1{
			ID: "UNIT TEST",
		},
	}
	wantResp := &event.Event{
		Context: &event.EventContextV1{ID: "UNIT TEST"},
	}

	fn, err := receiver(func(e event.Event) *event.Event {
		if diff := cmp.Diff(wantEvent, e); diff != "" {
			t.Errorf("unexpected event (-want, +got) = %v", diff)
		}

		return wantResp
	})
	if err != nil {
		t.Errorf("unexpected error, wanted nil got = %v", err)
	}

	resp, result := fn.invoke(ctx, &wantEvent)

	if diff := cmp.Diff(wantResp, resp); diff != "" {
		t.Errorf("unexpected response (-want, +got) = %v", diff)
	}

	if result != nil {
		t.Errorf("unexpected error (-want, +got) = %v", result)
	}
}

func TestReceiverFnInvoke_4(t *testing.T) {
	key := struct{}{}
	ctx := context.WithValue(context.TODO(), key, "UNIT TEST")
	wantResp := &event.Event{
		Context: &event.EventContextV1{ID: "UNIT TEST"},
	}
	wantResult := errors.New("UNIT TEST")

	fn, err := receiver(func() (*event.Event, protocol.Result) {
		return wantResp, wantResult
	})
	if err != nil {
		t.Errorf("unexpected error, wanted nil got = %v", err)
	}

	resp, result := fn.invoke(ctx, &event.Event{})

	if diff := cmp.Diff(wantResp, resp); diff != "" {
		t.Errorf("unexpected response (-want, +got) = %v", diff)
	}

	if diff := cmp.Diff(wantResult.Error(), result.Error()); diff != "" {
		t.Errorf("unexpected error (-want, +got) = %v", diff)
	}
}

func TestReceiverFnInvoke_5(t *testing.T) {
	key := struct{}{}
	ctx := context.WithValue(context.TODO(), key, "UNIT TEST")

	var wantResp *event.Event
	wantResult := errors.New("UNIT TEST")

	fn, err := receiver(func() protocol.Result {
		return wantResult
	})
	if err != nil {
		t.Errorf("unexpected error, wanted nil got = %v", err)
	}

	resp, result := fn.invoke(ctx, &event.Event{})

	if diff := cmp.Diff(wantResp, resp); diff != "" {
		t.Errorf("unexpected response (-want, +got) = %v", diff)
	}

	if diff := cmp.Diff(wantResult.Error(), result.Error()); diff != "" {
		t.Errorf("unexpected error (-want, +got) = %v", diff)
	}
}

type myErr struct {
}

func (m *myErr) Error() string {
	panic("implement me")
}

type myCtx struct {
}

func (m myCtx) Deadline() (deadline time.Time, ok bool) {
	panic("implement me")
}

func (m myCtx) Done() <-chan struct{} {
	panic("implement me")
}

func (m myCtx) Err() error {
	panic("implement me")
}

func (m myCtx) Value(key interface{}) interface{} {
	panic("implement me")
}
