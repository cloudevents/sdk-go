package json_test

import (
	"encoding/json"
	"fmt"
	cej "github.com/cloudevents/sdk-go/pkg/cloudevents/datacodec/json"
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

type BadMarshal struct{}

func (b BadMarshal) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("BadMashal Error")
}

func TestCodecDecode(t *testing.T) {
	now := time.Now()

	testCases := map[string]struct {
		in      interface{}
		want    interface{}
		wantErr string
	}{
		"empty": {},
		"out nil": {
			in:      "not nil",
			wantErr: "out is nil",
		},
		"not a []byte": {
			in: "something that is not a map",
			want: &map[string]string{
				"an": "error",
			},
			wantErr: `[json] found bytes ""something that is not a map"", but failed to unmarshal: json: cannot unmarshal string into Go value of type map[string]string`,
		},
		"BadMarshal": {
			in:      BadMarshal{},
			want:    &BadMarshal{},
			wantErr: "[json] failed to marshal in: json: error calling MarshalJSON for type json_test.BadMarshal: BadMashal Error",
		},
		"Bad Quotes": {
			in: []byte{'\\', '"'},
			want: &map[string]string{
				"an": "error",
			},
			wantErr: "[json] failed to unquote in: invalid syntax",
		},
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
		"simple array": {
			in: []byte(`["apple","banana"]`),
			want: &[]string{
				"apple",
				"banana",
			},
		},
		"simple quoted array": {
			in: []byte(`"[\"apple\",\"banana\"]"`),
			want: &[]string{
				"apple",
				"banana",
			},
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
			got, _ := types.Allocate(tc.want)

			err := cej.Decode(tc.in, got)
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

			if tc.want != nil {
				if diff := cmp.Diff(tc.want, got); diff != "" {
					t.Errorf("unexpected data (-want, +got) = %v", diff)
				}
			}
		})
	}
}

// TODO: test for bad []byte input?
func TestCodecEncode(t *testing.T) {
	now := time.Now()

	testCases := map[string]struct {
		in      interface{}
		want    []byte
		wantErr string
	}{
		"empty": {},
		"BadMarshal": {
			in:      BadMarshal{},
			wantErr: "json: error calling MarshalJSON for type json_test.BadMarshal: BadMashal Error",
		},
		"already encoded object": {
			in:   []byte(`{"a":"apple","b":"banana"}`),
			want: []byte(`{"a":"apple","b":"banana"}`),
		},
		"already encoded quote": {
			in:   []byte(`"{"a":"apple","b":"banana"}"`),
			want: []byte(`"{"a":"apple","b":"banana"}"`),
		},
		"already encoded slice": {
			in:   []byte(`["apple","banana"]`),
			want: []byte(`["apple","banana"]`),
		},
		"simple": {
			in: map[string]string{
				"a": "apple",
				"b": "banana",
			},
			want: []byte(`{"a":"apple","b":"banana"}`),
		},
		"complex empty": {
			in:   DataExample{},
			want: []byte(`{}`),
		},
		"simple array": {
			in: &[]string{
				"apple",
				"banana",
			},
			want: []byte(`["apple","banana"]`),
		},
		"complex filled": {
			in: &DataExample{
				AnInt: 42,
				AMap: map[string]map[string]int{
					"a": {"1": 1, "2": 2, "3": 3},
					"z": {"3": 3, "2": 2, "1": 1},
				},
				AString: "Hello, World!",
				ATime:   &now,
				AnArray: []string{"Anne", "Bob", "Chad"},
			},
			want: func() []byte {
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
		},
		"object in": {
			in: &DataExample{
				AnInt: 42,
			},
			want: []byte(`{"a":42}`),
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			got, err := cej.Encode(tc.in)
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

			if tc.want != nil {
				if diff := cmp.Diff(tc.want, got); diff != "" {
					t.Errorf("unexpected data (-want, +got) = %v", diff)
				}
			}
		})
	}
}
