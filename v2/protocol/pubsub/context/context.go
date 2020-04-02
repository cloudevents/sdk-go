// Package context provides the pubsub ProtocolContext.
package context

import (
	"context"
	"strings"
	"time"

	"cloud.google.com/go/pubsub"
)

// ProtocolContext allows a Receiver to understand the context of a request.
type ProtocolContext struct {
	ID           string
	PublishTime  time.Time
	Project      string
	Topic        string
	Subscription string
	Method       string // push or pull
}

// NewProtocolContext creates a new ProtocolContext from a pubsub.Message.
func NewProtocolContext(project, topic, subscription, method string, msg *pubsub.Message) ProtocolContext {
	var tx *ProtocolContext
	if msg != nil {
		tx = &ProtocolContext{
			ID:           msg.ID,
			PublishTime:  msg.PublishTime,
			Project:      project,
			Topic:        topic,
			Subscription: subscription,
			Method:       method,
		}
	} else {
		tx = &ProtocolContext{}
	}
	return *tx
}

// String generates a pretty-printed version of the resource as a string.
func (tx ProtocolContext) String() string {
	b := strings.Builder{}

	b.WriteString("Transport Context,\n")

	if tx.ID != "" {
		b.WriteString("  ID: " + tx.ID + "\n")
	}
	if !tx.PublishTime.IsZero() {
		b.WriteString("  PublishTime: " + tx.PublishTime.String() + "\n")
	}

	if tx.Project != "" {
		b.WriteString("  Project: " + tx.Project + "\n")
	}

	if tx.Topic != "" {
		b.WriteString("  Topic: " + tx.Topic + "\n")
	}

	if tx.Subscription != "" {
		b.WriteString("  Subscription: " + tx.Subscription + "\n")
	}

	if tx.Method != "" {
		b.WriteString("  Method: " + tx.Method + "\n")
	}

	return b.String()
}

// Opaque key type used to store ProtocolContext
type protocolContextKeyType struct{}

var protocolContextKey = protocolContextKeyType{}

// WithProtocolContext return a context with the given ProtocolContext into the provided context object.
func WithProtocolContext(ctx context.Context, tcxt ProtocolContext) context.Context {
	return context.WithValue(ctx, protocolContextKey, tcxt)
}

// ProtocolContextFrom pulls a ProtocolContext out of a context. Always
// returns a non-nil object.
func ProtocolContextFrom(ctx context.Context) ProtocolContext {
	tctx := ctx.Value(protocolContextKey)
	if tctx != nil {
		if tx, ok := tctx.(ProtocolContext); ok {
			return tx
		}
		if tx, ok := tctx.(*ProtocolContext); ok {
			return *tx
		}
	}
	return ProtocolContext{}
}
