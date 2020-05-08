package event_test

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/cloudevents/sdk-go/v2/event"

	"github.com/google/go-cmp/cmp"

	"github.com/cloudevents/sdk-go/v2/types"
)

func TestGetDataContentType(t *testing.T) {
	now := types.Timestamp{Time: time.Now()}

	testCases := map[string]struct {
		event event.Event
		want  string
	}{
		"min v03, blank": {
			event: event.Event{
				Context: MinEventContextV03(),
			},
			want: "",
		},
		"full v03, json": {
			event: event.Event{
				Context: FullEventContextV03(now),
			},
			want: "application/json",
		},
		"min v1, blank": {
			event: event.Event{
				Context: MinEventContextV1(),
			},
			want: "",
		},
		"full v1, json": {
			event: event.Event{
				Context: FullEventContextV1(now),
			},
			want: "application/json",
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			got, _ := tc.event.Context.GetDataMediaType()

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("unexpected (-want, +got) = %v", diff)
			}
		})
	}
}

func TestSource(t *testing.T) {
	now := types.Timestamp{Time: time.Now()}

	source := "http://example.com/source"

	testCases := map[string]struct {
		event event.Event
		want  string
	}{
		"min v03": {
			event: event.Event{
				Context: MinEventContextV03(),
			},
			want: source,
		},
		"full v03": {
			event: event.Event{
				Context: FullEventContextV03(now),
			},
			want: source,
		},
		"min v1": {
			event: event.Event{
				Context: MinEventContextV1(),
			},
			want: source,
		},
		"full v1": {
			event: event.Event{
				Context: FullEventContextV1(now),
			},
			want: source,
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			got := tc.event.Source()

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("unexpected (-want, +got) = %v", diff)
			}
		})
	}
}

func TestSchemaURL(t *testing.T) {
	now := types.Timestamp{Time: time.Now()}

	schema := "http://example.com/schema"

	testCases := map[string]struct {
		event event.Event
		want  string
	}{
		"min v03, empty schema": {
			event: event.Event{
				Context: MinEventContextV03(),
			},
			want: "",
		},
		"full v03, schema": {
			event: event.Event{
				Context: FullEventContextV03(now),
			},
			want: schema,
		},
		"min v1, empty schema": {
			event: event.Event{
				Context: MinEventContextV1(),
			},
			want: "",
		},
		"full v1, schema": {
			event: event.Event{
				Context: FullEventContextV1(now),
			},
			want: schema,
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			got := tc.event.DataSchema()

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("unexpected (-want, +got) = %v", diff)
			}
		})
	}
}

type DataExample struct {
	AnInt   int                       `json:"a,omitempty"`
	AString string                    `json:"b,omitempty"`
	AnArray []string                  `json:"c,omitempty"`
	AMap    map[string]map[string]int `json:"d,omitempty"`
	ATime   *time.Time                `json:"e,omitempty"`
}

func TestValidate(t *testing.T) {
	now := types.Timestamp{Time: time.Now()}

	testCases := map[string]struct {
		event event.Event
		want  []string
	}{
		"min valid v0.3": {
			event: event.Event{
				Context: MinEventContextV03(),
			},
		},
		"min valid v1.0": {
			event: event.Event{
				Context: MinEventContextV1(),
			},
		},
		"json valid, v0.3": {
			event: event.Event{
				Context:     FullEventContextV03(now),
				DataEncoded: []byte(`{"a":"apple","b":"banana"}`),
			},
		},
		"json valid, v1.0": {
			event: event.Event{
				Context:     FullEventContextV1(now),
				DataEncoded: []byte(`{"a":"apple","b":"banana"}`),
			},
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			got := tc.event.Validate()
			var gotErr string
			if got != nil {
				gotErr = got.Error()

				if len(tc.want) == 0 {
					t.Errorf("unexpected no error, got %q", gotErr)
				}
			}

			for _, want := range tc.want {
				if !strings.Contains(gotErr, want) {
					t.Errorf("unexpected error, expected to contain %q, got: %q ", want, gotErr)
				}
			}
		})
	}
}

func TestString(t *testing.T) {
	now := types.Timestamp{Time: time.Now()}

	testCases := map[string]struct {
		event event.Event
		want  string
	}{
		"min v0.3": {
			event: event.Event{
				Context: MinEventContextV03(),
			},
			want: `Validation: valid
Context Attributes,
  specversion: 0.3
  type: com.example.simple
  source: http://example.com/source
  id: ABC-123
`,
		},
		"min v1.0": {
			event: event.Event{
				Context: MinEventContextV1(),
			},
			want: `Validation: valid
Context Attributes,
  specversion: 1.0
  type: com.example.simple
  source: http://example.com/source
  id: ABC-123
`,
		},
		"json simple, v0.3": {
			event: event.Event{
				Context:     FullEventContextV03(now),
				DataEncoded: []byte(`{"a":"apple","b":"banana"}`),
			},
			want: fmt.Sprintf(`Validation: valid
Context Attributes,
  specversion: 0.3
  type: com.example.simple
  source: http://example.com/source
  subject: topic
  id: ABC-123
  time: %s
  schemaurl: http://example.com/schema
  datacontenttype: application/json
  datacontentencoding: base64
Extensions,
  anothertest: 1
  test: extended
Data,
  {
    "a": "apple",
    "b": "banana"
  }
`, now.String()),
		},
		"json simple, v1.0": {
			event: event.Event{
				Context:     FullEventContextV1(now),
				DataEncoded: []byte(`{"a":"apple","b":"banana"}`),
			},
			want: fmt.Sprintf(`Validation: valid
Context Attributes,
  specversion: 1.0
  type: com.example.simple
  source: http://example.com/source
  subject: topic
  id: ABC-123
  time: %s
  dataschema: http://example.com/schema
  datacontenttype: application/json
Extensions,
  anothertest: 1
  datacontentencoding: base64
  test: extended
Data,
  {
    "a": "apple",
    "b": "banana"
  }
`, now.String()),
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			got := tc.event.String()

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Log(got)
				t.Errorf("unexpected string (-want, +got) = %v", diff)
			}
		})
	}
}

// ExtensionAs is deprecated, replacement is TestExtensions below
func TestExtensionAs(t *testing.T) {
	now := types.Timestamp{Time: time.Now()}

	testCases := map[string]struct {
		event        event.Event
		extension    string
		want         string
		wantError    bool
		wantErrorMsg string
	}{
		"min v03, no extension": {
			event: event.Event{
				Context: MinEventContextV03(),
			},
			extension:    "test",
			wantError:    true,
			wantErrorMsg: `extension "test" does not exist`,
		},
		"full v03, test extension": {
			event: event.Event{
				Context: FullEventContextV03(now),
			},
			extension: "test",
			want:      "extended",
		},
		"full v03, anothertest extension invalid type": {
			event: event.Event{
				Context: FullEventContextV03(now),
			},
			extension:    "anothertest",
			wantError:    true,
			wantErrorMsg: `invalid type for extension "anothertest"`,
		},
		"full v1, test extension": {
			event: event.Event{
				Context: FullEventContextV1(now),
			},
			extension: "test",
			want:      "extended",
		},
		"full v1, anothertest extension invalid type": {
			event: event.Event{
				Context: FullEventContextV1(now),
			},
			extension:    "anothertest",
			wantError:    true,
			wantErrorMsg: `unknown extension type *string`,
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			var got string
			err := tc.event.Context.ExtensionAs(tc.extension, &got)

			if tc.wantError {
				if err == nil {
					t.Errorf("expected error %q, got nil", tc.wantErrorMsg)
				} else {
					if diff := cmp.Diff(tc.wantErrorMsg, err.Error()); diff != "" {
						t.Errorf("unexpected (-want, +got) = %v", diff)
					}
				}
			} else {
				if diff := cmp.Diff(tc.want, got); diff != "" {
					t.Errorf("unexpected (-want, +got) = %v", diff)
				}
			}
		})
	}
}

func TestExtensions(t *testing.T) {
	now := types.Timestamp{Time: time.Now()}

	testCases := map[string]struct {
		event        event.Event
		extension    string
		want         string
		wantError    bool
		wantErrorMsg string
	}{
		"full v03, test extension": {
			event: event.Event{
				Context: FullEventContextV03(now),
			},
			extension: "test",
			want:      "extended",
		},
		"full v03, anothertest extension invalid type": {
			event: event.Event{
				Context: FullEventContextV03(now),
			},
			extension:    "anothertest",
			wantError:    true,
			wantErrorMsg: "cannot convert 1 to string",
		},
		"full v1, test extension": {
			event: event.Event{
				Context: FullEventContextV1(now),
			},
			extension: "test",
			want:      "extended",
		},
		"full v1, anothertest extension invalid type": {
			event: event.Event{
				Context: FullEventContextV1(now),
			},
			extension:    "anothertest",
			wantError:    true,
			wantErrorMsg: "cannot convert 1 to string",
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			var got string
			got, err := types.ToString(tc.event.Context.GetExtensions()[tc.extension])
			if tc.wantError {
				if err == nil {
					t.Errorf("expected error %q, got nil", tc.wantErrorMsg)
				} else {
					if diff := cmp.Diff(tc.wantErrorMsg, err.Error()); diff != "" {
						t.Errorf("unexpected (-want, +got) = %v", diff)
					}
				}
			} else {
				if diff := cmp.Diff(tc.want, got); diff != "" {
					t.Errorf("unexpected (-want, +got) = %v", diff)
				}
			}
		})
	}
}

func strptr(s string) *string {
	return &s
}

func MinEventContextV03() *event.EventContextV03 {
	sourceUrl, _ := url.Parse("http://example.com/source")
	source := &types.URIRef{URL: *sourceUrl}

	return event.EventContextV03{
		Type:   "com.example.simple",
		Source: *source,
		ID:     "ABC-123",
	}.AsV03()
}

func MinEventContextV1() *event.EventContextV1 {
	sourceUrl, _ := url.Parse("http://example.com/source")
	source := &types.URIRef{URL: *sourceUrl}

	return event.EventContextV1{
		Type:   "com.example.simple",
		Source: *source,
		ID:     "ABC-123",
	}.AsV1()
}

func FullEventContextV03(now types.Timestamp) *event.EventContextV03 {
	sourceUrl, _ := url.Parse("http://example.com/source")
	source := &types.URIRef{URL: *sourceUrl}

	schemaUrl, _ := url.Parse("http://example.com/schema")
	schema := &types.URIRef{URL: *schemaUrl}

	eventContextV03 := event.EventContextV03{
		ID:                  "ABC-123",
		Time:                &now,
		Type:                "com.example.simple",
		SchemaURL:           schema,
		DataContentType:     event.StringOfApplicationJSON(),
		DataContentEncoding: event.StringOfBase64(),
		Source:              *source,
		Subject:             strptr("topic"),
	}
	_ = eventContextV03.SetExtension("test", "extended")
	_ = eventContextV03.SetExtension("anothertest", int32(1))
	return eventContextV03.AsV03()
}

func FullEventContextV1(now types.Timestamp) *event.EventContextV1 {
	sourceUrl, _ := url.Parse("http://example.com/source")
	source := &types.URIRef{URL: *sourceUrl}

	schemaUrl, _ := url.Parse("http://example.com/schema")
	schema := &types.URI{URL: *schemaUrl}

	eventContextV1 := event.EventContextV1{
		ID:              "ABC-123",
		Time:            &now,
		Type:            "com.example.simple",
		DataSchema:      schema,
		DataContentType: event.StringOfApplicationJSON(),
		Source:          *source,
		Subject:         strptr("topic"),
	}
	_ = eventContextV1.SetExtension("test", "extended")
	_ = eventContextV1.SetExtension("anothertest", 1)
	_ = eventContextV1.SetExtension(event.DataContentEncodingKey, event.Base64)
	return eventContextV1.AsV1()
}

func TestEvent_Clone(t *testing.T) {
	original := event.Event{
		Context: FullEventContextV1(types.Timestamp{Time: time.Now()}),
	}
	original.FieldErrors = map[string]error{
		"id": errors.New("an error"),
	}
	require.NoError(t, original.SetData(event.ApplicationJSON, "aaa"))

	clone := original.Clone()

	require.Equal(t, original.Context, clone.Context)
	require.NotSame(t, original.Context, clone.Context)
	require.Equal(t, original.Data(), clone.Data())
	require.NotSame(t, original.Data(), clone.Data())

	require.Equal(t, original.DataEncoded, clone.DataEncoded)
	require.Equal(t, original.DataBase64, clone.DataBase64)

	require.Equal(t, original.FieldErrors, clone.FieldErrors)
	require.NotSame(t, original.FieldErrors, clone.FieldErrors)

	require.NoError(t, clone.SetData(event.ApplicationJSON, "bbb"))

	require.Equal(t, []byte("\"aaa\""), original.Data())
	require.Equal(t, []byte("\"bbb\""), clone.Data())
}
