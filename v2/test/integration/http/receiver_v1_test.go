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
				Status:        fmt.Sprintf("%d %s", 415, http.StatusText(415)),
				ContentLength: 0,
			},
			receiverFuncFactory: func(cancelFunc context.CancelFunc) interface{} {
				return func(event cloudevents.Event) {
					cancelFunc()
				}
			},
		},
		"400 if the receiver is expecting an event but the received request doesn't contain a valid event": {
			now: now,
			request: func(url string) *http.Request {
				req, _ := http.NewRequest("POST", url, bytes.NewReader(toBytes(map[string]interface{}{"hello": "Francesco"})))
				req.Header.Set("content-type", cloudevents.ApplicationJSON)
				return req
			},
			asRecv: &TapValidation{
				Header:        map[string][]string{},
				Status:        fmt.Sprintf("%d %s", 400, http.StatusText(400)),
				ContentLength: 0,
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
				Status:        fmt.Sprintf("%d %s", 200, http.StatusText(200)),
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
