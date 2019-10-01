package http

import (
	"fmt"
	"testing"
	"time"

	"github.com/cloudevents/sdk-go"
)

func TestSenderReceiver_binary_v01(t *testing.T) {
	now := time.Now()

	testCases := DirectTapTestCases{
		"Binary v1.0": {
			now: now,
			event: &cloudevents.Event{
				Context: cloudevents.EventContextV1{
					ID:      "ABC-123",
					Type:    "unit.test.client.sent",
					Source:  *cloudevents.ParseURIRef("/unit/test/client"),
					Subject: strptr("resource"),
				}.AsV1(),
				Data: map[string]string{"hello": "unittest"},
			},
			want: &cloudevents.Event{
				Context: cloudevents.EventContextV1{
					ID:      "ABC-123",
					Type:    "unit.test.client.sent",
					Time:    &cloudevents.Timestamp{Time: now},
					Source:  *cloudevents.ParseURIRef("/unit/test/client"),
					Subject: strptr("resource"),
				}.AsV1(),
				Data: toBytes(map[string]interface{}{"hello": "unittest"}),
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

func TestSenderReceiver_structured_v01(t *testing.T) {
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
				_ = event.SetData(map[string]string{"hello": "unittest"})
				return &event
			}(),
			want: &cloudevents.Event{
				Context: cloudevents.EventContextV1{
					ID:      "ABC-123",
					Type:    "unit.test.client.sent",
					Time:    &cloudevents.Timestamp{Time: now},
					Source:  *cloudevents.ParseURIRef("/unit/test/client"),
					Subject: strptr("resource"),
				}.AsV1(),
				Data: toBytes(map[string]interface{}{"hello": "unittest"}),
			},
			asSent: &TapValidation{
				Method: "POST",
				URI:    "/",
				Header: map[string][]string{
					"content-type": {"application/cloudevents+json"},
				},
				Body:          fmt.Sprintf(`{"data":{"hello":"unittest"},"id":"ABC-123","source":"/unit/test/client","specversion":"1.0","subject":"resource","time":%q,"type":"unit.test.client.sent"}`, now.UTC().Format(time.RFC3339Nano)),
				ContentLength: 182,
			},
		},
	}

	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			ClientDirect(t, tc, cloudevents.WithStructuredEncoding())
		})
	}
}

func TestSenderReceiver_data_base64_v01(t *testing.T) {
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
				_ = event.SetData([]byte("hello: unittest"))
				return &event
			}(),
			want: &cloudevents.Event{
				Context: cloudevents.EventContextV1{
					ID:      "ABC-123",
					Type:    "unit.test.client.sent",
					Time:    &cloudevents.Timestamp{Time: now},
					Source:  *cloudevents.ParseURIRef("/unit/test/client"),
					Subject: strptr("resource"),
				}.AsV1(),
				Data: []byte("hello: unittest"),
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
			ClientDirect(t, tc, cloudevents.WithStructuredEncoding())
		})
	}
}
