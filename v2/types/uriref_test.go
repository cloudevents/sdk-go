package types_test

import (
	"encoding/xml"
	"net/url"
	"testing"

	"github.com/cloudevents/sdk-go/v2/types"
	"github.com/google/go-cmp/cmp"
)

func TestParseURIRef(t *testing.T) {
	testCases := map[string]struct {
		t    string
		want *types.URIRef
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
			want: func() *types.URIRef {
				u, _ := url.Parse("/path/to/something")
				return &types.URIRef{URL: *u}
			}(),
		},
		"url": {
			t: "http://path/to/something",
			want: func() *types.URIRef {
				u, _ := url.Parse("http://path/to/something")
				return &types.URIRef{URL: *u}
			}(),
		},
	}
	for n, tc := range testCases {
		tc := tc // Don't use range variable in func literal.
		t.Run(n, func(t *testing.T) {

			got := types.ParseURIRef(tc.t)

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("unexpected object (-want, +got) = %v", diff)
			}
		})
	}
}

func TestJsonMarshalURIRef(t *testing.T) {
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
		tc := tc // Don't use range variable in func literal.
		t.Run(n, func(t *testing.T) {

			var got []byte
			tt := types.ParseURIRef(tc.t)
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

func TestXMLMarshalURIRef(t *testing.T) {
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
			want: []byte(`<URIRef>not%20a%20url</URIRef>`),
		},
		"relative format": {
			t:    "/path/to/something",
			want: []byte(`<URIRef>/path/to/something</URIRef>`),
		},
	}
	for n, tc := range testCases {
		tc := tc // Don't use range variable in func literal.
		t.Run(n, func(t *testing.T) {

			var got []byte
			tt := types.ParseURIRef(tc.t)
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

func TestJsonUnmarshalURIRef(t *testing.T) {
	testCases := map[string]struct {
		b       []byte
		want    *types.URIRef
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
			want: func() *types.URIRef {
				u, _ := url.Parse("/path/to/something")
				return &types.URIRef{URL: *u}
			}(),
		},
		"url": {
			b: []byte(`"http://path/to/something"`),
			want: func() *types.URIRef {
				u, _ := url.Parse("http://path/to/something")
				return &types.URIRef{URL: *u}
			}(),
		},
	}
	for n, tc := range testCases {
		tc := tc // Don't use range variable in func literal.
		t.Run(n, func(t *testing.T) {

			got := &types.URIRef{}
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

func TestXMLUnmarshalURIRef(t *testing.T) {
	testCases := map[string]struct {
		b       []byte
		want    *types.URIRef
		wantErr string
	}{
		"empty": {
			wantErr: "EOF",
		},
		"invalid format": {
			b:    []byte(`<URIRef>%</URIRef>`),
			want: &types.URIRef{},
		},
		"bad xml": {
			b:       []byte(`<URIRef><bad>%<bad></URIRef>`),
			wantErr: "XML syntax error on line 1: element <bad> closed by </URIRef>",
		},
		"relative": {
			b: []byte(`<URIRef>/path/to/something</URIRef>`),
			want: func() *types.URIRef {
				u, _ := url.Parse("/path/to/something")
				return &types.URIRef{URL: *u}
			}(),
		},
		"url": {
			b: []byte(`<URIRef>http://path/to/something</URIRef>`),
			want: func() *types.URIRef {
				u, _ := url.Parse("http://path/to/something")
				return &types.URIRef{URL: *u}
			}(),
		},
	}
	for n, tc := range testCases {
		tc := tc // Don't use range variable in func literal.
		t.Run(n, func(t *testing.T) {

			got := &types.URIRef{}

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

func TestURIRefString(t *testing.T) {
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
		tc := tc // Don't use range variable in func literal.
		t.Run(n, func(t *testing.T) {

			tt := types.ParseURIRef(tc.t)
			got := tt.String()

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Logf("got: %s", got)
				t.Errorf("unexpected string (-want, +got) = %v", diff)
			}
		})
	}
}
