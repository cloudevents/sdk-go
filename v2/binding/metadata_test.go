/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package binding

import (
	"testing"

	"github.com/cloudevents/sdk-go/v2/binding/spec"
	"github.com/stretchr/testify/assert"
)

type mockMetadataReader struct {
	extensions map[string]interface{}
}

func (m *mockMetadataReader) GetAttribute(attributeKind spec.Kind) (spec.Attribute, interface{}) {
	return nil, nil
}

func (m *mockMetadataReader) GetExtension(name string) interface{} {
	return m.extensions[name]
}

type mockMetadataWriter struct {
	extensions map[string]interface{}
}

func (m *mockMetadataWriter) SetAttribute(attribute spec.Attribute, value interface{}) error {
	return nil
}

func (m *mockMetadataWriter) SetExtension(name string, value interface{}) error {
	if m.extensions == nil {
		m.extensions = make(map[string]interface{})
	}
	m.extensions[name] = value
	return nil
}

func TestExtractMetadata(t *testing.T) {
	reader := &mockMetadataReader{
		extensions: map[string]interface{}{
			"test1": "value1",
			"test2": "value2",
		},
	}

	var v1, v2, v3 string
	mapping := map[string]*string{
		"test1": &v1,
		"test2": &v2,
		"test3": &v3,
	}

	err := ExtractMetadata(reader, mapping)

	assert.NoError(t, err)
	assert.Equal(t, "value1", v1)
	assert.Equal(t, "value2", v2)
	assert.Equal(t, "", v3)
}

func TestAttachMetadata(t *testing.T) {
	writer := &mockMetadataWriter{}
	mapping := map[string]string{
		"test1": "value1",
		"test2": "value2",
		"test3": "",
	}

	err := AttachMetadata(writer, mapping)

	assert.NoError(t, err)
	assert.Equal(t, "value1", writer.extensions["test1"])
	assert.Equal(t, "value2", writer.extensions["test2"])
	_, ok := writer.extensions["test3"]
	assert.False(t, ok)
}
