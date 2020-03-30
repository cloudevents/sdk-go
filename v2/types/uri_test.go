package types_test

import (
	"encoding/xml"
	"net/url"
	"testing"

	"github.com/cloudevents/sdk-go/v2/types"
	"github.com/google/go-cmp/cmp"
)

func TestParseURL(t *testing.T) {
	testCases := map[string]struct {
		t    string
		want *types.URI
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
			want: func() *types.URI {
				u, _ := url.Parse("/path/to/something")
				return &types.URI{URL: *u}
			}(),
		},
		"url": {
			t: "http://path/to/something",
			want: func() *types.URI {
				u, _ := url.Parse("http://path/to/something")
				return &types.URI{URL: *u}
			}(),
		},
		"mailto": {
			t: "mailto:cncf-wg-serverless@lists.cncf.io",
			want: func() *types.URI {
				u, _ := url.Parse("mailto:cncf-wg-serverless@lists.cncf.io")
				return &types.URI{URL: *u}
			}(),
		},
		"urn": {
			t: "urn:uuid:6e8bc430-9c3a-11d9-9669-0800200c9a66",
			want: func() *types.URI {
				u, _ := url.Parse("urn:uuid:6e8bc430-9c3a-11d9-9669-0800200c9a66")
				return &types.URI{URL: *u}
			}(),
		},
		"id3": {
			t: "1-555-123-4567",
			want: func() *types.URI {
				u, _ := url.Parse("1-555-123-4567")
				return &types.URI{URL: *u}
			}(),
		},
	}
	for n, tc := range testCases {
		tc := tc // Don't use range variable in func literal.
		t.Run(n, func(t *testing.T) {

			got := types.ParseURI(tc.t)

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("unexpected object (-want, +got) = %v", diff)
			}
		})
	}
}

func TestJsonMarshalURL(t *testing.T) {
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
			tt := types.ParseURI(tc.t)
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

func TestXMLMarshalURI(t *testing.T) {
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
			want: []byte(`<URI>not%20a%20url</URI>`),
		},
		"relative format": {
			t:    "/path/to/something",
			want: []byte(`<URI>/path/to/something</URI>`),
		},
	}
	for n, tc := range testCases {
		tc := tc // Don't use range variable in func literal.
		t.Run(n, func(t *testing.T) {

			var got []byte
			tt := types.ParseURI(tc.t)
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

func TestJsonUnmarshalURI(t *testing.T) {
	testCases := map[string]struct {
		b       []byte
		want    *types.URI
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
			want: func() *types.URI {
				u, _ := url.Parse("/path/to/something")
				return &types.URI{URL: *u}
			}(),
		},
		"url": {
			b: []byte(`"http://path/to/something"`),
			want: func() *types.URI {
				u, _ := url.Parse("http://path/to/something")
				return &types.URI{URL: *u}
			}(),
		},
	}
	for n, tc := range testCases {
		tc := tc // Don't use range variable in func literal.
		t.Run(n, func(t *testing.T) {

			got := &types.URI{}
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

func TestXMLUnmarshalURI(t *testing.T) {
	testCases := map[string]struct {
		b       []byte
		want    *types.URI
		wantErr string
	}{
		"empty": {
			wantErr: "EOF",
		},
		"invalid format": {
			b:    []byte(`<URI>%</URI>`),
			want: &types.URI{},
		},
		"bad xml": {
			b:       []byte(`<URI><bad>%<bad></URI>`),
			wantErr: "XML syntax error on line 1: element <bad> closed by </URI>",
		},
		"relative": {
			b: []byte(`<URI>/path/to/something</URI>`),
			want: func() *types.URI {
				u, _ := url.Parse("/path/to/something")
				return &types.URI{URL: *u}
			}(),
		},
		"url": {
			b: []byte(`<URI>http://path/to/something</URI>`),
			want: func() *types.URI {
				u, _ := url.Parse("http://path/to/something")
				return &types.URI{URL: *u}
			}(),
		},
	}
	for n, tc := range testCases {
		tc := tc // Don't use range variable in func literal.
		t.Run(n, func(t *testing.T) {

			got := &types.URI{}

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

func TestURIString(t *testing.T) {
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

			tt := types.ParseURI(tc.t)
			got := tt.String()

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Logf("got: %s", got)
				t.Errorf("unexpected string (-want, +got) = %v", diff)
			}
		})
	}
}
