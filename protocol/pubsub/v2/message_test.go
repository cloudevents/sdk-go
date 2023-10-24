/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package pubsub

import (
	"context"
	"fmt"
	"testing"

	"cloud.google.com/go/pubsub"
	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/event"
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
