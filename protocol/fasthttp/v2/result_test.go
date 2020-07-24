package fasthttp

import (
	"errors"
	"io"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/cloudevents/sdk-go/v2/protocol"
)

func TestNil_Is(t *testing.T) {
	var err error
	if protocol.ResultIs(err, NewResult(200, "OK")) {
		t.Error("Did not expect error to be a Result")
	}
}

func TestError_Is(t *testing.T) {
	err := errors.New("some other error")
	if protocol.ResultIs(err, NewResult(200, "OK")) {
		t.Error("Did not expect error to be a Result")
	}
}

func TestNew_Is(t *testing.T) {
	err := NewResult(200, "this is an example message, %s", "yep")
	if !protocol.ResultIs(err, NewResult(200, "OK")) {
		t.Error("Expected error to be a 200 level result")
	}
}

func TestNewOtherType_Is(t *testing.T) {
	err := NewResult(404, "this is an example error, %s", "yep")
	if protocol.ResultIs(err, NewResult(200, "OK")) {
		t.Error("Expected error to be a [Normal, ExampleStatusFailed], filtered by eventtype failed")
	}
}

func TestNewWrappedErrors_Is(t *testing.T) {
	err := NewResult(500, "this is a wrapped error, %w", io.ErrUnexpectedEOF)
	if !protocol.ResultIs(err, io.ErrUnexpectedEOF) {
		t.Error("Result expected to be a wrapped ErrUnexpectedEOF but was not")
	}
}

func TestNewOtherStatus_Is(t *testing.T) {
	err := NewResult(403, "this is an example error, %s", "yep")
	if protocol.ResultIs(err, NewResult(200, "OK")) {
		t.Error("Did not expect event to be StatusCode=200")
	}
}

func TestNew_As(t *testing.T) {
	err := NewResult(404, "this is an example error, %s", "yep")

	var event *Result
	if !protocol.ResultAs(err, &event) {
		t.Errorf("Expected error to be a Result, is not")
	}

	if event.StatusCode != 404 {
		t.Errorf("Mismatched StatusCode")
	}
}

func TestNil_As(t *testing.T) {
	var err error

	var event *Result
	if protocol.ResultAs(err, &event) {
		t.Error("Did not expect error to be a Result")
	}
}

func TestNew_Error(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want string
	}{{
		name: "no wrapped error",
		err:  NewResult(500, "this is an example error, %s", "yep"),
		want: "500: this is an example error, yep",
	}, {
		name: "wrapped error",
		err:  NewResult(400, "outer error: %w", errors.New("inner error")),
		want: "400: outer error: inner error",
	}}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.err.Error()
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("Unexpected diff (-want, +got) = %v", diff)
			}
		})
	}
}
