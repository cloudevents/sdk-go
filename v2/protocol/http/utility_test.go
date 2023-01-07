/*
 Copyright 2022 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package http

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/cloudevents/sdk-go/v2/test"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestNewEventFromHttpRequest(t *testing.T) {
	tests := []struct {
		name     string
		encoding binding.Encoding
	}{{
		name:     "Structured encoding",
		encoding: binding.EncodingStructured,
	}, {
		name:     "Binary encoding",
		encoding: binding.EncodingBinary,
	}}

	for _, tt := range tests {
		test.EachEvent(t, test.Events(), func(t *testing.T, eventIn event.Event) {
			t.Run(tt.name, func(t *testing.T) {
				ctx := context.TODO()
				if tt.encoding == binding.EncodingStructured {
					ctx = binding.WithForceStructured(ctx)
				} else if tt.encoding == binding.EncodingBinary {
					ctx = binding.WithForceBinary(ctx)
				}

				req := httptest.NewRequest("POST", "http://localhost", nil)
				require.NoError(t, WriteRequest(ctx, (*binding.EventMessage)(&eventIn), req))

				got, err := NewEventFromHTTPRequest(req)
				require.NoError(t, err)
				test.AssertEvent(t, *got, test.IsValid())
			})
		})
	}
}

func TestNewEventFromHttpResponse(t *testing.T) {
	tests := []struct {
		name string
		resp *http.Response
	}{{
		name: "Structured encoding",
		resp: &http.Response{
			Header: http.Header{
				"Content-Type": {event.ApplicationCloudEventsJSON},
			},
			Body:          ioutil.NopCloser(bytes.NewReader([]byte(`{"data":"foo","datacontenttype":"application/json","id":"id","source":"source","specversion":"1.0","type":"type"}`))),
			ContentLength: 113,
		},
	}, {
		name: "Binary encoding",
		resp: &http.Response{
			Header: func() http.Header {
				h := http.Header{}
				h.Set("ce-specversion", "1.0")
				h.Set("ce-source", "unittest")
				h.Set("ce-type", "unittest")
				h.Set("ce-id", "unittest")
				h.Set("Content-Type", "application/json")
				return h
			}(),
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewEventFromHTTPResponse(tt.resp)
			require.NoError(t, err)
			test.AssertEvent(t, *got, test.IsValid())
		})
	}
}

func TestNewEventsFromHTTPRequest(t *testing.T) {
	req := httptest.NewRequest("POST", "http://localhost", bytes.NewReader([]byte(`[{"data":"foo","datacontenttype":"application/json","id":"id","source":"source","specversion":"1.0","type":"type"}]`)))
	req.Header.Set(ContentType, event.ApplicationCloudEventsBatchJSON)

	events, err := NewEventsFromHTTPRequest(req)
	require.NoError(t, err)
	require.Len(t, events, 1)
	test.AssertEvent(t, events[0], test.IsValid())
}

func TestNewEventsFromHTTPResponse(t *testing.T) {
	data := `[{"data":"foo","datacontenttype":"application/json","id":"id","source":"source","specversion":"1.0","type":"type"}]`
	resp := http.Response{
		Header: http.Header{
			"Content-Type": {event.ApplicationCloudEventsBatchJSON},
		},
		Body:          io.NopCloser(bytes.NewReader([]byte(data))),
		ContentLength: int64(len(data)),
	}
	events, err := NewEventsFromHTTPResponse(&resp)
	require.NoError(t, err)
	require.Len(t, events, 1)
	test.AssertEvent(t, events[0], test.IsValid())
}

func TestNewHTTPRequestFromEvents(t *testing.T) {
	var events []event.Event
	e := event.New()
	e.SetID(uuid.New().String())
	e.SetSource("example/uri")
	e.SetType("example.type")
	require.NoError(t, e.SetData(event.ApplicationJSON, map[string]string{"hello": "world"}))
	events = append(events, e.Clone())

	e.SetID(uuid.New().String())
	require.NoError(t, e.SetData(event.ApplicationJSON, map[string]string{"goodbye": "world"}))
	events = append(events, e)

	require.Len(t, events, 2)
	require.NotEqual(t, events[0], events[1])

	// echo back what we get, so we can compare events at either side.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(ContentType, r.Header.Get(ContentType))
		b, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		require.NoError(t, r.Body.Close())
		_, err = w.Write(b)
		require.NoError(t, err)
	}))
	defer ts.Close()

	req, err := NewHTTPRequestFromEvents(context.Background(), ts.URL, events)
	require.NoError(t, err)

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)

	result, err := NewEventsFromHTTPResponse(resp)
	require.NoError(t, err)

	require.Equal(t, events, result)
}
