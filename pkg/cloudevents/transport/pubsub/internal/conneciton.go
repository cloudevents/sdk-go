package internal

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	pscontext "github.com/cloudevents/sdk-go/pkg/cloudevents/transport/pubsub/context"
)

// Transport acts as both a pubsub topic and a pubsub subscription .
type Connection struct {
	// AllowCreateTopic controls if the transport can create a topic if it does
	// not exist.
	AllowCreateTopic bool

	// AllowCreateSubscription controls if the transport can create a
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
				err = fmt.Errorf("transport not allowed to create topic %q", c.TopicID)
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

func (c *Connection) DeleteTopic(ctx context.Context) error {
	if c.topicWasCreated {
		if err := c.topic.Delete(ctx); err != nil {
			return err
		}
		c.topic = nil
		c.topicWasCreated = false
		c.topicOnce = sync.Once{}
	}
	return errors.New("topic was not created by pubsub transport")
}

func (c *Connection) getOrCreateSubscription(ctx context.Context) (*pubsub.Subscription, error) {
	var err error
	c.subOnce.Do(func() {
		// Load the topic.
		var topic *pubsub.Topic
		topic, err = c.getOrCreateTopic(ctx)
		if err != nil {
			return
		}
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
				err = fmt.Errorf("transport not allowed to create subscription %q", c.SubscriptionID)
				return
			}
			// Create a new subscription to the previously created topic
			// with the given name.
			// TODO: allow to use push config + allow setting the SubscriptionConfig.
			sub, err = c.Client.CreateSubscription(ctx, c.SubscriptionID, pubsub.SubscriptionConfig{
				Topic:             topic,
				AckDeadline:       30 * time.Second,
				RetentionDuration: 25 * time.Hour,
			})
			if err != nil {
				_ = c.Client.Close()
				return
			}
			c.subWasCreated = true
		}
		// Success.
		c.sub = sub
	})
	if c.sub == nil {
		return nil, fmt.Errorf("unable to create sunscription %q, %v", c.SubscriptionID, err)
	}
	return c.sub, err
}

func (c *Connection) DeleteSubscription(ctx context.Context) error {
	if c.subWasCreated {
		if err := c.sub.Delete(ctx); err != nil {
			return err
		}
		c.sub = nil
		c.subWasCreated = false
		c.subOnce = sync.Once{}
	}
	return errors.New("subscription was not created by pubsub transport")
}

// Send
func (c *Connection) Publish(ctx context.Context, msg *pubsub.Message) (*cloudevents.Event, error) {
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
		log.Println("adding internal context.")
		ctx = pscontext.WithTransportContext(ctx, pscontext.NewTransportContext(c.ProjectID, c.TopicID, c.SubscriptionID, "pull", m))
		fn(ctx, m)
	})
}
