/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"context"
	"log"

	otelObs "github.com/cloudevents/sdk-go/observability/opentelemetry/v2/client"
	"github.com/cloudevents/sdk-go/samples/http/otel-sender-receiver/instrumentation"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/client"
	cehttp "github.com/cloudevents/sdk-go/v2/protocol/http"
)

const (
	serviceName = "cloudevents-client"
)

func main() {
	shutdown := instrumentation.InitOTelSdk(serviceName)
	defer shutdown()

	// create the cloudevents client instrumented with OpenTelemetry
	c, err := otelObs.NewClientHTTP([]cehttp.Option{}, []client.Option{})
	if err != nil {
		log.Fatalf("failed to create client, %v", err)
	}

	event := cloudevents.NewEvent()
	event.SetSource("example/uri")
	event.SetType("example.type")
	event.SetData(cloudevents.ApplicationJSON, map[string]string{"hello": "world"})

	ctx := cloudevents.ContextWithTarget(context.Background(), "http://localhost:8080/")

	if result := c.Send(ctx, event); cloudevents.IsUndelivered(result) {
		log.Fatalf("failed to send, %v", result)
	}
}
