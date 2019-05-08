package cloudevents_test

import (
	ce "github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/types"
	"github.com/google/go-cmp/cmp"
	"strings"
	"testing"
)

type DataTest struct {
	event   func(string) ce.Event
	set     interface{}
	want    interface{}
	wantErr string
}

func TestEventSetData_Json(t *testing.T) {
	// All version should be the same, so run through them all.

	versions := []string{ce.CloudEventsVersionV01, ce.CloudEventsVersionV02, ce.CloudEventsVersionV03}

	testCases := map[string]DataTest{
		"empty": {
			event: func(version string) ce.Event {
				return ce.New(version)
			},
			want: nil,
		},
		"defaults": {
			event: func(version string) ce.Event {
				return ce.New(version)
			},
			set: map[string]interface{}{
				"hello": "unittest",
			},
			want: []byte(`{"hello":"unittest"}`),
		},
		"text/json": {
			event: func(version string) ce.Event {
				e := ce.New(version)
				e.SetDataContentType("text/json")
				return e
			},
			set: map[string]interface{}{
				"hello": "unittest",
			},
			want: []byte(`{"hello":"unittest"}`),
		},
		"application/json": {
			event: func(version string) ce.Event {
				e := ce.New(version)
				e.SetDataContentType("application/json")
				return e
			},
			set: map[string]interface{}{
				"hello": "unittest",
			},
			want: []byte(`{"hello":"unittest"}`),
		},
		"application/json+base64": {
			event: func(version string) ce.Event {
				e := ce.New(version)
				e.SetDataContentType("application/json")
				e.SetDataContentEncoding(ce.Base64)
				return e
			},
			set: map[string]interface{}{
				"hello": "unittest",
			},
			want: `eyJoZWxsbyI6InVuaXR0ZXN0In0=`,
		},
	}
	for n, tc := range testCases {
		for _, version := range versions {
			t.Run(n+":"+version, func(t *testing.T) {
				// Make a versioned event.
				event := tc.event(version)

				if tc.set != nil {
					if err := event.SetData(tc.set); err != nil {
						t.Errorf("unexpected error, %v", err)
					}
				}
				got := event.Data

				as, _ := types.Allocate(tc.set)

				err := event.DataAs(&as)
				validateData(t, tc, got, as, err)
			})
		}
	}
}

type XmlExample struct {
	AnInt   int      `xml:"a,omitempty"`
	AString string   `xml:"b,omitempty"`
	AnArray []string `xml:"c,omitempty"`
}

func TestEventSetData_xml(t *testing.T) {
	// All version should be the same, so run through them all.

	versions := []string{ce.CloudEventsVersionV01, ce.CloudEventsVersionV02, ce.CloudEventsVersionV03}

	testCases := map[string]DataTest{
		"empty": {
			event: func(version string) ce.Event {
				e := ce.New(version)
				e.SetDataContentType("application/xml")
				return e
			},
			want: nil,
		},
		"text/xml": {
			event: func(version string) ce.Event {
				e := ce.New(version)
				e.SetDataContentType("text/xml")
				return e
			},
			set: &XmlExample{
				AnInt:   42,
				AString: "true fact",
				AnArray: versions,
			},
			want: []byte(`<XmlExample><a>42</a><b>true fact</b><c>0.1</c><c>0.2</c><c>0.3</c></XmlExample>`),
		},
		"applicaiton/xml": {
			event: func(version string) ce.Event {
				e := ce.New(version)
				e.SetDataContentType("application/xml")
				return e
			},
			set: &XmlExample{
				AnInt:   42,
				AString: "true fact",
				AnArray: versions,
			},
			want: []byte(`<XmlExample><a>42</a><b>true fact</b><c>0.1</c><c>0.2</c><c>0.3</c></XmlExample>`),
		},
		"applicaiton/xml+base64": {
			event: func(version string) ce.Event {
				e := ce.New(version)
				e.SetDataContentType("application/xml")
				e.SetDataContentEncoding(ce.Base64)
				return e
			},
			set: &XmlExample{
				AnInt:   42,
				AString: "true fact",
				AnArray: versions,
			},
			want: `PFhtbEV4YW1wbGU+PGE+NDI8L2E+PGI+dHJ1ZSBmYWN0PC9iPjxjPjAuMTwvYz48Yz4wLjI8L2M+PGM+MC4zPC9jPjwvWG1sRXhhbXBsZT4=`,
		},
	}
	for n, tc := range testCases {
		for _, version := range versions {
			t.Run(n+":"+version, func(t *testing.T) {
				// Make a versioned event.
				event := tc.event(version)

				if tc.set != nil {
					if err := event.SetData(tc.set); err != nil {
						t.Errorf("unexpected error, %v", err)
					}
				}
				got := event.Data

				as, _ := types.Allocate(tc.set)

				err := event.DataAs(&as)
				validateData(t, tc, got, as, err)
			})
		}
	}
}

func validateData(t *testing.T, tc DataTest, got, as interface{}, err error) {
	var gotErr string
	if err != nil {
		gotErr = err.Error()
		if tc.wantErr == "" {
			t.Errorf("unexpected no error, got %q", gotErr)
		}
	}
	if tc.wantErr != "" {
		if !strings.Contains(gotErr, tc.wantErr) {
			t.Errorf("unexpected error, expected to contain %q, got: %q ", tc.wantErr, gotErr)
		}
	}
	if diff := cmp.Diff(tc.want, got); diff != "" {
		t.Errorf("unexpected data (-want, +got) = %v", diff)
	}
	if diff := cmp.Diff(tc.set, as); diff != "" {
		t.Errorf("unexpected as (-want, +got) = %v", diff)
	}
}
