/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package event

import "testing"

func TestIsJSON(t *testing.T) {
	tests := []struct {
		contentType string
		want        bool
	}{
		// Legacy: unset type is treated as JSON.
		{"", true},
		{"   ", true},
		// Subtype exactly "json".
		{"application/json", true},
		{"text/json", true},
		{"application/json; charset=utf-8", true},
		{"  application/json  ; charset=utf-8 ", true},
		{"Application/JSON", true},
		// Final structured-syntax facet is "json" (RFC 6838: only the segment after the last "+" counts).
		{"application/cloudevents+json", true},
		{"application/cloudevents-batch+json", true},
		{"application/problem+json", true},
		{"text/json; charset=utf-8", true},
		{"application/vnd.custom+json+gzip", false},
		// Subtype ends with "json" but not as "/"json or final "+json".
		{"application/not-json", false},
		{"application/vnd.api+json-fork", false},
		{"text/plain", false},
		{"text/xml", false},
		{"application/xml", false},
		{"not-a-media-type", false},
		{"application/foojson", false},
		{"foojson", false},
		// Broken parameters: mediatype still returned with ErrInvalidMediaParameter.
		{"application/json; charset=", true},
	}
	for _, tt := range tests {
		if got := isJSON(tt.contentType); got != tt.want {
			t.Errorf("isJSON(%q) = %v, want %v", tt.contentType, got, tt.want)
		}
	}
}
