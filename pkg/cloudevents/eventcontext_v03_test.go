package cloudevents_test

import (
	ce "github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/types"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestValidateV03(t *testing.T) {
	now := types.Timestamp{Time: time.Now()}

	sourceUrl, _ := url.Parse("http://example.com/source")
	source := &types.URLRef{URL: *sourceUrl}

	schemaUrl, _ := url.Parse("http://example.com/schema")
	schema := &types.URLRef{URL: *schemaUrl}

	extensions := make(map[string]interface{})
	extensions["test"] = "extended"

	testCases := map[string]struct {
		ctx  ce.EventContextV03
		want []string
	}{
		"min valid": {
			ctx: ce.EventContextV03{
				SpecVersion: ce.CloudEventsVersionV02,
				ID:          "ABC-123",
				Type:        "com.example.simple",
				Source:      *source,
			},
		},
		"full valid": {
			ctx: ce.EventContextV03{
				SpecVersion:     ce.CloudEventsVersionV02,
				ID:              "ABC-123",
				Time:            &now,
				Type:            "com.example.simple",
				SchemaURL:       schema,
				DataContentType: ce.StringOfApplicationJSON(),
				Source:          *source,
				Extensions:      extensions,
			},
		},
		"no Type": {
			ctx: ce.EventContextV03{
				SpecVersion: ce.CloudEventsVersionV02,
				ID:          "ABC-123",
				Source:      *source,
			},
			want: []string{"type:"},
		},
		"non-empty SpecVersion": {
			ctx: ce.EventContextV03{
				SpecVersion: "",
				ID:          "ABC-123",
				Type:        "com.example.simple",
				Source:      *source,
			},
			want: []string{"specversion:"},
		},
		"missing source": {
			ctx: ce.EventContextV03{
				SpecVersion: ce.CloudEventsVersionV02,
				ID:          "ABC-123",
				Type:        "com.example.simple",
			},
			want: []string{"source:"},
		},
		"non-empty ID": {
			ctx: ce.EventContextV03{
				SpecVersion: ce.CloudEventsVersionV02,
				ID:          "",
				Type:        "com.example.simple",
				Source:      *source,
			},
			want: []string{"id:"},
		},
		"empty schemaURL": {
			ctx: ce.EventContextV03{
				SpecVersion: ce.CloudEventsVersionV02,
				ID:          "ABC-123",
				Type:        "com.example.simple",
				SchemaURL:   &types.URLRef{},
				Source:      *source,
			},
			want: []string{"schemaurl:"},
		},
		"non-empty contentType": {
			ctx: ce.EventContextV03{
				SpecVersion:     ce.CloudEventsVersionV02,
				ID:              "ABC-123",
				Type:            "com.example.simple",
				Source:          *source,
				DataContentType: strptr(""),
			},
			want: []string{"contenttype:"},
		},
		//"empty extensions": {
		//	ctx: ce.EventContextV03{
		//		SpecVersion: ce.CloudEventsVersionV02,
		//		ID:            "ABC-123",
		//		Type:          "com.example.simple",
		//		Source:             *source,
		//		Extensions:         make(map[string]interface{}),
		//	},
		//	want: []string{"extensions:"},
		//},
		"all errors": {
			ctx: ce.EventContextV03{
				SpecVersion:     "",
				ID:              "",
				SchemaURL:       &types.URLRef{},
				DataContentType: strptr(""),
				Extensions:      make(map[string]interface{}),
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
