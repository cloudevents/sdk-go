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
	return binding.GetOrDefaultFromCtx(ctx, withCustomAttributes{}, make(map[string]string)).(map[string]string)
}

// WithCustomAttributes sets Message Attributes without any CloudEvent logic.
// Note that this function is not intended for CloudEvent Extensions or any `ce-`-prefixed Attributes.
// For these please see `Event` and `Event.SetExtension`.
func WithCustomAttributes(ctx context.Context, attrs map[string]string) context.Context {
	return context.WithValue(ctx, withCustomAttributes{}, attrs)
}
