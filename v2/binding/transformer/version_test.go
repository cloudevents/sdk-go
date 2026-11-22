/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package transformer

import (
	"context"
	"errors"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/binding/spec"
	"github.com/cloudevents/sdk-go/v2/binding/test"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/cloudevents/sdk-go/v2/types"
)

func TestVersionTranscoder(t *testing.T) {
	var testEventV03 = event.Event{
		Context: event.EventContextV03{
			Source: types.URIRef{URL: url.URL{Path: "source"}},
			ID:     "id",
			Type:   "type",
		}.AsV03(),
	}

	var testEventV1 = testEventV03
	testEventV1.Context = testEventV03.Context.AsV1()

	data := []byte("\"data\"")
	err := testEventV03.SetData(event.ApplicationJSON, data)
	require.NoError(t, err)
	err = testEventV1.SetData(event.ApplicationJSON, data)
	require.NoError(t, err)

	test.RunTransformerTests(t, context.Background(), []test.TransformerTestArgs{
		{
			Name:         "V03 -> V1 with Mock Structured message",
			InputMessage: test.MustCreateMockStructuredMessage(t, testEventV03),
			WantEvent:    testEventV1,
			Transformers: binding.Transformers{Version(spec.V1)},
		},
		{
			Name:         "V03 -> V1 with Mock Binary message",
			InputMessage: test.MustCreateMockBinaryMessage(testEventV03),
			WantEvent:    testEventV1,
			Transformers: binding.Transformers{Version(spec.V1)},
		},
		{
			Name:         "V03 -> V1 with Event message",
			InputEvent:   testEventV03,
			WantEvent:    testEventV1,
			Transformers: binding.Transformers{Version(spec.V1)},
		},
	})
}

type errorWriter struct{ err error }

func (w *errorWriter) SetAttribute(attribute spec.Attribute, value interface{}) error { return w.err }
func (w *errorWriter) SetExtension(name string, value interface{}) error              { return nil }

type simpleReader struct {
	version spec.Version
	attrs   map[spec.Kind]interface{}
}

func (r *simpleReader) GetAttribute(k spec.Kind) (spec.Attribute, interface{}) {
	if a := r.version.AttributeFromKind(k); a != nil {
		return a, r.attrs[k]
	}
	return nil, nil
}
func (r *simpleReader) GetExtension(name string) interface{} { return nil }

func TestVersionTransformerPropagatesSetAttributeErrors(t *testing.T) {
	expectedErr := errors.New("set attribute failed")
	reader := &simpleReader{
		version: spec.V03,
		attrs: map[spec.Kind]interface{}{
			spec.ID:          "test-id",
			spec.Source:      "http://example.com/source",
			spec.SpecVersion: "0.3",
			spec.Type:        "test.type",
		},
	}

	err := Version(spec.V1).Transform(reader, &errorWriter{err: expectedErr})

	require.Error(t, err, "Version transformer must propagate errors from SetAttribute")
	require.True(t, errors.Is(err, expectedErr), "got: %v", err)
}
