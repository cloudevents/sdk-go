package xml_test

import (
	"encoding/xml"
	cex "github.com/cloudevents/sdk-go/pkg/cloudevents/datacodec/xml"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/types"
	"github.com/google/go-cmp/cmp"
	"reflect"
	"testing"
	"time"
)

type DataExample struct {
	AnInt   int        `xml:"a,omitempty"`
	AString string     `xml:"b,omitempty"`
	AnArray []string   `xml:"c,omitempty"`
	ATime   *time.Time `xml:"e,omitempty"`
}

func TestCodecDecode(t *testing.T) {
	now := time.Now()

	testCases := map[string]struct {
		in      interface{}
		want    interface{}
		wantErr error
	}{
		"empty": {},
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

			dataType := reflect.TypeOf(tc.want)

			t.Logf("got dataType: %s", dataType)

			got, _ := types.Allocate(dataType)

			err := cex.Decode(tc.in, got)

			if tc.wantErr != nil || err != nil {
				if diff := cmp.Diff(tc.wantErr, err); diff != "" {
					t.Errorf("unexpected error (-want, +got) = %v", diff)
				}
				return
			}

			if dataType != nil {
				if diff := cmp.Diff(tc.want, got); diff != "" {
					t.Errorf("unexpected data (-want, +got) = %v", diff)
				}
			}
		})
	}
}
