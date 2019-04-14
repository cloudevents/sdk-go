package cloudevents_test

import (
	ce "github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/types"
	"github.com/google/go-cmp/cmp"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestValidateV01(t *testing.T) {
	now := types.Timestamp{Time: time.Now()}

	sourceUrl, _ := url.Parse("http://example.com/source")
	source := &types.URLRef{URL: *sourceUrl}

	schemaUrl, _ := url.Parse("http://example.com/schema")
	schema := &types.URLRef{URL: *schemaUrl}

	extensions := make(map[string]interface{})
	extensions["test"] = "extended"

	testCases := map[string]struct {
		ctx  ce.EventContextV01
		want []string
	}{
		"min valid": {
			ctx: ce.EventContextV01{
				CloudEventsVersion: ce.CloudEventsVersionV01,
				EventID:            "ABC-123",
				EventType:          "com.example.simple",
				Source:             *source,
			},
		},
		"full valid": {
			ctx: ce.EventContextV01{
				CloudEventsVersion: ce.CloudEventsVersionV01,
				EventID:            "ABC-123",
				EventTime:          &now,
				EventType:          "com.example.simple",
				EventTypeVersion:   strptr("v1alpha1"),
				SchemaURL:          schema,
				ContentType:        ce.StringOfApplicationJSON(),
				Source:             *source,
				Extensions:         extensions,
			},
		},
		"no eventType": {
			ctx: ce.EventContextV01{
				CloudEventsVersion: ce.CloudEventsVersionV01,
				EventID:            "ABC-123",
				Source:             *source,
			},
			want: []string{"eventType:"},
		},
		"non-empty cloudEventsVersion": {
			ctx: ce.EventContextV01{
				CloudEventsVersion: "",
				EventID:            "ABC-123",
				EventType:          "com.example.simple",
				EventTypeVersion:   strptr("v1alpha1"),
				Source:             *source,
			},
			want: []string{"cloudEventsVersion:"},
		},
		"non-empty eventTypeVersion": {
			ctx: ce.EventContextV01{
				CloudEventsVersion: ce.CloudEventsVersionV01,
				EventID:            "ABC-123",
				EventType:          "com.example.simple",
				EventTypeVersion:   strptr(""),
				Source:             *source,
			},
			want: []string{"eventTypeVersion:"},
		},
		"missing source": {
			ctx: ce.EventContextV01{
				CloudEventsVersion: ce.CloudEventsVersionV01,
				EventID:            "ABC-123",
				EventType:          "com.example.simple",
			},
			want: []string{"source:"},
		},
		"non-empty eventID": {
			ctx: ce.EventContextV01{
				CloudEventsVersion: ce.CloudEventsVersionV01,
				EventID:            "",
				EventType:          "com.example.simple",
				Source:             *source,
			},
			want: []string{"eventID:"},
		},
		"empty schemaURL": {
			ctx: ce.EventContextV01{
				CloudEventsVersion: ce.CloudEventsVersionV01,
				EventID:            "ABC-123",
				EventType:          "com.example.simple",
				SchemaURL:          &types.URLRef{},
				Source:             *source,
			},
			want: []string{"schemaURL:"},
		},
		"non-empty contentType": {
			ctx: ce.EventContextV01{
				CloudEventsVersion: ce.CloudEventsVersionV01,
				EventID:            "ABC-123",
				EventType:          "com.example.simple",
				Source:             *source,
				ContentType:        strptr(""),
			},
			want: []string{"contentType:"},
		},
		"empty extensions": {
			ctx: ce.EventContextV01{
				CloudEventsVersion: ce.CloudEventsVersionV01,
				EventID:            "ABC-123",
				EventType:          "com.example.simple",
				Source:             *source,
				Extensions:         make(map[string]interface{}),
			},
			want: []string{"extensions:"},
		},
		"all errors": {
			ctx: ce.EventContextV01{
				CloudEventsVersion: "",
				EventID:            "",
				SchemaURL:          &types.URLRef{},
				ContentType:        strptr(""),
				Extensions:         make(map[string]interface{}),
			},
			want: []string{
				"eventType:",
				"eventID:",
				"extensions:",
				"cloudEventsVersion:",
				"source:",
				"contentType:",
			},
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			got := tc.ctx.Validate()
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

func TestGetMediaTypeV01(t *testing.T) {
	testCases := map[string]struct {
		t    string
		want string
	}{
		"nil": {
			want: "",
		},
		"just encoding": {
			t:    "charset=utf-8",
			want: "",
		},
		"text/html with encoding": {
			t:    "text/html; charset=utf-8",
			want: "text/html",
		},
		"application/json with encoding": {
			t:    "application/json; charset=utf-8",
			want: "application/json",
		},
		"application/json": {
			t:    "application/json",
			want: "application/json",
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			ec := ce.EventContextV01{}
			if tc.t != "" {
				ec.ContentType = &tc.t
			}
			got, _ := ec.GetDataMediaType()

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("unexpected  (-want, +got) = %v", diff)
			}
		})
	}
}
