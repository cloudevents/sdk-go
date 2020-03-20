package protocol

import (
	"errors"
	"io"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestNil_Is(t *testing.T) {
	var err error
	if ResultIs(err, NewResult("")) {
		t.Error("Did not expect error to be a ReconcilerResult")
	}
}

func TestError_Is(t *testing.T) {
	err := errors.New("some other error")
	if ResultIs(err, NewResult("")) {
		t.Error("Did not expect error to be a ReconcilerResult")
	}
}

func TestNewWrappedErrors_Is(t *testing.T) {
	err := NewResult("this is a wrapped error, %w", io.ErrUnexpectedEOF)
	if !ResultIs(err, io.ErrUnexpectedEOF) {
		t.Error("Result expected to be a wrapped ErrUnexpectedEOF but was not")
	}
}

func TestNewAnother_Is(t *testing.T) {
	err := NewResult("this is an example error, %s", "yep")
	if ResultIs(err, NewResult("")) {
		t.Error("Did not expect event to be failed")
	}
}

func TestNew_Error(t *testing.T) {
	err := NewResult("this is an example error, %s", "yep")

	const want = "this is an example error, yep"
	got := err.Error()
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Unexpected diff (-want, +got) = %v", diff)
	}
}
