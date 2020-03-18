package cloudevents_test

import (
	"testing"
	"time"

	ce "github.com/cloudevents/sdk-go/v1/cloudevents"
	"github.com/cloudevents/sdk-go/v1/cloudevents/types"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
)

func TestContextAsV01(t *testing.T) {
	now := types.Timestamp{Time: time.Now()}

	testCases := map[string]struct {
		event ce.Event
		want  *ce.EventContextV01
	}{
		"empty, no conversion": {
			event: ce.Event{
				Context: &ce.EventContextV01{},
			},
			want: &ce.EventContextV01{
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
		"min v03 -> v01": {
			event: ce.Event{
				Context: MinEventContextV03(),
			},
			want: MinEventContextV01(),
		},
		"full v03 -> v01": {
			event: ce.Event{
				Context: FullEventContextV03(now),
			},
			want: FullEventContextV01(now),
		},
		"min v1 -> v01": {
			event: ce.Event{
				Context: MinEventContextV1(),
			},
			want: MinEventContextV01(),
		},
		"full v1 -> v01": {
			event: ce.Event{
				Context: FullEventContextV1(now),
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
		want  *ce.EventContextV02
	}{
		"empty, no conversion": {
			event: ce.Event{
				Context: &ce.EventContextV02{},
			},
			want: &ce.EventContextV02{
				SpecVersion: "0.2",
			},
		},
		"min v01 -> v02": {
			event: ce.Event{
				Context: MinEventContextV01(),
			},
			want: MinEventContextV02(),
		},
		"full v01 -> v02": {
			event: ce.Event{
				Context: FullEventContextV01(now),
			},
			want: FullEventContextV02(now),
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
		"min v03 -> v02": {
			event: ce.Event{
				Context: MinEventContextV03(),
			},
			want: MinEventContextV02(),
		},
		"full v03 -> v02": {
			event: ce.Event{
				Context: FullEventContextV03(now),
			},
			want: FullEventContextV02(now),
		},

		"min v1 -> v02": {
			event: ce.Event{
				Context: MinEventContextV1(),
			},
			want: MinEventContextV02(),
		},
		"full v1 -> v02": {
			event: ce.Event{
				Context: FullEventContextV1(now),
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

func TestContextAsV03(t *testing.T) {
	now := types.Timestamp{Time: time.Now()}

	testCases := map[string]struct {
		event ce.Event
		want  *ce.EventContextV03
	}{
		"empty, no conversion": {
			event: ce.Event{
				Context: &ce.EventContextV03{},
			},
			want: &ce.EventContextV03{
				SpecVersion: "0.3",
			},
		},
		"min v01 -> v03": {
			event: ce.Event{
				Context: MinEventContextV01(),
			},
			want: MinEventContextV03(),
		},
		"full v01 -> v03": {
			event: ce.Event{
				Context: FullEventContextV01(now),
			},
			want: FullEventContextV03(now),
		},
		"min v02 -> v03": {
			event: ce.Event{
				Context: MinEventContextV02(),
			},
			want: MinEventContextV03(),
		},
		"full v02 -> v03": {
			event: ce.Event{
				Context: FullEventContextV02(now),
			},
			want: FullEventContextV03(now),
		},
		"min v03, no conversion": {
			event: ce.Event{
				Context: MinEventContextV03(),
			},
			want: MinEventContextV03(),
		},
		"full v03, no conversion": {
			event: ce.Event{
				Context: FullEventContextV03(now),
			},
			want: FullEventContextV03(now),
		},
		"min v1 -> v03": {
			event: ce.Event{
				Context: MinEventContextV1(),
			},
			want: MinEventContextV03(),
		},
		"full v1 -> v03": {
			event: ce.Event{
				Context: FullEventContextV1(now),
			},
			want: FullEventContextV03(now),
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			got := tc.event.Context.AsV03()

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("unexpected (-want, +got) = %v", diff)
			}
		})
	}
}

func TestContextAsV1(t *testing.T) {
	now := types.Timestamp{Time: time.Now()}

	testCases := map[string]struct {
		event ce.Event
		want  *ce.EventContextV1
	}{
		"empty, no conversion": {
			event: ce.Event{
				Context: &ce.EventContextV1{},
			},
			want: &ce.EventContextV1{
				SpecVersion: "1.0",
			},
		},
		"min v01 -> v1": {
			event: ce.Event{
				Context: MinEventContextV01(),
			},
			want: MinEventContextV1(),
		},
		"full v01 -> v1": {
			event: ce.Event{
				Context: FullEventContextV01(now),
			},
			want: FullEventContextV1(now),
		},
		"min v02 -> v1": {
			event: ce.Event{
				Context: MinEventContextV02(),
			},
			want: MinEventContextV1(),
		},
		"full v02 -> v1": {
			event: ce.Event{
				Context: FullEventContextV02(now),
			},
			want: FullEventContextV1(now),
		},
		"min v03 -> v1": {
			event: ce.Event{
				Context: MinEventContextV03(),
			},
			want: MinEventContextV1(),
		},
		"full v03 -> v1": {
			event: ce.Event{
				Context: FullEventContextV03(now),
			},
			want: FullEventContextV1(now),
		},
		"min v1, no conversion": {
			event: ce.Event{
				Context: MinEventContextV1(),
			},
			want: MinEventContextV1(),
		},
		"full v1, no conversion": {
			event: ce.Event{
				Context: FullEventContextV1(now),
			},
			want: FullEventContextV1(now),
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			got := tc.event.Context.AsV1()

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("unexpected (-want, +got) = %v", diff)
			}
		})
	}
}

func TestEventContextClone(t *testing.T) {
	tests := []struct {
		name    string
		context ce.EventContext
	}{
		{
			name:    "v0.1",
			context: FullEventContextV01(types.Timestamp{Time: time.Now()}),
		},
		{
			name:    "v0.2",
			context: FullEventContextV02(types.Timestamp{Time: time.Now()}),
		},
		{
			name:    "v0.3",
			context: FullEventContextV03(types.Timestamp{Time: time.Now()}),
		},
		{
			name:    "v1.0",
			context: FullEventContextV1(types.Timestamp{Time: time.Now()}),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			initial := test.context
			require.NoError(t, initial.SetExtension("aaa", "bbb"))

			clone := initial.Clone()
			require.NoError(t, clone.SetExtension("aaa", "ccc"))

			val, err := initial.GetExtension("aaa")
			require.NoError(t, err)
			require.Equal(t, "bbb", val)
		})
	}
}
