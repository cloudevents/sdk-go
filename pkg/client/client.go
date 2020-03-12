package client

import (
	"context"
	"fmt"
	"sync"

	"github.com/cloudevents/sdk-go/pkg/event"
	"github.com/cloudevents/sdk-go/pkg/extensions"
	"github.com/cloudevents/sdk-go/pkg/observability"
	"github.com/cloudevents/sdk-go/pkg/transport"
	"github.com/cloudevents/sdk-go/pkg/transport/http"
	"go.opencensus.io/trace"
)

// Client interface defines the runtime contract the CloudEvents client supports.
type Client interface {
	// Send will transmit the given event over the client's configured transport.
	Send(ctx context.Context, event event.Event) error

	// Request will transmit the given event over the client's configured
	// transport and return any response event.
	Request(ctx context.Context, event event.Event) (*event.Event, error)

	// StartReceiver will register the provided function for callback on receipt
	// of a cloudevent. It will also start the underlying transport as it has
	// been configured.
	// This call is blocking.
	// Valid fn signatures are:
	// * func()
	// * func() error
	// * func(context.Context)
	// * func(context.Context) transport.Result
	// * func(event.Event)
	// * func(event.Event) transport.Result
	// * func(context.Context, event.Event)
	// * func(context.Context, event.Event) transport.Result
	// * func(event.Event) *event.Event
	// * func(event.Event) (*event.Event, transport.Result)
	// * func(context.Context, event.Event) *event.Event
	// * func(context.Context, event.Event) (*event.Event, transport.Result)
	StartReceiver(ctx context.Context, fn interface{}) error
}

// New produces a new client with the provided transport object and applied
// client options.
func New(t transport.Transport, opts ...Option) (Client, error) {
	c := &ceClient{
		transport: t,
	}
	if err := c.applyOptions(opts...); err != nil {
		return nil, err
	}
	t.SetDelivery(c)
	return c, nil
}

// NewDefault provides the good defaults for the common case using an HTTP
// Protocol client. The http transport has had WithBinaryEncoding http
// transport option applied to it. The client will always send Binary
// encoding but will inspect the outbound event context and match the version.
// The WithTimeNow, WithUUIDs and WithDataContentType("application/json")
// client options are also applied to the client, all outbound events will have
// a time and id set if not already present.
func NewDefault() (Client, error) {
	p, err := http.NewProtocol()
	if err != nil {
		return nil, err
	}
	t, err := http.New(p, http.WithEncoding(http.Binary))
	if err != nil {
		return nil, err
	}
	c, err := New(t, WithTimeNow(), WithUUIDs(), WithDataContentType(event.ApplicationJSON))
	if err != nil {
		return nil, err
	}
	return c, nil
}

type ceClient struct {
	transport transport.Transport
	fn        *receiverFn

	receiverMu        sync.Mutex
	eventDefaulterFns []EventDefaulter

	disableTracePropagation bool
}

// Send transmits the provided event on a preconfigured Protocol. Send returns
// an error if there was an an issue validating the outbound event or the
// transport returns an error.
func (c *ceClient) Send(ctx context.Context, event event.Event) error {
	ctx, r := observability.NewReporter(ctx, reportSend)

	ctx, span := trace.StartSpan(ctx, clientSpanName, trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()
	if span.IsRecordingEvents() {
		span.AddAttributes(eventTraceAttributes(event.Context)...)
	}

	err := c.obsSend(ctx, event)
	if err != nil {
		r.Error()
	} else {
		r.OK()
	}
	return err
}

func (c *ceClient) obsSend(ctx context.Context, event event.Event) error {
	// Confirm we have a transport set.
	if c.transport == nil {
		return fmt.Errorf("client not ready, transport not initialized")
	}
	// Apply the defaulter chain to the incoming event.
	if len(c.eventDefaulterFns) > 0 {
		for _, fn := range c.eventDefaulterFns {
			event = fn(ctx, event)
		}
	}

	// Set distributed tracing extension.
	if !c.disableTracePropagation {
		if span := trace.FromContext(ctx); span != nil {
			event.Context = event.Context.Clone()
			if err := extensions.FromSpanContext(span.SpanContext()).AddTracingAttributes(event.Context); err != nil {
				return fmt.Errorf("error setting distributed tracing extension: %w", err)
			}
		}
	}

	// Validate the event conforms to the CloudEvents Spec.
	if err := event.Validate(); err != nil {
		return err
	}
	// Send the event over the transport.
	return c.transport.Send(ctx, event)
}

// Request transmits the provided event on a preconfigured Protocol. Request
// returns a response event if there is a response or an error if there was an
// an issue validating the outbound event or the transport returns an error.
func (c *ceClient) Request(ctx context.Context, event event.Event) (*event.Event, error) {
	ctx, r := observability.NewReporter(ctx, reportSend)

	ctx, span := trace.StartSpan(ctx, clientSpanName, trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()
	if span.IsRecordingEvents() {
		span.AddAttributes(eventTraceAttributes(event.Context)...)
	}

	resp, err := c.obsRequest(ctx, event)
	if err != nil {
		r.Error()
	} else {
		r.OK()
	}
	return resp, err
}

func (c *ceClient) obsRequest(ctx context.Context, event event.Event) (*event.Event, error) {
	// Confirm we have a transport set.
	if c.transport == nil {
		return nil, fmt.Errorf("client not ready, transport not initialized")
	}
	// Apply the defaulter chain to the incoming event.
	if len(c.eventDefaulterFns) > 0 {
		for _, fn := range c.eventDefaulterFns {
			event = fn(ctx, event)
		}
	}

	// Set distributed tracing extension.
	if !c.disableTracePropagation {
		if span := trace.FromContext(ctx); span != nil {
			event.Context = event.Context.Clone()
			if err := extensions.FromSpanContext(span.SpanContext()).AddTracingAttributes(event.Context); err != nil {
				return nil, fmt.Errorf("error setting distributed tracing extension: %w", err)
			}
		}
	}

	// Validate the event conforms to the CloudEvents Spec.
	if err := event.Validate(); err != nil {
		return nil, err
	}
	// Send the event over the transport.
	return c.transport.Request(ctx, event)
}

// Delivery is called from from the transport on event delivery.
func (c *ceClient) Delivery(ctx context.Context, e event.Event) (*event.Event, transport.Result) {
	ctx, r := observability.NewReporter(ctx, reportReceive)

	var span *trace.Span
	if !c.transport.HasTracePropagation() {
		if ext, ok := extensions.GetDistributedTracingExtension(e); ok {
			ctx, span = ext.StartChildSpan(ctx, clientSpanName, trace.WithSpanKind(trace.SpanKindServer))
		}
	}
	if span == nil {
		ctx, span = trace.StartSpan(ctx, clientSpanName, trace.WithSpanKind(trace.SpanKindServer))
	}
	defer span.End()
	if span.IsRecordingEvents() {
		span.AddAttributes(eventTraceAttributes(e.Context)...)
	}

	resp, result := c.obsDelivery(ctx, e)
	if result != nil { // TODO: test result for Ack/Nack
		r.Error()
	} else {
		r.OK()
	}
	return resp, result
}

func (c *ceClient) obsDelivery(ctx context.Context, e event.Event) (*event.Event, transport.Result) {
	if c.fn != nil {
		resp, err := c.fn.invoke(ctx, e)

		// Apply the defaulter chain to the outgoing event.
		if err == nil && resp != nil && len(c.eventDefaulterFns) > 0 {
			for _, fn := range c.eventDefaulterFns {
				*resp = fn(ctx, *resp)
			}
			// Validate the event conforms to the CloudEvents Spec.
			if verr := resp.Validate(); verr != nil {
				return nil, fmt.Errorf("cloudevent validation failed on response event: %v, %w", verr, err)
			}
		}
		return resp, err
	}
	return nil, nil
}

// StartReceiver sets up the given fn to handle Receive.
// See Client.StartReceiver for details. This is a blocking call.
func (c *ceClient) StartReceiver(ctx context.Context, fn interface{}) error {
	c.receiverMu.Lock()
	defer c.receiverMu.Unlock()

	if c.transport == nil {
		return fmt.Errorf("client not ready, transport not initialized")
	}
	if c.fn != nil {
		return fmt.Errorf("client already has a receiver")
	}

	if fn, err := receiver(fn); err != nil {
		return err
	} else {
		c.fn = fn
	}

	defer func() {
		c.fn = nil
	}()

	return c.transport.StartReceiver(ctx)
}

func (c *ceClient) applyOptions(opts ...Option) error {
	for _, fn := range opts {
		if err := fn(c); err != nil {
			return err
		}
	}
	return nil
}
