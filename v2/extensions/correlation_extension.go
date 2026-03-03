/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package extensions

import (
	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/event"
)

const (
	// CorrelationIDExtension is the CloudEvents extension attribute for correlationid.
	CorrelationIDExtension = "correlationid"
	// CausationIDExtension is the CloudEvents extension attribute for causationid.
	CausationIDExtension = "causationid"
)

// CorrelationExtension represents the correlation CloudEvents extension.
type CorrelationExtension struct {
	// CorrelationID is an identifier that groups related events within the same logical flow or business transaction.
	CorrelationID string `json:"correlationid"`
	// CausationID is the unique identifier of the event that directly caused this event to be generated.
	CausationID string `json:"causationid"`
}

// AddCorrelationAttributes adds the correlation attributes to the cloudevents context.
func (c CorrelationExtension) AddCorrelationAttributes(e event.EventWriter) {
	event.AttachExtensions(e, map[string]string{
		CorrelationIDExtension: c.CorrelationID,
		CausationIDExtension:   c.CausationID,
	})
}

// GetCorrelationExtension extracts the correlation extension from the event.
func GetCorrelationExtension(e event.Event) (CorrelationExtension, bool) {
	c := CorrelationExtension{}
	found := event.ExtractExtensions(e, map[string]*string{
		CorrelationIDExtension: &c.CorrelationID,
		CausationIDExtension:   &c.CausationID,
	})
	return c, found
}

// ReadTransformer returns a transformer that reads the correlation extension from the message metadata.
func (c *CorrelationExtension) ReadTransformer() binding.TransformerFunc {
	return func(reader binding.MessageMetadataReader, writer binding.MessageMetadataWriter) error {
		return binding.ExtractMetadata(reader, map[string]*string{
			CorrelationIDExtension: &c.CorrelationID,
			CausationIDExtension:   &c.CausationID,
		})
	}
}

// WriteTransformer returns a transformer that writes the correlation extension to the message metadata.
func (c CorrelationExtension) WriteTransformer() binding.TransformerFunc {
	return func(reader binding.MessageMetadataReader, writer binding.MessageMetadataWriter) error {
		return binding.AttachMetadata(writer, map[string]string{
			CorrelationIDExtension: c.CorrelationID,
			CausationIDExtension:   c.CausationID,
		})
	}
}
