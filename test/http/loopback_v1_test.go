package http

import (
	"testing"
	"time"

	cloudevents "github.com/cloudevents/sdk-go"
)

func TestClientLoopback_binary_v1tov01(t *testing.T) {
	now := time.Now()

	testCases := TapTestCases{
		"Loopback v1 -> v0.1": {
			now: now,
			event: &cloudevents.Event{
				Context: cloudevents.EventContextV1{
					ID:              "ABC-123",
					Type:            "unit.test.client.sent",
					Source:          *cloudevents.ParseURIRef("/unit/test/client"),
					Subject:         strptr("resource"),
					DataContentType: cloudevents.StringOfApplicationJSON(),
				}.AsV1(),
				Data: map[string]string{"hello": "unittest"},
			},
			resp: &cloudevents.Event{
				Context: cloudevents.EventContextV01{
					EventID:     "321-CBA",
					EventType:   "unit.test.client.response",
					Source:      *cloudevents.ParseURLRef("/unit/test/client"),
					ContentType: cloudevents.StringOfApplicationJSON(),
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

func TestClientLoopback_binary_v1tov02(t *testing.T) {
	now := time.Now()

	testCases := TapTestCases{
		"Loopback v1.0 -> v0.2": {
			now: now,
			event: &cloudevents.Event{
				Context: cloudevents.EventContextV1{
					ID:              "ABC-123",
					Type:            "unit.test.client.sent",
					Source:          *cloudevents.ParseURIRef("/unit/test/client"),
					Subject:         strptr("resource"),
					DataContentType: cloudevents.StringOfApplicationJSON(),
				}.AsV1(),
				Data: map[string]string{"hello": "unittest"},
			},
			resp: &cloudevents.Event{
				Context: cloudevents.EventContextV02{
					ID:          "321-CBA",
					Type:        "unit.test.client.response",
					Source:      *cloudevents.ParseURLRef("/unit/test/client"),
					ContentType: cloudevents.StringOfApplicationJSON(),
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
				Data: map[string]string{"hello": "unittest"},
			},
			resp: &cloudevents.Event{
				Context: cloudevents.EventContextV03{
					ID:              "321-CBA",
					Type:            "unit.test.client.response",
					Source:          *cloudevents.ParseURLRef("/unit/test/client"),
					DataContentType: cloudevents.StringOfApplicationJSON(),
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
				Data: map[string]string{"hello": "unittest"},
			},
			resp: &cloudevents.Event{
				Context: cloudevents.EventContextV1{
					ID:              "321-CBA",
					Type:            "unit.test.client.response",
					Source:          *cloudevents.ParseURIRef("/unit/test/client"),
					DataContentType: cloudevents.StringOfApplicationJSON(),
				}.AsV1(),
				Data: map[string]string{"unittest": "response"},
			},
			want: &cloudevents.Event{
				Context: cloudevents.EventContextV1{
					ID:              "321-CBA",
					Type:            "unit.test.client.response",
					Time:            &cloudevents.Timestamp{Time: now},
					Source:          *cloudevents.ParseURIRef("/unit/test/client"),
					DataContentType: cloudevents.StringOfApplicationJSON(),
				}.AsV1(),
				Data: map[string]string{"unittest": "response"},
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
	}

	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			ClientLoopback(t, tc)
		})
	}
}

// TODO: this test does not work because the test looks inside event.Data.
//       base64 content is now in DataBase64.
//func TestClientLoopback_structured_base64_v1tov1(t *testing.T) {
//	now := time.Now()
//
//	b64 := func(obj interface{}) string {
//		data, err := json.Marshal(obj)
//		if err != nil {
//			t.Error(err)
//		}
//		buf := make([]byte, base64.StdEncoding.EncodedLen(len(data)))
//		base64.StdEncoding.Encode(buf, data)
//		return string(buf)
//	}
//
//	testCases := TapTestCases{
//		"Loopback Base64 v1.0 -> v1.0": {
//			now: now,
//			event: &cloudevents.Event{
//				Context: cloudevents.EventContextV1{
//					ID:     "ABC-123",
//					Type:   "unit.test.client.sent",
//					Source: *cloudevents.ParseURIRef("/unit/test/client"),
//				}.AsV1(),
//				DataBase64: b64(map[string]string{"hello": "unittest"}),
//			},
//			resp: &cloudevents.Event{
//				Context: cloudevents.EventContextV1{
//					ID:     "321-CBA",
//					Type:   "unit.test.client.response",
//					Source: *cloudevents.ParseURIRef("/unit/test/client"),
//				}.AsV1(),
//				DataBase64: b64(map[string]string{"unittest": "response"}),
//			},
//			want: &cloudevents.Event{
//				Context: cloudevents.EventContextV1{
//					ID:              "321-CBA",
//					Type:            "unit.test.client.response",
//					Time:            &cloudevents.Timestamp{Time: now},
//					Source:          *cloudevents.ParseURIRef("/unit/test/client"),
//					DataContentType: cloudevents.StringOfApplicationJSON(),
//				}.AsV1(),
//				DataBase64: b64(map[string]string{"unittest": "response"}),
//			},
//			asSent: &TapValidation{
//				Method: "POST",
//				URI:    "/",
//				Header: map[string][]string{
//					"content-type": {"application/cloudevents+json"},
//				},
//				Body: fmt.Sprintf(`{"data_base64":"eyJoZWxsbyI6InVuaXR0ZXN0In0=","datacontenttype":"application/json","id":"ABC-123","source":"/unit/test/client","specversion":"1.0","time":%q,"type":"unit.test.client.sent"}`, now.UTC().Format(time.RFC3339Nano)),
//			},
//			asRecv: &TapValidation{
//				Header: map[string][]string{
//					"content-type": {"application/cloudevents+json"},
//				},
//				Body:   fmt.Sprintf(`{"data_base64":"eyJ1bml0dGVzdCI6InJlc3BvbnNlIn0=","datacontenttype":"application/json","id":"321-CBA","source":"/unit/test/client","specversion":"1.0","time":%q,"type":"unit.test.client.response"}`, now.UTC().Format(time.RFC3339Nano)),
//				Status: "200 OK",
//			},
//		},
//	}
//
//	for n, tc := range testCases {
//		t.Run(n, func(t *testing.T) {
//			// Time and Base64 can change the length...
//			tc.asSent.ContentLength = int64(len(tc.asSent.Body))
//			tc.asRecv.ContentLength = int64(len(tc.asRecv.Body))
//
//			ClientLoopback(t, tc, cloudevents.WithStructuredEncoding())
//		})
//	}
//}
