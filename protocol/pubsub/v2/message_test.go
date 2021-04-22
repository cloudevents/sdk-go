/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package pubsub

import (
	"context"
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
