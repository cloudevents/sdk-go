/*
 Copyright 2026 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package pubsub

import (
	"context"
	"encoding/json"
	"testing"

	"cloud.google.com/go/pubsub/v2"
	"github.com/stretchr/testify/require"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/binding/spec"
	bindingtest "github.com/cloudevents/sdk-go/v2/binding/test"
	"github.com/cloudevents/sdk-go/v2/binding/transformer"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/cloudevents/sdk-go/v2/test"
)

func TestWritePubSubMessageBinary(t *testing.T) {
	test.EachEvent(t, test.Events(), func(t *testing.T, eventIn event.Event) {
		eventIn = test.ConvertEventExtensionsToString(t, eventIn)
		ctx := binding.WithForceBinary(context.Background())
		pm := &pubsub.Message{Attributes: make(map[string]string)}

		require.NoError(t, WritePubSubMessage(ctx, binding.ToMessage(&eventIn), pm))
		require.NotContains(t, pm.Attributes, "content-type")
		if eventIn.DataContentType() == "" {
			require.NotContains(t, pm.Attributes, "ce-datacontenttype")
			require.NotContains(t, pm.Attributes, "Content-Type")
		} else {
			require.Equal(t, eventIn.DataContentType(), pm.Attributes["ce-datacontenttype"])
			require.Equal(t, eventIn.DataContentType(), pm.Attributes["Content-Type"])
		}
		require.Equal(t, eventIn.Data(), pm.Data)

		messageOut := NewMessage(pm)
		require.Equal(t, binding.EncodingBinary, messageOut.ReadEncoding())
		_, dataContentType := messageOut.GetAttribute(spec.DataContentType)
		require.Equal(t, eventIn.DataContentType(), dataContentType)
		eventOut, err := binding.ToEvent(ctx, messageOut)
		require.NoError(t, err)
		test.AssertEventEquals(t, eventIn, *eventOut)
	})
}

func TestWritePubSubMessageStructured(t *testing.T) {
	test.EachEvent(t, test.Events(), func(t *testing.T, eventIn event.Event) {
		eventIn = test.ConvertEventExtensionsToString(t, eventIn)
		pm := &pubsub.Message{}
		ctx := binding.WithForceStructured(context.Background())
		require.NoError(t, WritePubSubMessage(ctx, binding.ToMessage(&eventIn), pm))
		require.Equal(t, event.ApplicationCloudEventsJSON, pm.Attributes["content-type"])
		require.NotContains(t, pm.Attributes, "ce-datacontenttype")
		var body map[string]interface{}
		require.NoError(t, json.Unmarshal(pm.Data, &body))
		require.NotContains(t, body, "ce-datacontenttype")
		if eventIn.DataContentType() == "" {
			require.NotContains(t, body, "datacontenttype")
		} else {
			require.Equal(t, eventIn.DataContentType(), body["datacontenttype"])
		}
		msg := NewMessage(pm)
		require.Equal(t, binding.EncodingStructured, msg.ReadEncoding())
		eventOut, err := binding.ToEvent(ctx, msg)
		require.NoError(t, err)
		test.AssertEventEquals(t, eventIn, *eventOut)
	})
}

func TestWritePubSubMessageChangeEncoding(t *testing.T) {
	eventIn := test.ConvertEventExtensionsToString(t, test.FullEvent())
	pm := &pubsub.Message{Attributes: map[string]string{"custom": "preserved"}}
	for _, encoding := range []binding.Encoding{binding.EncodingBinary, binding.EncodingStructured, binding.EncodingBinary} {
		t.Run(encoding.String(), func(t *testing.T) {
			ctx := binding.WithPreferredEventEncoding(context.Background(), encoding)
			require.NoError(t, WritePubSubMessage(ctx, binding.ToMessage(&eventIn), pm))
			require.Equal(t, "preserved", pm.Attributes["custom"])
			msg := NewMessage(pm)
			require.Equal(t, encoding, msg.ReadEncoding())
			eventOut, err := binding.ToEvent(ctx, msg)
			require.NoError(t, err)
			test.AssertEventEquals(t, eventIn, *eventOut)
		})
	}
}

func TestWritePubSubMessageExistingDataContentType(t *testing.T) {
	tests := []struct {
		name       string
		attributes map[string]string
	}{
		{
			name:       "legacy attribute only",
			attributes: map[string]string{"Content-Type": "application/protobuf"},
		},
		{
			name:       "prefixed attribute only",
			attributes: map[string]string{"ce-datacontenttype": "application/protobuf"},
		},
		{
			name: "legacy and prefixed attributes with matching values",
			attributes: map[string]string{
				"Content-Type":       "application/protobuf",
				"ce-datacontenttype": "application/protobuf",
			},
		},
		{
			name: "legacy and prefixed attributes with conflicting values",
			attributes: map[string]string{
				"Content-Type":       "application/protobuf",
				"ce-datacontenttype": "text/plain",
			},
		},
	}
	test.EachEvent(t, test.AllVersions([]event.Event{test.FullEvent()}), func(t *testing.T, eventIn event.Event) {
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				eventIn := test.ConvertEventExtensionsToString(t, eventIn)
				ctx := binding.WithForceBinary(context.Background())
				pm := &pubsub.Message{Attributes: map[string]string{"custom": "preserved"}}
				for name, value := range tt.attributes {
					pm.Attributes[name] = value
				}

				require.NoError(t, WritePubSubMessage(ctx, binding.ToMessage(&eventIn), pm))
				require.Equal(t, eventIn.DataContentType(), pm.Attributes["ce-datacontenttype"])
				require.Equal(t, eventIn.DataContentType(), pm.Attributes["Content-Type"])
				require.Equal(t, "preserved", pm.Attributes["custom"])
				require.Equal(t, eventIn.Data(), pm.Data)

				eventOut, err := binding.ToEvent(ctx, NewMessage(pm))
				require.NoError(t, err)
				test.AssertEventEquals(t, eventIn, *eventOut)
			})
		}
	})
}

func TestWritePubSubMessageTransformDataContentType(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
	}{
		{name: "update", value: "application/json; charset=utf-8"},
		{name: "delete", value: nil},
	}
	test.EachEvent(t, test.AllVersions([]event.Event{test.FullEvent()}), func(t *testing.T, eventIn event.Event) {
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				ctx := binding.WithForceBinary(context.Background())
				pm := &pubsub.Message{Attributes: make(map[string]string)}
				err := WritePubSubMessage(ctx, bindingtest.MustCreateMockBinaryMessage(eventIn), pm,
					transformer.SetAttribute(spec.DataContentType, func(interface{}) (interface{}, error) {
						return tt.value, nil
					}))
				require.NoError(t, err)
				if tt.value == nil {
					require.NotContains(t, pm.Attributes, "ce-datacontenttype")
					require.NotContains(t, pm.Attributes, "Content-Type")
				} else {
					require.Equal(t, tt.value, pm.Attributes["ce-datacontenttype"])
					require.Equal(t, tt.value, pm.Attributes["Content-Type"])
				}
				require.Equal(t, eventIn.Data(), pm.Data)
			})
		}
	})
}
