/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package pubsub

import (
	"fmt"
	"os"

	"cloud.google.com/go/pubsub"
)

// Option is the function signature required to be considered an pubsub.Option.
type Option func(*Protocol) error

const (
	DefaultProjectEnvKey         = "GOOGLE_CLOUD_PROJECT"
	DefaultTopicEnvKey           = "PUBSUB_TOPIC"
	DefaultSubscriptionEnvKey    = "PUBSUB_SUBSCRIPTION"
	DefaultMessageOrderingEnvKey = "PUBSUB_MESSAGE_ORDERING"
)

// WithClient sets the pubsub client for pubsub transport. Use this for explicit
// auth setup. Otherwise the env var 'GOOGLE_APPLICATION_CREDENTIALS' is used.
// See https://cloud.google.com/docs/authentication/production for more details.
func WithClient(client *pubsub.Client) Option {
	return func(t *Protocol) error {
		t.client = client
		return nil
	}
}

// WithProjectID sets the project ID for pubsub transport.
func WithProjectID(projectID string) Option {
	return func(t *Protocol) error {
		t.projectID = projectID
		return nil
	}
}

// WithProjectIDFromEnv sets the project ID for pubsub transport from a
// given environment variable name.
func WithProjectIDFromEnv(key string) Option {
	return func(t *Protocol) error {
		v := os.Getenv(key)
		if v == "" {
			return fmt.Errorf("unable to load project id, %q environment variable not set", key)
		}
		t.projectID = v
		return nil
	}
}

// WithProjectIDFromDefaultEnv sets the project ID for pubsub transport from
// the environment variable named 'GOOGLE_CLOUD_PROJECT'.
func WithProjectIDFromDefaultEnv() Option {
	return WithProjectIDFromEnv(DefaultProjectEnvKey)
}

// WithTopicID sets the topic ID for pubsub transport.
func WithTopicID(topicID string) Option {
	return func(t *Protocol) error {
		t.topicID = topicID
		return nil
	}
}

// WithTopicIDFromEnv sets the topic ID for pubsub transport from a given
// environment variable name.
func WithTopicIDFromEnv(key string) Option {
	return func(t *Protocol) error {
		v := os.Getenv(key)
		if v == "" {
			return fmt.Errorf("unable to load topic id, %q environment variable not set", key)
		}
		t.topicID = v
		return nil
	}
}

// WithTopicIDFromDefaultEnv sets the topic ID for pubsub transport from the
// environment variable named 'PUBSUB_TOPIC'.
func WithTopicIDFromDefaultEnv() Option {
	return WithTopicIDFromEnv(DefaultTopicEnvKey)
}

// WithSubscriptionID sets the subscription ID for pubsub transport.
// This option can be used multiple times.
func WithSubscriptionID(subscriptionID string) Option {
	return func(t *Protocol) error {
		if t.subscriptions == nil {
			t.subscriptions = make([]subscriptionWithTopic, 0)
		}
		t.subscriptions = append(t.subscriptions, subscriptionWithTopic{
			subscriptionID: subscriptionID,
		})
		return nil
	}
}

// WithSubscriptionAndTopicID sets the subscription and topic IDs for pubsub transport.
// This option can be used multiple times.
func WithSubscriptionAndTopicID(subscriptionID, topicID string) Option {
	return func(t *Protocol) error {
		if t.subscriptions == nil {
			t.subscriptions = make([]subscriptionWithTopic, 0)
		}
		t.subscriptions = append(t.subscriptions, subscriptionWithTopic{
			subscriptionID: subscriptionID,
			topicID:        topicID,
		})
		return nil
	}
}

// WithSubscriptionIDAndFilter sets the subscription and topic IDs for pubsub transport.
// This option can be used multiple times.
func WithSubscriptionIDAndFilter(subscriptionID, filter string) Option {
	return func(t *Protocol) error {
		if t.subscriptions == nil {
			t.subscriptions = make([]subscriptionWithTopic, 0)
		}
		t.subscriptions = append(t.subscriptions, subscriptionWithTopic{
			subscriptionID: subscriptionID,
			filter:         filter,
		})
		return nil
	}
}

// WithSubscriptionTopicIDAndFilter sets the subscription with filter option and topic IDs for pubsub transport.
// This option can be used multiple times.
func WithSubscriptionTopicIDAndFilter(subscriptionID, topicID, filter string) Option {
	return func(t *Protocol) error {
		if t.subscriptions == nil {
			t.subscriptions = make([]subscriptionWithTopic, 0)
		}
		t.subscriptions = append(t.subscriptions, subscriptionWithTopic{
			subscriptionID: subscriptionID,
			topicID:        topicID,
			filter:         filter,
		})
		return nil
	}
}

// WithSubscriptionIDFromEnv sets the subscription ID for pubsub transport from
// a given environment variable name.
func WithSubscriptionIDFromEnv(key string) Option {
	return func(t *Protocol) error {
		v := os.Getenv(key)
		if v == "" {
			return fmt.Errorf("unable to load subscription id, %q environment variable not set", key)
		}

		opt := WithSubscriptionID(v)
		return opt(t)
	}
}

// WithSubscriptionIDFromDefaultEnv sets the subscription ID for pubsub
// transport from the environment variable named 'PUBSUB_SUBSCRIPTION'.
func WithSubscriptionIDFromDefaultEnv() Option {
	return WithSubscriptionIDFromEnv(DefaultSubscriptionEnvKey)
}

// WithFilter sets the subscription filter for pubsub transport.
func WithFilter(filter string) Option {
	return func(t *Protocol) error {
		if t.subscriptions == nil {
			t.subscriptions = make([]subscriptionWithTopic, 0)
		}
		t.subscriptions = append(t.subscriptions, subscriptionWithTopic{
			filter: filter,
		})
		return nil
	}
}

// WithFilterFromEnv sets the subscription filter for pubsub transport from
// a given environment variable name.
func WithFilterFromEnv(key string) Option {
	return func(t *Protocol) error {
		v := os.Getenv(key)
		if v == "" {
			return fmt.Errorf("unable to load subscription filter, %q environment variable not set", key)
		}

		opt := WithFilter(v)
		return opt(t)
	}
}

// AllowCreateTopic sets if the transport can create a topic if it does not
// exist.
func AllowCreateTopic(allow bool) Option {
	return func(t *Protocol) error {
		t.AllowCreateTopic = allow
		return nil
	}
}

// AllowCreateSubscription sets if the transport can create a subscription if
// it does not exist.
func AllowCreateSubscription(allow bool) Option {
	return func(t *Protocol) error {
		t.AllowCreateSubscription = allow
		return nil
	}
}

// WithReceiveSettings sets the Pubsub ReceiveSettings for pull subscriptions.
func WithReceiveSettings(rs *pubsub.ReceiveSettings) Option {
	return func(t *Protocol) error {
		t.ReceiveSettings = rs
		return nil
	}
}

// WithMessageOrdering enables message ordering for all topics and subscriptions.
func WithMessageOrdering() Option {
	return func(t *Protocol) error {
		t.MessageOrdering = true
		return nil
	}
}

// WithMessageOrderingFromEnv enables message ordering for all topics and
// subscriptions from a given environment variable name.
func WithMessageOrderingFromEnv(key string) Option {
	return func(t *Protocol) error {
		if v := os.Getenv(key); v != "" {
			t.MessageOrdering = true
		}
		return nil
	}
}

// WithMessageOrderingFromDefaultEnv enables message ordering for all topics and
// subscriptions from the environment variable named 'PUBSUB_MESSAGE_ORDERING'.
func WithMessageOrderingFromDefaultEnv() Option {
	return WithMessageOrderingFromEnv(DefaultMessageOrderingEnvKey)
}
