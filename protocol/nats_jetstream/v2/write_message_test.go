/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package nats_jetstream

import (
	"bytes"
	"context"
	"encoding/base64"
	"testing"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/test"
)

func Test_WriteMsg(t *testing.T) {
	gotHeader, err := WriteMsg(context.Background(), (*binding.EventMessage)(&testEvent), new(bytes.Buffer))
	if err != nil {
		t.Fatalf("WriteMsg() error = %v", err)
	}
	for key, want := range map[string]string{
		"ce-source":   test.Source.String(),
		"ce-time":     test.Timestamp.String(),
		"ce-exurl":    test.Source.String(),
		"ce-exbinary": base64.StdEncoding.EncodeToString([]byte{0, 1, 2, 3}),
	} {
		if got := gotHeader.Get(key); got != want {
			t.Errorf("WriteMsg() key %v got = %v, want %v", key, got, want)
		}
	}
}

// Forwarding a message that arrived in binary mode hands the writer a string for
// every attribute, including time.
func Test_WriteMsg_ForwardBinaryMessage(t *testing.T) {
	gotHeader, err := WriteMsg(context.Background(), NewMessage(binaryConsumerMessage), new(bytes.Buffer))
	if err != nil {
		t.Fatalf("WriteMsg() error = %v", err)
	}
	if got, want := gotHeader.Get("ce-time"), test.Timestamp.String(); got != want {
		t.Errorf("WriteMsg() ce-time got = %v, want %v", got, want)
	}
	if got, want := gotHeader.Get("ce-id"), testEvent.ID(); got != want {
		t.Errorf("WriteMsg() ce-id got = %v, want %v", got, want)
	}
}
