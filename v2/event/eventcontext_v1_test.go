package event_test

import (
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/cloudevents/sdk-go/v2/types"
)

func TestValidateV1(t *testing.T) {
	now := types.Timestamp{Time: time.Now()}

	source := types.ParseURIRef("http://example.com/source")
	subject := "a subject"
	dataSchema := types.ParseURI("http://example.com/schema")

	extensions := make(map[string]interface{})
	extensions["test"] = "extended"

	testCases := map[string]struct {
		ctx  event.EventContextV1
		want []string
	}{
		"min valid": {
			ctx: event.EventContextV1{
				ID:     "ABC-123",
				Type:   "com.example.simple",
				Source: *source,
			},
		},
		"full valid": {
			ctx: event.EventContextV1{
				ID:              "ABC-123",
				Time:            &now,
				Type:            "com.example.simple",
				DataSchema:      dataSchema,
				DataContentType: event.StringOfApplicationJSON(),
				Source:          *source,
				Subject:         &subject,
				Extensions:      extensions,
			},
		},
		"no Type": {
			ctx: event.EventContextV1{
				ID:     "ABC-123",
				Source: *source,
			},
			want: []string{"type:"},
		},
		"missing source": {
			ctx: event.EventContextV1{
				ID:   "ABC-123",
				Type: "com.example.simple",
			},
			want: []string{"source:"},
		},
		"non-empty subject": {
			ctx: event.EventContextV1{
				ID:      "",
				Type:    "com.example.simple",
				Source:  *source,
				Subject: strptr("  "),
			},
			want: []string{"subject:"},
		},
		"non-empty ID": {
			ctx: event.EventContextV1{
				ID:     "",
				Type:   "com.example.simple",
				Source: *source,
			},
			want: []string{"id:"},
		},
		"empty DataSchema": {
			ctx: event.EventContextV1{
				ID:         "ABC-123",
				Type:       "com.example.simple",
				DataSchema: &types.URI{},
				Source:     *source,
			},
			want: []string{"dataschema:"},
		},
		"non absolute DataSchema": {
			ctx: event.EventContextV1{
				ID:         "ABC-123",
				Type:       "com.example.simple",
				DataSchema: types.ParseURI("/hello"),
				Source:     *source,
			},
			want: []string{"dataschema:"},
		},
		"non-empty contentType": {
			ctx: event.EventContextV1{
				ID:              "ABC-123",
				Type:            "com.example.simple",
				Source:          *source,
				DataContentType: strptr(""),
			},
			want: []string{"datacontenttype:"},
		},
		"invalid contentType": {
			ctx: event.EventContextV1{
				ID:              "ABC-123",
				Type:            "com.example.simple",
				Source:          *source,
				DataContentType: strptr("bogus ;========="),
			},
			want: []string{"datacontenttype:"},
		},

		"all errors": {
			ctx: event.EventContextV1{
				ID:              "",
				DataSchema:      &types.URI{},
				DataContentType: strptr(""),
				Extensions:      make(map[string]interface{}),
			},
			want: []string{
				"type:",
				"id:",
				"source:",
				"contenttype:",
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

func TestGetMediaTypeV1(t *testing.T) {
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

			ec := event.EventContextV1{}
			if tc.t != "" {
				ec.DataContentType = &tc.t
			}
			got, _ := ec.GetDataMediaType()

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("unexpected  (-want, +got) = %v", diff)
			}
		})
	}
}
