/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package nats

import (
	"reflect"
	"testing"
)

func TestWithQueueSubscriber(t *testing.T) {
	type args struct {
		consumer *Consumer
		queue    string
	}
	type wants struct {
		err      error
		consumer *Consumer
	}
	tests := []struct {
		name  string
		args  args
		wants wants
	}{
		{
			name: "valid queue",
			args: args{
				consumer: &Consumer{},
				queue:    "my-queue",
			},
			wants: wants{
				err: nil,
				consumer: &Consumer{
					Subscriber: &QueueSubscriber{Queue: "my-queue"},
				},
			},
		},
		{
			name: "invalid queue",
			args: args{
				consumer: &Consumer{},
				queue:    "",
			},
			wants: wants{
				err:      ErrInvalidQueueName,
				consumer: &Consumer{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := tt.args.consumer.applyOptions(WithQueueSubscriber(tt.args.queue))
			if gotErr != tt.wants.err {
				t.Errorf("applyOptions(WithQueueSubscriber()) = %v, want %v", gotErr, tt.wants.err)
			}

			if !reflect.DeepEqual(tt.args.consumer, tt.wants.consumer) {
				t.Errorf("p = %v, want %v", tt.args.consumer, tt.wants.consumer)
			}
		})
	}
}
