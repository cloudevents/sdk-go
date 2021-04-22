/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/cloudevents/sdk-go/v2/client"

	cloudevents "github.com/cloudevents/sdk-go/v2"
)

func TestClientLoopback_binary_v1tov03(t *testing.T) {
	now := time.Now()

	testCases := TapTestCases{
		"Loopback v1.0 -> v0.3": {
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
			resp: &cloudevents.Event{
				Context: cloudevents.EventContextV03{
					ID:              "321-CBA",
					Type:            "unit.test.client.response",
					Source:          *cloudevents.ParseURIRef("/unit/test/client"),
					DataContentType: cloudevents.StringOfApplicationJSON(),
				}.AsV03(),
				DataEncoded: toBytes(map[string]interface{}{"unittest": "response"}),
			},
			want: &cloudevents.Event{
				Context: cloudevents.EventContextV03{
					ID:              "321-CBA",
					Type:            "unit.test.client.response",
					Time:            &cloudevents.Timestamp{Time: now},
					Source:          *cloudevents.ParseURIRef("/unit/test/client"),
					DataContentType: cloudevents.StringOfApplicationJSON(),
				}.AsV03(),
				DataEncoded: toBytes(map[string]interface{}{"unittest": "response"}),
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
			asRecv: &TapValidation{
				Header: map[string][]string{
					"ce-specversion": {"0.3"},
					"ce-id":          {"321-CBA"},
					"ce-time":        {now.UTC().Format(time.RFC3339Nano)},
					"ce-type":        {"unit.test.client.response"},
					"ce-source":      {"/unit/test/client"},
					"content-type":   {"application/json"},
				},
				Body:          `{"unittest":"response"}`,
				Status:        "200 OK",
				ContentLength: 23,
			},
		},
	}

	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			ClientLoopback(t, tc)
		})
	}
}

func TestClientLoopback_binary_v1tov1(t *testing.T) {
	now := time.Now()

	testCases := TapTestCases{
		"Loopback v1.0 -> v1.0": {
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
			resp: &cloudevents.Event{
				Context: cloudevents.EventContextV1{
					ID:              "321-CBA",
					Type:            "unit.test.client.response",
					Source:          *cloudevents.ParseURIRef("/unit/test/client"),
					DataContentType: cloudevents.StringOfApplicationJSON(),
				}.AsV1(),
				DataEncoded: toBytes(map[string]interface{}{"unittest": "response"}),
			},
			want: &cloudevents.Event{
				Context: cloudevents.EventContextV1{
					ID:              "321-CBA",
					Type:            "unit.test.client.response",
					Time:            &cloudevents.Timestamp{Time: now},
					Source:          *cloudevents.ParseURIRef("/unit/test/client"),
					DataContentType: cloudevents.StringOfApplicationJSON(),
				}.AsV1(),
				DataEncoded: toBytes(map[string]interface{}{"unittest": "response"}),
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
			asRecv: &TapValidation{
				Header: map[string][]string{
					"ce-specversion": {"1.0"},
					"ce-id":          {"321-CBA"},
					"ce-time":        {now.UTC().Format(time.RFC3339Nano)},
					"ce-type":        {"unit.test.client.response"},
					"ce-source":      {"/unit/test/client"},
					"content-type":   {"application/json"},
				},
				Body:          `{"unittest":"response"}`,
				Status:        "200 OK",
				ContentLength: 23,
			},
		},
		"Loopback v1.0 with error": {
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
			result: cloudevents.NewHTTPResult(http.StatusForbidden, "unit test %s", http.StatusText(http.StatusForbidden)),
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
			asRecv: &TapValidation{
				Header: http.Header{},
				Status: "403 Forbidden",
			},
			wantResult: cloudevents.NewHTTPResult(http.StatusForbidden, "unit test %s", http.StatusText(http.StatusForbidden)),
		},
	}

	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			ClientLoopback(t, tc)
		})
	}
}

func TestClientLoopback_structured_base64_v03tov1(t *testing.T) {
	t.Skip("TODO: bindings does not yet support base64")
	now := time.Now()

	bytes := func(obj interface{}) []byte {
		data, err := json.Marshal(obj)
		if err != nil {
			t.Error(err)
		}
		return data
	}

	testCases := TapTestCases{
		"Loopback Base64 v0.3 -> v1.0": {
			now: now,
			event: &cloudevents.Event{
				Context: cloudevents.EventContextV03{
					ID:                  "ABC-123",
					Type:                "unit.test.client.sent",
					Source:              *cloudevents.ParseURIRef("/unit/test/client"),
					DataContentEncoding: cloudevents.StringOfBase64(),
					DataContentType:     cloudevents.StringOfApplicationJSON(),
				}.AsV03(),
				DataEncoded: toBytes(map[string]interface{}{"hello": "unittest"}),
			},
			resp: &cloudevents.Event{
				Context: cloudevents.EventContextV1{
					ID:     "321-CBA",
					Type:   "unit.test.client.response",
					Source: *cloudevents.ParseURIRef("/unit/test/client"),
				}.AsV1(),
				DataEncoded: bytes(map[string]string{"unittest": "response"}),
				DataBase64:  true,
			},
			want: &cloudevents.Event{
				Context: cloudevents.EventContextV1{
					ID:     "321-CBA",
					Type:   "unit.test.client.response",
					Time:   &cloudevents.Timestamp{Time: now},
					Source: *cloudevents.ParseURIRef("/unit/test/client"),
				}.AsV1(),
				DataEncoded: toBytes(map[string]interface{}{"unittest": "response"}),
			},
			asSent: &TapValidation{
				Method: "POST",
				URI:    "/",
				Header: map[string][]string{
					"content-type": {"application/cloudevents+json"},
				},
				Body: fmt.Sprintf(`{"data":"eyJoZWxsbyI6InVuaXR0ZXN0In0=","datacontentencoding":"base64","datacontenttype":"application/json","id":"ABC-123","source":"/unit/test/client","specversion":"0.3","time":%q,"type":"unit.test.client.sent"}`, now.UTC().Format(time.RFC3339Nano)),
			},
			asRecv: &TapValidation{
				Header: map[string][]string{
					"content-type": {"application/cloudevents+json"},
				},
				Body:   fmt.Sprintf(`{"data_base64":"eyJ1bml0dGVzdCI6InJlc3BvbnNlIn0=","id":"321-CBA","source":"/unit/test/client","specversion":"1.0","time":%q,"type":"unit.test.client.response"}`, now.UTC().Format(time.RFC3339Nano)),
				Status: "200 OK",
			},
		},
	}

	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			// Time and Base64 can change the length...
			tc.asSent.ContentLength = int64(len(tc.asSent.Body))
			tc.asRecv.ContentLength = int64(len(tc.asRecv.Body))

			ClientLoopback(t, tc, client.WithForceStructured())
		})
	}
}

func TestClientLoopback_structured_base64_v1tov1(t *testing.T) {
	t.Skip("TODO: bindings does not yet support base64")

	now := time.Now()

	testCases := TapTestCases{
		"Loopback Base64 v1.0 -> v1.0": {
			now: now,
			event: &cloudevents.Event{
				Context: cloudevents.EventContextV1{
					ID:     "ABC-123",
					Type:   "unit.test.client.sent",
					Source: *cloudevents.ParseURIRef("/unit/test/client"),
				}.AsV1(),
				DataEncoded: toBytes(map[string]interface{}{"hello": "unittest"}),
				DataBase64:  true,
			},
			resp: &cloudevents.Event{
				Context: cloudevents.EventContextV1{
					ID:     "321-CBA",
					Type:   "unit.test.client.response",
					Source: *cloudevents.ParseURIRef("/unit/test/client"),
				}.AsV1(),
				DataEncoded: toBytes(map[string]interface{}{"hello": "unittest"}),
				DataBase64:  true,
			},
			want: &cloudevents.Event{
				Context: cloudevents.EventContextV1{
					ID:     "321-CBA",
					Type:   "unit.test.client.response",
					Time:   &cloudevents.Timestamp{Time: now},
					Source: *cloudevents.ParseURIRef("/unit/test/client"),
				}.AsV1(),
				DataEncoded: toBytes(map[string]interface{}{"unittest": "response"}),
			},
			asSent: &TapValidation{
				Method: "POST",
				URI:    "/",
				Header: map[string][]string{
					"content-type": {"application/cloudevents+json"},
				},
				Body: fmt.Sprintf(`{"data_base64":"eyJoZWxsbyI6InVuaXR0ZXN0In0=","id":"ABC-123","source":"/unit/test/client","specversion":"1.0","time":%q,"type":"unit.test.client.sent"}`, now.UTC().Format(time.RFC3339Nano)),
			},
			asRecv: &TapValidation{
				Header: map[string][]string{
					"content-type": {"application/cloudevents+json"},
				},
				Body:   fmt.Sprintf(`{"data_base64":"eyJ1bml0dGVzdCI6InJlc3BvbnNlIn0=","id":"321-CBA","source":"/unit/test/client","specversion":"1.0","time":%q,"type":"unit.test.client.response"}`, now.UTC().Format(time.RFC3339Nano)),
				Status: "200 OK",
			},
		},
	}

	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			// Time and Base64 can change the length...
			tc.asSent.ContentLength = int64(len(tc.asSent.Body))
			tc.asRecv.ContentLength = int64(len(tc.asRecv.Body))

			ClientLoopback(t, tc, client.WithForceStructured())
		})
	}
}
