package rocketmq

import (
	"context"

	rocketmq "github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
)

// The Subscriber interface allows us to configure how the subscription is created
type Subscriber interface {
	Subscribe(consumer rocketmq.PushConsumer, topic string, selector consumer.MessageSelector,
		f func(context.Context, ...*primitive.MessageExt) (consumer.ConsumeResult, error)) error
}

// RegularSubscriber creates regular subscriptions
type RegularSubscriber struct {
}

// Subscribe implements Subscriber.Subscribe
func (s *RegularSubscriber) Subscribe(pc rocketmq.PushConsumer, topic string, selector consumer.MessageSelector,
	f func(context.Context, ...*primitive.MessageExt) (consumer.ConsumeResult, error)) error {
	return pc.Subscribe(topic, selector, f)
}

var _ Subscriber = (*RegularSubscriber)(nil)
