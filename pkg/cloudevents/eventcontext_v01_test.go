package cloudevents_test

import (
	ce "github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/types"
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
				ContentType:        strptr("application/json"),
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
				Source:             *source,
			},
			want: []string{"cloudEventsVersion:"},
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
