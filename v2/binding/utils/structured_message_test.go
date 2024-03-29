/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package utils_test

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/binding/format"
	"github.com/cloudevents/sdk-go/v2/binding/utils"
	"github.com/cloudevents/sdk-go/v2/test"
)

func TestNewStructuredMessage(t *testing.T) {
	testEvent := test.ConvertEventExtensionsToString(t, test.FullEvent())
	jsonBytes := test.MustJSON(t, testEvent)

	message := utils.NewStructuredMessage(format.JSON, io.NopCloser(bytes.NewReader(jsonBytes)))

	require.Equal(t, binding.EncodingStructured, message.ReadEncoding())

	event := test.MustToEvent(t, context.TODO(), message)
	test.AssertEventEquals(t, testEvent, event)

	require.NoError(t, message.Finish(nil))
}
