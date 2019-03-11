package cloudevents

import (
	"github.com/google/go-cmp/cmp"
	"testing"
)

func TestEventResponse_RespondWith(t *testing.T) {
	testCases := map[string]struct {
		t      *EventResponse
		e      *Event
		status int
		want   *EventResponse
	}{
		"nil": {},
		"valid": {
			t:      &EventResponse{},
			e:      &Event{Data: "unit test"},
			status: 200,
			want: &EventResponse{
				Status: 200,
				Event:  &Event{Data: "unit test"},
			},
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			tc.t.RespondWith(tc.status, tc.e)

			got := tc.t

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("unexpected  (-want, +got) = %v", diff)
			}
		})
	}
}

func TestEventResponse_Error(t *testing.T) {
	testCases := map[string]struct {
		t      *EventResponse
		msg    string
		status int
		want   *EventResponse
	}{
		"nil": {},
		"valid": {
			t:      &EventResponse{},
			msg:    "unit test",
			status: 400,
			want: &EventResponse{
				Status: 400,
				Reason: "unit test",
			},
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			tc.t.Error(tc.status, tc.msg)

			got := tc.t

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("unexpected  (-want, +got) = %v", diff)
			}
		})
	}
}
