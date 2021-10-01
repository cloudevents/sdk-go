/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package instrumentation

import (
	"context"
	"log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

const (
	JaegerEndpoint = "http://localhost:14268/api/traces"
)

func InitOTelSdk(serviceName string) func() {
	ctx := context.Background()

	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(JaegerEndpoint)))
	if err != nil {
		return func() { log.Printf("Failed to create the trace exporter: %v", err) }
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
		)),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	return func() {
		if tp == nil {
			return
		}
		if err := tp.Shutdown(ctx); err != nil {
			log.Printf("Error shutting down the tracer provider: %v", err)
		}
	}
}
