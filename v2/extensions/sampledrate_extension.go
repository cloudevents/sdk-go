/*
 Copyright 2026 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package extensions

import (
	"fmt"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/cloudevents/sdk-go/v2/types"
)

const (
	SampledRateExtensionKey = "sampledrate"
)

// SampledRateExtension represents the sampledrate extension as defined in
// https://github.com/cloudevents/spec/blob/56b54136a6088c47058ed43c2df45e433b4868f0/cloudevents/extensions/sampledrate.md
//
// A zero SampledRate means the extension is not used. The spec requires the
// rate to be greater than zero, so negative values are rejected.
type SampledRateExtension struct {
	SampledRate int32 `json:"sampledrate"`
}

func (e SampledRateExtension) validate() error {
	if e.SampledRate < 0 {
		return fmt.Errorf("%s must be greater than zero, got %d", SampledRateExtensionKey, e.SampledRate)
	}
	return nil
}

// AddSampledRate sets the sampledrate extension on the event.
func (e SampledRateExtension) AddSampledRate(ev event.EventWriter) error {
	if err := e.validate(); err != nil {
		return err
	}
	if e.SampledRate == 0 {
		return nil
	}
	ev.SetExtension(SampledRateExtensionKey, e.SampledRate)
	return nil
}

// GetSampledRate retrieves the sampledrate extension from an event.
// It reports false when the extension is absent or not a positive integer.
func GetSampledRate(ev event.Event) (SampledRateExtension, bool) {
	if ext, ok := ev.Extensions()[SampledRateExtensionKey]; ok {
		if r, err := types.ToInteger(ext); err == nil && r > 0 {
			return SampledRateExtension{SampledRate: r}, true
		}
	}
	return SampledRateExtension{}, false
}

func (e *SampledRateExtension) ReadTransformer() binding.TransformerFunc {
	return func(reader binding.MessageMetadataReader, writer binding.MessageMetadataWriter) error {
		ext := reader.GetExtension(SampledRateExtensionKey)
		// Some readers report an absent extension as an empty string rather than nil.
		if ext == nil || ext == "" {
			return nil
		}
		r, err := types.ToInteger(ext)
		if err != nil {
			return err
		}
		if r <= 0 {
			return fmt.Errorf("%s must be greater than zero, got %d", SampledRateExtensionKey, r)
		}
		e.SampledRate = r
		return nil
	}
}

func (e *SampledRateExtension) WriteTransformer() binding.TransformerFunc {
	return func(reader binding.MessageMetadataReader, writer binding.MessageMetadataWriter) error {
		if err := e.validate(); err != nil {
			return err
		}
		if e.SampledRate == 0 {
			return nil
		}
		return writer.SetExtension(SampledRateExtensionKey, e.SampledRate)
	}
}
