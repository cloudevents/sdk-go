package http

import (
	"errors"
	"fmt"
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

func TestNilResult_IsACK(t *testing.T) {
	var err *Result
	if !protocol.ResultIs(err, protocol.ResultACK) {
		t.Error("Expected result to be a ACK")
	}
}

func TestResult_IsACK(t *testing.T) {
	err := NewResult(200, "%w", protocol.ResultACK)
	if !protocol.ResultIs(err, protocol.ResultACK) {
		t.Error("Expected result to be a ACK")
	}
}

func TestNil_IsUndelivered(t *testing.T) {
	{
		var err error
		if protocol.IsUndelivered(err) {
			t.Error("Expected nil result to be ACK == delivered, Got IsUndelivered.")
		}
	}
	{
		var err *Result
		if protocol.IsUndelivered(err) {
			t.Error("Expected nil result to be ACK == delivered, Got IsUndelivered.")
		}
	}
}

func Test_IsUndelivered(t *testing.T) {
	tests := []struct {
		name   string
		result error
		want   bool
	}{{
		name: "Nil error should be considered delivered",
		want: false,
	}, {
		name: "Nil *Result should be considered delivered",
		result: func() error {
			var err *Result
			return err
		}(),
		want: false,
	}, {
		name:   "EOF should be considered undelivered",
		result: io.ErrUnexpectedEOF,
		want:   true,
	}, {
		name:   "ACK should be considered delivered",
		result: protocol.ResultACK,
		want:   false,
	}, {
		name:   "NACK should be considered delivered",
		result: protocol.ResultNACK,
		want:   false,
	}, {
		name:   "200 - should be considered delivered",
		result: NewResult(200, "OK - %w", protocol.ResultACK),
		want:   false,
	}, {
		name:   "500 should be considered delivered",
		result: NewResult(500, "MY_BAD - %w", protocol.ResultNACK),
		want:   false,
	}}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := protocol.IsUndelivered(tc.result)
			if got != tc.want {
				t.Error(fmt.Sprintf("%v, expected result to be IsUndelivered == %t, got %t", tc.name, tc.want, got))
			}
		})
	}
}

func TestError_Is(t *testing.T) {
	err := errors.New("some other error")
	if protocol.ResultIs(err, NewResult(200, "OK")) {
		t.Error("Did not expect error to be a Result")
	}
}

func TestError_Is_not(t *testing.T) {
	var err *Result
	if protocol.ResultIs(err, io.ErrUnexpectedEOF) {
		t.Error("Did not expect nil *Result to be a ErrUnexpectedEOF")
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
