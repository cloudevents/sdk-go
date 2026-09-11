/*
 Copyright 2024 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package extensions

import (
	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/cloudevents/sdk-go/v2/types"
)

const SequenceExtensionKey = "sequence"

// SequenceExtension represents the CloudEvents Sequence extension,
// which describes the position of an event in the ordered sequence
// of events produced by a unique event source.
// See https://github.com/cloudevents/spec/blob/main/cloudevents/extensions/sequence.md
// for more info.
type SequenceExtension struct {
	Sequence string `json:"sequence"`
}

// AddSequenceExtension sets the sequence extension attribute on the event.
// The value MUST be a non-empty, lexicographically-orderable string.
func AddSequenceExtension(e *event.Event, sequence string) {
	if sequence != "" {
		e.SetExtension(SequenceExtensionKey, sequence)
	}
}

// GetSequenceExtension extracts the sequence extension from a CloudEvent.
// It returns the extension and true if found, or an empty extension and false otherwise.
func GetSequenceExtension(e event.Event) (SequenceExtension, bool) {
	if val, ok := e.Extensions()[SequenceExtensionKey]; ok {
		if s, err := types.ToString(val); err == nil && s != "" {
			return SequenceExtension{Sequence: s}, true
		}
	}
	return SequenceExtension{}, false
}

// ReadTransformer returns a binding.TransformerFunc that reads the sequence
// extension from a message into this SequenceExtension struct.
func (s *SequenceExtension) ReadTransformer() binding.TransformerFunc {
	return func(reader binding.MessageMetadataReader, writer binding.MessageMetadataWriter) error {
		val := reader.GetExtension(SequenceExtensionKey)
		if val != nil {
			formatted, err := types.Format(val)
			if err != nil {
				return err
			}
			s.Sequence = formatted
		}
		return nil
	}
}

// WriteTransformer returns a binding.TransformerFunc that writes the sequence
// extension from this SequenceExtension struct into a message.
func (s *SequenceExtension) WriteTransformer() binding.TransformerFunc {
	return func(reader binding.MessageMetadataReader, writer binding.MessageMetadataWriter) error {
		if s.Sequence != "" {
			return writer.SetExtension(SequenceExtensionKey, s.Sequence)
		}
		return nil
	}
}
