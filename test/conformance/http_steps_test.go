/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package conformance

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"

	nethttp "net/http"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/protocol/http"
	"github.com/cucumber/godog"
)

var request *nethttp.Request

func HTTPFeatureContext(ctx *godog.ScenarioContext) {
	ctx.BeforeScenario(func(*godog.Scenario) {
		request = nil
	})

	ctx.Step(`^HTTP Protocol Binding is supported$`, func() error {
		return nil
	})

	ctx.Step(`^an HTTP request$`, func(rawRequest *godog.DocString) error {
		// Create a proper HTTP request with body
		requestLines := strings.Split(rawRequest.Content, "\n")

		// Find the empty line that separates headers from body
		var headerEnd int
		for i, line := range requestLines {
			if line == "" {
				headerEnd = i
				break
			}
		}

		// Extract headers and body
		headerLines := strings.Join(requestLines[:headerEnd], "\r\n")
		var body string
		if headerEnd < len(requestLines)-1 {
			body = strings.Join(requestLines[headerEnd+1:], "\n")
		}

		// Create a complete HTTP request string
		completeRequest := headerLines + "\r\n\r\n" + body

		// Parse the HTTP request
		parsedRequest, err := nethttp.ReadRequest(bufio.NewReader(strings.NewReader(completeRequest)))
		if err != nil {
			return fmt.Errorf("failed to parse HTTP request: %v", err)
		}

		// Set the body manually to ensure it's properly set
		if body != "" {
			parsedRequest.Body = io.NopCloser(strings.NewReader(body))
			parsedRequest.ContentLength = int64(len(body))
		}

		request = parsedRequest
		return nil
	})

	ctx.Step(`^parsed as HTTP request$`, func() error {
		message := http.NewMessageFromHttpRequest(request)
		event, err := binding.ToEvent(context.TODO(), message)
		if err != nil {
			return fmt.Errorf("failed to convert to event: %v", err)
		}

		if event == nil {
			return fmt.Errorf("event is nil")
		}

		currentEvent = event
		return nil
	})
}
