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
