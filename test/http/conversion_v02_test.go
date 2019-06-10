package http

import (
	"net/http"
	"testing"
	"time"

	"github.com/cloudevents/sdk-go"
)

func TestClientConversion_v02(t *testing.T) {
	now := time.Now()

	testCases := ConversionTestCases{
		"Conversion v0.2": {
			now:       now,
			convertFn: UnitTestConvert,
			data:      map[string]string{"hello": "unittest"},
			want: &cloudevents.Event{
				Context: cloudevents.EventContextV02{
					ID:     "321-CBA",
					Type:   "io.cloudevents.conversion.http.post",
					Source: *cloudevents.ParseURLRef("github.com/cloudevents/test/http/conversion"),
				}.AsV02(),
				Data: map[string]string{"unittest": "response"},
			},
			asSent: &TapValidation{
				Method: "POST",
				URI:    "/",
				Header: map[string][]string{
					"content-type": {"application/json"},
				},
				Body:          `{"hello":"unittest"}`,
				ContentLength: 20,
			},
			asRecv: &TapValidation{
				Header:        http.Header{},
				Status:        "202 Accepted",
				ContentLength: 0,
			},
		},
	}

	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			ClientConversion(t, tc)
		})
	}
}
