package event_test

import (
	"context"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"

	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/cloudevents/sdk-go/v2/event/datacodec"
	"github.com/cloudevents/sdk-go/v2/types"
)

type DataTest struct {
	event   func(string) event.Event
	set     interface{}
	want    interface{}
	wantErr string
}

func TestEventSetData_Jsonv03(t *testing.T) {
	testCases := map[string]DataTest{
		"empty": {
			event: func(version string) event.Event {
				return event.New(version)
			},
			want: []uint8(nil),
		},
		"defaults": {
			event: func(version string) event.Event {
				return event.New(version)
			},
			set: map[string]interface{}{
				"hello": "unittest",
			},
			want: []byte(`{"hello":"unittest"}`),
		},
		"text/json": {
			event: func(version string) event.Event {
				e := event.New(version)
				e.SetDataContentType("text/json")
				return e
			},
			set: map[string]interface{}{
				"hello": "unittest",
			},
			want: []byte(`{"hello":"unittest"}`),
		},
		"application/json": {
			event: func(version string) event.Event {
				e := event.New(version)
				e.SetDataContentType("application/json")
				return e
			},
			set: map[string]interface{}{
				"hello": "unittest",
			},
			want: []byte(`{"hello":"unittest"}`),
		},
		"application/json+base64": {
			event: func(version string) event.Event {
				e := event.New(version)
				e.SetDataContentEncoding(event.Base64)
				return e
			},
			set: map[string]interface{}{
				"hello": "unittest",
			},
			want: []byte(`eyJoZWxsbyI6InVuaXR0ZXN0In0=`),
		},
	}
	for n, tc := range testCases {
		version := event.CloudEventsVersionV03
		t.Run(n+":"+version, func(t *testing.T) {
			// Make a versioned event.
			e := tc.event(version)

			if tc.set != nil {
				if err := e.SetData(event.ApplicationJSON, tc.set); err != nil {
					t.Errorf("unexpected error, %v", err)
				}
			}
			got := e.Data()

			as, _ := types.Allocate(tc.set)

			err := e.DataAs(&as)
			validateData(t, tc, got, as, err)
		})
	}
}

func TestEventSetData_Jsonv1(t *testing.T) {
	testCases := map[string]DataTest{
		"empty": {
			event: func(version string) event.Event {
				return event.New(version)
			},
			want: []uint8(nil),
		},
		"defaults": {
			event: func(version string) event.Event {
				return event.New(version)
			},
			set: map[string]interface{}{
				"hello": "unittest",
			},
			want: []byte(`{"hello":"unittest"}`),
		},
		"text/json": {
			event: func(version string) event.Event {
				e := event.New(version)
				e.SetDataContentType("text/json")
				return e
			},
			set: map[string]interface{}{
				"hello": "unittest",
			},
			want: []byte(`{"hello":"unittest"}`),
		},
		"application/json": {
			event: func(version string) event.Event {
				e := event.New(version)
				e.SetDataContentType("application/json")
				return e
			},
			set: map[string]interface{}{
				"hello": "unittest",
			},
			want: []byte(`{"hello":"unittest"}`),
		},
	}
	for n, tc := range testCases {
		version := event.CloudEventsVersionV1
		t.Run(n+":"+version, func(t *testing.T) {
			// Make a versioned event.
			e := tc.event(version)

			if tc.set != nil {
				if err := e.SetData(e.DataContentType(), tc.set); err != nil {
					t.Errorf("unexpected error, %v", err)
				}
			}
			got := e.Data()

			as, _ := types.Allocate(tc.set)

			err := e.DataAs(&as)
			validateData(t, tc, got, as, err)
		})
	}
}

func TestEventSetData_binary_v1(t *testing.T) {
	e := event.New(event.CloudEventsVersionV1)

	decodedPayload := map[string]interface{}{"hello": "world"}
	encodedPayload := mustEncodeWithDataCodec(t, event.ApplicationJSON, decodedPayload)

	require.NoError(t, e.SetData(event.ApplicationJSON, encodedPayload))

	require.True(t, e.DataBase64)
	require.Equal(t, encodedPayload, e.Data())

	actual := map[string]interface{}{}
	require.NoError(t, e.DataAs(&actual))

	require.Equal(t, decodedPayload, actual)
}

type XmlExample struct {
	AnInt   int      `xml:"a,omitempty"`
	AString string   `xml:"b,omitempty"`
	AnArray []string `xml:"c,omitempty"`
}

func TestEventSetData_xml(t *testing.T) {
	// All version should be the same, so run through them all.

	versions := []string{event.CloudEventsVersionV03, event.CloudEventsVersionV1}

	testCases := map[string]DataTest{
		"empty": {
			event: func(version string) event.Event {
				e := event.New(version)
				e.SetDataContentType("application/xml")
				return e
			},
			want: []uint8(nil),
		},
		"text/xml": {
			event: func(version string) event.Event {
				e := event.New(version)
				e.SetDataContentType("text/xml")
				return e
			},
			set: &XmlExample{
				AnInt:   42,
				AString: "true fact",
				AnArray: versions,
			},
			want: []byte(`<XmlExample><a>42</a><b>true fact</b><c>0.3</c><c>1.0</c></XmlExample>`),
		},
		"application/xml": {
			event: func(version string) event.Event {
				e := event.New(version)
				e.SetDataContentType("application/xml")
				return e
			},
			set: &XmlExample{
				AnInt:   42,
				AString: "true fact",
				AnArray: versions,
			},
			want: []byte(`<XmlExample><a>42</a><b>true fact</b><c>0.3</c><c>1.0</c></XmlExample>`),
		},
	}
	for n, tc := range testCases {
		for _, version := range versions {
			t.Run(n+":"+version, func(t *testing.T) {
				// Make a versioned event.
				e := tc.event(version)

				if tc.set != nil {
					if err := e.SetData(e.DataContentType(), tc.set); err != nil {
						t.Errorf("unexpected error, %v", err)
					}
				}
				got := e.Data()

				as, _ := types.Allocate(tc.set)

				err := e.DataAs(&as)
				validateData(t, tc, got, as, err)
			})
		}
	}
}

func TestEventSetData_xml_base64(t *testing.T) {
	// All version should be the same, so run through them all.

	versions := []string{event.CloudEventsVersionV03}

	testCases := map[string]DataTest{
		"application/xml+base64": {
			event: func(version string) event.Event {
				e := event.New(version)
				e.SetDataContentType("application/xml")
				e.SetDataContentEncoding(event.Base64)
				return e
			},
			set: &XmlExample{
				AnInt:   42,
				AString: "true fact",
				AnArray: versions,
			},
			want: []byte(`PFhtbEV4YW1wbGU+PGE+NDI8L2E+PGI+dHJ1ZSBmYWN0PC9iPjxjPjAuMzwvYz48L1htbEV4YW1wbGU+`),
		},
	}
	for n, tc := range testCases {
		for _, version := range versions {
			t.Run(n+":"+version, func(t *testing.T) {
				// Make a versioned event.
				e := tc.event(version)

				if tc.set != nil {
					if err := e.SetData(e.DataContentType(), tc.set); err != nil {
						t.Errorf("unexpected error, %v", err)
					}
				}
				got := e.Data()

				as, _ := types.Allocate(tc.set)

				err := e.DataAs(&as)
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

func mustEncodeWithDataCodec(tb testing.TB, ct string, in interface{}) []byte {
	data, err := datacodec.Encode(context.TODO(), ct, in)
	require.NoError(tb, err)
	return data
}
