/*
 Copyright 2026 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package extensions

import (
	"time"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/cloudevents/sdk-go/v2/types"
)

const (
	ExpiryTimeExtensionKey = "expirytime"
)

// ExpiryTimeExtension represents the expirytime extension as defined in
// https://github.com/cloudevents/spec/blob/main/cloudevents/extensions/expirytime.md
type ExpiryTimeExtension struct {
	ExpiryTime time.Time `json:"expirytime"`
}

// AddExpiryTime sets the expirytime extension on the event.
func (e ExpiryTimeExtension) AddExpiryTime(ev event.EventWriter) {
	if !e.ExpiryTime.IsZero() {
		ev.SetExtension(ExpiryTimeExtensionKey, types.Timestamp{Time: e.ExpiryTime})
	}
}

// GetExpiryTime retrieves the expirytime extension from an event.
func GetExpiryTime(ev event.Event) (ExpiryTimeExtension, bool) {
	if ext, ok := ev.Extensions()[ExpiryTimeExtensionKey]; ok {
		if t, err := types.ToTime(ext); err == nil {
			return ExpiryTimeExtension{ExpiryTime: t}, true
		}
	}
	return ExpiryTimeExtension{}, false
}

// IsExpired checks whether the event has expired relative to the given time.
func IsExpired(ev event.Event, now time.Time) bool {
	if ext, ok := GetExpiryTime(ev); ok {
		return now.After(ext.ExpiryTime)
	}
	return false
}

func (e *ExpiryTimeExtension) ReadTransformer() binding.TransformerFunc {
	return func(reader binding.MessageMetadataReader, writer binding.MessageMetadataWriter) error {
		ext := reader.GetExtension(ExpiryTimeExtensionKey)
		if ext != nil {
			formatted, err := types.Format(ext)
			if err != nil {
				return err
			}
			if formatted != "" {
				t, err := types.ParseTime(formatted)
				if err != nil {
					return err
				}
				e.ExpiryTime = t
			}
		}
		return nil
	}
}

func (e *ExpiryTimeExtension) WriteTransformer() binding.TransformerFunc {
	return func(reader binding.MessageMetadataReader, writer binding.MessageMetadataWriter) error {
		if !e.ExpiryTime.IsZero() {
			return writer.SetExtension(ExpiryTimeExtensionKey, types.Timestamp{Time: e.ExpiryTime})
		}
		return nil
	}
}
