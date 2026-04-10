/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package event

import (
	"errors"
	"mime"
	"strings"
)

const (
	TextPlain                       = "text/plain"
	TextJSON                        = "text/json"
	ApplicationJSON                 = "application/json"
	ApplicationXML                  = "application/xml"
	ApplicationCloudEventsJSON      = "application/cloudevents+json"
	ApplicationCloudEventsBatchJSON = "application/cloudevents-batch+json"
)

// isJSON reports whether contentType denotes JSON: subtype "json" (e.g. application/json, text/json) or a subtype
// whose final "+" facet is "json" per RFC 6838 (e.g. application/cloudevents+json). Parameters are ignored.
// Empty or whitespace-only contentType is treated as JSON for backward compatibility.
func isJSON(contentType string) bool {
	if strings.TrimSpace(contentType) == "" {
		return true
	}

	mediaType, _, err := mime.ParseMediaType(contentType)

	// if err is for an invalid media parameter, we can still check the mediatype, since we don't use parameters:
	if err == nil || errors.Is(err, mime.ErrInvalidMediaParameter) {
		if _, subtype, ok := strings.Cut(mediaType, "/"); ok {
			parts := strings.Split(subtype, "+")
			return parts[len(parts)-1] == "json"
		}
	}

	// fallback to checking the suffix of the entire Content-Type:
	return strings.HasSuffix(contentType, "/json")
}

// StringOfApplicationJSON returns a string pointer to "application/json"
func StringOfApplicationJSON() *string {
	a := ApplicationJSON
	return &a
}

// StringOfApplicationXML returns a string pointer to "application/xml"
func StringOfApplicationXML() *string {
	a := ApplicationXML
	return &a
}

// StringOfTextPlain returns a string pointer to "text/plain"
func StringOfTextPlain() *string {
	a := TextPlain
	return &a
}

// StringOfApplicationCloudEventsJSON  returns a string pointer to
// "application/cloudevents+json"
func StringOfApplicationCloudEventsJSON() *string {
	a := ApplicationCloudEventsJSON
	return &a
}

// StringOfApplicationCloudEventsBatchJSON returns a string pointer to
// "application/cloudevents-batch+json"
func StringOfApplicationCloudEventsBatchJSON() *string {
	a := ApplicationCloudEventsBatchJSON
	return &a
}
