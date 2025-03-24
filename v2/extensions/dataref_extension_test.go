/*
 Copyright 2024 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package extensions

import (
	"testing"

	"github.com/cloudevents/sdk-go/v2/event"
)

func TestAddDataRefExtension(t *testing.T) {
	// MUST HAVE AT LEAST ONE FAILING TEST as that'll also test to make sure
	// that the failed add() call won't actually set anything

	tests := []struct {
		dataref string
		pass    bool
	}{
		{"https://example.com/data", true},
		{"://invalid-url", false},
	}

	for _, test := range tests {
		e := event.New()
		err := AddDataRefExtension(&e, test.dataref)

		// Make sure adding it passed/fails appropriately
		if test.pass && err != nil {
			t.Fatalf("Failed to add DataRefExtension with valid URL(%s): %s",
				test.dataref, err)
		}
		if !test.pass && err == nil {
			t.Fatalf("Expected not to find DataRefExtension (%s), but did",
				test.dataref)
		}

		// Now make sure it's actually there in the 'pass' cases, but
		// missing in the failed cases.
		dr, ok := GetDataRefExtension(e)
		if test.pass {
			if !ok || dr.DataRef == "" {
				t.Fatalf("Dataref (%s) is missing after being set",
					test.dataref)
			}
			if dr.DataRef != test.dataref {
				t.Fatalf("Retrieved dataref(%v) doesn't match set value(%s)",
					dr, test.dataref)
			}
		} else {
			if ok || dr.DataRef != "" {
				t.Fatalf("Expected not to find DataRefExtension, but did(%s)",
					test.dataref)
			}
		}
	}
}

func TestGetDataRefExtensionNotFound(t *testing.T) {
	e := event.New()

	// Make sure there's no dataref by default
	dr, ok := GetDataRefExtension(e)
	if ok || dr.DataRef != "" {
		t.Fatal("Expected not to find DataRefExtension, but did")
	}
}
