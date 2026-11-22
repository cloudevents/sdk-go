/*
 Copyright 2023 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package mqtt_paho

import (
	"context"
	"testing"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/binding/spec"
	"github.com/cloudevents/sdk-go/v2/binding/transformer"
	"github.com/eclipse/paho.golang/paho"
	"github.com/stretchr/testify/require"
)

// A binary PUBLISH carrying the required attributes and a data content type, which MQTT
// transports in the PUBLISH Content Type property rather than in a user property.
func requiredOnlyPublish(specVersion string) *paho.Publish {
	return &paho.Publish{
		Payload: []byte("{hello:world}"),
		Properties: &paho.PublishProperties{
			ContentType: "application/json",
			User: []paho.UserProperty{
				{Key: "ce-specversion", Value: specVersion},
				{Key: "ce-type", Value: "binary.test"},
				{Key: "ce-source", Value: "test-source"},
				{Key: "ce-id", Value: "ABC-123"},
			},
		},
	}
}

// binding.MessageMetadataReader documents that GetAttribute returns "attribute, nil" when
// the spec version supports the attribute kind but the message carries no value for it.
func TestGetAttributeReturnsNilWhenPropertyAbsent(t *testing.T) {
	m := NewMessage(requiredOnlyPublish("1.0"))
	require.Equal(t, binding.EncodingBinary, m.ReadEncoding())

	attr, value := m.GetAttribute(spec.Time)
	require.NotNil(t, attr)
	require.Nil(t, value)

	require.Nil(t, m.GetExtension("notthere"))
}

// ReadBinary maps PublishProperties.ContentType onto the datacontenttype attribute and
// SetAttribute writes it back to the same place, so GetAttribute has to read it there too.
func TestGetAttributeReadsTheDataContentType(t *testing.T) {
	m := NewMessage(requiredOnlyPublish("1.0"))

	attr, value := m.GetAttribute(spec.DataContentType)
	require.NotNil(t, attr)
	require.Equal(t, "application/json", value)
}

// transformer.AddTimeNow stamps the time attribute when the source message has none. It
// reaches the message through WritePubMessage, binding.Write and the binary transformer
// path, so a reader reporting "" instead of nil leaves its only branch unreachable.
func TestWritePubMessageAddTimeNow(t *testing.T) {
	out := &paho.Publish{}

	err := WritePubMessage(context.Background(), NewMessage(requiredOnlyPublish("1.0")), out, transformer.AddTimeNow)
	require.NoError(t, err)

	require.NotEmpty(t, out.Properties.User.Get(prefix+"time"), "ce-time was not added")
}

// transformer.Version rewrites an attribute only when GetAttribute reports a value. A
// reader that reports "" for everything therefore rewrites the attributes the event does
// not carry as empty user properties, and rewrites the data content type from the empty
// string it read out of the wrong property. CloudEvents v1.0.2 says of subject "If
// present, MUST be a non-empty string" and of dataschema "If present, MUST be a non-empty
// URI", so those belong absent rather than present and empty.
func TestWritePubMessageVersionKeepsContentTypeAndSkipsAbsentAttributes(t *testing.T) {
	out := &paho.Publish{}

	err := WritePubMessage(context.Background(), NewMessage(requiredOnlyPublish("0.3")), out, transformer.Version(spec.V1))
	require.NoError(t, err)

	require.Equal(t, "1.0", out.Properties.User.Get(prefix+"specversion"))
	require.Equal(t, "application/json", out.Properties.ContentType)

	var empty []string
	for _, p := range out.Properties.User {
		if p.Value == "" {
			empty = append(empty, p.Key)
		}
	}
	require.Empty(t, empty, "user properties written with an empty value")
}
