/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package event

import (
	"errors"
	"fmt"
	"strings"

	"github.com/cloudevents/sdk-go/v2/types"
)

const (
	// DataContentEncodingKey is the key to DeprecatedDataContentEncoding for versions that do not support data content encoding
	// directly.
	DataContentEncodingKey = "datacontentencoding"
)

var (
	// This determines the behavior of validateExtensionName(). For MaxExtensionNameLength > 0, an error will be returned,
	// if len(key) > MaxExtensionNameLength
	MaxExtensionNameLength = 0
)

func caseInsensitiveSearch(key string, space map[string]interface{}) (interface{}, bool) {
	lkey := strings.ToLower(key)
	for k, v := range space {
		if strings.EqualFold(lkey, strings.ToLower(k)) {
			return v, true
		}
	}
	return nil, false
}

func IsExtensionNameValid(key string) bool {
	if err := validateExtensionName(key); err != nil {
		return false
	}
	return true
}

func validateExtensionName(key string) error {
	if len(key) < 1 {
		return errors.New("bad key, CloudEvents attribute names MUST NOT be empty")
	}
	if MaxExtensionNameLength > 0 && len(key) > MaxExtensionNameLength {
		return fmt.Errorf("bad key, CloudEvents attribute name '%s' is longer than %d characters", key, MaxExtensionNameLength)
	}

	for _, c := range key {
		if (c < 'a' || c > 'z') && (c < 'A' || c > 'Z') && (c < '0' || c > '9') {
			return errors.New("bad key, CloudEvents attribute names MUST consist of lower-case letters ('a' to 'z'), upper-case letters ('A' to 'Z') or digits ('0' to '9') from the ASCII character set")
		}
	}
	return nil
}

// ExtractExtensions reads multiple extension attributes from an EventReader into the provided mapping.
// It returns true if at least one extension was found and successfully mapped.
func ExtractExtensions[T ~string](reader EventReader, mapping map[string]*T) bool {
	found := false
	extensions := reader.Extensions()
	for name, target := range mapping {
		v, ok := extensions[name]
		if !ok {
			continue
		}
		s, err := types.ToString(v)
		if err != nil {
			continue
		}
		*target = T(s)
		found = true
	}
	return found
}

// AttachExtensions sets multiple extension attributes on an EventWriter using the provided mapping.
// It skips empty values to ensure only valid data is written.
func AttachExtensions[T ~string](writer EventWriter, mapping map[string]T) {
	for name, value := range mapping {
		if value == "" {
			continue
		}
		writer.SetExtension(name, string(value))
	}
}
