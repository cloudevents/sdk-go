/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package opentelemetry

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"

	otelObs "github.com/cloudevents/sdk-go/observability/opentelemetry/v2/client"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	event "github.com/cloudevents/sdk-go/v2/event"
	"github.com/cloudevents/sdk-go/v2/extensions"
	"github.com/cloudevents/sdk-go/v2/protocol"
	"github.com/cloudevents/sdk-go/v2/protocol/http"
)

var (
	expectedEvent cloudevents.Event = createCloudEvent(extensions.DistributedTracingExtension{})
)

func TestRecordSendingEvent(t *testing.T) {
	tests := []struct {
		name             string
		expectedSpanName string
		expectedStatus   codes.Code
		expectedAttrs    []attribute.KeyValue
		expectedResult   protocol.Result
		expectedSpanKind trace.SpanKind
		nameFormatter    func(cloudevents.Event) string
		attributesGetter func(cloudevents.Event) []attribute.KeyValue
	}{

		{
			name:             "send with default options",
			expectedSpanName: "cloudevents.client.example.type send",
			expectedStatus:   codes.Unset,
			expectedAttrs:    otelObs.GetDefaultSpanAttributes(&expectedEvent, "RecordSendingEvent"),
			expectedSpanKind: trace.SpanKindProducer,
			nameFormatter:    nil,
		},
		{
			name:             "send with custom span name",
			expectedSpanName: "test.example.type send",
			expectedStatus:   codes.Unset,
			expectedAttrs:    otelObs.GetDefaultSpanAttributes(&expectedEvent, "RecordSendingEvent"),
			expectedSpanKind: trace.SpanKindProducer,
			nameFormatter: func(e cloudevents.Event) string {
				return "test." + e.Context.GetType()
			},
		},
		{
			name:             "send with custom attributes",
			expectedSpanName: "test.example.type send",
			expectedStatus:   codes.Unset,
			expectedAttrs:    append(otelObs.GetDefaultSpanAttributes(&expectedEvent, "RecordSendingEvent"), attribute.String("my-attr", "some-value")),
			expectedSpanKind: trace.SpanKindProducer,
			nameFormatter: func(e cloudevents.Event) string {
				return "test." + e.Context.GetType()
			},
			attributesGetter: func(cloudevents.Event) []attribute.KeyValue {
				return []attribute.KeyValue{
					attribute.String("my-attr", "some-value"),
				}
			},
		},
		{
			name:             "send with error response",
			expectedSpanName: "cloudevents.client.example.type send",
			expectedStatus:   codes.Unset,
			expectedAttrs:    otelObs.GetDefaultSpanAttributes(&expectedEvent, "RecordSendingEvent"),
			expectedSpanKind: trace.SpanKindProducer,
			expectedResult:   protocol.NewReceipt(false, "some error here"),
		},
		{
			name:             "send with http error response",
			expectedSpanName: "cloudevents.client.example.type send",
			expectedStatus:   codes.Error,
			expectedAttrs:    otelObs.GetDefaultSpanAttributes(&expectedEvent, "RecordSendingEvent"),
			expectedSpanKind: trace.SpanKindProducer,
			expectedResult:   http.NewResult(500, "some server error"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			sr, _ := configureOtelTestSdk()
			ctx := context.Background()

			os := otelObs.NewOTelObservabilityService(
				otelObs.WithSpanNameFormatter(tc.nameFormatter),
				otelObs.WithSpanAttributesGetter(tc.attributesGetter))

			// act
			ctx, cb := os.RecordSendingEvent(ctx, expectedEvent)
			cb(tc.expectedResult)

			spans := sr.Ended()

			// since the obs service started a span, the context should have the spancontext
			assert.NotNil(t, trace.SpanContextFromContext(ctx))
			assert.Equal(t, 1, len(spans))

			span := spans[0]
			assert.Equal(t, tc.expectedSpanName, span.Name())
			assert.Equal(t, tc.expectedStatus, span.Status().Code)
			assert.Equal(t, tc.expectedSpanKind, span.SpanKind())

			if !reflect.DeepEqual(span.Attributes(), tc.expectedAttrs) {
				t.Errorf("p = %v, want %v", span.Attributes(), tc.expectedAttrs)
			}

			if tc.expectedResult != nil {
				assert.Equal(t, 1, len(span.Events()))
				assert.Equal(t, semconv.ExceptionEventName, span.Events()[0].Name)

				attrsMap := getSpanEventMap(span.Events()[0].Attributes)
				assert.Equal(t, tc.expectedResult.Error(), attrsMap[string(semconv.ExceptionMessageKey)])
			}
		})
	}
}

func TestRecordRequestEvent(t *testing.T) {
	tests := []struct {
		name             string
		expectedSpanName string
		expectedStatus   codes.Code
		expectedAttrs    []attribute.KeyValue
		expectedResult   protocol.Result
		expectedSpanKind trace.SpanKind
		nameFormatter    func(cloudevents.Event) string
		attributesGetter func(cloudevents.Event) []attribute.KeyValue
	}{

		{
			name:             "request with default options",
			expectedSpanName: "cloudevents.client.example.type send",
			expectedStatus:   codes.Unset,
			expectedAttrs:    otelObs.GetDefaultSpanAttributes(&expectedEvent, "RecordRequestEvent"),
			expectedSpanKind: trace.SpanKindProducer,
			nameFormatter:    nil,
		},
		{
			name:             "request with custom span name",
			expectedSpanName: "test.example.type send",
			expectedStatus:   codes.Unset,
			expectedAttrs:    otelObs.GetDefaultSpanAttributes(&expectedEvent, "RecordRequestEvent"),
			expectedSpanKind: trace.SpanKindProducer,
			nameFormatter: func(e cloudevents.Event) string {
				return "test." + e.Context.GetType()
			},
		},
		{
			name:             "request with custom attributes",
			expectedSpanName: "test.example.type send",
			expectedStatus:   codes.Unset,
			expectedAttrs:    append(otelObs.GetDefaultSpanAttributes(&expectedEvent, "RecordRequestEvent"), attribute.String("my-attr", "some-value")),
			expectedSpanKind: trace.SpanKindProducer,
			nameFormatter: func(e cloudevents.Event) string {
				return "test." + e.Context.GetType()
			},
			attributesGetter: func(cloudevents.Event) []attribute.KeyValue {
				return []attribute.KeyValue{
					attribute.String("my-attr", "some-value"),
				}
			},
		},
		{
			name:             "send with error response",
			expectedSpanName: "cloudevents.client.example.type send",
			expectedStatus:   codes.Unset,
			expectedAttrs:    otelObs.GetDefaultSpanAttributes(&expectedEvent, "RecordRequestEvent"),
			expectedSpanKind: trace.SpanKindProducer,
			expectedResult:   protocol.NewReceipt(false, "some error here"),
		},
		{
			name:             "request with http error response",
			expectedSpanName: "cloudevents.client.example.type send",
			expectedStatus:   codes.Error,
			expectedAttrs:    otelObs.GetDefaultSpanAttributes(&expectedEvent, "RecordRequestEvent"),
			expectedSpanKind: trace.SpanKindProducer,
			expectedResult:   http.NewResult(500, "some server error"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			sr, _ := configureOtelTestSdk()
			ctx := context.Background()

			os := otelObs.NewOTelObservabilityService(
				otelObs.WithSpanNameFormatter(tc.nameFormatter),
				otelObs.WithSpanAttributesGetter(tc.attributesGetter))

			// act
			ctx, cb := os.RecordRequestEvent(ctx, expectedEvent)
			cb(tc.expectedResult, &expectedEvent)

			spans := sr.Ended()

			// since the obs service started a span, the context should have the spancontext
			assert.NotNil(t, trace.SpanContextFromContext(ctx))
			assert.Equal(t, 1, len(spans))

			span := spans[0]
			assert.Equal(t, tc.expectedSpanName, span.Name())
			assert.Equal(t, tc.expectedStatus, span.Status().Code)
			assert.Equal(t, tc.expectedSpanKind, span.SpanKind())

			if !reflect.DeepEqual(span.Attributes(), tc.expectedAttrs) {
				t.Errorf("p = %v, want %v", span.Attributes(), tc.expectedAttrs)
			}

			if tc.expectedResult != nil {
				assert.Equal(t, 1, len(span.Events()))
				assert.Equal(t, semconv.ExceptionEventName, span.Events()[0].Name)

				attrsMap := getSpanEventMap(span.Events()[0].Attributes)
				assert.Equal(t, tc.expectedResult.Error(), attrsMap[string(semconv.ExceptionMessageKey)])
			}
		})
	}
}

func TestRecordCallingInvoker(t *testing.T) {
	tests := []struct {
		name             string
		expectedSpanName string
		expectedStatus   codes.Code
		expectedAttrs    []attribute.KeyValue
		expectedResult   protocol.Result
		expectedSpanKind trace.SpanKind
		nameFormatter    func(cloudevents.Event) string
		attributesGetter func(cloudevents.Event) []attribute.KeyValue
	}{

		{
			name:             "invoker with default options",
			expectedSpanName: "cloudevents.client.example.type process",
			expectedStatus:   codes.Unset,
			expectedAttrs:    otelObs.GetDefaultSpanAttributes(&expectedEvent, "RecordCallingInvoker"),
			expectedSpanKind: trace.SpanKindConsumer,
			nameFormatter:    nil,
		},
		{
			name:             "invoker with custom span name",
			expectedSpanName: "test.example.type process",
			expectedStatus:   codes.Unset,
			expectedAttrs:    otelObs.GetDefaultSpanAttributes(&expectedEvent, "RecordCallingInvoker"),
			expectedSpanKind: trace.SpanKindConsumer,
			nameFormatter: func(e cloudevents.Event) string {
				return "test." + e.Context.GetType()
			},
		},
		{
			name:             "invoker with custom attributes",
			expectedSpanName: "test.example.type process",
			expectedStatus:   codes.Unset,
			expectedAttrs:    append(otelObs.GetDefaultSpanAttributes(&expectedEvent, "RecordCallingInvoker"), attribute.String("my-attr", "some-value")),
			expectedSpanKind: trace.SpanKindConsumer,
			nameFormatter: func(e cloudevents.Event) string {
				return "test." + e.Context.GetType()
			},
			attributesGetter: func(cloudevents.Event) []attribute.KeyValue {
				return []attribute.KeyValue{
					attribute.String("my-attr", "some-value"),
				}
			},
		},
		{
			name:             "invoker with error response",
			expectedSpanName: "cloudevents.client.example.type process",
			expectedStatus:   codes.Unset,
			expectedAttrs:    otelObs.GetDefaultSpanAttributes(&expectedEvent, "RecordCallingInvoker"),
			expectedSpanKind: trace.SpanKindConsumer,
			expectedResult:   protocol.NewReceipt(false, "some error here"),
		},
		{
			name:             "invoker with http error response",
			expectedSpanName: "cloudevents.client.example.type process",
			expectedStatus:   codes.Error,
			expectedAttrs:    otelObs.GetDefaultSpanAttributes(&expectedEvent, "RecordCallingInvoker"),
			expectedSpanKind: trace.SpanKindConsumer,
			expectedResult:   http.NewResult(500, "some server error"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			sr, _ := configureOtelTestSdk()
			ctx := context.Background()

			os := otelObs.NewOTelObservabilityService(
				otelObs.WithSpanNameFormatter(tc.nameFormatter),
				otelObs.WithSpanAttributesGetter(tc.attributesGetter))

			// act
			ctx, cb := os.RecordCallingInvoker(ctx, &expectedEvent)
			cb(tc.expectedResult)

			spans := sr.Ended()

			// since the obs service started a span, the context should have the spancontext
			assert.NotNil(t, trace.SpanContextFromContext(ctx))
			assert.Equal(t, 1, len(spans))

			span := spans[0]
			assert.Equal(t, tc.expectedSpanName, span.Name())
			assert.Equal(t, tc.expectedStatus, span.Status().Code)
			assert.Equal(t, tc.expectedSpanKind, span.SpanKind())

			if !reflect.DeepEqual(span.Attributes(), tc.expectedAttrs) {
				t.Errorf("p = %v, want %v", span.Attributes(), tc.expectedAttrs)
			}

			if tc.expectedResult != nil {
				assert.Equal(t, 1, len(span.Events()))
				assert.Equal(t, semconv.ExceptionEventName, span.Events()[0].Name)

				attrsMap := getSpanEventMap(span.Events()[0].Attributes)
				assert.Equal(t, tc.expectedResult.Error(), attrsMap[string(semconv.ExceptionMessageKey)])
			}
		})
	}
}

func TestRecordReceivedMalformedEvent(t *testing.T) {
	tests := []struct {
		name             string
		expectedSpanName string
		expectedStatus   codes.Code
		expectedAttrs    []attribute.KeyValue
		expectedResult   protocol.Result
		expectedSpanKind trace.SpanKind
	}{

		{
			name:             "received simple error",
			expectedSpanName: "cloudevents.client.malformed receive",
			expectedStatus:   codes.Unset,
			expectedAttrs: []attribute.KeyValue{
				attribute.String(string(semconv.CodeFunctionKey), "RecordReceivedMalformedEvent"),
			},
			expectedSpanKind: trace.SpanKindConsumer,
			expectedResult:   fmt.Errorf("unrecognized event version 0.1.1"),
		},
		{
			name:             "received validation error",
			expectedSpanName: "cloudevents.client.malformed receive",
			expectedStatus:   codes.Unset,
			expectedAttrs: []attribute.KeyValue{
				attribute.String(string(semconv.CodeFunctionKey), "RecordReceivedMalformedEvent"),
			},
			expectedSpanKind: trace.SpanKindConsumer,
			expectedResult:   event.ValidationError{"specversion": fmt.Errorf("missing Event.Context")},
		},
		{
			name:             "received http error",
			expectedSpanName: "cloudevents.client.malformed receive",
			expectedStatus:   codes.Error,
			expectedAttrs: []attribute.KeyValue{
				attribute.String(string(semconv.CodeFunctionKey), "RecordReceivedMalformedEvent"),
			},
			expectedSpanKind: trace.SpanKindConsumer,
			expectedResult:   http.NewResult(400, "malformed event"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			sr, _ := configureOtelTestSdk()
			ctx := context.Background()

			os := otelObs.NewOTelObservabilityService()

			// act
			os.RecordReceivedMalformedEvent(ctx, tc.expectedResult)

			spans := sr.Ended()

			// since the obs service started a span, the context should have the spancontext
			assert.NotNil(t, trace.SpanContextFromContext(ctx))
			assert.Equal(t, 1, len(spans))

			span := spans[0]
			assert.Equal(t, tc.expectedSpanName, span.Name())
			assert.Equal(t, tc.expectedStatus, span.Status().Code)
			assert.Equal(t, tc.expectedSpanKind, span.SpanKind())

			if !reflect.DeepEqual(span.Attributes(), tc.expectedAttrs) {
				t.Errorf("p = %v, want %v", span.Attributes(), tc.expectedAttrs)
			}

			if tc.expectedResult != nil {
				assert.Equal(t, 1, len(span.Events()))
				assert.Equal(t, semconv.ExceptionEventName, span.Events()[0].Name)

				attrsMap := getSpanEventMap(span.Events()[0].Attributes)
				assert.Equal(t, tc.expectedResult.Error(), attrsMap[string(semconv.ExceptionMessageKey)])
			}
		})
	}
}

func getSpanEventMap(evtAttrs []attribute.KeyValue) map[string]string {
	attr := map[string]string{}
	for _, v := range evtAttrs {
		attr[string(v.Key)] = v.Value.AsString()
	}
	return attr
}
