/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package pubsub

import (
	"context"

	"github.com/cloudevents/sdk-go/v2/binding"
)

type withCustomAttributes struct{}

func AttributesFrom(ctx context.Context) map[string]string {
	ctxVal := binding.GetOrDefaultFromCtx(ctx, withCustomAttributes{}, nil)
	if ctxVal == nil {
		return make(map[string]string, 0)
	}

	m := ctxVal.(map[string]string)

	// Since it is possible that we get the same map from one ctx multiple times
	// we need to make sure, that it is race-free to modify returned map.
	cp := make(map[string]string, len(m))
	for k, v := range m {
		cp[k] = v
	}
	return cp
}

// WithCustomAttributes sets Message Attributes without any CloudEvent logic.
// Note that this function is not intended for CloudEvent Extensions or any `ce-`-prefixed Attributes.
// For these please see `Event` and `Event.SetExtension`.
func WithCustomAttributes(ctx context.Context, attrs map[string]string) context.Context {
	if attrs != nil {
		// Since it is likely that the map gets used in another goroutine
		// ensure that modifying passed in map is race-free.
		cp := make(map[string]string, len(attrs))
		for k, v := range attrs {
			cp[k] = v
		}
		attrs = cp
	}
	return context.WithValue(ctx, withCustomAttributes{}, attrs)
}
