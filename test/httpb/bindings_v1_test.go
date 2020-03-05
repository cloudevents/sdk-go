package httpb

import (
	"encoding/json"
	"fmt"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport/httpb"
	"github.com/cloudevents/sdk-go/test/http"
	"testing"
	"time"

	"github.com/cloudevents/sdk-go"
	"github.com/cloudevents/sdk-go/pkg/types"
)

func TestSenderReceiver_bindings_binary_v1(t *testing.T) {
	now := time.Now()

	testCases := http.DirectTapTestCases{
		"Binary v1.0": {
			Now: now,
			Event: &cloudevents.Event{
				Context: cloudevents.EventContextV1{
					ID:      "ABC-123",
					Type:    "unit.test.client.sent",
					Source:  *cloudevents.ParseURIRef("/unit/test/client"),
					Subject: strptr("resource"),
				}.AsV1(),
				Data: map[string]string{"hello": "unittest"},
			},
			Want: &cloudevents.Event{
				Context: cloudevents.EventContextV1{
					ID:      "ABC-123",
					Type:    "unit.test.client.sent",
					Time:    &cloudevents.Timestamp{Time: now},
					Source:  *cloudevents.ParseURIRef("/unit/test/client"),
					Subject: strptr("resource"),
				}.AsV1(),
				Data: toBytes(map[string]interface{}{"hello": "unittest"}),
			},
			AsSent: &http.TapValidation{
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
			ClientBindingsDirect(t, tc)
		})
	}
}

func TestSenderReceiver_bindings_structured_v1(t *testing.T) {
	now := time.Now()

	testCases := http.DirectTapTestCases{
		"Structured v1.0": {
			Now: now,
			Event: func() *cloudevents.Event {
				event := cloudevents.NewEvent(cloudevents.VersionV1)
				event.SetID("ABC-123")
				event.SetType("unit.test.client.sent")
				event.SetSource("/unit/test/client")
				event.SetSubject("resource")
				_ = event.SetData(map[string]string{"hello": "unittest"})
				return &event
			}(),
			Want: &cloudevents.Event{
				Context: cloudevents.EventContextV1{
					ID:      "ABC-123",
					Type:    "unit.test.client.sent",
					Time:    &cloudevents.Timestamp{Time: now},
					Source:  *cloudevents.ParseURIRef("/unit/test/client"),
					Subject: strptr("resource"),
				}.AsV1(),
				Data: toBytes(map[string]interface{}{"hello": "unittest"}),
			},
			AsSent: &http.TapValidation{
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
			ClientBindingsDirect(t, tc, httpb.WithStructuredEncoding())
		})
	}
}

func TestSenderReceiver_bindings_data_base64_v1(t *testing.T) {
	t.Skipf("bindings need work to support base64 encoded messages.")

	now := time.Now()

	testCases := http.DirectTapTestCases{
		"Structured v1.0": {
			Now: now,
			Event: func() *cloudevents.Event {
				event := cloudevents.NewEvent(cloudevents.VersionV1)
				event.SetID("ABC-123")
				event.SetType("unit.test.client.sent")
				event.SetSource("/unit/test/client")
				event.SetSubject("resource")
				_ = event.SetData([]byte("hello: unittest"))
				return &event
			}(),
			Want: &cloudevents.Event{
				Context: cloudevents.EventContextV1{
					ID:      "ABC-123",
					Type:    "unit.test.client.sent",
					Time:    &cloudevents.Timestamp{Time: now},
					Source:  *cloudevents.ParseURIRef("/unit/test/client"),
					Subject: strptr("resource"),
				}.AsV1(),
				Data: []byte("hello: unittest"),
			},
			AsSent: &http.TapValidation{
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
			ClientBindingsDirect(t, tc, httpb.WithStructuredEncoding())
		})
	}
}

func toBytes(body map[string]interface{}) []byte {
	b, err := json.Marshal(body)
	if err != nil {
		return []byte(fmt.Sprintf(`{"error":%q}`, err.Error()))
	}
	return b
}

func strptr(s string) *string {
	return &s
}
