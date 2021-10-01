/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package opentelemetry

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdkTrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"

	otelObs "github.com/cloudevents/sdk-go/observability/opentelemetry/v2/client"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/extensions"
)

var (
	traceparent = http.CanonicalHeaderKey("traceparent")
	tracestate  = http.CanonicalHeaderKey("tracestate")

	prop           = propagation.TraceContext{}
	eventTraceID   = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	eventSpanID    = "bbbbbbbbbbbbbbbb"
	distributedExt = extensions.DistributedTracingExtension{
		TraceParent: fmt.Sprintf("00-%s-%s-00", eventTraceID, eventSpanID),
		TraceState:  "key1=value1,key2=value2",
	}
)

func TestExtractContextWithTraceContext(t *testing.T) {
	type testcase struct {
		name   string
		event  cloudevents.Event
		header http.Header
		want   string
	}

	tests := []testcase{
		{
			name:  "tracecontext in the context is overwritten by the one from the event",
			event: createCloudEvent(distributedExt),
			header: http.Header{
				traceparent: []string{"00-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa-aaaaaaaaaaaaaaaa-00"},
			},
			want: "00-bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb-bbbbbbbbbbbbbbbb-00",
		},
		{
			name:  "context with tracecontext and event with invalid tracecontext",
			event: createCloudEventWithInvalidTraceParent(),
			header: http.Header{
				traceparent: []string{"00-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa-aaaaaaaaaaaaaaaa-00"},
			},
			want: "00-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa-aaaaaaaaaaaaaaaa-00",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Simulates a case of an auto-instrumented client where the context
			// has the incoming parent span
			incomingCtx := prop.Extract(context.Background(), propagation.HeaderCarrier(tc.header))

			// act
			newCtx := otelObs.ExtractDistributedTracingExtension(incomingCtx, tc.event)

			prop := propagation.TraceContext{}
			carrier := otelObs.NewCloudEventCarrier()
			prop.Inject(newCtx, carrier)

			// the newCtx contains the expected traceparent
			assert.Equal(t, tc.want, carrier.Extension.TraceParent)
		})
	}
}

func TestExtractContextWithoutTraceContext(t *testing.T) {
	type testcase struct {
		name   string
		event  cloudevents.Event
		header http.Header
	}
	_, _ = configureOtelTestSdk()
	tests := []testcase{
		{
			name:  "context without tracecontext",
			event: createCloudEvent(distributedExt),
		},
		{
			name:  "context with invalid tracecontext and event with valid tracecontext",
			event: createCloudEvent(distributedExt),
			header: http.Header{
				traceparent: []string{"00-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa-1-00"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			incomingCtx := context.Background()

			if tc.header != nil {
				incomingCtx = prop.Extract(incomingCtx, propagation.HeaderCarrier(tc.header))
			}

			// act
			newCtx := otelObs.ExtractDistributedTracingExtension(incomingCtx, tc.event)
			sc := trace.SpanContextFromContext(newCtx)

			// the new context should be different since it was enriched with the tracecontext from the event
			assert.NotEqual(t, trace.SpanContextFromContext(incomingCtx), sc)

			// make sure the IDs are as expected
			assert.Equal(t, eventTraceID, sc.TraceID().String())
			assert.Equal(t, eventSpanID, sc.SpanID().String())
			assert.Equal(t, distributedExt.TraceState, sc.TraceState().String())
		})
	}
}

func TestInjectDistributedTracingExtension(t *testing.T) {
	type testcase struct {
		name   string
		event  cloudevents.Event
		header http.Header
		want   extensions.DistributedTracingExtension
	}
	tests := []testcase{
		{
			name:  "inject tracecontext into event",
			event: createCloudEvent(extensions.DistributedTracingExtension{}),
			header: http.Header{
				traceparent: []string{"00-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa-aaaaaaaaaaaaaaaa-00"},
				tracestate:  []string{"key1=value1,key2=value2"},
			},
			want: extensions.DistributedTracingExtension{
				TraceParent: "00-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa-aaaaaaaaaaaaaaaa-00",
				TraceState:  "key1=value1,key2=value2",
			},
		},
		{
			name:  "overwrite tracecontext in the event from the context",
			event: createCloudEvent(distributedExt),
			header: http.Header{
				traceparent: []string{"00-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa-aaaaaaaaaaaaaaaa-00"},
				tracestate:  []string{"key1=value1,key2=value2,key3=value3"},
			},
			want: extensions.DistributedTracingExtension{
				TraceParent: "00-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa-aaaaaaaaaaaaaaaa-00",
				TraceState:  "key1=value1,key2=value2,key3=value3",
			},
		},
		{
			name:  "context without tracecontext",
			event: createCloudEvent(distributedExt),
			want:  distributedExt,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			ctx = prop.Extract(ctx, propagation.HeaderCarrier(tc.header))

			// act
			otelObs.InjectDistributedTracingExtension(ctx, tc.event)

			actual, ok := extensions.GetDistributedTracingExtension(tc.event)
			assert.True(t, ok)
			assert.Equal(t, tc.want, actual)
		})
	}

}

func createCloudEvent(distributedExt extensions.DistributedTracingExtension) cloudevents.Event {
	event := cloudevents.NewEvent()
	event.SetSource("example/uri")
	event.SetType("example.type")
	event.SetData(cloudevents.ApplicationJSON, map[string]string{"hello": "world"})

	if distributedExt.TraceParent != "" {
		distributedExt.AddTracingAttributes(&event)
	}

	return event
}

func createCloudEventWithInvalidTraceParent() cloudevents.Event {
	event := cloudevents.NewEvent()
	event.SetSource("example/uri")
	event.SetType("example.type")
	event.SetData(cloudevents.ApplicationJSON, map[string]string{"hello": "world"})

	// set directly to force an invalid value
	event.SetExtension(extensions.TraceParentExtension, 123)

	return event
}

func configureOtelTestSdk() (*tracetest.SpanRecorder, trace.Tracer) {
	sr := tracetest.NewSpanRecorder()
	provider := sdkTrace.NewTracerProvider(sdkTrace.WithSpanProcessor(sr), sdkTrace.WithSampler(sdkTrace.AlwaysSample()))
	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	return sr, provider.Tracer("test-tracer")
}
