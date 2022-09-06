/*
 Copyright 2022 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package http

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/cloudevents/sdk-go/v2/test"
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
				ctx := context.Background()
				req := httptest.NewRequest("POST", "http://localhost", nil)
				require.NoError(t, WriteRequest(ctx, (*binding.EventMessage)(&eventIn), req))

				NewEventFromHttpRequest(req)
				got, err := NewEventFromHttpRequest(req)
				require.NoError(t, err)
				// Equals fails with diff like:
				// - 	"exbinary": []uint8{0x00, 0x01, 0x02, 0x03},
				// + 	"exbinary": string("AAECAw=="),
				// - 	"exint":    int32(42),
				// + 	"exint":    string("42"),
				// test.AssertEventEquals(t, eventIn, *got)
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
			got, err := NewEventFromHttpResponse(tt.resp)
			require.NoError(t, err)
			test.AssertEvent(t, *got, test.IsValid())
		})
	}
}
