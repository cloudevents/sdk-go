package types_test

import (
	"github.com/cloudevents/sdk-go/pkg/cloudevents/types"
	"github.com/google/go-cmp/cmp"
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

func TestAllocate(t *testing.T) {

	emptyString := ""
	exampleString := "howdy"

	testCases := map[string]struct {
		obj  interface{}
		want interface{}
	}{
		"nil": {
			obj:  nil,
			want: nil,
		},
		"map": {
			obj: map[string]string{
				"test": "case",
			},
			want: map[string]string{},
		},
		"slice": {
			obj: []string{
				"test",
				"case",
			},
			want: []string{},
		},
		"string": {
			obj:  "hello",
			want: "",
		},
		"string ptr": {
			obj:  &exampleString,
			want: &emptyString,
		},
		"struct": {
			obj: DataExample{
				AnInt: 42,
			},
			want: &DataExample{},
		},
		"pointer": {
			obj: &DataExample{
				AnInt: 42,
			},
			want: &DataExample{},
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			got, _ := types.Allocate(tc.obj)

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("unexpected object (-want, +got) = %v", diff)
			}
		})
	}
}
