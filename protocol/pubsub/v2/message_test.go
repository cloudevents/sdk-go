/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package pubsub

import (
	"context"
	"fmt"
	"testing"

	"cloud.google.com/go/pubsub/v2"
	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/cloudevents/sdk-go/v2/protocol"
)

func TestReadStructured(t *testing.T) {
	tests := []struct {
		name    string
		pm      *pubsub.Message
		wantErr error
	}{
		{
			name: "nil format",
			pm: &pubsub.Message{
				ID: "testid",
			},
			wantErr: binding.ErrNotStructured,
		},
		{
			name: "json format",
			pm: &pubsub.Message{
				ID:         "testid",
				Attributes: map[string]string{contentType: event.ApplicationCloudEventsJSON},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			msg := NewMessage(tc.pm)
			err := msg.ReadStructured(context.Background(), (*pubsubMessagePublisher)(tc.pm))
			if err != tc.wantErr {
				t.Errorf("Error unexpected. got: %v, want: %v", err, tc.wantErr)
			}
		})
	}
}

func TestFinish(t *testing.T) {
	tests := []struct {
		name    string
		pm      *pubsub.Message
		err     error
		wantErr bool
	}{
		{
			name: "return error",
			pm: &pubsub.Message{
				ID: "testid",
			},
			err:     fmt.Errorf("error"),
			wantErr: true,
		},
		{
			name: "result not acked",
			pm: &pubsub.Message{
				ID: "testid",
			},
			err:     protocol.NewReceipt(false, "error"),
			wantErr: true,
		},
		{
			name: "result acked",
			pm: &pubsub.Message{
				ID: "testid",
			},
			err:     protocol.NewReceipt(true, "error"),
			wantErr: true,
		},
		{
			name: "no errors",
			pm: &pubsub.Message{
				ID: "testid",
			},
			wantErr: false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			msg := NewMessage(tc.pm)
			err := msg.Finish(tc.err)
			if tc.wantErr {
				if err != tc.err {
					t.Errorf("Error mismatch. got: %v, want: %v", err, tc.err)
				}
			}
			if !tc.wantErr && err != nil {
				t.Errorf("Should not error but got: %v", err)
			}
		})
	}
}

func TestWritePubSubMessageNilAttributes(t *testing.T) {
	e := event.New()
	e.SetID("testid")
	e.SetSource("test/source")
	e.SetType("test.type")
	if err := e.SetData(event.ApplicationJSON, map[string]string{"hello": "world"}); err != nil {
		t.Fatalf("failed to set data: %v", err)
	}

	msg := (*binding.EventMessage)(&e)

	// A pubsub.Message with nil Attributes must not panic when the writer
	// assigns into the attribute map during binary encoding.
	pm := &pubsub.Message{ID: "testid"}
	if err := WritePubSubMessage(context.Background(), msg, pm); err != nil {
		t.Errorf("unexpected error writing pubsub message: %v", err)
	}
	if pm.Attributes == nil {
		t.Errorf("expected Attributes to be initialized, got nil")
	}
}
