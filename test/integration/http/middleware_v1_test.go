/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package http

import (
	"testing"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
)

func TestClientMiddleware_binary_v1(t *testing.T) {
	now := time.Now()

	testCases := TapTestCases{
		"Middleware v1.0 -> v1.0": {
			now: now,
			event: func() *cloudevents.Event {
				e := cloudevents.NewEvent(cloudevents.VersionV1)
				e.SetID("ABC-123")
				e.SetType("unit.test.client.sent")
				e.SetSource("/unit/test/client")
				e.SetSubject("resource")
				_ = e.SetData(cloudevents.ApplicationJSON, map[string]string{"hello": "unittest"})
				e.SetExtension("number", "4002909746823859279")
				return &e
			}(),
			want: func() *cloudevents.Event {
				e := cloudevents.NewEvent(cloudevents.VersionV1)
				e.SetID("ABC-123")
				e.SetType("unit.test.client.sent")
				e.SetTime(now)
				e.SetSource("/unit/test/client")
				e.SetSubject("resource")
				_ = e.SetData(cloudevents.ApplicationJSON, map[string]string{"hello": "unittest"})
				e.SetExtension("number", "4002909746823859279")
				return &e
			}(),
			asSent: &TapValidation{
				Method: "POST",
				URI:    "/",
				Header: map[string][]string{
					"ce-specversion": {"1.0"},
					"ce-id":          {"ABC-123"},
					"ce-number":      {"4002909746823859279"},
					"ce-time":        {now.UTC().Format(time.RFC3339Nano)},
					"ce-type":        {"unit.test.client.sent"},
					"ce-source":      {"/unit/test/client"},
					"ce-subject":     {"resource"},
					"content-type":   {"application/json"},
				},
				Body:          `{"hello":"unittest"}`,
				ContentLength: 20,
			},
			asRecv: &TapValidation{
				Header: map[string][]string{
					"ce-specversion": {"1.0"},
					"ce-id":          {"ABC-123"},
					"ce-number":      {"4002909746823859279"},
					"ce-time":        {now.UTC().Format(time.RFC3339Nano)},
					"ce-type":        {"unit.test.client.sent"},
					"ce-source":      {"/unit/test/client"},
					"ce-subject":     {"resource"},
					"content-type":   {"application/json"},
				},
				Body:          `{"hello":"unittest"}`,
				Status:        "200 OK",
				ContentLength: 20,
			},
		},
	}

	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			ClientMiddleware(t, tc)
		})
	}
}
