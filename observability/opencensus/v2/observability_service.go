package v2

import (
	"context"

	"go.opencensus.io/trace"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/client"
	"github.com/cloudevents/sdk-go/v2/observability"

	"github.com/cloudevents/sdk-go/v2/protocol"
)

type opencensusObservabilityService struct{}

func (o opencensusObservabilityService) RecordReceivedMalformedEvent(ctx context.Context, err error) {
	ctx, r := NewReporter(ctx, reportReceive)
	r.Error()
}

func (o opencensusObservabilityService) RecordReceivedEvent(ctx context.Context, event cloudevents.Event) (context.Context, func(errOrResult error)) {
	ctx, r := NewReporter(ctx, reportReceive)
	return ctx, func(errOrResult error) {
		if protocol.IsACK(errOrResult) {
			r.OK()
		} else {
			r.Error()
		}
	}
}

func (o opencensusObservabilityService) RecordSendingEvent(ctx context.Context, event cloudevents.Event) (context.Context, func(errOrResult error)) {
	ctx, r := NewReporter(ctx, reportSend)
	ctx, span := trace.StartSpan(ctx, observability.ClientSpanName, trace.WithSpanKind(trace.SpanKindClient))
	if span.IsRecordingEvents() {
		span.AddAttributes(EventTraceAttributes(&event)...)
	}

	return ctx, func(errOrResult error) {
		span.End()
		if protocol.IsACK(errOrResult) {
			r.OK()
		} else {
			r.Error()
		}
	}
}

func (o opencensusObservabilityService) RecordRequestEvent(ctx context.Context, event cloudevents.Event) (context.Context, func(errOrResult error, event *cloudevents.Event)) {
	ctx, r := NewReporter(ctx, reportSend)
	ctx, span := trace.StartSpan(ctx, observability.ClientSpanName, trace.WithSpanKind(trace.SpanKindClient))
	if span.IsRecordingEvents() {
		span.AddAttributes(EventTraceAttributes(&event)...)
	}

	return ctx, func(errOrResult error, event *cloudevents.Event) {
		span.End()
		if protocol.IsACK(errOrResult) {
			r.OK()
		} else {
			r.Error()
		}
	}
}

func New() client.ObservabilityService {
	return opencensusObservabilityService{}
}
