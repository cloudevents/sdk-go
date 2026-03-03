/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package event

import (
	"errors"
	"testing"
)

func TestEvent_validateExtensionName(t *testing.T) {
	testCases := map[string]struct {
		key  string
		want error
	}{
		"empty key": {
			key:  "",
			want: errors.New("bad key, CloudEvents attribute names MUST NOT be empty"),
		},
		"invalid character": {
			key:  "invalid_key",
			want: errors.New("bad key, CloudEvents attribute names MUST consist of lower-case letters ('a' to 'z'), upper-case letters ('A' to 'Z') or digits ('0' to '9') from the ASCII character set"),
		},
		"valid key": {
			key:  "validkey123",
			want: nil,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			err := validateExtensionName(tc.key)
			if err != nil && err.Error() != tc.want.Error() || err == nil && tc.want != nil {
				t.Errorf("unexpected error, expected: %v, actual: %v", tc.want, err)
			}
		})
	}
}

func TestExtractExtensions(t *testing.T) {
	e := New()
	e.SetExtension("test1", "value1")
	e.SetExtension("test2", "value2")

	var v1, v2, v3 string
	mapping := map[string]*string{
		"test1": &v1,
		"test2": &v2,
		"test3": &v3,
	}

	found := ExtractExtensions(&e, mapping)

	if !found {
		t.Errorf("expected found to be true")
	}
	if v1 != "value1" {
		t.Errorf("expected v1 to be value1, got %s", v1)
	}
	if v2 != "value2" {
		t.Errorf("expected v2 to be value2, got %s", v2)
	}
	if v3 != "" {
		t.Errorf("expected v3 to be empty, got %s", v3)
	}
}

func TestExtractExtensions_NotFound(t *testing.T) {
	e := New()

	var v1 string
	mapping := map[string]*string{
		"test1": &v1,
	}

	found := ExtractExtensions(&e, mapping)

	if found {
		t.Errorf("expected found to be false")
	}
	if v1 != "" {
		t.Errorf("expected v1 to be empty, got %s", v1)
	}
}

func TestAttachExtensions(t *testing.T) {
	e := New()
	mapping := map[string]string{
		"test1": "value1",
		"test2": "value2",
		"test3": "",
	}

	AttachExtensions(&e, mapping)

	if e.Extensions()["test1"] != "value1" {
		t.Errorf("expected test1 to be value1, got %v", e.Extensions()["test1"])
	}
	if e.Extensions()["test2"] != "value2" {
		t.Errorf("expected test2 to be value2, got %v", e.Extensions()["test2"])
	}
	if _, ok := e.Extensions()["test3"]; ok {
		t.Errorf("expected test3 to be missing")
	}
}
