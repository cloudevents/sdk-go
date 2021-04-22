/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package stan

import (
	"reflect"
	"testing"

	"github.com/nats-io/stan.go"
)

func TestWithQueueSubscriber(t *testing.T) {
	type args struct {
		queue string
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
				queue: "my-queue",
			},
			wants: wants{
				err: nil,
				consumer: &Consumer{
					Subscriber: &QueueSubscriber{QueueGroup: "my-queue"},
				},
			},
		},
		{
			name: "invalid queue",
			args: args{
				queue: "",
			},
			wants: wants{
				err:      ErrInvalidQueueName,
				consumer: &Consumer{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotP := &Consumer{}
			gotErr := gotP.applyOptions(WithQueueSubscriber(tt.args.queue))
			if gotErr != tt.wants.err {
				t.Errorf("applyOptions(WithQueueSubscriber()) = %v, want %v", gotErr, tt.wants.err)
			}

			if !reflect.DeepEqual(gotP, tt.wants.consumer) {
				t.Errorf("p = %v, want %v", gotP, tt.wants.consumer)
			}
		})
	}
}

func TestWithSubscriptionOptions(t *testing.T) {
	type args struct {
		stanSubscriptionOptions []stan.SubscriptionOption
	}
	type wants struct {
		err    error
		length int
	}
	tests := []struct {
		name  string
		args  args
		wants wants
	}{
		{
			name: "valid options",
			args: args{
				stanSubscriptionOptions: []stan.SubscriptionOption{
					stan.DurableName("my-durable-sub"),
				},
			},
			wants: wants{
				err:    nil,
				length: 1,
			},
		},
		{
			name: "no options",
			args: args{
				stanSubscriptionOptions: []stan.SubscriptionOption{},
			},
			wants: wants{
				err:    nil,
				length: 0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotC := &Consumer{}
			gotErr := gotC.applyOptions(WithSubscriptionOptions(tt.args.stanSubscriptionOptions...))
			if gotErr != tt.wants.err {
				t.Errorf("applyOptions(WithSubscriptionOptions()) = %v, want %v", gotErr, tt.wants.err)
			}

			if len(gotC.subscriptionOptions) != tt.wants.length {
				t.Errorf("len(p.subscriptionOptions) = %v, want %v", len(gotC.subscriptionOptions), tt.wants.length)
			}
		})
	}
}

func TestWithUnsubscribeOnClose(t *testing.T) {
	type args struct {
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
			name: "default",
			args: args{},
			wants: wants{
				err: nil,
				consumer: &Consumer{
					UnsubscribeOnClose: true,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotP := &Consumer{}
			gotErr := gotP.applyOptions(WithUnsubscribeOnClose())
			if gotErr != tt.wants.err {
				t.Errorf("applyOptions(WithUnsubscribeOnClose()) = %v, want %v", gotErr, tt.wants.err)
			}

			if !reflect.DeepEqual(gotP, tt.wants.consumer) {
				t.Errorf("p = %v, want %v", gotP, tt.wants.consumer)
			}
		})
	}
}
