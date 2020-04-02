package context

import (
	"context"
	"net/url"

	"github.com/cloudevents/sdk-go/v2/types"
)

// Opaque key type used to store target
type targetKeyType struct{}

var targetKey = targetKeyType{}

// WithTarget returns back a new context with the given target. Target is intended to be transport dependent.
// For http transport, `target` should be a full URL and will be injected into the outbound http request.
func WithTarget(ctx context.Context, target string) context.Context {
	return context.WithValue(ctx, targetKey, target)
}

// TargetFrom looks in the given context and returns `target` as a parsed url if found and valid, otherwise nil.
func TargetFrom(ctx context.Context) *url.URL {
	c := ctx.Value(targetKey)
	if c != nil {
		if s, ok := c.(string); ok && s != "" {
			if target, err := url.Parse(s); err == nil {
				return target
			}
		}
	}
	return nil
}

// Opaque key type used to store topic
type topicKeyType struct{}

var topicKey = topicKeyType{}

// WithTopic returns back a new context with the given topic. Topic is intended to be transport dependent.
// For pubsub transport, `topic` should be a Pub/Sub Topic ID.
func WithTopic(ctx context.Context, topic string) context.Context {
	return context.WithValue(ctx, topicKey, topic)
}

// TopicFrom looks in the given context and returns `topic` as a string if found and valid, otherwise "".
func TopicFrom(ctx context.Context) string {
	c := ctx.Value(topicKey)
	if c != nil {
		if s, ok := c.(string); ok {
			return s
		}
	}
	return ""
}

// Opaque key type used to store backoff parameters
type retryKeyType struct{}

var retryKey = retryKeyType{}

// WithRetry returns back a new context with the retry backoff parameters.
func WithRetry(ctx context.Context, backoff types.Backoff) context.Context {
	return context.WithValue(ctx, retryKey, backoff)
}

// RetryFrom looks in the given context and returns the backoff parameters if found, otherwise nil
func RetryFrom(ctx context.Context) *types.Backoff {
	c := ctx.Value(retryKey)
	if c != nil {
		if s, ok := c.(types.Backoff); ok {
			return &s
		}
	}
	return nil
}
