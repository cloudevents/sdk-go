/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package http

import (
	"fmt"
	"github.com/cloudevents/sdk-go/v2/client"
	"testing"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
)

func TestClientLoopback_setters_binary_json(t *testing.T) {
	now := time.Now()

	versions := []string{cloudevents.VersionV03, cloudevents.VersionV1}

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
				if err := event.SetData(cloudevents.ApplicationJSON, map[string]string{"hello": "unittest"}); err != nil {
					t.Fatal(err)
				}
				return &event
			},
			resp: func(version string) *cloudevents.Event {
				event := cloudevents.NewEvent(version)
				event.SetID("321-CBA")
				event.SetType("unit.test.client.response")
				event.SetSource("/unit/test/client")
				if err := event.SetData(cloudevents.ApplicationJSON, map[string]string{"unittest": "response"}); err != nil {
					t.Fatal(err)
				}
				return &event
			},
			want: map[string]*cloudevents.Event{
				cloudevents.VersionV1: {
					Context: cloudevents.EventContextV1{
						ID:              "321-CBA",
						Type:            "unit.test.client.response",
						Time:            &cloudevents.Timestamp{Time: now},
						Source:          *cloudevents.ParseURIRef("/unit/test/client"),
						DataContentType: cloudevents.StringOfApplicationJSON(),
					}.AsV1(),
					DataEncoded: toBytes(map[string]interface{}{"unittest": "response"}),
				},
				cloudevents.VersionV03: {
					Context: cloudevents.EventContextV03{
						ID:              "321-CBA",
						Type:            "unit.test.client.response",
						Time:            &cloudevents.Timestamp{Time: now},
						Source:          *cloudevents.ParseURIRef("/unit/test/client"),
						DataContentType: cloudevents.StringOfApplicationJSON(),
					}.AsV03(),
					DataEncoded: toBytes(map[string]interface{}{"unittest": "response"}),
				},
			},
			asSent: map[string]*TapValidation{
				cloudevents.VersionV1: {
					Method: "POST",
					URI:    "/",
					Header: map[string][]string{
						"ce-specversion": {"1.0"},
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
				cloudevents.VersionV1: {
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

	versions := []string{cloudevents.VersionV1, cloudevents.VersionV03}

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
				event.SetDataContentType(cloudevents.ApplicationJSON)
				return &event
			},
			resp: func(version string) *cloudevents.Event {
				event := cloudevents.NewEvent(version)
				event.SetID("321-CBA")
				event.SetType("unit.test.client.response")
				event.SetSource("/unit/test/client")
				event.SetDataContentType(cloudevents.ApplicationJSON)
				return &event
			},
			want: map[string]*cloudevents.Event{
				cloudevents.VersionV1: {
					Context: cloudevents.EventContextV1{
						ID:              "321-CBA",
						Type:            "unit.test.client.response",
						Time:            &cloudevents.Timestamp{Time: now},
						Source:          *cloudevents.ParseURIRef("/unit/test/client"),
						DataContentType: cloudevents.StringOfApplicationJSON(),
					}.AsV1(),
				},
				cloudevents.VersionV03: {
					Context: cloudevents.EventContextV03{
						ID:              "321-CBA",
						Type:            "unit.test.client.response",
						Time:            &cloudevents.Timestamp{Time: now},
						Source:          *cloudevents.ParseURIRef("/unit/test/client"),
						DataContentType: cloudevents.StringOfApplicationJSON(),
					}.AsV03(),
				},
			},
			asSent: map[string]*TapValidation{
				cloudevents.VersionV1: {
					Method: "POST",
					URI:    "/",
					Header: map[string][]string{
						"ce-specversion": {"1.0"},
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
				cloudevents.VersionV1: {
					Header: map[string][]string{
						"ce-specversion": {"1.0"},
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

	versions := []string{cloudevents.VersionV1, cloudevents.VersionV03}

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
				event.SetDataContentType(cloudevents.ApplicationJSON)
				if err := event.SetData(cloudevents.ApplicationJSON, map[string]string{"hello": "unittest"}); err != nil {
					t.Fatal(err)
				}
				return &event
			},
			resp: func(version string) *cloudevents.Event {
				event := cloudevents.NewEvent(version)
				event.SetID("321-CBA")
				event.SetType("unit.test.client.response")
				event.SetSource("/unit/test/client")
				event.SetDataContentType(cloudevents.ApplicationJSON)
				if err := event.SetData(cloudevents.ApplicationJSON, map[string]string{"unittest": "response"}); err != nil {
					t.Fatal(err)
				}
				return &event
			},
			want: map[string]*cloudevents.Event{
				cloudevents.VersionV1: {
					Context: cloudevents.EventContextV1{
						ID:              "321-CBA",
						Type:            "unit.test.client.response",
						Time:            &cloudevents.Timestamp{Time: now},
						Source:          *cloudevents.ParseURIRef("/unit/test/client"),
						DataContentType: cloudevents.StringOfApplicationJSON(),
					}.AsV1(),
					DataEncoded: toBytes(map[string]interface{}{"unittest": "response"}),
				},
				cloudevents.VersionV03: {
					Context: cloudevents.EventContextV03{
						ID:              "321-CBA",
						Type:            "unit.test.client.response",
						Time:            &cloudevents.Timestamp{Time: now},
						Source:          *cloudevents.ParseURIRef("/unit/test/client"),
						DataContentType: cloudevents.StringOfApplicationJSON(),
					}.AsV03(),
					DataEncoded: toBytes(map[string]interface{}{"unittest": "response"}),
				},
			},
			asSent: map[string]*TapValidation{
				cloudevents.VersionV1: {
					Method: "POST",
					URI:    "/",
					Header: map[string][]string{
						"content-type": {"application/cloudevents+json"},
					},
					Body: fmt.Sprintf(`{"data":{"hello":"unittest"},"datacontenttype":"application/json","id":"ABC-123","source":"/unit/test/client","specversion":"1.0","time":%q,"type":"unit.test.client.sent"}`, now.UTC().Format(time.RFC3339Nano)),
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
				cloudevents.VersionV1: {
					Header: map[string][]string{
						"content-type": {"application/cloudevents+json"},
					},
					Body: fmt.Sprintf(`{"data":{"unittest":"response"},"datacontenttype":"application/json","id":"321-CBA","source":"/unit/test/client","specversion":"1.0","time":%q,"type":"unit.test.client.response"}`, now.UTC().Format(time.RFC3339Nano)),

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

				ClientLoopback(t, testcase, client.WithForceStructured())
			})
		}
	}
}

func TestClientLoopback_setters_structured_json_base64(t *testing.T) {
	t.Skip("TODO: bindings does not yet support base64")

	now := time.Now()

	versions := []string{cloudevents.VersionV03}

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
				if err := event.SetData(cloudevents.ApplicationJSON, map[string]string{"hello": "unittest"}); err != nil {
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
				if err := event.SetData(cloudevents.ApplicationJSON, map[string]string{"unittest": "response"}); err != nil {
					t.Fatal(err)
				}
				return &event
			},
			want: map[string]*cloudevents.Event{
				cloudevents.VersionV03: {
					Context: cloudevents.EventContextV03{
						ID:                  "321-CBA",
						Type:                "unit.test.client.response",
						Time:                &cloudevents.Timestamp{Time: now},
						Source:              *cloudevents.ParseURIRef("/unit/test/client"),
						DataContentType:     cloudevents.StringOfApplicationJSON(),
						DataContentEncoding: cloudevents.StringOfBase64(),
					}.AsV03(),
					DataEncoded: toBytes(map[string]interface{}{"unittest": "response"}),
				},
			},
			asSent: map[string]*TapValidation{
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

				ClientLoopback(t, testcase, client.WithForceStructured())
			})
		}
	}
}
