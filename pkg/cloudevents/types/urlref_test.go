package types_test

import (
	"encoding/xml"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/types"
	"github.com/google/go-cmp/cmp"
	"net/url"
	"testing"
)

func TestParseURLRef(t *testing.T) {
	testCases := map[string]struct {
		t    string
		want *types.URLRef
	}{
		"empty": {
			want: nil,
		},
		"empty string": {
			t:    "",
			want: nil,
		},
		"invalid format": {
			t:    "💩://error",
			want: nil,
		},
		"relative": {
			t: "/path/to/something",
			want: func() *types.URLRef {
				u, _ := url.Parse("/path/to/something")
				return &types.URLRef{URL: *u}
			}(),
		},
		"url": {
			t: "http://path/to/something",
			want: func() *types.URLRef {
				u, _ := url.Parse("http://path/to/something")
				return &types.URLRef{URL: *u}
			}(),
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			got := types.ParseURLRef(tc.t)

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("unexpected object (-want, +got) = %v", diff)
			}
		})
	}
}

func TestJsonMarshalURLRef(t *testing.T) {
	testCases := map[string]struct {
		t    string
		want []byte
	}{
		"empty": {},
		"empty string": {
			t: "",
		},
		"invalid url": {
			t:    "not a url",
			want: []byte(`"not%20a%20url"`),
		},
		"relative format": {
			t:    "/path/to/something",
			want: []byte(`"/path/to/something"`),
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			var got []byte
			tt := types.ParseURLRef(tc.t)
			if tt != nil {
				got, _ = tt.MarshalJSON()
			}

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Logf("got: %s", string(got))
				t.Errorf("unexpected object (-want, +got) = %v", diff)
			}
		})
	}
}

func TestXMLMarshalURLRef(t *testing.T) {
	testCases := map[string]struct {
		t    string
		want []byte
	}{
		"empty": {},
		"empty string": {
			t: "",
		},
		"invalid url": {
			t:    "not a url",
			want: []byte(`<URLRef>not%20a%20url</URLRef>`),
		},
		"relative format": {
			t:    "/path/to/something",
			want: []byte(`<URLRef>/path/to/something</URLRef>`),
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			var got []byte
			tt := types.ParseURLRef(tc.t)
			if tt != nil {
				got, _ = xml.Marshal(tt)
			}

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Logf("got: %s", string(got))
				t.Errorf("unexpected object (-want, +got) = %v", diff)
			}
		})
	}
}

func TestJsonUnmarshalURLRef(t *testing.T) {
	testCases := map[string]struct {
		b       []byte
		want    *types.URLRef
		wantErr string
	}{
		"empty": {
			wantErr: "unexpected end of JSON input",
		},
		"invalid format": {
			b:       []byte("%"),
			wantErr: "invalid character '%' looking for beginning of value",
		},
		"relative": {
			b: []byte(`"/path/to/something"`),
			want: func() *types.URLRef {
				u, _ := url.Parse("/path/to/something")
				return &types.URLRef{URL: *u}
			}(),
		},
		"url": {
			b: []byte(`"http://path/to/something"`),
			want: func() *types.URLRef {
				u, _ := url.Parse("http://path/to/something")
				return &types.URLRef{URL: *u}
			}(),
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			got := &types.URLRef{}
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

func TestXMLUnmarshalURLRef(t *testing.T) {
	testCases := map[string]struct {
		b       []byte
		want    *types.URLRef
		wantErr string
	}{
		"empty": {
			wantErr: "EOF",
		},
		"invalid format": {
			b:    []byte(`<URLRef>%</URLRef>`),
			want: &types.URLRef{},
		},
		"bad xml": {
			b:       []byte(`<URLRef><bad>%<bad></URLRef>`),
			wantErr: "XML syntax error on line 1: element <bad> closed by </URLRef>",
		},
		"relative": {
			b: []byte(`<URLRef>/path/to/something</URLRef>`),
			want: func() *types.URLRef {
				u, _ := url.Parse("/path/to/something")
				return &types.URLRef{URL: *u}
			}(),
		},
		"url": {
			b: []byte(`<URLRef>http://path/to/something</URLRef>`),
			want: func() *types.URLRef {
				u, _ := url.Parse("http://path/to/something")
				return &types.URLRef{URL: *u}
			}(),
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			got := &types.URLRef{}

			err := xml.Unmarshal(tc.b, got)

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

func TestURLRefString(t *testing.T) {
	testCases := map[string]struct {
		t    string
		want string
	}{
		"empty": {
			want: "",
		},
		"relative": {
			t:    "/path/to/something",
			want: "/path/to/something",
		},
		"url": {
			t:    "http://path/to/something",
			want: "http://path/to/something",
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			tt := types.ParseURLRef(tc.t)
			got := tt.String()

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Logf("got: %s", string(got))
				t.Errorf("unexpected string (-want, +got) = %v", diff)
			}
		})
	}
}

// TODO: Test xml:  MarshalXML(e *xml.Encoder, start xml.StartElement) error {
