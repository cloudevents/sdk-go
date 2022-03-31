/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package http

import (
	"fmt"
	"github.com/cloudevents/sdk-go/v2/client"
	"net/http"
	"testing"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/types"
)

func TestSenderReceiver_binary_v1(t *testing.T) {
	now := time.Now()

	testCases := DirectTapTestCases{
		"Binary v1.0": {
			now: now,
			event: &cloudevents.Event{
				Context: cloudevents.EventContextV1{
					ID:              "ABC-123",
					Type:            "unit.test.client.sent",
					Source:          *cloudevents.ParseURIRef("/unit/test/client"),
					Subject:         strptr("resource"),
					DataContentType: cloudevents.StringOfApplicationJSON(),
				}.AsV1(),
				DataEncoded: toBytes(map[string]interface{}{"hello": "unittest"}),
			},
			want: &cloudevents.Event{
				Context: cloudevents.EventContextV1{
					ID:              "ABC-123",
					Type:            "unit.test.client.sent",
					Time:            &cloudevents.Timestamp{Time: now},
					Source:          *cloudevents.ParseURIRef("/unit/test/client"),
					Subject:         strptr("resource"),
					DataContentType: cloudevents.StringOfApplicationJSON(),
				}.AsV1(),
				DataEncoded: toBytes(map[string]interface{}{"hello": "unittest"}),
			},
			asSent: &TapValidation{
				Method: "POST",
				URI:    "/",
				Header: map[string][]string{
					"ce-specversion": {"1.0"},
					"ce-id":          {"ABC-123"},
					"ce-time":        {now.UTC().Format(time.RFC3339Nano)},
					"ce-type":        {"unit.test.client.sent"},
					"ce-source":      {"/unit/test/client"},
					"ce-subject":     {"resource"},
					"content-type":   {"application/json"},
				},
				Body:          `{"hello":"unittest"}`,
				ContentLength: 20,
			},
		},
		"Binary v1.0 Sender Result Are NACK For Non-2XX": {
			now: now,
			event: &cloudevents.Event{
				Context: cloudevents.EventContextV1{
					ID:              "ABC-123",
					Type:            "unit.test.client.sent",
					Source:          *cloudevents.ParseURIRef("/unit/test/client"),
					Subject:         strptr("resource"),
					DataContentType: cloudevents.StringOfApplicationJSON(),
				}.AsV1(),
				DataEncoded: toBytes(map[string]interface{}{"hello": "unittest"}),
			},
			serverReturnedStatusCode: http.StatusInternalServerError,
			want:                     nil,
			wantResult:               cloudevents.ResultNACK,
			asSent: &TapValidation{
				Method: "POST",
				URI:    "/",
				Header: map[string][]string{
					"ce-specversion": {"1.0"},
					"ce-id":          {"ABC-123"},
					"ce-time":        {now.UTC().Format(time.RFC3339Nano)},
					"ce-type":        {"unit.test.client.sent"},
					"ce-source":      {"/unit/test/client"},
					"ce-subject":     {"resource"},
					"content-type":   {"application/json"},
				},
				Body:          `{"hello":"unittest"}`,
				ContentLength: 20,
			},
		},
	}

	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			ClientDirect(t, tc)
		})
	}
}

func TestSenderReceiver_structured_v1(t *testing.T) {
	now := time.Now()

	testCases := DirectTapTestCases{
		"Structured v1.0": {
			now: now,
			event: func() *cloudevents.Event {
				event := cloudevents.NewEvent(cloudevents.VersionV1)
				event.SetID("ABC-123")
				event.SetType("unit.test.client.sent")
				event.SetSource("/unit/test/client")
				event.SetSubject("resource")
				_ = event.SetData(cloudevents.ApplicationJSON, map[string]string{"hello": "unittest"})
				return &event
			}(),
			want: &cloudevents.Event{
				Context: cloudevents.EventContextV1{
					ID:              "ABC-123",
					Type:            "unit.test.client.sent",
					Time:            &cloudevents.Timestamp{Time: now},
					Source:          *cloudevents.ParseURIRef("/unit/test/client"),
					Subject:         strptr("resource"),
					DataContentType: cloudevents.StringOfApplicationJSON(),
				}.AsV1(),
				DataEncoded: toBytes(map[string]interface{}{"hello": "unittest"}),
			},
			asSent: &TapValidation{
				Method: "POST",
				URI:    "/",
				Header: map[string][]string{
					"content-type": {"application/cloudevents+json"},
				},
				Body:          fmt.Sprintf(`{"data":{"hello":"unittest"},"id":"ABC-123","source":"/unit/test/client","specversion":"1.0","subject":"resource","time":%q,"type":"unit.test.client.sent"}`, types.FormatTime(now.UTC())),
				ContentLength: 182,
			},
		},
	}

	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			ClientDirect(t, tc, client.WithForceStructured())
		})
	}
}

func TestSenderReceiver_data_base64_v1(t *testing.T) {

	now := time.Now()

	testCases := DirectTapTestCases{
		"Structured v1.0": {
			now: now,
			event: func() *cloudevents.Event {
				event := cloudevents.NewEvent(cloudevents.VersionV1)
				event.SetID("ABC-123")
				event.SetType("unit.test.client.sent")
				event.SetSource("/unit/test/client")
				event.SetSubject("resource")
				_ = event.SetData(cloudevents.TextPlain, []byte("hello: unittest"))
				return &event
			}(),
			want: &cloudevents.Event{
				Context: cloudevents.EventContextV1{
					ID:              "ABC-123",
					Type:            "unit.test.client.sent",
					Time:            &cloudevents.Timestamp{Time: now},
					Source:          *cloudevents.ParseURIRef("/unit/test/client"),
					Subject:         strptr("resource"),
					DataContentType: cloudevents.StringOfTextPlain(),
				}.AsV1(),
				DataEncoded: []byte("hello: unittest"),
			},
			asSent: &TapValidation{
				Method: "POST",
				URI:    "/",
				Header: map[string][]string{
					"content-type": {"application/cloudevents+json"},
				},
				Body:          fmt.Sprintf(`{"data_base64":"aGVsbG86IHVuaXR0ZXN0","id":"ABC-123","source":"/unit/test/client","specversion":"1.0","subject":"resource","time":%q,"type":"unit.test.client.sent"}`, now.UTC().Format(time.RFC3339Nano)),
				ContentLength: 191,
			},
		},
	}

	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			ClientDirect(t, tc, client.WithForceStructured())
		})
	}
}
