package http

import (
	"encoding/json"
	"fmt"
	"github.com/cloudevents/sdk-go"
	"testing"
	"time"
)

func TestClientLoopback_setters_binary_json(t *testing.T) {
	now := time.Now()

	versions := []string{cloudevents.VersionV01, cloudevents.VersionV02, cloudevents.VersionV03}

	testCases := map[string]struct {
		event  func(string) *cloudevents.Event
		resp   func(string) *cloudevents.Event
		want   map[string]*cloudevents.Event
		asSent map[string]*TapValidation
		asRecv map[string]*TapValidation
	}{
		"Loopback": {
			event: func(version string) *cloudevents.Event {
				event := cloudevents.NewEvent(version)
				event.SetID("ABC-123")
				event.SetType("unit.test.client.sent")
				event.SetSource("/unit/test/client")
				if err := event.SetData(map[string]string{"hello": "unittest"}); err != nil {
					t.Fatal(err)
				}
				return &event
			},
			resp: func(version string) *cloudevents.Event {
				event := cloudevents.NewEvent(version)
				event.SetID("321-CBA")
				event.SetType("unit.test.client.response")
				event.SetSource("/unit/test/client")
				if err := event.SetData(map[string]string{"unittest": "response"}); err != nil {
					t.Fatal(err)
				}
				return &event
			},
			want: map[string]*cloudevents.Event{
				cloudevents.VersionV01: {
					Context: cloudevents.EventContextV01{
						EventID:     "321-CBA",
						EventType:   "unit.test.client.response",
						EventTime:   &cloudevents.Timestamp{Time: now},
						Source:      *cloudevents.ParseURLRef("/unit/test/client"),
						ContentType: cloudevents.StringOfApplicationJSON(),
					}.AsV01(),
					Data: map[string]string{"unittest": "response"},
				},
				cloudevents.VersionV02: {
					Context: cloudevents.EventContextV02{
						ID:          "321-CBA",
						Type:        "unit.test.client.response",
						Time:        &cloudevents.Timestamp{Time: now},
						Source:      *cloudevents.ParseURLRef("/unit/test/client"),
						ContentType: cloudevents.StringOfApplicationJSON(),
					}.AsV02(),
					Data: map[string]string{"unittest": "response"},
				},
				cloudevents.VersionV03: {
					Context: cloudevents.EventContextV03{
						ID:              "321-CBA",
						Type:            "unit.test.client.response",
						Time:            &cloudevents.Timestamp{Time: now},
						Source:          *cloudevents.ParseURLRef("/unit/test/client"),
						DataContentType: cloudevents.StringOfApplicationJSON(),
					}.AsV03(),
					Data: map[string]string{"unittest": "response"},
				},
			},
			asSent: map[string]*TapValidation{
				cloudevents.VersionV01: {
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
				cloudevents.VersionV02: {
					Method: "POST",
					URI:    "/",
					Header: map[string][]string{
						"ce-specversion": {"0.2"},
						"ce-id":          {"ABC-123"},
						"ce-time":        {now.UTC().Format(time.RFC3339Nano)},
						"ce-type":        {"unit.test.client.sent"},
						"ce-source":      {"/unit/test/client"},
						"content-type":   {"application/json"},
					},
					Body:          `{"hello":"unittest"}`,
					ContentLength: 20,
				},
				cloudevents.VersionV03: {
					Method: "POST",
					URI:    "/",
					Header: map[string][]string{
						"ce-specversion": {"0.3"},
						"ce-id":          {"ABC-123"},
						"ce-time":        {now.UTC().Format(time.RFC3339Nano)},
						"ce-type":        {"unit.test.client.sent"},
						"ce-source":      {"/unit/test/client"},
						"content-type":   {"application/json"},
					},
					Body:          `{"hello":"unittest"}`,
					ContentLength: 20,
				},
			},
			asRecv: map[string]*TapValidation{
				cloudevents.VersionV01: {
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
				cloudevents.VersionV02: {
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
				cloudevents.VersionV03: {
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
		},
	}

	for n, tc := range testCases {
		for _, version := range versions {
			t.Run(n+version+" -> "+version, func(t *testing.T) {

				testcase := TapTest{
					now:    now,
					event:  tc.event(version),
					resp:   tc.resp(version),
					want:   tc.want[version],
					asSent: tc.asSent[version],
					asRecv: tc.asRecv[version],
				}
				ClientLoopback(t, testcase)
			})
		}
	}
}

func TestClientLoopback_setters_binary_json_noBody(t *testing.T) {
	now := time.Now()

	versions := []string{cloudevents.VersionV01, cloudevents.VersionV02, cloudevents.VersionV03}

	testCases := map[string]struct {
		event  func(string) *cloudevents.Event
		resp   func(string) *cloudevents.Event
		want   map[string]*cloudevents.Event
		asSent map[string]*TapValidation
		asRecv map[string]*TapValidation
	}{
		"Bodiless Loopback": {
			event: func(version string) *cloudevents.Event {
				event := cloudevents.NewEvent(version)
				event.SetID("ABC-123")
				event.SetType("unit.test.client.sent")
				event.SetSource("/unit/test/client")
				return &event
			},
			resp: func(version string) *cloudevents.Event {
				event := cloudevents.NewEvent(version)
				event.SetID("321-CBA")
				event.SetType("unit.test.client.response")
				event.SetSource("/unit/test/client")
				return &event
			},
			want: map[string]*cloudevents.Event{
				cloudevents.VersionV01: {
					Context: cloudevents.EventContextV01{
						EventID:     "321-CBA",
						EventType:   "unit.test.client.response",
						EventTime:   &cloudevents.Timestamp{Time: now},
						Source:      *cloudevents.ParseURLRef("/unit/test/client"),
						ContentType: cloudevents.StringOfApplicationJSON(),
					}.AsV01(),
					Data: map[string]string{},
				},
				cloudevents.VersionV02: {
					Context: cloudevents.EventContextV02{
						ID:          "321-CBA",
						Type:        "unit.test.client.response",
						Time:        &cloudevents.Timestamp{Time: now},
						Source:      *cloudevents.ParseURLRef("/unit/test/client"),
						ContentType: cloudevents.StringOfApplicationJSON(),
					}.AsV02(),
					Data: map[string]string{},
				},
				cloudevents.VersionV03: {
					Context: cloudevents.EventContextV03{
						ID:              "321-CBA",
						Type:            "unit.test.client.response",
						Time:            &cloudevents.Timestamp{Time: now},
						Source:          *cloudevents.ParseURLRef("/unit/test/client"),
						DataContentType: cloudevents.StringOfApplicationJSON(),
					}.AsV03(),
					Data: map[string]string{},
				},
			},
			asSent: map[string]*TapValidation{
				cloudevents.VersionV01: {
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
					ContentLength: 0,
				},
				cloudevents.VersionV02: {
					Method: "POST",
					URI:    "/",
					Header: map[string][]string{
						"ce-specversion": {"0.2"},
						"ce-id":          {"ABC-123"},
						"ce-time":        {now.UTC().Format(time.RFC3339Nano)},
						"ce-type":        {"unit.test.client.sent"},
						"ce-source":      {"/unit/test/client"},
						"content-type":   {"application/json"},
					},
					ContentLength: 0,
				},
				cloudevents.VersionV03: {
					Method: "POST",
					URI:    "/",
					Header: map[string][]string{
						"ce-specversion": {"0.3"},
						"ce-id":          {"ABC-123"},
						"ce-time":        {now.UTC().Format(time.RFC3339Nano)},
						"ce-type":        {"unit.test.client.sent"},
						"ce-source":      {"/unit/test/client"},
						"content-type":   {"application/json"},
					},
					ContentLength: 0,
				},
			},
			asRecv: map[string]*TapValidation{
				cloudevents.VersionV01: {
					Header: map[string][]string{
						"ce-cloudeventsversion": {"0.1"},
						"ce-eventid":            {"321-CBA"},
						"ce-eventtime":          {now.UTC().Format(time.RFC3339Nano)},
						"ce-eventtype":          {"unit.test.client.response"},
						"ce-source":             {"/unit/test/client"},
						"content-type":          {"application/json"},
					},
					Status:        "200 OK",
					ContentLength: 0,
				},
				cloudevents.VersionV02: {
					Header: map[string][]string{
						"ce-specversion": {"0.2"},
						"ce-id":          {"321-CBA"},
						"ce-time":        {now.UTC().Format(time.RFC3339Nano)},
						"ce-type":        {"unit.test.client.response"},
						"ce-source":      {"/unit/test/client"},
						"content-type":   {"application/json"},
					},
					Status:        "200 OK",
					ContentLength: 0,
				},
				cloudevents.VersionV03: {
					Header: map[string][]string{
						"ce-specversion": {"0.3"},
						"ce-id":          {"321-CBA"},
						"ce-time":        {now.UTC().Format(time.RFC3339Nano)},
						"ce-type":        {"unit.test.client.response"},
						"ce-source":      {"/unit/test/client"},
						"content-type":   {"application/json"},
					},
					Status:        "200 OK",
					ContentLength: 0,
				},
			},
		},
	}

	for n, tc := range testCases {
		for _, version := range versions {
			t.Run(n+version+" -> "+version, func(t *testing.T) {

				testcase := TapTest{
					now:    now,
					event:  tc.event(version),
					resp:   tc.resp(version),
					want:   tc.want[version],
					asSent: tc.asSent[version],
					asRecv: tc.asRecv[version],
				}
				ClientLoopback(t, testcase)
			})
		}
	}
}

func TestClientLoopback_setters_structured_json(t *testing.T) {
	now := time.Now()

	versions := []string{cloudevents.VersionV01, cloudevents.VersionV02, cloudevents.VersionV03}

	testCases := map[string]struct {
		event  func(string) *cloudevents.Event
		resp   func(string) *cloudevents.Event
		want   map[string]*cloudevents.Event
		asSent map[string]*TapValidation
		asRecv map[string]*TapValidation
	}{
		"Loopback": {
			event: func(version string) *cloudevents.Event {
				event := cloudevents.NewEvent(version)
				event.SetID("ABC-123")
				event.SetType("unit.test.client.sent")
				event.SetSource("/unit/test/client")
				if err := event.SetData(map[string]string{"hello": "unittest"}); err != nil {
					t.Fatal(err)
				}
				return &event
			},
			resp: func(version string) *cloudevents.Event {
				event := cloudevents.NewEvent(version)
				event.SetID("321-CBA")
				event.SetType("unit.test.client.response")
				event.SetSource("/unit/test/client")
				if err := event.SetData(map[string]string{"unittest": "response"}); err != nil {
					t.Fatal(err)
				}
				return &event
			},
			want: map[string]*cloudevents.Event{
				cloudevents.VersionV01: {
					Context: cloudevents.EventContextV01{
						EventID:     "321-CBA",
						EventType:   "unit.test.client.response",
						EventTime:   &cloudevents.Timestamp{Time: now},
						Source:      *cloudevents.ParseURLRef("/unit/test/client"),
						ContentType: cloudevents.StringOfApplicationJSON(),
					}.AsV01(),
					Data: map[string]string{"unittest": "response"},
				},
				cloudevents.VersionV02: {
					Context: cloudevents.EventContextV02{
						ID:          "321-CBA",
						Type:        "unit.test.client.response",
						Time:        &cloudevents.Timestamp{Time: now},
						Source:      *cloudevents.ParseURLRef("/unit/test/client"),
						ContentType: cloudevents.StringOfApplicationJSON(),
					}.AsV02(),
					Data: map[string]string{"unittest": "response"},
				},
				cloudevents.VersionV03: {
					Context: cloudevents.EventContextV03{
						ID:              "321-CBA",
						Type:            "unit.test.client.response",
						Time:            &cloudevents.Timestamp{Time: now},
						Source:          *cloudevents.ParseURLRef("/unit/test/client"),
						DataContentType: cloudevents.StringOfApplicationJSON(),
					}.AsV03(),
					Data: map[string]string{"unittest": "response"},
				},
			},
			asSent: map[string]*TapValidation{
				cloudevents.VersionV01: {
					Method: "POST",
					URI:    "/",
					Header: map[string][]string{
						"content-type": {"application/cloudevents+json"},
					},
					Body: fmt.Sprintf(`{"cloudEventsVersion":"0.1","contentType":"application/json","data":{"hello":"unittest"},"eventID":"ABC-123","eventTime":%q,"eventType":"unit.test.client.sent","source":"/unit/test/client"}`, now.UTC().Format(time.RFC3339Nano)),
				},
				cloudevents.VersionV02: {
					Method: "POST",
					URI:    "/",
					Header: map[string][]string{
						"content-type": {"application/cloudevents+json"},
					},
					Body: fmt.Sprintf(`{"contenttype":"application/json","data":{"hello":"unittest"},"id":"ABC-123","source":"/unit/test/client","specversion":"0.2","time":%q,"type":"unit.test.client.sent"}`, now.UTC().Format(time.RFC3339Nano)),
				},
				cloudevents.VersionV03: {
					Method: "POST",
					URI:    "/",
					Header: map[string][]string{
						"content-type": {"application/cloudevents+json"},
					},
					Body: fmt.Sprintf(`{"data":{"hello":"unittest"},"datacontenttype":"application/json","id":"ABC-123","source":"/unit/test/client","specversion":"0.3","time":%q,"type":"unit.test.client.sent"}`, now.UTC().Format(time.RFC3339Nano)),
				},
			},
			asRecv: map[string]*TapValidation{
				cloudevents.VersionV01: {
					Header: map[string][]string{
						"content-type": {"application/cloudevents+json"},
					},
					Body:   fmt.Sprintf(`{"cloudEventsVersion":"0.1","contentType":"application/json","data":{"unittest":"response"},"eventID":"321-CBA","eventTime":%q,"eventType":"unit.test.client.response","source":"/unit/test/client"}`, now.UTC().Format(time.RFC3339Nano)),
					Status: "200 OK",
				},
				cloudevents.VersionV02: {
					Header: map[string][]string{
						"content-type": {"application/cloudevents+json"},
					},
					Body:   fmt.Sprintf(`{"contenttype":"application/json","data":{"unittest":"response"},"id":"321-CBA","source":"/unit/test/client","specversion":"0.2","time":%q,"type":"unit.test.client.response"}`, now.UTC().Format(time.RFC3339Nano)),
					Status: "200 OK",
				},
				cloudevents.VersionV03: {
					Header: map[string][]string{
						"content-type": {"application/cloudevents+json"},
					},
					Body:   fmt.Sprintf(`{"data":{"unittest":"response"},"datacontenttype":"application/json","id":"321-CBA","source":"/unit/test/client","specversion":"0.3","time":%q,"type":"unit.test.client.response"}`, now.UTC().Format(time.RFC3339Nano)),
					Status: "200 OK",
				},
			},
		},
	}

	for n, tc := range testCases {
		for _, version := range versions {
			t.Run(n+version+" -> "+version, func(t *testing.T) {

				testcase := TapTest{
					now:    now,
					event:  tc.event(version),
					resp:   tc.resp(version),
					want:   tc.want[version],
					asSent: tc.asSent[version],
					asRecv: tc.asRecv[version],
				}

				testcase.asSent.ContentLength = int64(len(testcase.asSent.Body))
				testcase.asRecv.ContentLength = int64(len(testcase.asRecv.Body))

				ClientLoopback(t, testcase, cloudevents.WithStructuredEncoding())
			})
		}
	}
}

func TestClientLoopback_setters_structured_json_base64(t *testing.T) {
	now := time.Now()

	versions := []string{cloudevents.VersionV01, cloudevents.VersionV02, cloudevents.VersionV03}

	testCases := map[string]struct {
		event  func(string) *cloudevents.Event
		resp   func(string) *cloudevents.Event
		want   map[string]*cloudevents.Event
		asSent map[string]*TapValidation
		asRecv map[string]*TapValidation
	}{
		"Loopback": {
			event: func(version string) *cloudevents.Event {
				event := cloudevents.NewEvent(version)
				event.SetID("ABC-123")
				event.SetType("unit.test.client.sent")
				event.SetSource("/unit/test/client")
				event.SetDataContentEncoding(cloudevents.Base64)
				if err := event.SetData(map[string]string{"hello": "unittest"}); err != nil {
					t.Fatal(err)
				}
				return &event
			},
			resp: func(version string) *cloudevents.Event {
				event := cloudevents.NewEvent(version)
				event.SetID("321-CBA")
				event.SetType("unit.test.client.response")
				event.SetSource("/unit/test/client")
				event.SetDataContentEncoding(cloudevents.Base64)
				if err := event.SetData(map[string]string{"unittest": "response"}); err != nil {
					t.Fatal(err)
				}
				return &event
			},
			want: map[string]*cloudevents.Event{
				cloudevents.VersionV01: {
					Context: cloudevents.EventContextV01{
						EventID:     "321-CBA",
						EventType:   "unit.test.client.response",
						EventTime:   &cloudevents.Timestamp{Time: now},
						Source:      *cloudevents.ParseURLRef("/unit/test/client"),
						ContentType: cloudevents.StringOfApplicationJSON(),
						Extensions: map[string]interface{}{
							"datacontentencoding": "base64",
						},
					}.AsV01(),
					Data: map[string]string{"unittest": "response"},
				},
				cloudevents.VersionV02: {
					Context: cloudevents.EventContextV02{
						ID:          "321-CBA",
						Type:        "unit.test.client.response",
						Time:        &cloudevents.Timestamp{Time: now},
						Source:      *cloudevents.ParseURLRef("/unit/test/client"),
						ContentType: cloudevents.StringOfApplicationJSON(),
						Extensions: map[string]interface{}{
							"datacontentencoding": json.RawMessage(`"base64"`),
						},
					}.AsV02(),
					Data: map[string]string{"unittest": "response"},
				},
				cloudevents.VersionV03: {
					Context: cloudevents.EventContextV03{
						ID:                  "321-CBA",
						Type:                "unit.test.client.response",
						Time:                &cloudevents.Timestamp{Time: now},
						Source:              *cloudevents.ParseURLRef("/unit/test/client"),
						DataContentType:     cloudevents.StringOfApplicationJSON(),
						DataContentEncoding: cloudevents.StringOfBase64(),
					}.AsV03(),
					Data: map[string]string{"unittest": "response"},
				},
			},
			asSent: map[string]*TapValidation{
				cloudevents.VersionV01: {
					Method: "POST",
					URI:    "/",
					Header: map[string][]string{
						"content-type": {"application/cloudevents+json"},
					},
					Body: fmt.Sprintf(`{"cloudEventsVersion":"0.1","contentType":"application/json","data":"eyJoZWxsbyI6InVuaXR0ZXN0In0=","eventID":"ABC-123","eventTime":%q,"eventType":"unit.test.client.sent","extensions":{"datacontentencoding":"base64"},"source":"/unit/test/client"}`, now.UTC().Format(time.RFC3339Nano)),
				},
				cloudevents.VersionV02: {
					Method: "POST",
					URI:    "/",
					Header: map[string][]string{
						"content-type": {"application/cloudevents+json"},
					},
					Body: fmt.Sprintf(`{"contenttype":"application/json","data":"eyJoZWxsbyI6InVuaXR0ZXN0In0=","datacontentencoding":"base64","id":"ABC-123","source":"/unit/test/client","specversion":"0.2","time":%q,"type":"unit.test.client.sent"}`, now.UTC().Format(time.RFC3339Nano)),
				},
				cloudevents.VersionV03: {
					Method: "POST",
					URI:    "/",
					Header: map[string][]string{
						"content-type": {"application/cloudevents+json"},
					},
					Body: fmt.Sprintf(`{"data":"eyJoZWxsbyI6InVuaXR0ZXN0In0=","datacontentencoding":"base64","datacontenttype":"application/json","id":"ABC-123","source":"/unit/test/client","specversion":"0.3","time":%q,"type":"unit.test.client.sent"}`, now.UTC().Format(time.RFC3339Nano)),
				},
			},
			asRecv: map[string]*TapValidation{
				cloudevents.VersionV01: {
					Header: map[string][]string{
						"content-type": {"application/cloudevents+json"},
					},
					Body:   fmt.Sprintf(`{"cloudEventsVersion":"0.1","contentType":"application/json","data":"eyJ1bml0dGVzdCI6InJlc3BvbnNlIn0=","eventID":"321-CBA","eventTime":%q,"eventType":"unit.test.client.response","extensions":{"datacontentencoding":"base64"},"source":"/unit/test/client"}`, now.UTC().Format(time.RFC3339Nano)),
					Status: "200 OK",
				},
				cloudevents.VersionV02: {
					Header: map[string][]string{
						"content-type": {"application/cloudevents+json"},
					},
					Body:   fmt.Sprintf(`{"contenttype":"application/json","data":"eyJ1bml0dGVzdCI6InJlc3BvbnNlIn0=","datacontentencoding":"base64","id":"321-CBA","source":"/unit/test/client","specversion":"0.2","time":%q,"type":"unit.test.client.response"}`, now.UTC().Format(time.RFC3339Nano)),
					Status: "200 OK",
				},
				cloudevents.VersionV03: {
					Header: map[string][]string{
						"content-type": {"application/cloudevents+json"},
					},
					Body:   fmt.Sprintf(`{"data":"eyJ1bml0dGVzdCI6InJlc3BvbnNlIn0=","datacontentencoding":"base64","datacontenttype":"application/json","id":"321-CBA","source":"/unit/test/client","specversion":"0.3","time":%q,"type":"unit.test.client.response"}`, now.UTC().Format(time.RFC3339Nano)),
					Status: "200 OK",
				},
			},
		},
	}

	for n, tc := range testCases {
		for _, version := range versions {
			t.Run(n+version+" -> "+version, func(t *testing.T) {

				testcase := TapTest{
					now:    now,
					event:  tc.event(version),
					resp:   tc.resp(version),
					want:   tc.want[version],
					asSent: tc.asSent[version],
					asRecv: tc.asRecv[version],
				}

				testcase.asSent.ContentLength = int64(len(testcase.asSent.Body))
				testcase.asRecv.ContentLength = int64(len(testcase.asRecv.Body))

				ClientLoopback(t, testcase, cloudevents.WithStructuredEncoding())
			})
		}
	}
}
