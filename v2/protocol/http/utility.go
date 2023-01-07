/*
 Copyright 2022 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package http

import (
	"bytes"
	"context"
	"encoding/json"
	nethttp "net/http"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/event"
)

// NewEventFromHTTPRequest returns an Event.
func NewEventFromHTTPRequest(req *nethttp.Request) (*event.Event, error) {
	msg := NewMessageFromHttpRequest(req)
	return binding.ToEvent(context.Background(), msg)
}

// NewEventFromHTTPResponse returns an Event.
func NewEventFromHTTPResponse(resp *nethttp.Response) (*event.Event, error) {
	msg := NewMessageFromHttpResponse(resp)
	return binding.ToEvent(context.Background(), msg)
}

// NewEventsFromHTTPRequest returns a batched set of Events from a http.Request
func NewEventsFromHTTPRequest(req *nethttp.Request) ([]event.Event, error) {
	msg := NewMessageFromHttpRequest(req)
	return binding.ToEvents(context.Background(), msg, msg.BodyReader)
}

// NewEventsFromHTTPResponse returns a batched set of Events from a http.Response
func NewEventsFromHTTPResponse(resp *nethttp.Response) ([]event.Event, error) {
	msg := NewMessageFromHttpResponse(resp)
	return binding.ToEvents(context.Background(), msg, msg.BodyReader)
}

// NewHTTPRequestFromEvents creates a http.Request object that can be used with any http.Client.
func NewHTTPRequestFromEvents(ctx context.Context, url string, events []event.Event) (*nethttp.Request, error) {
	// Sending batch events is quite straightforward, as there is only JSON format, so a simple implementation.
	for _, e := range events {
		if err := e.Validate(); err != nil {
			return nil, err
		}
	}
	var buffer bytes.Buffer
	err := json.NewEncoder(&buffer).Encode(events)
	if err != nil {
		return nil, err
	}

	request, err := nethttp.NewRequestWithContext(ctx, nethttp.MethodPost, url, &buffer)
	if err != nil {
		return nil, err
	}

	request.Header.Set(ContentType, event.ApplicationCloudEventsBatchJSON)

	return request, nil
}
