package event_test

import (
	"strings"
	"testing"
	"time"

	event "github.com/cloudevents/sdk-go/v2/event"

	"github.com/google/go-cmp/cmp"
)

type ReadWriteTest struct {
	event     event.Event
	set       string
	want      interface{}
	corrected interface{} // used in corrected tests.
	wantErr   string
}

func TestEventRW_Type(t *testing.T) {
	testCases := map[string]ReadWriteTest{
		"v03": {
			event: event.New("0.3"),
			set:   "type.0.3",
			want:  "type.0.3",
		},
		"v1": {
			event: event.New("1.0"),
			set:   "type.1.0",
			want:  "type.1.0",
		},
		"spaced v03": {
			event: event.New("0.3"),
			set:   "   type.0.3   ",
			want:  "type.0.3",
		},
		"spaced v1": {
			event: event.New("1.0"),
			set:   "   type.1.0   ",
			want:  "type.1.0",
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			var got interface{}

			tc.event.SetType(tc.set)
			got = tc.event.Type()

			err := tc.event.Validate()
			validateReaderWriter(t, tc, got, err)
		})
	}
}

func TestEventRW_ID(t *testing.T) {
	testCases := map[string]ReadWriteTest{
		"v03 blank": {
			event:   event.New("0.3"),
			set:     "",
			want:    "",
			wantErr: "id is required to be a non-empty string",
		},
		"v1 blank": {
			event:   event.New("1.0"),
			set:     "",
			want:    "",
			wantErr: "id is required to be a non-empty string",
		},
		"v03": {
			event: event.New("0.3"),
			set:   "id.0.3",
			want:  "id.0.3",
		},
		"v1": {
			event: event.New("1.0"),
			set:   "id.1.0",
			want:  "id.1.0",
		},
		"spaced v03": {
			event: event.New("0.3"),
			set:   "   id.0.3   ",
			want:  "id.0.3",
		},
		"spaced v1": {
			event: event.New("1.0"),
			set:   "  id.1.0  ",
			want:  "id.1.0",
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			var got interface{}

			tc.event.SetID(tc.set)
			got = tc.event.ID()

			var err error
			if tc.wantErr != "" {
				err = event.ValidationError(tc.event.FieldErrors)
			}
			validateReaderWriter(t, tc, got, err)
		})
	}
}

func TestEventRW_Source(t *testing.T) {
	testCases := map[string]ReadWriteTest{
		"v03": {
			event: event.New("0.3"),
			set:   "http://example/",
			want:  "http://example/",
		},
		"v1": {
			event: event.New("1.0"),
			set:   "http://example/",
			want:  "http://example/",
		},
		"invalid v03": {
			event:   event.New("0.3"),
			set:     "%",
			want:    "",
			wantErr: "invalid URL escape",
		},
		"invalid v1": {
			event:   event.New("1.0"),
			set:     "%",
			want:    "",
			wantErr: "invalid URL escape",
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			var got interface{}

			tc.event.SetSource(tc.set)
			got = tc.event.Source()

			var err error
			if tc.wantErr != "" {
				err = event.ValidationError(tc.event.FieldErrors)
			}
			validateReaderWriter(t, tc, got, err)
		})
	}
}

// Set will be split on pipe, set1|set2
func TestEventRW_Corrected_Source(t *testing.T) {
	testCases := map[string]ReadWriteTest{
		"corrected v03": {
			event:     event.New("0.3"),
			set:       "%|http://good",
			want:      "",
			corrected: "http://good",
			wantErr:   "invalid URL escape",
		},
		"corrected v1": {
			event:     event.New("1.0"),
			set:       "%|http://good",
			want:      "",
			corrected: "http://good",
			wantErr:   "invalid URL escape",
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			var got interface{}
			var err error

			// Split set on pipe.
			set := strings.Split(tc.set, "|")

			// Set

			tc.event.SetSource(set[0])
			got = tc.event.Source()
			if tc.wantErr != "" {
				err = event.ValidationError(tc.event.FieldErrors)
			}
			validateReaderWriter(t, tc, got, err)

			// Correct

			tc.event.SetSource(set[1])
			got = tc.event.Source()
			if tc.wantErr != "" {
				err = event.ValidationError(tc.event.FieldErrors)
			}
			validateReaderWriterCorrected(t, tc, got, err)
		})
	}
}

func TestEventRW_Subject(t *testing.T) {
	testCases := map[string]ReadWriteTest{
		"v03": {
			event: event.New("0.3"),
			set:   "subject.0.3",
			want:  "subject.0.3",
		},
		"v1": {
			event: event.New("1.0"),
			set:   "subject.1.0",
			want:  "subject.1.0",
		},
		"spaced v03": {
			event: event.New("0.3"),
			set:   "   subject.0.3   ",
			want:  "subject.0.3",
		},
		"spaced v1": {
			event: event.New("1.0"),
			set:   "  subject.1.0  ",
			want:  "subject.1.0",
		},
		"nilled v03": {
			event: func() event.Event {
				e := event.New("0.3")
				e.SetSource("should nil")
				return e
			}(),
			want: "",
		},
		"nilled v1": {
			event: func() event.Event {
				e := event.New("1.0")
				e.SetSource("should nil")
				return e
			}(),
			want: "",
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			var got interface{}

			tc.event.SetSubject(tc.set)
			got = tc.event.Subject()

			err := tc.event.Validate()
			validateReaderWriter(t, tc, got, err)
		})
	}
}

func TestEventRW_Time(t *testing.T) {
	now := time.Now()

	testCases := map[string]ReadWriteTest{
		"v03": {
			event: event.New("0.3"),
			set:   "now", // hack
			want:  now,
		},
		"v1": {
			event: event.New("1.0"),
			set:   "now", // hack
			want:  now,
		},
		"nilled v03": {
			event: func() event.Event {
				e := event.New("0.3")
				e.SetTime(now)
				return e
			}(),
			want: time.Time{},
		},
		"nilled v1": {
			event: func() event.Event {
				e := event.New("1.0")
				e.SetTime(now)
				return e
			}(),
			want: time.Time{},
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			var got interface{}

			if tc.set == "now" {
				tc.event.SetTime(now) // pull now from outer test.
			} else {
				tc.event.SetTime(time.Time{}) // pull now from outer test.
			}
			got = tc.event.Time()

			err := tc.event.Validate()
			validateReaderWriter(t, tc, got, err)
		})
	}
}

func TestEventRW_SchemaURL(t *testing.T) {
	testCases := map[string]ReadWriteTest{
		"v03": {
			event: event.New("0.3"),
			set:   "http://example/",
			want:  "http://example/",
		},
		"v1": {
			event: event.New("1.0"),
			set:   "http://example/",
			want:  "http://example/",
		},
		"invalid v03": {
			event:   event.New("0.3"),
			set:     "%",
			want:    "",
			wantErr: "invalid URL escape",
		},
		"invalid v1": {
			event:   event.New("1.0"),
			set:     "%",
			want:    "",
			wantErr: "invalid URL escape",
		},
		"nilled v03": {
			event: func() event.Event {
				e := event.New("0.3")
				e.SetDataSchema("should nil")
				return e
			}(),
			want: "",
		},
		"nilled v1": {
			event: func() event.Event {
				e := event.New("1.0")
				e.SetDataSchema("should nil")
				return e
			}(),
			want: "",
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			var got interface{}

			tc.event.SetDataSchema(tc.set)
			got = tc.event.DataSchema()

			var err error
			if tc.wantErr != "" {
				err = event.ValidationError(tc.event.FieldErrors)
			}
			validateReaderWriter(t, tc, got, err)
		})
	}
}

func TestEventRW_DataContentType(t *testing.T) {
	testCases := map[string]ReadWriteTest{
		"v03": {
			event: event.New("0.3"),
			set:   "application/json",
			want:  "application/json",
		},
		"v1": {
			event: event.New("1.0"),
			set:   "application/json",
			want:  "application/json",
		},
		"spaced v03": {
			event: event.New("0.3"),
			set:   "   application/json   ",
			want:  "application/json",
		},
		"spaced v1": {
			event: event.New("1.0"),
			set:   "  application/json  ",
			want:  "application/json",
		},
		"nilled v03": {
			event: func() event.Event {
				e := event.New("0.3")
				e.SetDataContentType("application/json")
				return e
			}(),
			want: "",
		},
		"nilled v1": {
			event: func() event.Event {
				e := event.New("1.0")
				e.SetDataContentType("application/json")
				return e
			}(),
			want: "",
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			var got interface{}

			tc.event.SetDataContentType(tc.set)
			got = tc.event.DataContentType()

			err := tc.event.Validate()
			validateReaderWriter(t, tc, got, err)
		})
	}
}

func TestEventRW_DataContentEncoding(t *testing.T) {
	testCases := map[string]ReadWriteTest{
		"v03": {
			event: event.New("0.3"),
			set:   "base64",
			want:  "base64",
		},
		"v1": {
			event: event.New("1.0"),
			set:   "base64",
			want:  "",
		},
		"spaced v03": {
			event: event.New("0.3"),
			set:   "   base64   ",
			want:  "base64",
		},
		"spaced v1": {
			event: event.New("1.0"),
			set:   "  base64  ",
			want:  "",
		},
		"cased v03": {
			event: event.New("0.3"),
			set:   "   BaSe64   ",
			want:  "base64",
		},
		"cased v1": {
			event: event.New("1.0"),
			set:   "  BaSe64  ",
			want:  "",
		},
		"nilled v03": {
			event: func() event.Event {
				e := event.New("0.3")
				e.SetDataContentEncoding("base64")
				return e
			}(),
			want: "",
		},
		"nilled v1": {
			event: func() event.Event {
				e := event.New("1.0")
				e.SetDataContentEncoding("base64")
				return e
			}(),
			want: "",
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			var got interface{}

			tc.event.SetDataContentEncoding(tc.set)
			got = tc.event.DeprecatedDataContentEncoding()

			err := tc.event.Validate()
			validateReaderWriter(t, tc, got, err)
		})
	}
}

func validateReaderWriter(t *testing.T, tc ReadWriteTest, got interface{}, err error) {
	var gotErr string
	if err != nil {
		gotErr = err.Error()
	}
	if tc.wantErr != "" {
		if !strings.Contains(gotErr, tc.wantErr) {
			t.Errorf("unexpected error, expected to contain %q, got: %q ", tc.wantErr, gotErr)
		}
	}
	if diff := cmp.Diff(tc.want, got); diff != "" {
		t.Errorf("unexpected (-want, +got) = %v", diff)
	}
}

func validateReaderWriterCorrected(t *testing.T, tc ReadWriteTest, got interface{}, err error) {
	var gotErr string
	if err != nil {
		gotErr = err.Error()
	}
	if tc.wantErr != "" {
		if strings.Contains(gotErr, tc.wantErr) {
			t.Errorf("unexpected error, expected to NOT contain %q, got: %q ", tc.wantErr, gotErr)
		}
	}
	if diff := cmp.Diff(tc.corrected, got); diff != "" {
		t.Errorf("unexpected (-want, +got) = %v", diff)
	}
}
