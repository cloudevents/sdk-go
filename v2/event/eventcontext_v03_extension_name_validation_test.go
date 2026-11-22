/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package event_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/cloudevents/sdk-go/v2/types"
)

// CloudEvents v0.3 attribute names are ASCII letters or digits and MUST begin
// with a letter. Unlike EventContextV1.SetExtension, EventContextV03 does not
// lower-case the name: mixed-case input is stored and read back unchanged.
// ExtensionAs is case-sensitive on v0.3.
func TestEventContextV03_SetExtension_ValidatesName(t *testing.T) {
	testCases := map[string]struct {
		name    string
		wantErr string
	}{
		"empty name": {
			name:    "",
			wantErr: "MUST NOT be empty",
		},
		"leading digit": {
			name:    "1abc",
			wantErr: "MUST begin with a letter",
		},
		"underscore in name": {
			name:    "trace_parent",
			wantErr: "MUST consist of",
		},
		"hyphen in name": {
			name:    "my-extension",
			wantErr: "MUST consist of",
		},
		"space in name": {
			name:    "has space",
			wantErr: "MUST consist of",
		},
		"non-ASCII": {
			name:    "café",
			wantErr: "MUST consist of",
		},
		"data_base64 is not a legal 0.3 extension name": {
			name:    "data_base64",
			wantErr: "MUST consist of",
		},
		"reserved id": {
			name:    "id",
			wantErr: "MUST NOT be overwritten",
		},
		"reserved type": {
			name:    "type",
			wantErr: "MUST NOT be overwritten",
		},
		"reserved source": {
			name:    "source",
			wantErr: "MUST NOT be overwritten",
		},
		"reserved subject": {
			name:    "subject",
			wantErr: "MUST NOT be overwritten",
		},
		"reserved time": {
			name:    "time",
			wantErr: "MUST NOT be overwritten",
		},
		"reserved schemaurl": {
			name:    "schemaurl",
			wantErr: "MUST NOT be overwritten",
		},
		"reserved datacontenttype": {
			name:    "datacontenttype",
			wantErr: "MUST NOT be overwritten",
		},
		"reserved datacontentencoding": {
			name:    "datacontentencoding",
			wantErr: "MUST NOT be overwritten",
		},
		"reserved mixed-case spec attribute": {
			name:    "SchemaURL",
			wantErr: "MUST NOT be overwritten",
		},
		"valid key": {
			name: "validkey",
		},
		"valid key with digits": {
			name: "key123",
		},
		"upper-case name is stored unchanged": {
			name: "VALIDKEY",
		},
		"mixed-case name is stored unchanged": {
			name: "ValidKey",
		},
	}

	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			ctx := event.EventContextV03{
				ID:     "ABC-123",
				Type:   "com.example.simple",
				Source: *types.ParseURIRef("http://example.com/source"),
			}
			if err := ctx.SetExtension("keep", "yes"); err != nil {
				t.Fatalf("SetExtension(%q) error = %v", "keep", err)
			}

			err := ctx.SetExtension(tc.name, "somevalue")
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("SetExtension(%q) error = %v", tc.name, err)
				}
				if got, ok := ctx.Extensions[tc.name]; !ok || got != "somevalue" {
					t.Fatalf("Extensions[%q] = %#v, want %q", tc.name, ctx.Extensions[tc.name], "somevalue")
				}
				var read string
				if err := ctx.ExtensionAs(tc.name, &read); err != nil {
					t.Fatalf("ExtensionAs(%q) error = %v", tc.name, err)
				}
				if read != "somevalue" {
					t.Fatalf("ExtensionAs(%q) = %q, want %q", tc.name, read, "somevalue")
				}
				if lower := strings.ToLower(tc.name); lower != tc.name {
					var ignored string
					if err := ctx.ExtensionAs(lower, &ignored); err == nil {
						t.Fatalf("ExtensionAs(%q) succeeded; v0.3 names are not lower-cased", lower)
					}
					if _, ok := ctx.Extensions[lower]; ok {
						t.Fatalf("Extensions[%q] was stored; v0.3 names are not lower-cased", lower)
					}
				}
			} else {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) || !strings.Contains(err.Error(), "bad key") {
					t.Fatalf("SetExtension(%q) error = %v, want a %q error containing %q", tc.name, err, "bad key", tc.wantErr)
				}
				if _, ok := ctx.Extensions[tc.name]; ok {
					t.Fatalf("SetExtension(%q) stored an invalid name: %#v", tc.name, ctx.Extensions)
				}
			}
			if ctx.Extensions["keep"] != "yes" {
				t.Fatalf("existing extension was modified: %#v", ctx.Extensions)
			}
		})
	}
}

// JSON decoding keeps a 1.0 data_base64 member of a 0.3 event as an extension,
// including when it appears before specversion (TestUnmarshalWithOrdering).
// Validate matches v1.0 and does not reject that key.
func TestEventContextV03_UnmarshalDataBase64Extension(t *testing.T) {
	testCases := map[string]string{
		"data_base64 before specversion": `{"data_base64":"foo","id":"ABC-123","source":"http://example.com/source","type":"com.example.simple","specversion":"0.3"}`,
		"data_base64 after specversion":  `{"specversion":"0.3","id":"ABC-123","source":"http://example.com/source","type":"com.example.simple","data_base64":"foo"}`,
	}
	for n, body := range testCases {
		t.Run(n, func(t *testing.T) {
			assertV03DataBase64Extension(t, body)

			var got event.Event
			if err := json.Unmarshal([]byte(body), &got); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			raw, err := json.Marshal(got)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			assertV03DataBase64Extension(t, string(raw))
		})
	}
}

func assertV03DataBase64Extension(t *testing.T, body string) {
	t.Helper()
	var got event.Event
	if err := json.Unmarshal([]byte(body), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if err := got.Validate(); err != nil {
		t.Fatalf("Validate() = %v, want nil", err)
	}
	if got.Extensions()["data_base64"] != "foo" {
		t.Fatalf("Extensions()[data_base64] = %#v, want %q", got.Extensions()["data_base64"], "foo")
	}

	// The same key assigned without SetExtension must still validate. v1.0
	// Validate does not inspect extension names, and v0.3 matches that.
	ctx := got.Context.AsV03()
	ctx.Extensions["trace_parent"] = "00"
	if err := ctx.Validate(); err != nil {
		t.Fatalf("Validate() with existing illegal extension keys = %v, want nil", err)
	}
}

func TestUnmarshalV03_RejectsIllegalExtensionName(t *testing.T) {
	testCases := map[string]string{
		"before specversion": `{"trace_parent":"00","id":"ABC-123","source":"http://example.com/source","type":"com.example.simple","specversion":"0.3"}`,
		"after specversion":  `{"specversion":"0.3","id":"ABC-123","source":"http://example.com/source","type":"com.example.simple","trace_parent":"00"}`,
	}
	for n, body := range testCases {
		t.Run(n, func(t *testing.T) {
			var got event.Event
			err := json.Unmarshal([]byte(body), &got)
			if err == nil || !strings.Contains(err.Error(), "bad key") {
				t.Fatalf("Unmarshal error = %v, want a %q error", err, "bad key")
			}
		})
	}
}
