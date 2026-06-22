/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package binding

import (
	"github.com/cloudevents/sdk-go/v2/types"
)

// ExtractMetadata reads metadata extensions from a MessageMetadataReader and maps them to the target pointers.
// It skips empty values to ensure only valid data is written.
func ExtractMetadata[T ~string](reader MessageMetadataReader, mapping map[string]*T) error {
	for name, target := range mapping {
		v := reader.GetExtension(name)
		if v == nil {
			continue
		}
		s, err := types.Format(v)
		if err != nil {
			return err
		}
		*target = T(s)
	}
	return nil
}

// AttachMetadata sets metadata extensions on a MessageMetadataWriter using the provided mapping.
// It skips empty values to ensure only valid data is written.
func AttachMetadata[T ~string](writer MessageMetadataWriter, mapping map[string]T) error {
	for name, value := range mapping {
		if value == "" {
			continue
		}
		if err := writer.SetExtension(name, value); err != nil {
			return err
		}
	}
	return nil
}
