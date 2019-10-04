package pubsub

import (
	"fmt"
	"os"

	"cloud.google.com/go/pubsub"
)

// Option is the function signature required to be considered an pubsub.Option.
type Option func(*Transport) error

const (
	DefaultProjectEnvKey      = "GOOGLE_CLOUD_PROJECT"
	DefaultTopicEnvKey        = "PUBSUB_TOPIC"
	DefaultSubscriptionEnvKey = "PUBSUB_SUBSCRIPTION"
)

// WithEncoding sets the encoding for pubsub transport.
func WithEncoding(encoding Encoding) Option {
	return func(t *Transport) error {
		t.Encoding = encoding
		return nil
	}
}

// WithDefaultEncodingSelector sets the encoding selection strategy for
// default encoding selections based on Event.
func WithDefaultEncodingSelector(fn EncodingSelector) Option {
	return func(t *Transport) error {
		if t == nil {
			return fmt.Errorf("http default encoding selector option can not set nil transport")
		}
		if fn != nil {
			t.DefaultEncodingSelectionFn = fn
			return nil
		}
		return fmt.Errorf("pubsub fn for DefaultEncodingSelector was nil")
	}
}

// WithBinaryEncoding sets the encoding selection strategy for
// default encoding selections based on Event, the encoded event will be the
// given version in Binary form.
func WithBinaryEncoding() Option {
	return func(t *Transport) error {
		if t == nil {
			return fmt.Errorf("pubsub binary encoding option can not set nil transport")
		}

		t.DefaultEncodingSelectionFn = DefaultBinaryEncodingSelectionStrategy
		return nil
	}
}

// WithStructuredEncoding sets the encoding selection strategy for
// default encoding selections based on Event, the encoded event will be the
// given version in Structured form.
func WithStructuredEncoding() Option {
	return func(t *Transport) error {
		if t == nil {
			return fmt.Errorf("pubsub structured encoding option can not set nil transport")
		}

		t.DefaultEncodingSelectionFn = DefaultStructuredEncodingSelectionStrategy
		return nil
	}
}

// WithClient sets the pubsub client for pubsub transport. Use this for explicit
// auth setup. Otherwise the env var 'GOOGLE_APPLICATION_CREDENTIALS' is used.
// See https://cloud.google.com/docs/authentication/production for more details.
func WithClient(client *pubsub.Client) Option {
	return func(t *Transport) error {
		t.client = client
		return nil
	}
}

// WithProjectID sets the project ID for pubsub transport.
func WithProjectID(projectID string) Option {
	return func(t *Transport) error {
		t.projectID = projectID
		return nil
	}
}

// WithProjectIDFromEnv sets the project ID for pubsub transport from a
// given environment variable name.
func WithProjectIDFromEnv(key string) Option {
	return func(t *Transport) error {
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
	return func(t *Transport) error {
		t.topicID = topicID
		return nil
	}
}

// WithTopicIDFromEnv sets the topic ID for pubsub transport from a given
// environment variable name.
func WithTopicIDFromEnv(key string) Option {
	return func(t *Transport) error {
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
	return func(t *Transport) error {
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
	return func(t *Transport) error {
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

// WithSubscriptionIDFromEnv sets the subscription ID for pubsub transport from
// a given environment variable name.
func WithSubscriptionIDFromEnv(key string) Option {
	return func(t *Transport) error {
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

// AllowCreateTopic sets if the transport can create a topic if it does not
// exist.
func AllowCreateTopic(allow bool) Option {
	return func(t *Transport) error {
		t.AllowCreateTopic = allow
		return nil
	}
}

// AllowCreateSubscription sets if the transport can create a subscription if
// it does not exist.
func AllowCreateSubscription(allow bool) Option {
	return func(t *Transport) error {
		t.AllowCreateSubscription = allow
		return nil
	}
}
