package cloudevents_test

import (
	"strings"
	"testing"
	"time"

	ce "github.com/cloudevents/sdk-go/v1/cloudevents"
	"github.com/google/go-cmp/cmp"
)

type ReadWriteTest struct {
	event     ce.Event
	set       string
	want      interface{}
	corrected interface{} // used in corrected tests.
	wantErr   string
}

func TestEventRW_SpecVersion(t *testing.T) {
	testCases := map[string]ReadWriteTest{
		"empty v01": {
			event:   ce.New(),
			want:    "1.0",
			set:     "0.1",
			wantErr: "invalid version",
		},
		"empty v02": {
			event:   ce.New(),
			want:    "1.0",
			set:     "0.2",
			wantErr: "invalid version",
		},
		"empty v03": {
			event:   ce.New(),
			want:    "1.0",
			set:     "0.3",
			wantErr: "invalid version",
		},
		"empty v1": {
			event: ce.New(),
			set:   "1.0",
			want:  "1.0",
		},
		"v01": {
			event: ce.New("0.1"),
			set:   "0.1",
			want:  "0.1",
		},
		"v02": {
			event: ce.New("0.2"),
			set:   "0.2",
			want:  "0.2",
		},
		"v03": {
			event: ce.New("0.3"),
			set:   "0.3",
			want:  "0.3",
		},
		"v1": {
			event: ce.New("1.0"),
			set:   "1.0",
			want:  "1.0",
		},
		"invalid v01": {
			event:   ce.New("0.1"),
			want:    "0.1",
			set:     "1.1",
			wantErr: "invalid version",
		},
		"invalid v02": {
			event:   ce.New("0.2"),
			want:    "0.2",
			set:     "1.2",
			wantErr: "invalid version",
		},
		"invalid v03": {
			event:   ce.New("0.3"),
			want:    "0.3",
			set:     "1.3",
			wantErr: "invalid version",
		},
		"invalid v1": {
			event:   ce.New("1.0"),
			want:    "1.0",
			set:     "1.3",
			wantErr: "invalid version",
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			var got interface{}

			tc.event.SetSpecVersion(tc.set)
			got = tc.event.SpecVersion()

			err := tc.event.Validate()
			validateReaderWriter(t, tc, got, err)
		})
	}
}

func TestEventRW_Type(t *testing.T) {
	testCases := map[string]ReadWriteTest{
		"v01": {
			event: ce.New("0.1"),
			set:   "type.0.1",
			want:  "type.0.1",
		},
		"v02": {
			event: ce.New("0.2"),
			set:   "type.0.2",
			want:  "type.0.2",
		},
		"v03": {
			event: ce.New("0.3"),
			set:   "type.0.3",
			want:  "type.0.3",
		},
		"spaced v01": {
			event: ce.New("0.1"),
			set:   "  type.0.1  ",
			want:  "type.0.1",
		},
		"spaced v02": {
			event: ce.New("0.2"),
			set:   "  type.0.2  ",
			want:  "type.0.2",
		},
		"spaced v03": {
			event: ce.New("0.3"),
			set:   "   type.0.3   ",
			want:  "type.0.3",
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
		"v01": {
			event: ce.New("0.1"),
			set:   "id.0.1",
			want:  "id.0.1",
		},
		"v02": {
			event: ce.New("0.2"),
			set:   "id.0.2",
			want:  "id.0.2",
		},
		"v03": {
			event: ce.New("0.3"),
			set:   "id.0.3",
			want:  "id.0.3",
		},
		"spaced v01": {
			event: ce.New("0.1"),
			set:   "  id.0.1  ",
			want:  "id.0.1",
		},
		"spaced v02": {
			event: ce.New("0.2"),
			set:   "  id.0.2  ",
			want:  "id.0.2",
		},
		"spaced v03": {
			event: ce.New("0.3"),
			set:   "   id.0.3   ",
			want:  "id.0.3",
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			var got interface{}

			tc.event.SetID(tc.set)
			got = tc.event.ID()

			err := tc.event.Validate()
			validateReaderWriter(t, tc, got, err)
		})
	}
}

func TestEventRW_Source(t *testing.T) {
	testCases := map[string]ReadWriteTest{
		"v01": {
			event: ce.New("0.1"),
			set:   "http://example/",
			want:  "http://example/",
		},
		"v02": {
			event: ce.New("0.2"),
			set:   "http://example/",
			want:  "http://example/",
		},
		"v03": {
			event: ce.New("0.3"),
			set:   "http://example/",
			want:  "http://example/",
		},
		"invalid v01": {
			event:   ce.New("0.1"),
			set:     "%",
			want:    "",
			wantErr: "invalid URL escape",
		},
		"invalid v02": {
			event:   ce.New("0.2"),
			set:     "%",
			want:    "",
			wantErr: "invalid URL escape",
		},
		"invalid v03": {
			event:   ce.New("0.3"),
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

			err := tc.event.Validate()
			validateReaderWriter(t, tc, got, err)
		})
	}
}

// Set will be split on pipe, set1|set2
func TestEventRW_Corrected_Source(t *testing.T) {
	testCases := map[string]ReadWriteTest{
		"corrected v01": {
			event:     ce.New("0.1"),
			set:       "%|http://good",
			want:      "",
			corrected: "http://good",
			wantErr:   "invalid URL escape",
		},
		"corrected v02": {
			event:     ce.New("0.2"),
			set:       "%|http://good",
			want:      "",
			corrected: "http://good",
			wantErr:   "invalid URL escape",
		},
		"corrected v03": {
			event:     ce.New("0.3"),
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
			err = tc.event.Validate()
			validateReaderWriter(t, tc, got, err)

			// Correct

			tc.event.SetSource(set[1])
			got = tc.event.Source()
			err = tc.event.Validate()
			validateReaderWriterCorrected(t, tc, got, err)
		})
	}
}

func TestEventRW_Subject(t *testing.T) {
	testCases := map[string]ReadWriteTest{
		"v01": {
			event: ce.New("0.1"),
			set:   "subject.0.1",
			want:  "subject.0.1",
		},
		"v02": {
			event: ce.New("0.2"),
			set:   "subject.0.2",
			want:  "subject.0.2",
		},
		"v03": {
			event: ce.New("0.3"),
			set:   "subject.0.3",
			want:  "subject.0.3",
		},
		"spaced v01": {
			event: ce.New("0.1"),
			set:   "  subject.0.1  ",
			want:  "subject.0.1",
		},
		"spaced v02": {
			event: ce.New("0.2"),
			set:   "  subject.0.2  ",
			want:  "subject.0.2",
		},
		"spaced v03": {
			event: ce.New("0.3"),
			set:   "   subject.0.3   ",
			want:  "subject.0.3",
		},
		"nilled v01": {
			event: func() ce.Event {
				e := ce.New("0.1")
				e.SetSource("should nil")
				return e
			}(),
			want: "",
		},
		"nilled v02": {
			event: func() ce.Event {
				e := ce.New("0.2")
				e.SetSource("should nil")
				return e
			}(),
			want: "",
		},
		"nilled v03": {
			event: func() ce.Event {
				e := ce.New("0.3")
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
		"v01": {
			event: ce.New("0.1"),
			set:   "now", // hack
			want:  now,
		},
		"v02": {
			event: ce.New("0.2"),
			set:   "now", // hack
			want:  now,
		},
		"v03": {
			event: ce.New("0.3"),
			set:   "now", // hack
			want:  now,
		},
		"nilled v01": {
			event: func() ce.Event {
				e := ce.New("0.1")
				e.SetTime(now)
				return e
			}(),
			want: time.Time{},
		},
		"nilled v02": {
			event: func() ce.Event {
				e := ce.New("0.2")
				e.SetTime(now)
				return e
			}(),
			want: time.Time{},
		},
		"nilled v03": {
			event: func() ce.Event {
				e := ce.New("0.3")
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
		"v01": {
			event: ce.New("0.1"),
			set:   "http://example/",
			want:  "http://example/",
		},
		"v02": {
			event: ce.New("0.2"),
			set:   "http://example/",
			want:  "http://example/",
		},
		"v03": {
			event: ce.New("0.3"),
			set:   "http://example/",
			want:  "http://example/",
		},
		"invalid v01": {
			event:   ce.New("0.1"),
			set:     "%",
			want:    "",
			wantErr: "invalid URL escape",
		},
		"invalid v02": {
			event:   ce.New("0.2"),
			set:     "%",
			want:    "",
			wantErr: "invalid URL escape",
		},
		"invalid v03": {
			event:   ce.New("0.3"),
			set:     "%",
			want:    "",
			wantErr: "invalid URL escape",
		},
		"nilled v01": {
			event: func() ce.Event {
				e := ce.New("0.1")
				e.SetDataSchema("should nil")
				return e
			}(),
			want: "",
		},
		"nilled v02": {
			event: func() ce.Event {
				e := ce.New("0.2")
				e.SetDataSchema("should nil")
				return e
			}(),
			want: "",
		},
		"nilled v03": {
			event: func() ce.Event {
				e := ce.New("0.3")
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

			err := tc.event.Validate()
			validateReaderWriter(t, tc, got, err)
		})
	}
}

func TestEventRW_DataContentType(t *testing.T) {
	testCases := map[string]ReadWriteTest{
		"v01": {
			event: ce.New("0.1"),
			set:   "application/json",
			want:  "application/json",
		},
		"v02": {
			event: ce.New("0.2"),
			set:   "application/json",
			want:  "application/json",
		},
		"v03": {
			event: ce.New("0.3"),
			set:   "application/json",
			want:  "application/json",
		},
		"spaced v01": {
			event: ce.New("0.1"),
			set:   "  application/json  ",
			want:  "application/json",
		},
		"spaced v02": {
			event: ce.New("0.2"),
			set:   "  application/json  ",
			want:  "application/json",
		},
		"spaced v03": {
			event: ce.New("0.3"),
			set:   "   application/json   ",
			want:  "application/json",
		},
		"nilled v01": {
			event: func() ce.Event {
				e := ce.New("0.1")
				e.SetDataContentType("application/json")
				return e
			}(),
			want: "",
		},
		"nilled v02": {
			event: func() ce.Event {
				e := ce.New("0.2")
				e.SetDataContentType("application/json")
				return e
			}(),
			want: "",
		},
		"nilled v03": {
			event: func() ce.Event {
				e := ce.New("0.3")
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
		"v01": {
			event: ce.New("0.1"),
			set:   "base64",
			want:  "base64",
		},
		"v02": {
			event: ce.New("0.2"),
			set:   "base64",
			want:  "base64",
		},
		"v03": {
			event: ce.New("0.3"),
			set:   "base64",
			want:  "base64",
		},
		"spaced v01": {
			event: ce.New("0.1"),
			set:   "  base64  ",
			want:  "base64",
		},
		"spaced v02": {
			event: ce.New("0.2"),
			set:   "  base64  ",
			want:  "base64",
		},
		"spaced v03": {
			event: ce.New("0.3"),
			set:   "   base64   ",
			want:  "base64",
		},
		"cased v01": {
			event: ce.New("0.1"),
			set:   "  BaSe64  ",
			want:  "base64",
		},
		"cased v02": {
			event: ce.New("0.2"),
			set:   "  BaSe64  ",
			want:  "base64",
		},
		"cased v03": {
			event: ce.New("0.3"),
			set:   "   BaSe64   ",
			want:  "base64",
		},
		"nilled v01": {
			event: func() ce.Event {
				e := ce.New("0.1")
				e.SetDataContentEncoding("base64")
				return e
			}(),
			want: "",
		},
		"nilled v02": {
			event: func() ce.Event {
				e := ce.New("0.2")
				e.SetDataContentEncoding("base64")
				return e
			}(),
			want: "",
		},
		"nilled v03": {
			event: func() ce.Event {
				e := ce.New("0.3")
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
