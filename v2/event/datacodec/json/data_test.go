/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package json_test

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
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

// Helper function to create string pointers for tests
func stringPtr(s string) *string {
	return &s
}

func TestCodecDecode(t *testing.T) {
	now := time.Now()

	testCases := map[string]struct {
		in      []byte
		want    interface{}
		wantErr string
	}{
		"empty": {
			in:      []byte{},
			wantErr: "out is nil",
		},
		"null": {
			in:   []byte(`null`),
			want: nil,
			wantErr: "out is nil",
		},
		"error": {
			in:      []byte(`"this is not valid json"`),
			want:    &map[string]string{},
			wantErr: `[json] found bytes ""this is not valid json"", but failed to unmarshal:`,
		},
		"text": {
			in:   []byte(`"hello"`),
			want: stringPtr("hello"),
		},
		"blank text": {
			in:   []byte(`""`),
			want: stringPtr(""),
		},
		"number": {
			in:   []byte(`100`),
			want: 100,
		},
		"zero": {
			in:   []byte(`0`),
			want: 0,
		},
		"bool true": {
			in:   []byte(`true`),
			want: true,
		},
		"bool false": {
			in:   []byte(`false`),
			want: false,
		},
		"out nil": {
			in:      []byte{},
			wantErr: "out is nil",
		},
		"wrong unmarshalling receiver": {
			in: []byte("\"something that is not a map\""),
			want: &map[string]string{
				"an": "error",
			},
			wantErr: `[json] found bytes ""something that is not a map"", but failed to unmarshal:`,
		},
		"wrong string": {
			in:      []byte("a non json string"),
			want:    "a non json string",
			wantErr: `[json] found bytes "a non json string", but failed to unmarshal:`,
		},
		"Bad Quotes": {
			in: []byte("\""),
			want: &map[string]string{
				"an": "error",
			},
			wantErr: `[json] found bytes """, but failed to unmarshal:`,
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

			err := cej.Decode(context.TODO(), tc.in, got)

			if tc.wantErr != "" && err != nil {
				errMsg := err.Error()
				if !strings.Contains(errMsg, tc.wantErr) {
					t.Errorf("unexpected error. expected to contain: %q, got: %q", tc.wantErr, errMsg)
				}
			} else if (tc.wantErr == "" && err != nil) || (tc.wantErr != "" && err == nil) {
				t.Errorf("unexpected error. expected: %q, got: %v", tc.wantErr, err)
			}

			if tc.wantErr == "" && tc.want != nil {
				// For simple types, goccy/go-json may return pointers while standard json returns values
				// Need to dereference pointers for comparison
				var gotValue interface{}
				switch v := got.(type) {
				case *int:
					if v != nil {
						gotValue = *v
					} else {
						gotValue = got
					}
				case *bool:
					if v != nil {
						gotValue = *v
					} else {
						gotValue = got
					}
				default:
					gotValue = got
				}

				switch tc.want.(type) {
				case int, bool:
					if diff := cmp.Diff(tc.want, gotValue); diff != "" {
						t.Errorf("unexpected data (-want, +got) = %v", diff)
					}
				default:
					if tc.want != nil && got != nil {
						if diff := cmp.Diff(tc.want, got); diff != "" {
							t.Errorf("unexpected data (-want, +got) = %v", diff)
						}
					}
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
			wantErr: "BadMashal Error",
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

			if tc.wantErr != "" && err != nil {
				errMsg := err.Error()
				if !strings.Contains(errMsg, tc.wantErr) {
					t.Errorf("unexpected error. expected to contain: %q, got: %q", tc.wantErr, errMsg)
				}
			} else if (tc.wantErr == "" && err != nil) || (tc.wantErr != "" && err == nil) {
				t.Errorf("unexpected error. expected: %q, got: %v", tc.wantErr, err)
			}

			if tc.want != nil {
				if diff := cmp.Diff(tc.want, got); diff != "" {
					t.Errorf("unexpected data (-want, +got) = %v", diff)
				}
			}
		})
	}
}
