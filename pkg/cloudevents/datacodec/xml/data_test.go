package xml_test

import (
	"encoding/xml"
	"fmt"
	cex "github.com/cloudevents/sdk-go/pkg/cloudevents/datacodec/xml"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/types"
	"github.com/google/go-cmp/cmp"
	"strings"
	"testing"
	"time"
)

type DataExample struct {
	AnInt   int        `xml:"a,omitempty"`
	AString string     `xml:"b,omitempty"`
	AnArray []string   `xml:"c,omitempty"`
	ATime   *time.Time `xml:"e,omitempty"`
}

type BadDataExample struct {
	AnInt int `xml:"a,omitempty"`
}

func (b BadDataExample) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return fmt.Errorf("unit test")
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
		wantErr string
	}{
		"empty": {},
		"not bytes": {
			in:      &BadDataExample{},
			wantErr: "[xml] failed to marshal in",
		},
		"structured type encoding, escaped": {
			in:   []byte(`"<Example><Sequence>7</Sequence><Message>Hello, Structured Encoding v0.2!</Message></Example>"`),
			want: &Example{Sequence: 7, Message: "Hello, Structured Encoding v0.2!"},
		},
		"structured type encoding, escaped error": {
			in:      []byte(`"<Example><Sequence>7</Sequence></Message>Hello, Structured Encoding v0.2!</Message></Example>"`),
			wantErr: "[xml] found bytes, but failed to unmarshal",
		},
		"structured type encoding, base64": {
			in:   []byte(`"PEV4YW1wbGU+PFNlcXVlbmNlPjc8L1NlcXVlbmNlPjxNZXNzYWdlPkhlbGxvLCBTdHJ1Y3R1cmVkIEVuY29kaW5nIHYwLjIhPC9NZXNzYWdlPjwvRXhhbXBsZT4="`),
			want: &Example{Sequence: 7, Message: "Hello, Structured Encoding v0.2!"},
		},
		"structured type encoding, bad quote base64": {
			in:      []byte(`"PEV4YW1wbGU+PFNlcXVlbmNlPjc8L1NlcXVlbmNlPjxNZXNzYWdlPkhlbGxvLCBTdHJ1Y3R1cmVkIEVuY29kaW5nIHYwLjIhPC9NZXNzYWdlPjwvRXhhbXBsZT4=`),
			wantErr: "[xml] failed to unquote quoted data",
		},
		"structured type encoding, bad base64": {
			in:      []byte(`"?EV4YW1wbGU+PFNlcXVlbmNlPjc8L1NlcXVlbmNlPjxNZXNzYWdlPkhlbGxvLCBTdHJ1Y3R1cmVkIEVuY29kaW5nIHYwLjIhPC9NZXNzYWdlPjwvRXhhbXBsZT4="`),
			wantErr: "[xml] failed to decode base64 encoded string",
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

			if tc.wantErr != "" {
				if err != nil {
					gotErr := err.Error()
					if !strings.Contains(gotErr, tc.wantErr) {
						t.Errorf("unexpected error, expected to contain %q, got: %q", tc.wantErr, gotErr)
					}
				} else {
					t.Errorf("expected error to contain %q, got: nil", tc.wantErr)
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
