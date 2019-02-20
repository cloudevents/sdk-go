package types_test

import (
	"github.com/cloudevents/sdk-go/pkg/cloudevents/types"
	"github.com/google/go-cmp/cmp"
	"testing"
	"time"
)

func TestParseTimestamp(t *testing.T) {
	testCases := map[string]struct {
		t    string
		want *types.Timestamp
	}{
		"empty": {
			want: nil,
		},
		"empty string": {
			t:    "",
			want: nil,
		},
		"invalid format": {
			t:    "2019-02-28",
			want: nil,
		},
		"RFC3339 format": {
			t: "1984-02-28T15:04:05Z",
			want: func() *types.Timestamp {
				t, _ := time.Parse(time.RFC3339, "1984-02-28T15:04:05Z")
				return &types.Timestamp{Time: t}
			}(),
		},
		"RFC3339Nano format": {
			t: "1984-02-28T15:04:05.999999999Z",
			want: func() *types.Timestamp {
				t, _ := time.Parse(time.RFC3339Nano, "1984-02-28T15:04:05.999999999Z")
				return &types.Timestamp{Time: t}
			}(),
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			got := types.ParseTimestamp(tc.t)

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("unexpected object (-want, +got) = %v", diff)
			}
		})
	}
}

func TestJsonMarshalTimestamp(t *testing.T) {
	testCases := map[string]struct {
		t    string
		want []byte
	}{
		"empty": {
			want: []byte(`""`),
		},
		"empty string": {
			t:    "",
			want: []byte(`""`),
		},
		"invalid format": {
			t:    "2019-02-28",
			want: []byte(`""`),
		},
		"RFC3339 format": {
			t:    "1984-02-28T15:04:05Z",
			want: []byte(`"1984-02-28T15:04:05Z"`),
		},
		"RFC3339Nano format": {
			t:    "1984-02-28T15:04:05.999999999Z",
			want: []byte(`"1984-02-28T15:04:05.999999999Z"`),
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			tt := types.ParseTimestamp(tc.t)
			got, _ := tt.MarshalJSON()

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Logf("got: %s", string(got))
				t.Errorf("unexpected object (-want, +got) = %v", diff)
			}
		})
	}
}

func TestJsonUnmarshalTimestamp(t *testing.T) {
	testCases := map[string]struct {
		b       []byte
		want    *types.Timestamp
		wantErr string
	}{
		"empty": {
			wantErr: "unexpected end of JSON input",
		},
		"invalid format": {
			b:       []byte("2019-02-28"),
			wantErr: "invalid character '-' after top-level value",
		},
		"RFC3339 format": {
			b: []byte(`"1984-02-28T15:04:05Z"`),
			want: func() *types.Timestamp {
				t, _ := time.Parse(time.RFC3339, "1984-02-28T15:04:05Z")
				return &types.Timestamp{Time: t}
			}(),
		},
		"RFC3339Nano format": {
			b: []byte(`"1984-02-28T15:04:05.999999999Z"`),
			want: func() *types.Timestamp {
				t, _ := time.Parse(time.RFC3339Nano, "1984-02-28T15:04:05.999999999Z")
				return &types.Timestamp{Time: t}
			}(),
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			got := &types.Timestamp{}
			err := got.UnmarshalJSON(tc.b)

			if tc.wantErr != "" || err != nil {
				var gotErr string
				if err != nil {
					gotErr = err.Error()
				}
				if diff := cmp.Diff(tc.wantErr, gotErr); diff != "" {
					t.Errorf("unexpected error (-want, +got) = %v", diff)
				}
				return
			}

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("unexpected object (-want, +got) = %v", diff)
			}
		})
	}
}

func TestTimestampString(t *testing.T) {
	testCases := map[string]struct {
		t    string
		want string
	}{
		"empty": {
			want: "0001-01-01T00:00:00Z",
		},
		"RFC3339 format": {
			t:    "1984-02-28T15:04:05Z",
			want: `1984-02-28T15:04:05Z`,
		},
		"RFC3339Nano format": {
			t:    "1984-02-28T15:04:05.999999999Z",
			want: `1984-02-28T15:04:05.999999999Z`,
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			tt := types.ParseTimestamp(tc.t)
			got := tt.String()

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Logf("got: %s", string(got))
				t.Errorf("unexpected string (-want, +got) = %v", diff)
			}
		})
	}
}
