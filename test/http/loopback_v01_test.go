package http

import (
	"testing"
	"time"

	"github.com/cloudevents/sdk-go"
)

func TestClientLoopback_binary_v01tov01(t *testing.T) {
	now := time.Now()

	testCases := TapTestCases{
		"Loopback v0.1 -> v0.1": {
			now: now,
			event: &cloudevents.Event{
				Context: cloudevents.EventContextV02{
					ID:     "ABC-123",
					Type:   "unit.test.client.sent",
					Source: *cloudevents.ParseURLRef("/unit/test/client"),
				}.AsV01(),
				Data: map[string]string{"hello": "unittest"},
			},
			resp: &cloudevents.Event{
				Context: cloudevents.EventContextV01{
					EventID:   "321-CBA",
					EventType: "unit.test.client.response",
					Source:    *cloudevents.ParseURLRef("/unit/test/client"),
				}.AsV01(),
				Data: map[string]string{"unittest": "response"},
			},
			want: &cloudevents.Event{
				Context: cloudevents.EventContextV01{
					EventID:     "321-CBA",
					EventType:   "unit.test.client.response",
					EventTime:   &cloudevents.Timestamp{Time: now},
					Source:      *cloudevents.ParseURLRef("/unit/test/client"),
					ContentType: cloudevents.StringOfApplicationJSON(),
				}.AsV01(),
				Data: map[string]string{"unittest": "response"},
			},
			asSent: &TapValidation{
				Method: "POST",
				URI:    "/",
				Header: map[string][]string{
					"ce-cloudeventsversion": {"0.1"},
					"ce-eventid":            {"ABC-123"},
					"ce-eventtime":          {now.UTC().Format(time.RFC3339Nano)},
					"ce-eventtype":          {"unit.test.client.sent"},
					"ce-source":             {"/unit/test/client"},
					"content-type":          {"application/json"},
				},
				Body:          `{"hello":"unittest"}`,
				ContentLength: 20,
			},
			asRecv: &TapValidation{
				Header: map[string][]string{
					"ce-cloudeventsversion": {"0.1"},
					"ce-eventid":            {"321-CBA"},
					"ce-eventtime":          {now.UTC().Format(time.RFC3339Nano)},
					"ce-eventtype":          {"unit.test.client.response"},
					"ce-source":             {"/unit/test/client"},
					"content-type":          {"application/json"},
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

func TestClientLoopback_binary_v01tov02(t *testing.T) {
	now := time.Now()

	testCases := TapTestCases{
		"Loopback v0.2 -> v0.2": {
			now: now,
			event: &cloudevents.Event{
				Context: cloudevents.EventContextV02{
					ID:     "ABC-123",
					Type:   "unit.test.client.sent",
					Source: *cloudevents.ParseURLRef("/unit/test/client"),
				}.AsV01(),
				Data: map[string]string{"hello": "unittest"},
			},
			resp: &cloudevents.Event{
				Context: cloudevents.EventContextV02{
					ID:     "321-CBA",
					Type:   "unit.test.client.response",
					Source: *cloudevents.ParseURLRef("/unit/test/client"),
				}.AsV02(),
				Data: map[string]string{"unittest": "response"},
			},
			want: &cloudevents.Event{
				Context: cloudevents.EventContextV02{
					ID:          "321-CBA",
					Type:        "unit.test.client.response",
					Time:        &cloudevents.Timestamp{Time: now},
					Source:      *cloudevents.ParseURLRef("/unit/test/client"),
					ContentType: cloudevents.StringOfApplicationJSON(),
				}.AsV02(),
				Data: map[string]string{"unittest": "response"},
			},
			asSent: &TapValidation{
				Method: "POST",
				URI:    "/",
				Header: map[string][]string{
					"ce-cloudeventsversion": {"0.1"},
					"ce-eventid":            {"ABC-123"},
					"ce-eventtime":          {now.UTC().Format(time.RFC3339Nano)},
					"ce-eventtype":          {"unit.test.client.sent"},
					"ce-source":             {"/unit/test/client"},
					"content-type":          {"application/json"},
				},
				Body:          `{"hello":"unittest"}`,
				ContentLength: 20,
			},
			asRecv: &TapValidation{
				Header: map[string][]string{
					"ce-specversion": {"0.2"},
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

func TestClientLoopback_binary_v01tov03(t *testing.T) {
	now := time.Now()

	testCases := TapTestCases{
		"Loopback v0.2 -> v0.3": {
			now: now,
			event: &cloudevents.Event{
				Context: cloudevents.EventContextV02{
					ID:     "ABC-123",
					Type:   "unit.test.client.sent",
					Source: *cloudevents.ParseURLRef("/unit/test/client"),
				}.AsV01(),
				Data: map[string]string{"hello": "unittest"},
			},
			resp: &cloudevents.Event{
				Context: cloudevents.EventContextV03{
					ID:     "321-CBA",
					Type:   "unit.test.client.response",
					Source: *cloudevents.ParseURLRef("/unit/test/client"),
				}.AsV03(),
				Data: map[string]string{"unittest": "response"},
			},
			want: &cloudevents.Event{
				Context: cloudevents.EventContextV03{
					ID:              "321-CBA",
					Type:            "unit.test.client.response",
					Time:            &cloudevents.Timestamp{Time: now},
					Source:          *cloudevents.ParseURLRef("/unit/test/client"),
					DataContentType: cloudevents.StringOfApplicationJSON(),
				}.AsV03(),
				Data: map[string]string{"unittest": "response"},
			},
			asSent: &TapValidation{
				Method: "POST",
				URI:    "/",
				Header: map[string][]string{
					"ce-cloudeventsversion": {"0.1"},
					"ce-eventid":            {"ABC-123"},
					"ce-eventtime":          {now.UTC().Format(time.RFC3339Nano)},
					"ce-eventtype":          {"unit.test.client.sent"},
					"ce-source":             {"/unit/test/client"},
					"content-type":          {"application/json"},
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
