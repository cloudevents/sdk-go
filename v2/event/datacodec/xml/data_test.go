package xml_test

import (
	"context"
	"encoding/xml"
	"fmt"
	"testing"
	"time"

	cex "github.com/cloudevents/sdk-go/v2/event/datacodec/xml"
	"github.com/cloudevents/sdk-go/v2/types"
	"github.com/google/go-cmp/cmp"
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
		in      []byte
		want    interface{}
		wantErr string
	}{
		"empty": {},
		"structured type encoding": {
			in:   []byte(`<Example><Sequence>7</Sequence><Message>Hello, Structured Encoding v0.2!</Message></Example>`),
			want: &Example{Sequence: 7, Message: "Hello, Structured Encoding v0.2!"},
		},
		"structured type encoding, escaped error": {
			in:      []byte(`"<Example><Sequence>7</Sequence></Message>Hello, Structured Encoding v0.2!</Message></Example>"`),
			wantErr: `[xml] found bytes, but failed to unmarshal: non-pointer passed to Unmarshal "<Example><Sequence>7</Sequence></Message>Hello, Structured Encoding v0.2!</Message></Example>"`,
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
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			got, _ := types.Allocate(tc.want)
			gotObs, _ := types.Allocate(tc.want)

			err := cex.Decode(context.TODO(), tc.in, got)
			errObs := cex.DecodeObserved(context.TODO(), tc.in, gotObs)

			if diff := cmpErrors(tc.wantErr, errObs); diff != "" {
				t.Errorf("obs unexpected error (-want, +got) = %v", diff)
			}

			if diff := cmpErrors(tc.wantErr, err); diff != "" {
				t.Errorf("unexpected error (-want, +got) = %v", diff)
			}

			if diff := cmp.Diff(gotObs, got); diff != "" {
				t.Errorf("obs unexpected obj diff between observed and direct (-want, +got) = %v", diff)
			}

			if tc.wantErr == "" && tc.want != nil {
				if diff := cmp.Diff(tc.want, got); diff != "" {
					t.Errorf("unexpected data (-want, +got) = %v", diff)
				}
			}
		})
	}
}

func TestCodecEncode(t *testing.T) {
	testCases := map[string]struct {
		in      interface{}
		want    interface{}
		wantErr string
	}{
		"empty": {},
		"not bytes": {
			in:      &BadDataExample{},
			wantErr: "unit test",
		},
		"bytes": {
			in:   []byte(`"<pre>Value</pre>"`),
			want: []byte(`"<pre>Value</pre>"`),
		},
		"structured type encoding, escaped": {
			in:   &Example{Sequence: 7, Message: "Hello, Structured Encoding v0.2!"},
			want: []byte(`<Example><Sequence>7</Sequence><Message>Hello, Structured Encoding v0.2!</Message></Example>`),
		},
		"complex filled": {
			in: &DataExample{
				AnInt:   42,
				AString: "Hello, World!",
				AnArray: []string{"Anne", "Bob", "Chad"},
			},
			want: []byte("<DataExample><a>42</a><b>Hello, World!</b><c>Anne</c><c>Bob</c><c>Chad</c></DataExample>"),
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			got, err := cex.Encode(context.TODO(), tc.in)

			gotObs, errObs := cex.EncodeObserved(context.TODO(), tc.in)

			if diff := cmpErrors(tc.wantErr, errObs); diff != "" {
				t.Errorf("obs unexpected error (-want, +got) = %v", diff)
			}
			if diff := cmp.Diff(gotObs, got); diff != "" {
				t.Errorf("obs unexpected obj diff between observed and direct (-want, +got) = %v", diff)
			}

			if diff := cmpErrors(tc.wantErr, err); diff != "" {
				t.Errorf("unexpected error (-want, +got) = %v", diff)
			}

			if tc.want != nil {
				if diff := cmp.Diff(tc.want, got); diff != "" {
					t.Errorf("unexpected data (-want, +got) = %v", diff)
				}
			}
		})
	}
}

func cmpErrors(want string, err error) string {
	if want != "" || err != nil {
		var gotErr string
		if err != nil {
			gotErr = err.Error()
		}
		return cmp.Diff(want, gotErr)
	}
	return ""
}
