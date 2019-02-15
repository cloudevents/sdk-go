package xml_test

import (
	"encoding/xml"
	cex "github.com/cloudevents/sdk-go/pkg/cloudevents/datacodec/xml"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/types"
	"github.com/google/go-cmp/cmp"
	"testing"
	"time"
)

type DataExample struct {
	AnInt   int        `xml:"a,omitempty"`
	AString string     `xml:"b,omitempty"`
	AnArray []string   `xml:"c,omitempty"`
	ATime   *time.Time `xml:"e,omitempty"`
}

// Basic data struct.
type Example struct {
	Sequence int    `json:"id"`
	Message  string `json:"message"`
}

func TestCodecDecode(t *testing.T) {
	now := time.Now()

	testCases := map[string]struct {
		in      interface{}
		want    interface{}
		wantErr error
	}{
		"empty": {},
		"structured type encoding, escaped": {
			in:   []byte(`"<Example><Sequence>7</Sequence><Message>Hello, Structured Encoding v0.2!</Message></Example>"`),
			want: &Example{Sequence: 7, Message: "Hello, Structured Encoding v0.2!"},
		},
		"structured type encoding, base64": {
			in:   []byte(`"PEV4YW1wbGU+PFNlcXVlbmNlPjc8L1NlcXVlbmNlPjxNZXNzYWdlPkhlbGxvLCBTdHJ1Y3R1cmVkIEVuY29kaW5nIHYwLjIhPC9NZXNzYWdlPjwvRXhhbXBsZT4="`),
			want: &Example{Sequence: 7, Message: "Hello, Structured Encoding v0.2!"},
		},
		"complex filled": {
			in: func() []byte {
				data := &DataExample{
					AnInt:   42,
					AString: "Hello, World!",
					ATime:   &now,
					AnArray: []string{"Anne", "Bob", "Chad"},
				}

				j, err := xml.Marshal(data)

				if err != nil {
					t.Errorf("failed to marshal test data: %s", err.Error())
				}
				return j
			}(),
			want: &DataExample{
				AnInt:   42,
				AString: "Hello, World!",
				ATime:   &now,
				AnArray: []string{"Anne", "Bob", "Chad"},
			},
		},
		"object in": {
			in: &DataExample{
				AnInt: 42,
			},
			want: &DataExample{
				AnInt: 42,
			},
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			got, _ := types.Allocate(tc.want)

			err := cex.Decode(tc.in, got)

			if tc.wantErr != nil || err != nil {
				if diff := cmp.Diff(tc.wantErr, err); diff != "" {
					t.Errorf("unexpected error (-want, +got) = %v", diff)
				}
				return
			}

			if tc.want != nil {
				if diff := cmp.Diff(tc.want, got); diff != "" {
					t.Errorf("unexpected data (-want, +got) = %v", diff)
				}
			}
		})
	}
}
