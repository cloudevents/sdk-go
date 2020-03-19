package internal

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/cloudevents/sdk-go/v2/binding"
	pscontext "github.com/cloudevents/sdk-go/v2/protocol/pubsub/context"
)

// Connection acts as either a pubsub topic or a pubsub subscription .
type Connection struct {
	// AllowCreateTopic controls if the protocol can create a topic if it does
	// not exist.
	AllowCreateTopic bool

	// AllowCreateSubscription controls if the protocol can create a
	// subscription if it does not exist.
	AllowCreateSubscription bool

	ProjectID string

	Client *pubsub.Client

	TopicID         string
	topic           *pubsub.Topic
	topicWasCreated bool
	topicOnce       sync.Once

	SubscriptionID string
	sub            *pubsub.Subscription
	subWasCreated  bool
	subOnce        sync.Once

	// ReceiveSettings is used to configure Pubsub pull subscription.
	ReceiveSettings *pubsub.ReceiveSettings

	// AckDeadline is Pub/Sub AckDeadline.
	// Default is 30 seconds.
	AckDeadline *time.Duration
	// RetentionDuration is Pub/Sub RetentionDuration.
	// Default is 25 hours.
	RetentionDuration *time.Duration
}

const (
	DefaultAckDeadline       = 30 * time.Second
	DefaultRetentionDuration = 25 * time.Hour
)

var DefaultReceiveSettings = pubsub.ReceiveSettings{
	// Pubsub default receive settings will fill in other values.
	// https://godoc.org/cloud.google.com/go/pubsub#Client.Subscription

	// Override the default number of goroutines.
	// This is a magical number now. This has shown throughput improvements empirically
	// by at least 10x (compared to the default value).
	NumGoroutines: 1000,
	Synchronous:   false,
}

func (c *Connection) getOrCreateTopic(ctx context.Context) (*pubsub.Topic, error) {
	var err error
	c.topicOnce.Do(func() {
		var ok bool
		// Load the topic.
		topic := c.Client.Topic(c.TopicID)
		ok, err = topic.Exists(ctx)
		if err != nil {
			return
		}
		// If the topic does not exist, create a new topic with the given name.
		if !ok {
			if !c.AllowCreateTopic {
				err = fmt.Errorf("protocol not allowed to create topic %q", c.TopicID)
				return
			}
			topic, err = c.Client.CreateTopic(ctx, c.TopicID)
			if err != nil {
				return
			}
			c.topicWasCreated = true
		}
		// Success.
		c.topic = topic
	})
	if c.topic == nil {
		return nil, fmt.Errorf("unable to create topic %q, %v", c.TopicID, err)
	}
	return c.topic, err
}

// DeleteTopic
func (c *Connection) DeleteTopic(ctx context.Context) error {
	if !c.topicWasCreated {
		return errors.New("topic was not created by pubsub protocol")
	}
	if err := c.topic.Delete(ctx); err != nil {
		return err
	}
	c.topic = nil
	c.topicWasCreated = false
	c.topicOnce = sync.Once{}
	return nil
}

func (c *Connection) getOrCreateSubscription(ctx context.Context) (*pubsub.Subscription, error) {
	var err error
	c.subOnce.Do(func() {
		// Load the subscription.
		var ok bool
		sub := c.Client.Subscription(c.SubscriptionID)
		ok, err = sub.Exists(ctx)
		if err != nil {
			return
		}
		// If subscription doesn't exist, create it.
		if !ok {
			if !c.AllowCreateSubscription {
				err = fmt.Errorf("protocol not allowed to create subscription %q", c.SubscriptionID)
				return
			}

			// Load the topic.
			var topic *pubsub.Topic
			topic, err = c.getOrCreateTopic(ctx)
			if err != nil {
				return
			}
			// Default the ack deadline and retention duration config.
			if c.AckDeadline == nil {
				ackDeadline := DefaultAckDeadline
				c.AckDeadline = &(ackDeadline)
			}
			if c.RetentionDuration == nil {
				retentionDuration := DefaultRetentionDuration
				c.RetentionDuration = &retentionDuration
			}

			// Create a new subscription to the previously created topic
			// with the given name.
			// TODO: allow to use push config + allow setting the SubscriptionConfig.
			sub, err = c.Client.CreateSubscription(ctx, c.SubscriptionID, pubsub.SubscriptionConfig{
				Topic:             topic,
				AckDeadline:       *c.AckDeadline,
				RetentionDuration: *c.RetentionDuration,
			})
			if err != nil {
				_ = c.Client.Close()
				return
			}
			if c.ReceiveSettings == nil {
				sub.ReceiveSettings = DefaultReceiveSettings
			} else {
				sub.ReceiveSettings = *c.ReceiveSettings
			}
			c.subWasCreated = true
		}
		// Success.
		c.sub = sub
	})
	if c.sub == nil {
		return nil, fmt.Errorf("unable to create subscription %q, %v", c.SubscriptionID, err)
	}
	return c.sub, err
}

// DeleteSubscription
func (c *Connection) DeleteSubscription(ctx context.Context) error {
	if !c.subWasCreated {
		return errors.New("subscription was not created by pubsub protocol")
	}
	if err := c.sub.Delete(ctx); err != nil {
		return err
	}
	c.sub = nil
	c.subWasCreated = false
	c.subOnce = sync.Once{}
	return nil
}

// Publish
func (c *Connection) Publish(ctx context.Context, msg *pubsub.Message) (*binding.Message, error) {
	topic, err := c.getOrCreateTopic(ctx)
	if err != nil {
		return nil, err
	}

	r := topic.Publish(ctx, msg)
	_, err = r.Get(ctx)
	return nil, err
}

// Start
// NOTE: This is a blocking call.
func (c *Connection) Receive(ctx context.Context, fn func(context.Context, *pubsub.Message)) error {
	sub, err := c.getOrCreateSubscription(ctx)
	if err != nil {
		return err
	}
	// Ok, ready to start pulling.
	return sub.Receive(ctx, func(ctx context.Context, m *pubsub.Message) {
		ctx = pscontext.WithProtocolContext(ctx, pscontext.NewProtocolContext(c.ProjectID, c.TopicID, c.SubscriptionID, "pull", m))
		fn(ctx, m)
	})
}
