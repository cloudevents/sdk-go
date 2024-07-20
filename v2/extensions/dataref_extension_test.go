/*
 Copyright 2024 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package extensions

import (
	"testing"

	"github.com/cloudevents/sdk-go/v2/event"
)

func TestAddDataRefExtensionValidURL(t *testing.T) {
	e := event.New()
	expectedDataRef := "https://example.com/data"

	err := AddDataRefExtension(&e, expectedDataRef)
	if err != nil {
		t.Fatalf("Failed to add DataRefExtension with valid URL: %s", err)
	}
}

func TestAddDataRefExtensionInvalidURL(t *testing.T) {
	e := event.New()
	invalidDataRef := "://invalid-url"

	err := AddDataRefExtension(&e, invalidDataRef)
	if err == nil {
		t.Fatal("Expected error when adding DataRefExtension with invalid URL, but got none")
	}
}

func TestGetDataRefExtensionNotFound(t *testing.T) {
	e := event.New()

	_, ok := GetDataRefExtension(e)
	if ok {
		t.Fatal("Expected not to find DataRefExtension, but did")
	}
}
