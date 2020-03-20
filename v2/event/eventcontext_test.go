package event_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/google/go-cmp/cmp"

	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/cloudevents/sdk-go/v2/types"
)

func TestContextAsV03(t *testing.T) {
	now := types.Timestamp{Time: time.Now()}

	testCases := map[string]struct {
		event event.Event
		want  *event.EventContextV03
	}{
		"min v03, no conversion": {
			event: event.Event{
				Context: MinEventContextV03(),
			},
			want: MinEventContextV03(),
		},
		"full v03, no conversion": {
			event: event.Event{
				Context: FullEventContextV03(now),
			},
			want: FullEventContextV03(now),
		},
		"min v1 -> v03": {
			event: event.Event{
				Context: MinEventContextV1(),
			},
			want: MinEventContextV03(),
		},
		"full v1 -> v03": {
			event: event.Event{
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
		event event.Event
		want  *event.EventContextV1
	}{
		"min v03 -> v1": {
			event: event.Event{
				Context: MinEventContextV03(),
			},
			want: MinEventContextV1(),
		},
		"full v03 -> v1": {
			event: event.Event{
				Context: FullEventContextV03(now),
			},
			want: FullEventContextV1(now),
		},
		"min v1, no conversion": {
			event: event.Event{
				Context: MinEventContextV1(),
			},
			want: MinEventContextV1(),
		},
		"full v1, no conversion": {
			event: event.Event{
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

func TestCloneEventContextV03(t *testing.T) {
	tests := []struct {
		name    string
		context event.EventContext
	}{
		{
			name:    "v0.3 min",
			context: MinEventContextV03(),
		},
		{
			name:    "v0.3 full",
			context: FullEventContextV03(types.Timestamp{Time: time.Now()}),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			initial := test.context
			require.NoError(t, initial.SetExtension("aaa", "bbb"))

			clone := initial.Clone()

			initialv03 := initial.(*event.EventContextV03)
			clonev03 := clone.(*event.EventContextV03)
			require.True(t, reflect.DeepEqual(initialv03, clonev03))
			require.NotSame(t, &initialv03.Source.URL, &clonev03.Source.URL)
			if initialv03.Time != nil {
				require.NotSame(t, &initialv03.Time.Time, &clonev03.Time.Time)
			}
			if initialv03.SchemaURL != nil {
				require.NotSame(t, &initialv03.SchemaURL.URL, &clonev03.SchemaURL.URL)
			}

			// Test mutate extensions
			require.NoError(t, clone.SetExtension("aaa", "ccc"))
			val, err := initial.GetExtension("aaa")
			require.NoError(t, err)
			require.Equal(t, "bbb", val)
		})
	}
}

func TestCloneEventContextV1(t *testing.T) {
	tests := []struct {
		name    string
		context event.EventContext
	}{
		{
			name:    "v1.0 min",
			context: MinEventContextV1(),
		},
		{
			name:    "v1.0 full",
			context: FullEventContextV1(types.Timestamp{Time: time.Now()}),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			initial := test.context
			require.NoError(t, initial.SetExtension("aaa", "bbb"))

			clone := initial.Clone()

			initialv1 := initial.(*event.EventContextV1)
			clonev1 := clone.(*event.EventContextV1)
			require.True(t, reflect.DeepEqual(initialv1, clonev1))
			require.NotSame(t, initialv1, clonev1)
			require.NotSame(t, &initialv1.Source.URL, &clonev1.Source.URL)
			if initialv1.Time != nil {
				require.NotSame(t, &initialv1.Time.Time, &clonev1.Time.Time)
			}
			if initialv1.DataSchema != nil {
				require.NotSame(t, &initialv1.DataSchema.URL, &clonev1.DataSchema.URL)
			}

			// Test mutate extensions
			require.NoError(t, clone.SetExtension("aaa", "ccc"))
			val, err := initial.GetExtension("aaa")
			require.NoError(t, err)
			require.Equal(t, "bbb", val)
		})
	}
}
