package client

import (
	"context"
	"github.com/cloudevents/sdk-go/pkg/event"
	"github.com/cloudevents/sdk-go/pkg/observability"
	"go.opencensus.io/trace"
)

// New produces a new client with the provided transport object and applied
// client options.
func NewObserved(protocol interface{}, opts ...Option) (Client, error) {
	client, err := New(protocol, opts...)
	if err != nil {
		return nil, err
	}

	return &obsClient{client: client}, nil
}

type obsClient struct {
	client Client

	disableTracePropagation bool // TODO?
}

//
//func (c *ceClient) applyOptions(opts ...Option) error {
//	for _, fn := range opts {
//		if err := fn(c); err != nil {
//			return err
//		}
//	}
//	return nil
//}

// Send transmits the provided event on a preconfigured Protocol. Send returns
// an error if there was an an issue validating the outbound event or the
// transport returns an error.
func (c *obsClient) Send(ctx context.Context, e event.Event) error {
	ctx, r := observability.NewReporter(ctx, reportSend)
	ctx, span := trace.StartSpan(ctx, clientSpanName, trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()
	if span.IsRecordingEvents() {
		span.AddAttributes(eventTraceAttributes(e.Context)...)
	}

	err := c.client.Send(ctx, e)

	if err != nil {
		r.Error()
	} else {
		r.OK()
	}
	return err
}

func (c *obsClient) Request(ctx context.Context, e event.Event) (*event.Event, error) {
	ctx, r := observability.NewReporter(ctx, reportRequest)
	ctx, span := trace.StartSpan(ctx, clientSpanName, trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()
	if span.IsRecordingEvents() {
		span.AddAttributes(eventTraceAttributes(e.Context)...)
	}

	resp, err := c.client.Request(ctx, e)

	if err != nil {
		r.Error()
	} else {
		r.OK()
	}
	return resp, err
}

// StartReceiver sets up the given fn to handle Receive.
// See Client.StartReceiver for details. This is a blocking call.
func (c *obsClient) StartReceiver(ctx context.Context, fn interface{}) error {
	ctx, r := observability.NewReporter(ctx, reportStartReceiver)

	err := c.client.StartReceiver(ctx, fn)

	if err != nil {
		r.Error()
	} else {
		r.OK()
	}
	return err
}
