/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"context"
	"log"
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	otelObs "github.com/cloudevents/sdk-go/observability/opentelemetry/v2/client"
	"github.com/cloudevents/sdk-go/samples/http/otel-sender-receiver/instrumentation"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	ceclient "github.com/cloudevents/sdk-go/v2/client"
	"github.com/cloudevents/sdk-go/v2/protocol"
	cehttp "github.com/cloudevents/sdk-go/v2/protocol/http"
)

var tracer trace.Tracer

const (
	serviceName = "cloudevents-server"
)

func main() {
	shutdown := instrumentation.InitOTelSdk(serviceName)
	tracer = otel.Tracer(serviceName + "-main")
	defer shutdown()

	// create the cloudevents client instrumented with OpenTelemetry
	c, err := otelObs.NewClientHTTP([]cehttp.Option{}, []ceclient.Option{})
	if err != nil {
		log.Fatalf("failed to create client, %v", err)
	}

	log.Fatal(c.StartReceiver(context.Background(), handleReceivedEvent))
}

func handleReceivedEvent(ctx context.Context, event cloudevents.Event) protocol.Result {

	// Showcase injecting the incoming tracecontext into the event as a DistributedTraceExtension
	otelObs.InjectDistributedTracingExtension(ctx, event)

	// Showcase extracting the tracecontext from the event into a context in order to continue the trace.
	// This is useful for cases where events are read from a queue and no context is present.
	ctx = otelObs.ExtractDistributedTracingExtension(ctx, event)

	// manually start a span for this http request
	ctx, childSpan := tracer.Start(ctx, "externalHttpCall", trace.WithAttributes(attribute.String("id", "123")))
	defer childSpan.End()

	// manually creating a http client instrumented with OpenTelemetry to make an external request
	client := http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}

	req, _ := http.NewRequestWithContext(ctx, "GET", "https://cloudevents.io/", nil)

	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	_ = res.Body.Close()

	return nil
}
