package json_test

import (
	"encoding/json"
	cej "github.com/cloudevents/sdk-go/pkg/cloudevents/datacodec/json"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/types"
	"github.com/google/go-cmp/cmp"
	"reflect"
	"testing"
	"time"
)

type DataExample struct {
	AnInt   int                       `json:"a,omitempty"`
	AString string                    `json:"b,omitempty"`
	AnArray []string                  `json:"c,omitempty"`
	AMap    map[string]map[string]int `json:"d,omitempty"`
	ATime   *time.Time                `json:"e,omitempty"`
}

func TestCodecDecode(t *testing.T) {
	now := time.Now()

	testCases := map[string]struct {
		in      interface{}
		want    interface{}
		wantErr error
	}{
		"empty": {},
		"simple": {
			in: []byte(`{"a":"apple","b":"banana"}`),
			want: &map[string]string{
				"a": "apple",
				"b": "banana",
			},
		},
		"complex empty": {
			in:   []byte(`{}`),
			want: &DataExample{},
		},
		"complex filled": {
			in: func() []byte {
				data := &DataExample{
					AnInt: 42,
					AMap: map[string]map[string]int{
						"a": {"1": 1, "2": 2, "3": 3},
						"z": {"3": 3, "2": 2, "1": 1},
					},
					AString: "Hello, World!",
					ATime:   &now,
					AnArray: []string{"Anne", "Bob", "Chad"},
				}

				j, err := json.Marshal(data)
				if err != nil {
					t.Errorf("failed to marshal test data: %s", err.Error())
				}
				return j
			}(),
			want: &DataExample{
				AnInt: 42,
				AMap: map[string]map[string]int{
					"a": {"1": 1, "2": 2, "3": 3},
					"z": {"3": 3, "2": 2, "1": 1},
				},
				AString: "Hello, World!",
				ATime:   &now,
				AnArray: []string{"Anne", "Bob", "Chad"},
			},
		},
		"object in": {
			in: &DataExample{
				AnInt: 42,
			},
			want: &DataExample{
				AnInt: 42,
			},
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			dataType := reflect.TypeOf(tc.want)

			t.Logf("got dataType: %s", dataType)

			got, _ := types.Allocate(dataType)

			err := cej.Decode(tc.in, got)

			if tc.wantErr != nil || err != nil {
				if diff := cmp.Diff(tc.wantErr, err); diff != "" {
					t.Errorf("unexpected error (-want, +got) = %v", diff)
				}
				return
			}

			if dataType != nil {
				if diff := cmp.Diff(tc.want, got); diff != "" {
					t.Errorf("unexpected data (-want, +got) = %v", diff)
				}
			}
		})
	}
}
