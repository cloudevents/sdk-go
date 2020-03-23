package stan

import (
	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/binding/transformer"
	"github.com/nats-io/stan.go"
	"reflect"
	"testing"
)

func TestWithQueueSubscriber(t *testing.T) {
	baseConsumer := Consumer{}
	type args struct {
		queue string
	}
	type wants struct {
		err      error
		consumer Consumer
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
				consumer: Consumer{
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
				consumer: Consumer{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotP := baseConsumer
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
	baseConsumer := Consumer{}
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
			gotC := baseConsumer
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

func TestWithTransformer(t *testing.T) {
	baseSender := Sender{}
	type args struct {
		transformer binding.TransformerFactory
	}
	type wants struct {
		err    error
		sender Sender
	}
	tests := []struct {
		name  string
		args  args
		wants wants
	}{
		{
			name: "nested transformer factories",
			args: args{
				transformer: binding.TransformerFactories{transformer.SetUUID, transformer.AddTimeNow},
			},
			wants: wants{
				err: nil,
				sender: Sender{
					Transformers: binding.TransformerFactories{
						binding.TransformerFactories{transformer.SetUUID, transformer.AddTimeNow},
					},
				},
			},
		},
		{
			name: "empty transformers",
			args: args{
				transformer: transformer.SetUUID,
			},
			wants: wants{
				err: nil,
				sender: Sender{
					Transformers: binding.TransformerFactories{transformer.SetUUID},
				},
			},
		},
		{
			name: "no transformer",
			args: args{
				transformer: nil,
			},
			wants: wants{
				err: nil,
				sender: Sender{
					Transformers: binding.TransformerFactories{nil},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotP := baseSender
			gotErr := gotP.applyOptions(WithTransformer(tt.args.transformer))
			if gotErr != tt.wants.err {
				t.Errorf("applyOptions(WithTransformer()) = %v, want %v", gotErr, tt.wants.err)
			}

			if !reflect.DeepEqual(gotP, tt.wants.sender) {
				t.Errorf("p = %v, want %v", gotP, tt.wants.sender)
			}
		})
	}
}

func TestWithUnsubscribeOnClose(t *testing.T) {
	baseConsumer := Consumer{}
	type args struct {
	}
	type wants struct {
		err      error
		consumer Consumer
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
				consumer: Consumer{
					UnsubscribeOnClose: true,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotP := baseConsumer
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
