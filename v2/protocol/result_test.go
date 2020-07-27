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
		t.Error("Did not expect error to be a Result")
	}
}

func TestNilReceipt_Is_ACK(t *testing.T) {
	var err *Receipt
	if !ResultIs(err, ResultACK) {
		t.Error("Expected nil Receipt to be ACK")
	}
}

func TestReceipt_Is_ACK(t *testing.T) {
	err := NewReceipt(true, "test message")
	if !ResultIs(err, ResultACK) {
		t.Error("Expected ACK Receipt to be ACK")
	}
}

func TestReceipt_Is_NACK(t *testing.T) {
	err := NewReceipt(false, "test message")
	if !ResultIs(err, ResultNACK) {
		t.Error("Expected NACK Receipt to be NACK")
	}
}

func TestWrappedReceipt_Is_ACK(t *testing.T) {
	err := NewResult("%w", NewReceipt(true, "test message"))
	if !ResultIs(err, ResultACK) {
		t.Error("Expected wrapped ACK Receipt to be ACK")
	}
}

func TestWrappedReceipt_Is_NACK(t *testing.T) {
	err := NewResult("%w", NewReceipt(false, "test message"))
	if !ResultIs(err, ResultNACK) {
		t.Error("Expected wrapped NACK Receipt to be NACK")
	}
}

func TestNilReceipt_IsACK(t *testing.T) {
	var err *Receipt
	if !IsACK(err) {
		t.Error("Expected nil Receipt to be ACK")
	}
}

func TestNilReceipt_IsNACK(t *testing.T) {
	var err *Receipt
	if IsNACK(err) {
		t.Error("Expected nil Receipt to not be NACK")
	}
}

func TestReceipt_IsACK(t *testing.T) {
	err := NewReceipt(true, "test message")
	if !IsACK(err) {
		t.Error("Expected ACK Receipt to be ACK")
	}
}

func TestReceipt_IsNACK(t *testing.T) {
	err := NewReceipt(false, "test message")
	if !IsNACK(err) {
		t.Error("Expected NACK Receipt to be NACK")
	}
}

func TestError_Is(t *testing.T) {
	err := errors.New("some other error")
	if ResultIs(err, NewResult("")) {
		t.Error("Did not expect error to be a Result")
	}
}

func TestNewWrappedReceipt_Is(t *testing.T) {
	err := NewReceipt(true, "this is a wrapped error, %w", io.ErrUnexpectedEOF)
	if !ResultIs(err, io.ErrUnexpectedEOF) {
		t.Error("Result expected to be a wrapped ErrUnexpectedEOF but was not")
	}
}

func TestNilReceipt_Is(t *testing.T) {
	var err *Receipt
	if ResultIs(err, io.ErrUnexpectedEOF) {
		t.Error("Did not expected nil Receipt to be a wrapped ErrUnexpectedEOF, but was")
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
