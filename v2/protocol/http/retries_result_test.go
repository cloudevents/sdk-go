package http

import (
	"errors"
	"io"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/cloudevents/sdk-go/v2/protocol"
)

func TestRetriesNil_Is(t *testing.T) {
	var err error
	if protocol.ResultIs(err, NewRetriesResult(nil, 0, time.Now(), nil)) {
		t.Error("Did not expect error to be a NewRetriesResult")
	}
}

func TestRetriesError_Is(t *testing.T) {
	err := errors.New("some other error")
	if protocol.ResultIs(err, NewRetriesResult(nil, 0, time.Now(), nil)) {
		t.Error("Did not expect error to be a Result")
	}
}

func TestRetriesNew_Is(t *testing.T) {
	err := NewRetriesResult(NewResult(200, "this is an example message, %s", "yep"), 0, time.Now(), nil)
	if !protocol.ResultIs(err, NewResult(200, "OK")) {
		t.Error("Expected error to be a 200 level result")
	}
}

func TestRetriesNewOtherType_Is(t *testing.T) {
	err := NewRetriesResult(NewResult(404, "this is an example error, %s", "yep"), 0, time.Now(), nil)
	if protocol.ResultIs(err, NewResult(200, "OK")) {
		t.Error("Expected error to be a [Normal, ExampleStatusFailed], filtered by eventtype failed")
	}
}

func TestRetriesNewWrappedErrors_Is(t *testing.T) {
	err := NewRetriesResult(NewResult(500, "this is a wrapped error, %w", io.ErrUnexpectedEOF), 0, time.Now(), nil)
	if !protocol.ResultIs(err, io.ErrUnexpectedEOF) {
		t.Error("Result expected to be a wrapped ErrUnexpectedEOF but was not")
	}
}

func TestRetriesNewOtherStatus_Is(t *testing.T) {
	err := NewRetriesResult(NewResult(403, "this is an example error, %s", "yep"), 0, time.Now(), nil)
	if protocol.ResultIs(err, NewRetriesResult(NewResult(200, "OK"), 0, time.Now(), nil)) {
		t.Error("Did not expect event to be StatusCode=200")
	}
}

func TestRetriesNew_As(t *testing.T) {
	err := NewRetriesResult(NewResult(404, "this is an example error, %s", "yep"), 5, time.Now(), nil)

	var event *RetriesResult
	if !protocol.ResultAs(err, &event) {
		t.Errorf("Expected error to be a NewRetriesResult, is not")
	}

	if event.Retries != 5 {
		t.Errorf("Mismatched retries")
	}
}

func TestRetriesNil_As(t *testing.T) {
	var err error

	var event *RetriesResult
	if protocol.ResultAs(err, &event) {
		t.Error("Did not expect error to be a Result")
	}
}

func TestRetriesNew_Error(t *testing.T) {
	err := NewRetriesResult(NewResult(500, "this is an example error, %s", "yep"), 0, time.Now(), nil)

	const want = "500: this is an example error, yep"
	got := err.Error()
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Unexpected diff (-want, +got) = %v", diff)
	}
}

func TestRetriesNew_ErrorWithRetries(t *testing.T) {
	err := NewRetriesResult(NewResult(500, "this is an example error, %s", "yep"), 10, time.Now(), nil)

	const want = "500: this is an example error, yep (10x)"
	got := err.Error()
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Unexpected diff (-want, +got) = %v", diff)
	}
}
