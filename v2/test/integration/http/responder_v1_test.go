package http

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
)

func TestClientResponder_Empty(t *testing.T) {
	now := time.Now()

	template := func(statusCode int) TapTest {
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
		}
	}

	testCases := TapTestCases{
		"Responder v1.0 - 200": template(http.StatusOK),
		"Responder v1.0 - 202": template(http.StatusAccepted),
		"Responder v1.0 - 204": template(http.StatusNoContent),
		"Responder v1.0 - 400": template(http.StatusBadRequest),
		"Responder v1.0 - 401": template(http.StatusUnauthorized),
		"Responder v1.0 - 404": template(http.StatusNotFound),
		"Responder v1.0 - 500": template(http.StatusInternalServerError),
	}

	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			ClientLoopback(t, tc)
		})
	}
}

func TestClientResponder_Response(t *testing.T) {
	now := time.Now()

	template := func(statusCode int) TapTest {
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
		}
	}

	testCases := TapTestCases{
		"Responder v1.0 - 200": template(http.StatusOK),
		"Responder v1.0 - 202": template(http.StatusAccepted),
		"Responder v1.0 - 204": template(http.StatusNoContent),
		"Responder v1.0 - 400": template(http.StatusBadRequest),
		"Responder v1.0 - 401": template(http.StatusUnauthorized),
		"Responder v1.0 - 404": template(http.StatusNotFound),
		"Responder v1.0 - 500": template(http.StatusInternalServerError),
	}

	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			ClientLoopback(t, tc)
		})
	}
}
