package client

import (
	"context"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/cloudevents/sdk-go/v2/extensions"
	"github.com/cloudevents/sdk-go/v2/observability"
	"github.com/cloudevents/sdk-go/v2/protocol"

	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
)

// ObservabilityService is an interface users can implement to record metrics, create tracing spans, and plug other observability tools in the Client
type ObservabilityService interface {
	// This is invoked before the user function is invoked.
	// The returned callback will be invoked after the user finishes to process the event with the eventual processing error
	RecordReceivedEvent(ctx context.Context, event event.Event) (context.Context, func(result protocol.Result))
	// This is invoked before the event is sent.
	// The returned callback will be invoked when the response is received
	RecordSendingEvent(ctx context.Context, event event.Event) (context.Context, func(result protocol.Result))

	// This is invoked before the event is requested.
	// The returned callback will be invoked when the response is received
	RecordRequestEvent(ctx context.Context, event event.Event) (context.Context, func(result protocol.Result, event *event.Event))
}

type noopObservabilityService struct{}

func (n noopObservabilityService) RecordReceivedEvent(ctx context.Context, event event.Event) (context.Context, func(result protocol.Result)) {
	return ctx, func(result protocol.Result) {}
}

func (n noopObservabilityService) RecordSendingEvent(ctx context.Context, event event.Event) (context.Context, func(result protocol.Result)) {
	return ctx, func(result protocol.Result) {}
}

func (n noopObservabilityService) RecordRequestEvent(ctx context.Context, e event.Event) (context.Context, func(result protocol.Result, event *event.Event)) {
	return ctx, func(result protocol.Result, event *event.Event) {}
}

// TODO remove all of this stuff below, unless some useful constants

var (
	// LatencyMs measures the latency in milliseconds for the CloudEvents
	// client methods.
	LatencyMs = stats.Float64("cloudevents.io/sdk-go/client/latency", "The latency in milliseconds for the CloudEvents client methods.", "ms")
)

var (
	// LatencyView is an OpenCensus view that shows client method latency.
	LatencyView = &view.View{
		Name:        "client/latency",
		Measure:     LatencyMs,
		Description: "The distribution of latency inside of client for CloudEvents.",
		Aggregation: view.Distribution(0, .01, .1, 1, 10, 100, 1000, 10000),
		TagKeys:     observability.LatencyTags(),
	}
)

type observed int32

// Adheres to Observable
var _ observability.Observable = observed(0)

const (
	specversionAttr     = "cloudevents.specversion"
	idAttr              = "cloudevents.id"
	typeAttr            = "cloudevents.type"
	sourceAttr          = "cloudevents.source"
	subjectAttr         = "cloudevents.subject"
	datacontenttypeAttr = "cloudevents.datacontenttype"

	reportSend observed = iota
	reportRequest
	reportStartReceiver
)

// MethodName implements Observable.MethodName
func (o observed) MethodName() string {
	switch o {
	case reportSend:
		return "send"
	case reportRequest:
		return "request"
	case reportStartReceiver:
		return "start_receiver"
	default:
		return "unknown"
	}
}

// LatencyMs implements Observable.LatencyMs
func (o observed) LatencyMs() *stats.Float64Measure {
	return LatencyMs
}

func EventTraceAttributes(e event.EventReader) []trace.Attribute {
	as := []trace.Attribute{
		trace.StringAttribute(specversionAttr, e.SpecVersion()),
		trace.StringAttribute(idAttr, e.ID()),
		trace.StringAttribute(typeAttr, e.Type()),
		trace.StringAttribute(sourceAttr, e.Source()),
	}
	if sub := e.Subject(); sub != "" {
		as = append(as, trace.StringAttribute(subjectAttr, sub))
	}
	if dct := e.DataContentType(); dct != "" {
		as = append(as, trace.StringAttribute(datacontenttypeAttr, dct))
	}
	return as
}

// TraceSpan returns context and trace.Span based on event. Caller must call span.End()
func TraceSpan(ctx context.Context, e event.Event) (context.Context, *trace.Span) {
	var span *trace.Span
	if ext, ok := extensions.GetDistributedTracingExtension(e); ok {
		ctx, span = ext.StartChildSpan(ctx, observability.ClientSpanName, trace.WithSpanKind(trace.SpanKindServer))
	}
	if span == nil {
		ctx, span = trace.StartSpan(ctx, observability.ClientSpanName, trace.WithSpanKind(trace.SpanKindServer))
	}
	if span.IsRecordingEvents() {
		span.AddAttributes(EventTraceAttributes(&e)...)
	}
	return ctx, span
}
