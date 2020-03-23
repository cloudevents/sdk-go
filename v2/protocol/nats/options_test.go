package nats

import (
	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/binding/transformer"
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

func TestWithTransformer(t *testing.T) {
	type args struct {
		sender      *Sender
		transformer binding.TransformerFactory
	}
	type wants struct {
		err    error
		sender *Sender
	}
	tests := []struct {
		name  string
		args  args
		wants wants
	}{
		{
			name: "nested transformer factories",
			args: args{
				sender:      &Sender{},
				transformer: binding.TransformerFactories{transformer.SetUUID, transformer.AddTimeNow},
			},
			wants: wants{
				err: nil,
				sender: &Sender{
					Transformers: binding.TransformerFactories{
						binding.TransformerFactories{transformer.SetUUID, transformer.AddTimeNow},
					},
				},
			},
		},
		{
			name: "empty transformers",
			args: args{
				sender:      &Sender{},
				transformer: transformer.SetUUID,
			},
			wants: wants{
				err: nil,
				sender: &Sender{
					Transformers: binding.TransformerFactories{transformer.SetUUID},
				},
			},
		},
		{
			name: "no transformer",
			args: args{
				sender:      &Sender{},
				transformer: nil,
			},
			wants: wants{
				err: nil,
				sender: &Sender{
					Transformers: binding.TransformerFactories{nil},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := tt.args.sender.applyOptions(WithTransformer(tt.args.transformer))
			if gotErr != tt.wants.err {
				t.Errorf("applyOptions(WithTransformer()) = %v, want %v", gotErr, tt.wants.err)
			}

			if !reflect.DeepEqual(tt.args.sender, tt.wants.sender) {
				t.Errorf("p = %v, want %v", tt.args.sender, tt.wants.sender)
			}
		})
	}
}
