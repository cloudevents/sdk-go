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
	TraceParentExtension = "traceparent"
	TraceStateExtension  = "tracestate"
)

// DistributedTracingExtension represents the extension for cloudevents context
type DistributedTracingExtension struct {
	TraceParent string `json:"traceparent"`
	TraceState  string `json:"tracestate"`
}

// AddTracingAttributes adds the tracing attributes traceparent and tracestate to the cloudevents context
func (d DistributedTracingExtension) AddTracingAttributes(e event.EventWriter) {
	if d.TraceParent != "" {
		event.AttachExtensions(e, map[string]string{
			TraceParentExtension: d.TraceParent,
			TraceStateExtension:  d.TraceState,
		})
	}
}

func GetDistributedTracingExtension(e event.Event) (DistributedTracingExtension, bool) {
	d := DistributedTracingExtension{}
	ok := event.ExtractExtensions(e, map[string]*string{
		TraceParentExtension: &d.TraceParent,
		TraceStateExtension:  &d.TraceState,
	})
	return d, ok
}

func (d *DistributedTracingExtension) ReadTransformer() binding.TransformerFunc {
	return func(reader binding.MessageMetadataReader, writer binding.MessageMetadataWriter) error {
		return binding.ExtractMetadata(reader, map[string]*string{
			TraceParentExtension: &d.TraceParent,
			TraceStateExtension:  &d.TraceState,
		})
	}
}

func (d *DistributedTracingExtension) WriteTransformer() binding.TransformerFunc {
	return func(reader binding.MessageMetadataReader, writer binding.MessageMetadataWriter) error {
		return binding.AttachMetadata(writer, map[string]string{
			TraceParentExtension: d.TraceParent,
			TraceStateExtension:  d.TraceState,
		})
	}
}
