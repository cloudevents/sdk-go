package json_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	cej "github.com/cloudevents/sdk-go/v2/event/datacodec/json"
	"github.com/cloudevents/sdk-go/v2/types"
	"github.com/google/go-cmp/cmp"
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
		in      []byte
		want    interface{}
		wantErr string
	}{
		"empty": {},
		"out nil": {
			in:      []byte{},
			wantErr: "out is nil",
		},
		"wrong unmarshalling receiver": {
			in: []byte("\"something that is not a map\""),
			want: &map[string]string{
				"an": "error",
			},
			wantErr: `[json] found bytes ""something that is not a map"", but failed to unmarshal: json: cannot unmarshal string into Go value of type map[string]string`,
		},
		"wrong string": {
			in:      []byte("a non json string"),
			want:    "a non json string",
			wantErr: `[json] found bytes "a non json string", but failed to unmarshal: invalid character 'a' looking for beginning of value`,
		},
		"Bad Quotes": {
			in: []byte("\""),
			want: &map[string]string{
				"an": "error",
			},
			wantErr: `[json] found bytes """, but failed to unmarshal: unexpected end of JSON input`,
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
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			got, _ := types.Allocate(tc.want)
			gotObs, _ := types.Allocate(tc.want)

			err := cej.Decode(context.TODO(), tc.in, got)
			errObs := cej.DecodeObserved(context.TODO(), tc.in, gotObs)

			if diff := cmpErrors(tc.wantErr, errObs); diff != "" {
				t.Errorf("obs unexpected error (-want, +got) = %v", diff)
			}

			if diff := cmpErrors(tc.wantErr, err); diff != "" {
				t.Errorf("unexpected error (-want, +got) = %v", diff)
			}

			if diff := cmp.Diff(gotObs, got); diff != "" {
				t.Errorf("obs unexpected obj diff between observed and direct (-want, +got) = %v", diff)
			}

			if tc.wantErr == "" && tc.want != nil {
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
			got, err := cej.Encode(context.TODO(), tc.in)

			gotObs, errObs := cej.EncodeObserved(context.TODO(), tc.in)

			if diff := cmpErrors(tc.wantErr, errObs); diff != "" {
				t.Errorf("obs unexpected error (-want, +got) = %v", diff)
			}
			if diff := cmp.Diff(gotObs, got); diff != "" {
				t.Errorf("obs unexpected obj diff between observed and direct (-want, +got) = %v", diff)
			}

			if diff := cmpErrors(tc.wantErr, err); diff != "" {
				t.Errorf("unexpected error (-want, +got) = %v", diff)
			}

			if tc.want != nil {
				if diff := cmp.Diff(tc.want, got); diff != "" {
					t.Errorf("unexpected data (-want, +got) = %v", diff)
				}
			}
		})
	}
}

func cmpErrors(want string, err error) string {
	if want != "" || err != nil {
		var gotErr string
		if err != nil {
			gotErr = err.Error()
		}
		return cmp.Diff(want, gotErr)
	}
	return ""
}
