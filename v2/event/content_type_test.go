/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package event_test

import (
	"testing"

	"github.com/cloudevents/sdk-go/v2/event"

	"github.com/google/go-cmp/cmp"
)

func TestStringOfApplicationJSON(t *testing.T) {
	want := strptr("application/json")
	got := event.StringOfApplicationJSON()

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("unexpected string (-want, +got) = %v", diff)
	}
}

func TestStringOfApplicationXML(t *testing.T) {
	want := strptr("application/xml")
	got := event.StringOfApplicationXML()

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("unexpected string (-want, +got) = %v", diff)
	}
}

func TestStringOfTextPlain(t *testing.T) {
	want := strptr("text/plain")
	got := event.StringOfTextPlain()

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("unexpected string (-want, +got) = %v", diff)
	}
}

func TestStringOfApplicationCloudEventsJSON(t *testing.T) {
	want := strptr("application/cloudevents+json")
	got := event.StringOfApplicationCloudEventsJSON()

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("unexpected string (-want, +got) = %v", diff)
	}
}

func TestStringOfApplicationCloudEventsBatchJSON(t *testing.T) {
	want := strptr("application/cloudevents-batch+json")
	got := event.StringOfApplicationCloudEventsBatchJSON()

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("unexpected string (-want, +got) = %v", diff)
	}
}

func TestContentTypeIsJSON(t *testing.T) {
	tests := []struct {
		name     string
		ct       event.ContentType
		expected bool
	}{
		{name: "Empty", ct: "", expected: true},
		{name: "ApplicationJSON", ct: event.ApplicationJSON, expected: true},
		{name: "TextJSON", ct: event.TextJSON, expected: true},
		{name: "ApplicationCloudEventsJSON", ct: event.ApplicationCloudEventsJSON, expected: true},
		{name: "ApplicationCloudEventsBatchJSON", ct: event.ApplicationCloudEventsBatchJSON, expected: true},
		{name: "ApplicationXML", ct: event.ApplicationXML, expected: false},
		{name: "TextPlain", ct: event.TextPlain, expected: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.ct.IsJSON(); got != tc.expected {
				t.Errorf("ContentType.IsJSON() = %v, want %v", got, tc.expected)
			}
		})
	}
}
