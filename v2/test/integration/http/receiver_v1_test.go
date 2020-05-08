package http

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
)

func TestClientReceiver_Status_Codes(t *testing.T) {
	now := time.Now()

	testCases := ReceiverTapTestCases{
		"415 if the receiver is expecting an event but the received request doesn't contain an event": {
			now: now,
			request: func(url string) *http.Request {
				req, _ := http.NewRequest("POST", url, bytes.NewReader(toBytes(map[string]interface{}{"hello": "Francesco"})))
				req.Header.Set("content-type", "application/json")
				return req
			},
			asRecv: &TapValidation{
				Header:        map[string][]string{},
				Status:        fmt.Sprintf("%d %s", http.StatusUnsupportedMediaType, http.StatusText(http.StatusUnsupportedMediaType)),
				ContentLength: 0,
			},
			receiverFuncFactory: func(cancelFunc context.CancelFunc) interface{} {
				return func(event cloudevents.Event) {
					cancelFunc()
				}
			},
		},
		"400 if the receiver is expecting an event but the received request doesn't contain a valid event without spec version": {
			now: now,
			request: func(url string) *http.Request {
				req, _ := http.NewRequest("POST", url, bytes.NewReader(toBytes(map[string]interface{}{"hello": "Francesco"})))
				req.Header.Set("content-type", cloudevents.ApplicationCloudEventsJSON)
				return req
			},
			asRecv: &TapValidation{
				Header:        http.Header{"content-type": {"text/plain"}},
				Status:        fmt.Sprintf("%d %s", http.StatusBadRequest, http.StatusText(http.StatusBadRequest)),
				ContentLength: 0,
				Body:          "specversion: unknown : \"\"\n",
				BodyContains: []string{
					"specversion: unknown : \"\"",
				},
			},
			receiverFuncFactory: func(cancelFunc context.CancelFunc) interface{} {
				return func(event cloudevents.Event) {
					cancelFunc()
				}
			},
		},
		"400 if the receiver is expecting an event but the received request doesn't contain a valid event with spec version": {
			now: now,
			request: func(url string) *http.Request {
				req, _ := http.NewRequest("POST", url, bytes.NewReader(toBytes(map[string]interface{}{"specversion": "1.0"})))
				req.Header.Set("content-type", cloudevents.ApplicationCloudEventsJSON)
				return req
			},
			asRecv: &TapValidation{
				Header:        http.Header{"content-type": {"text/plain"}},
				Status:        fmt.Sprintf("%d %s", http.StatusBadRequest, http.StatusText(http.StatusBadRequest)),
				ContentLength: 0,
				BodyContains: []string{
					"type: MUST be a non-empty string",
					"id: MUST be a non-empty string",
					"source: REQUIRED",
				},
			},
			receiverFuncFactory: func(cancelFunc context.CancelFunc) interface{} {
				return func(event cloudevents.Event) {
					cancelFunc()
				}
			},
		},
		"200 if the receiver is not expecting an event and the received request doesn't contain an event": {
			now: now,
			request: func(url string) *http.Request {
				req, _ := http.NewRequest("POST", url, bytes.NewReader(toBytes(map[string]interface{}{"hello": "Francesco"})))
				req.Header.Set("content-type", "application/json")
				return req
			},
			asRecv: &TapValidation{
				Header:        map[string][]string{},
				Status:        fmt.Sprintf("%d %s", http.StatusOK, http.StatusText(http.StatusOK)),
				ContentLength: 0,
			},
			receiverFuncFactory: func(cancelFunc context.CancelFunc) interface{} {
				return func() *cloudevents.Event {
					defer cancelFunc()
					return nil
				}
			},
		},
	}

	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			ClientReceiver(t, tc)
		})
	}
}
