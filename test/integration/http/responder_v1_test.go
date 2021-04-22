/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package http

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/protocol"
)

func TestClientResponder_Empty(t *testing.T) {
	now := time.Now()

	template := func(statusCode int, wantResult protocol.Result) TapTest {
		return TapTest{
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
			result: cloudevents.NewHTTPResult(statusCode, "unit test %s", http.StatusText(statusCode)),
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
				Header:        map[string][]string{},
				Status:        fmt.Sprintf("%d %s", statusCode, http.StatusText(statusCode)),
				ContentLength: 0,
			},
			wantResult: wantResult,
		}
	}

	testCases := TapTestCases{
		// For 2xx, results should be ACK.
		"Responder v1.0 - 200": template(
			http.StatusOK,
			protocol.ResultACK,
		),
		"Responder v1.0 - 202": template(
			http.StatusAccepted,
			protocol.ResultACK,
		),
		"Responder v1.0 - 204": template(
			http.StatusNoContent,
			protocol.ResultACK,
		),
		// For 4xx/5xx, http results with status code should be returned.
		"Responder v1.0 - 400": template(
			http.StatusBadRequest,
			cloudevents.NewHTTPResult(http.StatusBadRequest, "unit test %s", http.StatusText(http.StatusBadRequest)),
		),
		"Responder v1.0 - 401": template(
			http.StatusUnauthorized,
			cloudevents.NewHTTPResult(http.StatusUnauthorized, "unit test %s", http.StatusText(http.StatusUnauthorized)),
		),
		"Responder v1.0 - 404": template(
			http.StatusNotFound,
			cloudevents.NewHTTPResult(http.StatusNotFound, "unit test %s", http.StatusText(http.StatusNotFound)),
		),
		"Responder v1.0 - 500": template(
			http.StatusInternalServerError,
			cloudevents.NewHTTPResult(http.StatusInternalServerError, "unit test %s", http.StatusText(http.StatusInternalServerError)),
		),
	}

	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			ClientLoopback(t, tc)
		})
	}
}

func TestClientResponder_Response(t *testing.T) {
	now := time.Now()

	template := func(statusCode int, wantResult protocol.Result) TapTest {
		tt := TapTest{
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
			result: cloudevents.NewHTTPResult(statusCode, "unit test %s", http.StatusText(statusCode)),
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
			resp: &cloudevents.Event{
				Context: cloudevents.EventContextV1{
					ID:              "321-CBA",
					Type:            "unit.test.client.response",
					Source:          *cloudevents.ParseURIRef("/unit/test/client"),
					DataContentType: cloudevents.StringOfApplicationJSON(),
				}.AsV1(),
				DataEncoded: toBytes(map[string]interface{}{"unittest": "response"}),
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
				Status:        fmt.Sprintf("%d %s", statusCode, http.StatusText(statusCode)),
				ContentLength: 23,
			},
			wantResult: wantResult,
		}
		tt.want = tt.resp

		// When status is NoContent, no payload will be received.
		// So unset the following fields.
		if statusCode == http.StatusNoContent {
			tt.want.DataEncoded = nil
			tt.asRecv.Body = ""
		}

		return tt
	}

	testCases := TapTestCases{
		// 2xx should receive responded event with ACK result.
		"Responder v1.0 - 200": template(
			http.StatusOK,
			protocol.ResultACK,
		),
		"Responder v1.0 - 202": template(
			http.StatusAccepted,
			protocol.ResultACK,
		),
		"Responder v1.0 - 204": template(
			http.StatusNoContent,
			protocol.ResultACK,
		),
		// 4xx/5xx should receive nil event and http results with status code.
		"Responder v1.0 - 400": template(
			http.StatusBadRequest,
			cloudevents.NewHTTPResult(http.StatusBadRequest, "unit test %s", http.StatusText(http.StatusBadRequest)),
		),
		"Responder v1.0 - 401": template(
			http.StatusUnauthorized,
			cloudevents.NewHTTPResult(http.StatusUnauthorized, "unit test %s", http.StatusText(http.StatusUnauthorized)),
		),
		"Responder v1.0 - 404": template(
			http.StatusNotFound,
			cloudevents.NewHTTPResult(http.StatusNotFound, "unit test %s", http.StatusText(http.StatusNotFound)),
		),
		"Responder v1.0 - 500": template(
			http.StatusInternalServerError,
			cloudevents.NewHTTPResult(http.StatusInternalServerError, "unit test %s", http.StatusText(http.StatusInternalServerError)),
		),
	}

	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			ClientLoopback(t, tc)
		})
	}
}
