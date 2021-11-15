/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

// Package internal provides the internal pubsub Connection type.
package internal

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"cloud.google.com/go/pubsub"
	pscontext "github.com/cloudevents/sdk-go/protocol/pubsub/v2/context"
	"github.com/cloudevents/sdk-go/v2/binding"
)

type topicInfo struct {
	topic      *pubsub.Topic
	wasCreated bool
	once       sync.Once
	err        error
}

type subInfo struct {
	sub        *pubsub.Subscription
	wasCreated bool
	once       sync.Once
	err        error
}

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

	TopicID   string
	topicInfo *topicInfo

	SubscriptionID string
	subInfo        *subInfo

	// Held when reading or writing topicInfo and subInfo. This is only
	// held while reading the pointer, the structure internally manage
	// their own internal concurrency.  This also controls
	// the update of AckDeadline and RetentionDuration if those are
	// nil on start.
	initLock sync.Mutex

	// ReceiveSettings is used to configure Pubsub pull subscription.
	ReceiveSettings *pubsub.ReceiveSettings

	// AckDeadline is Pub/Sub AckDeadline.
	// Default is 30 seconds.
	// This can only be set prior to first call of any function.
	AckDeadline *time.Duration
	// RetentionDuration is Pub/Sub RetentionDuration.
	// Default is 25 hours.
	// This can only be set prior to first call of any function.
	RetentionDuration *time.Duration
	// MessageOrdering enables message ordering for all topics and subscriptions.
	// This can only be set prior to first call of any function.
	MessageOrdering bool
	// Filter is an expression written in the Cloud Pub/Sub filter language. If
	// non-empty, then only `PubsubMessage`s whose `attributes` field matches the
	// filter are delivered on this subscription. If empty, then no messages are
	// filtered out. Cannot be changed after the subscription is created.
	// This can only be set prior to first call of any function.
	Filter string
}

const (
	DefaultAckDeadline       = 30 * time.Second
	DefaultRetentionDuration = 25 * time.Hour
)

var DefaultReceiveSettings = pubsub.ReceiveSettings{
	// Pubsub default receive settings will fill in other values.
	// https://godoc.org/cloud.google.com/go/pubsub#Client.Subscription

	Synchronous: false,
}

func (c *Connection) getOrCreateTopicInfo(ctx context.Context, getAlreadyOpenOnly bool) (*topicInfo, error) {
	// See if a topic has already been created or is in the process of being created.
	// If not, start creating one.
	c.initLock.Lock()
	ti := c.topicInfo
	if ti == nil && !getAlreadyOpenOnly {
		c.topicInfo = &topicInfo{}
		ti = c.topicInfo
	}
	c.initLock.Unlock()
	if ti == nil {
		return nil, fmt.Errorf("no already open topic")
	}

	// Make sure the topic structure is initialized at most once.
	ti.once.Do(func() {
		var ok bool
		// Load the topic.
		topic := c.Client.Topic(c.TopicID)
		ok, ti.err = topic.Exists(ctx)
		if ti.err != nil {
			return
		}
		// If the topic does not exist, create a new topic with the given name.
		if !ok {
			if !c.AllowCreateTopic {
				ti.err = fmt.Errorf("protocol not allowed to create topic %q", c.TopicID)
				return
			}
			topic, ti.err = c.Client.CreateTopic(ctx, c.TopicID)
			if ti.err != nil {
				return
			}
			ti.wasCreated = true
		}
		// Success.
		ti.topic = topic

		// EnableMessageOrdering is a runtime parameter only and not part of the topic
		// Pub/Sub configuration. The Pub/Sub SDK requires this to be set to accept Pub/Sub
		// messages with an ordering key set.
		ti.topic.EnableMessageOrdering = c.MessageOrdering
	})
	if ti.topic == nil {
		// Initialization failed, remove this attempt so that future callers
		// will try to initialize again.
		c.initLock.Lock()
		if c.topicInfo == ti {
			c.topicInfo = nil
		}
		c.initLock.Unlock()

		return nil, fmt.Errorf("unable to get or create topic %q, %v", c.TopicID, ti.err)
	}

	return ti, nil
}

func (c *Connection) getOrCreateTopic(ctx context.Context, getAlreadyOpenOnly bool) (*pubsub.Topic, error) {
	ti, err := c.getOrCreateTopicInfo(ctx, getAlreadyOpenOnly)
	if ti != nil {
		return ti.topic, nil
	} else {
		return nil, err
	}
}

// DeleteTopic deletes the connection's topic
func (c *Connection) DeleteTopic(ctx context.Context) error {
	ti, err := c.getOrCreateTopicInfo(ctx, true)

	if err != nil {
		return errors.New("topic not open")
	}
	if !ti.wasCreated {
		return errors.New("topic was not created by pubsub protocol")
	}
	if err := ti.topic.Delete(ctx); err != nil {
		return err
	}

	ti.topic.Stop()

	c.initLock.Lock()
	if ti == c.topicInfo {
		c.topicInfo = nil
	}
	c.initLock.Unlock()

	return nil
}

func (c *Connection) getOrCreateSubscriptionInfo(ctx context.Context, getAlreadyOpenOnly bool) (*subInfo, error) {
	c.initLock.Lock()
	// Default the ack deadline and retention duration config.
	// We only do this once.
	if c.AckDeadline == nil {
		ackDeadline := DefaultAckDeadline
		c.AckDeadline = &(ackDeadline)
	}
	if c.RetentionDuration == nil {
		retentionDuration := DefaultRetentionDuration
		c.RetentionDuration = &retentionDuration
	}
	// See if a subscription has already been created or is in the process of being created.
	// If not, start creating one.
	si := c.subInfo
	if si == nil && !getAlreadyOpenOnly {
		c.subInfo = &subInfo{}
		si = c.subInfo
	}
	c.initLock.Unlock()
	if si == nil {
		return nil, fmt.Errorf("no already open subscription")
	}

	// Make sure the subscription structure is initialized at most once.
	si.once.Do(func() {
		// Load the subscription.
		var ok bool
		sub := c.Client.Subscription(c.SubscriptionID)
		ok, si.err = sub.Exists(ctx)
		if si.err != nil {
			return
		}
		// If subscription doesn't exist, create it.
		if !ok {
			if !c.AllowCreateSubscription {
				si.err = fmt.Errorf("protocol not allowed to create subscription %q", c.SubscriptionID)
				return
			}

			// Load the topic.
			var topic *pubsub.Topic
			topic, si.err = c.getOrCreateTopic(ctx, false)
			if si.err != nil {
				return
			}

			// Create a new subscription to the previously created topic
			// with the given name.
			// TODO: allow to use push config + allow setting the SubscriptionConfig.
			sub, si.err = c.Client.CreateSubscription(ctx, c.SubscriptionID, pubsub.SubscriptionConfig{
				Topic:                 topic,
				AckDeadline:           *c.AckDeadline,
				RetentionDuration:     *c.RetentionDuration,
				EnableMessageOrdering: c.MessageOrdering,
				Filter:                c.Filter,
			})
			if si.err != nil {
				return
			}

			si.wasCreated = true
		}
		if c.ReceiveSettings == nil {
			sub.ReceiveSettings = DefaultReceiveSettings
		} else {
			sub.ReceiveSettings = *c.ReceiveSettings
		}
		// Success.
		si.sub = sub
	})
	if si.sub == nil {
		// Initialization failed, remove this attempt so that future callers
		// will try to initialize again.
		c.initLock.Lock()
		if c.subInfo == si {
			c.subInfo = nil
		}
		c.initLock.Unlock()
		return nil, fmt.Errorf("unable to create subscription %q, %v", c.SubscriptionID, si.err)
	}
	return si, nil
}

func (c *Connection) getOrCreateSubscription(ctx context.Context, getAlreadyOpenOnly bool) (*pubsub.Subscription, error) {
	si, err := c.getOrCreateSubscriptionInfo(ctx, getAlreadyOpenOnly)
	if si != nil {
		return si.sub, nil
	} else {
		return nil, err
	}
}

// DeleteSubscription delete's the connection's subscription
func (c *Connection) DeleteSubscription(ctx context.Context) error {
	si, err := c.getOrCreateSubscriptionInfo(ctx, true)

	if err != nil {
		return errors.New("subscription not open")
	}

	if !si.wasCreated {
		return errors.New("subscription was not created by pubsub protocol")
	}
	if err := si.sub.Delete(ctx); err != nil {
		return err
	}

	c.initLock.Lock()
	if si == c.subInfo {
		c.subInfo = nil
	}
	c.initLock.Unlock()

	return nil
}

// Publish publishes a message to the connection's topic
func (c *Connection) Publish(ctx context.Context, msg *pubsub.Message) (*binding.Message, error) {
	topic, err := c.getOrCreateTopic(ctx, false)
	if err != nil {
		return nil, err
	}

	r := topic.Publish(ctx, msg)
	_, err = r.Get(ctx)
	return nil, err
}

// Receive begins pulling messages.
// NOTE: This is a blocking call.
func (c *Connection) Receive(ctx context.Context, fn func(context.Context, *pubsub.Message)) error {
	sub, err := c.getOrCreateSubscription(ctx, false)
	if err != nil {
		return err
	}
	// Ok, ready to start pulling.
	return sub.Receive(ctx, func(ctx context.Context, m *pubsub.Message) {
		ctx = pscontext.WithProtocolContext(ctx, pscontext.NewProtocolContext(c.ProjectID, c.TopicID, c.SubscriptionID, "pull", m))
		fn(ctx, m)
	})
}
