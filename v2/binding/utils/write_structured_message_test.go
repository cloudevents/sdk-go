package utils_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/binding/format"
	"github.com/cloudevents/sdk-go/v2/binding/utils"
	"github.com/cloudevents/sdk-go/v2/test"
)

func TestWriteStructured(t *testing.T) {
	testEvent := test.ConvertEventExtensionsToString(t, test.FullEvent())
	testMessage := binding.ToMessage(&testEvent)

	var buffer bytes.Buffer
	err := utils.WriteStructured(context.TODO(), testMessage, &buffer)
	require.NoError(t, err)

	haveEvent := cloudevents.Event{}
	require.NoError(t, json.Unmarshal(buffer.Bytes(), &haveEvent))
	test.AssertEventEquals(t, testEvent, haveEvent)
}

func TestPipeStructured(t *testing.T) {
	testEvent := test.ConvertEventExtensionsToString(t, test.FullEvent())
	jsonBytes := test.MustJSON(t, testEvent)

	message := utils.NewStructuredMessage(format.JSON, bytes.NewReader(jsonBytes))
	defer message.Finish(nil)

	var buffer bytes.Buffer
	err := utils.WriteStructured(context.TODO(), message, &buffer)
	require.NoError(t, err)

	haveEvent := cloudevents.Event{}
	require.NoError(t, json.Unmarshal(buffer.Bytes(), &haveEvent))
	test.AssertEventEquals(t, testEvent, haveEvent)
}

func TestWriteStructuredWithWriteCloser(t *testing.T) {
	wantErr := errors.New("writer mock error")

	testEvent := test.ConvertEventExtensionsToString(t, test.FullEvent())
	testMessage := binding.ToMessage(&testEvent)

	haveErr := utils.WriteStructured(context.TODO(), testMessage, writeCloserMock{wantErr})
	require.Equal(t, wantErr, haveErr)
}

type writeCloserMock struct {
	error
}

func (w writeCloserMock) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (w writeCloserMock) Close() error {
	return w.error
}
