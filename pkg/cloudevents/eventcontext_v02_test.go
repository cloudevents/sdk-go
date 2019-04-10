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

func TestValidateV02(t *testing.T) {
	now := types.Timestamp{Time: time.Now()}

	sourceUrl, _ := url.Parse("http://example.com/source")
	source := &types.URLRef{URL: *sourceUrl}

	schemaUrl, _ := url.Parse("http://example.com/schema")
	schema := &types.URLRef{URL: *schemaUrl}

	extensions := make(map[string]interface{})
	extensions["test"] = "extended"

	testCases := map[string]struct {
		ctx  ce.EventContextV02
		want []string
	}{
		"min valid": {
			ctx: ce.EventContextV02{
				SpecVersion: ce.CloudEventsVersionV02,
				ID:          "ABC-123",
				Type:        "com.example.simple",
				Source:      *source,
			},
		},
		"full valid": {
			ctx: ce.EventContextV02{
				SpecVersion: ce.CloudEventsVersionV02,
				ID:          "ABC-123",
				Time:        &now,
				Type:        "com.example.simple",
				SchemaURL:   schema,
				ContentType: ce.StringOfApplicationJSON(),
				Source:      *source,
				Extensions:  extensions,
			},
		},
		"no Type": {
			ctx: ce.EventContextV02{
				SpecVersion: ce.CloudEventsVersionV02,
				ID:          "ABC-123",
				Source:      *source,
			},
			want: []string{"type:"},
		},
		"non-empty SpecVersion": {
			ctx: ce.EventContextV02{
				SpecVersion: "",
				ID:          "ABC-123",
				Type:        "com.example.simple",
				Source:      *source,
			},
			want: []string{"specversion:"},
		},
		"missing source": {
			ctx: ce.EventContextV02{
				SpecVersion: ce.CloudEventsVersionV02,
				ID:          "ABC-123",
				Type:        "com.example.simple",
			},
			want: []string{"source:"},
		},
		"non-empty ID": {
			ctx: ce.EventContextV02{
				SpecVersion: ce.CloudEventsVersionV02,
				ID:          "",
				Type:        "com.example.simple",
				Source:      *source,
			},
			want: []string{"id:"},
		},
		"empty schemaURL": {
			ctx: ce.EventContextV02{
				SpecVersion: ce.CloudEventsVersionV02,
				ID:          "ABC-123",
				Type:        "com.example.simple",
				SchemaURL:   &types.URLRef{},
				Source:      *source,
			},
			want: []string{"schemaurl:"},
		},
		"non-empty contentType": {
			ctx: ce.EventContextV02{
				SpecVersion: ce.CloudEventsVersionV02,
				ID:          "ABC-123",
				Type:        "com.example.simple",
				Source:      *source,
				ContentType: strptr(""),
			},
			want: []string{"contenttype:"},
		},
		//"empty extensions": {
		//	ctx: ce.EventContextV02{
		//		SpecVersion: ce.CloudEventsVersionV02,
		//		ID:            "ABC-123",
		//		Type:          "com.example.simple",
		//		Source:             *source,
		//		Extensions:         make(map[string]interface{}),
		//	},
		//	want: []string{"extensions:"},
		//},
		"all errors": {
			ctx: ce.EventContextV02{
				SpecVersion: "",
				ID:          "",
				SchemaURL:   &types.URLRef{},
				ContentType: strptr(""),
				Extensions:  make(map[string]interface{}),
			},
			want: []string{
				"type:",
				"id:",
				"specversion:",
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

func TestGetMediaTypeV02(t *testing.T) {
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

			ec := ce.EventContextV02{}
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
