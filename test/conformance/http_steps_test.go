/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package conformance

import (
	"bufio"
	"context"
	"strings"

	nethttp "net/http"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/protocol/http"
	"github.com/cucumber/godog"
	messages "github.com/cucumber/messages-go/v10"
)

var request *nethttp.Request

func HTTPFeatureContext(s *godog.Suite) {
	s.BeforeScenario(func(message *messages.Pickle) {
		request = nil
	})

	s.Step(`^HTTP Protocol Binding is supported$`, func() error {
		return nil
	})

	s.Step(`^an HTTP request$`, func(rawRequest *messages.PickleStepArgument_PickleDocString) error {
		parsedRequest, err := nethttp.ReadRequest(bufio.NewReader(strings.NewReader(rawRequest.Content)))
		if err != nil {
			return err
		}

		request = parsedRequest

		return nil
	})

	s.Step(`^parsed as HTTP request$`, func() error {
		message := http.NewMessageFromHttpRequest(request)

		event, err := binding.ToEvent(context.TODO(), message)

		if err != nil {
			return err
		}

		currentEvent = event

		return err
	})
}
