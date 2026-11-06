/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package http

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/binding/spec"
	bindingtest "github.com/cloudevents/sdk-go/v2/binding/test"
	"github.com/cloudevents/sdk-go/v2/binding/transformer"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/cloudevents/sdk-go/v2/test"
)

// Mirror of TestMessageTransformDeleteExtension in message_test.go, which covers
// the request writer, for the response writer.
func TestWriteHttpResponseWriter_delete_extension(t *testing.T) {
	eventIn := test.ConvertEventExtensionsToString(t, test.FullEvent())
	messageIn := bindingtest.MustCreateMockBinaryMessage(eventIn)

	res := httptest.NewRecorder()
	require.NoError(t, WriteResponseWriter(context.TODO(), messageIn, 202, res,
		transformer.DeleteExtension("exstring")))

	response := res.Result()
	require.Equal(t, 202, response.StatusCode)

	messageOut := NewMessageFromHttpResponse(response)
	require.Equal(t, binding.EncodingBinary, messageOut.ReadEncoding())
	require.Equal(t, nil, binding.MessageMetadataReader(messageOut).GetExtension("exstring"))
}

func TestWriteHttpResponseWriter_delete_attribute(t *testing.T) {
	eventIn := test.ConvertEventExtensionsToString(t, test.FullEvent())
	messageIn := bindingtest.MustCreateMockBinaryMessage(eventIn)

	res := httptest.NewRecorder()
	require.NoError(t, WriteResponseWriter(context.TODO(), messageIn, 202, res,
		transformer.DeleteAttribute(spec.Subject)))

	response := res.Result()
	require.Equal(t, 202, response.StatusCode)

	messageOut := NewMessageFromHttpResponse(response)
	_, subject := binding.MessageMetadataReader(messageOut).GetAttribute(spec.Subject)
	require.Nil(t, subject)

	eventOut, err := binding.ToEvent(context.TODO(), messageOut)
	require.NoError(t, err)
	require.Equal(t, "", eventOut.Subject())
	require.Equal(t, event.CloudEventsVersionV1, eventOut.SpecVersion())
}
