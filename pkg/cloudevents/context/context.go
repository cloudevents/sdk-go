package context

import (
	"context"
	"net/url"
)

// Opaque key type used to store target
type targetKeyType struct{}

var targetKey = targetKeyType{}

func ContextWithTarget(ctx context.Context, target string) context.Context {
	return context.WithValue(ctx, targetKey, target)
}

func TargetFromContext(ctx context.Context) *url.URL {
	c := ctx.Value(targetKey)
	if c != nil {
		if target, err := url.Parse(c.(string)); err == nil {
			return target
		}
	}
	return nil
}
