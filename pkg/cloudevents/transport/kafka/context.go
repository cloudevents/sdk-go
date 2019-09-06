package kafka

import (
	"context"
	"github.com/confluentinc/confluent-kafka-go-dev/kafka"
	"strings"
	"time"
)

// TransportContext allows a Receiver to understand the context of a request.
type TransportContext struct {
	PublishTime time.Time
	Topic       string
	Method      string // push or pull
}

// NewTransportContext creates a new TransportContext from a kafka.Message.
func NewTransportContext(topic, method string, msg *kafka.Message) TransportContext {
	var tx *TransportContext
	if msg != nil {
		tx = &TransportContext{
			PublishTime: msg.Timestamp,
			Topic:       topic,
			Method:      method,
		}
	} else {
		tx = &TransportContext{}
	}
	return *tx
}

// String generates a pretty-printed version of the resource as a string.
func (tx TransportContext) String() string {
	b := strings.Builder{}

	b.WriteString("Transport Context,\n")

	if !tx.PublishTime.IsZero() {
		b.WriteString("  PublishTime: " + tx.PublishTime.String() + "\n")
	}

	if tx.Topic != "" {
		b.WriteString("  Topic: " + tx.Topic + "\n")
	}

	if tx.Method != "" {
		b.WriteString("  Method: " + tx.Method + "\n")
	}

	return b.String()
}

// Opaque key type used to store TransportContext
type transportContextKeyType struct{}

var transportContextKey = transportContextKeyType{}

// WithTransportContext return a context with the given TransportContext into the provided context object.
func WithTransportContext(ctx context.Context, tcxt TransportContext) context.Context {
	return context.WithValue(ctx, transportContextKey, tcxt)
}

// TransportContextFrom pulls a TransportContext out of a context. Always
// returns a non-nil object.
func TransportContextFrom(ctx context.Context) TransportContext {
	tctx := ctx.Value(transportContextKey)
	if tctx != nil {
		if tx, ok := tctx.(TransportContext); ok {
			return tx
		}
		if tx, ok := tctx.(*TransportContext); ok {
			return *tx
		}
	}
	return TransportContext{}
}
