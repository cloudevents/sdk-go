package http

import (
	"context"
	"testing"
	"time"
)

func TestReadTimeout(t *testing.T) {
	// Test case 1: Value is present in the context
	ctxWithValue := context.WithValue(context.Background(), InboundReadTimeout{}, 1000*time.Second)
	resultWithValue := readTimeout(ctxWithValue)
	expectedValue := 1000 * time.Second

	if resultWithValue != expectedValue {
		t.Errorf("Expected timeout value %v, but got %v", expectedValue, resultWithValue)
	}

	// Test case 2: Invalid value is present in the context
	ctxWithValue = context.WithValue(context.Background(), InboundReadTimeout{}, "invalid")
	resultWithValue = readTimeout(ctxWithValue)

	if resultWithValue != DefaultTimeout {
		t.Errorf("Expected default timeout value %v, but got %v", DefaultTimeout, resultWithValue)
	}

	// Test case 3: Value is not present in the context
	ctxWithoutValue := context.Background()
	resultWithoutValue := readTimeout(ctxWithoutValue)

	if resultWithoutValue != DefaultTimeout {
		t.Errorf("Expected default timeout value %v, but got %v", DefaultTimeout, resultWithoutValue)
	}
}

func TestWriteTimeout(t *testing.T) {
	// Test case 1: Value is present in the context
	ctxWithValue := context.WithValue(context.Background(), InboundWriteTimeout{}, 1000*time.Second)
	resultWithValue := writeTimeout(ctxWithValue)
	expectedValue := 1000 * time.Second

	if resultWithValue != expectedValue {
		t.Errorf("Expected timeout value %v, but got %v", expectedValue, resultWithValue)
	}

	// Test case 2: Invalid value is present in the context
	ctxWithValue = context.WithValue(context.Background(), InboundWriteTimeout{}, "invalid")
	resultWithValue = writeTimeout(ctxWithValue)

	if resultWithValue != DefaultTimeout {
		t.Errorf("Expected default timeout value %v, but got %v", DefaultTimeout, resultWithValue)
	}

	// Test case 3: Value is not present in the context
	ctxWithoutValue := context.Background()
	resultWithoutValue := writeTimeout(ctxWithoutValue)

	if resultWithoutValue != DefaultTimeout {
		t.Errorf("Expected default timeout value %v, but got %v", DefaultTimeout, resultWithoutValue)
	}
}
