package datacodec_test

import (
	"fmt"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/datacodec"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/types"
	"github.com/google/go-cmp/cmp"
	"strings"
	"testing"
)

type Example struct {
	Sequence int    `json:"id"`
	Message  string `json:"message"`
}

func TestCodecDecode(t *testing.T) {
	testCases := map[string]struct {
		contentType string
		decoder     datacodec.Decoder
		in          interface{}
		want        interface{}
		wantErr     string
	}{
		"empty": {},
		"invalid content type": {
			contentType: "unit/testing-invalid",
			wantErr:     `[decode] unsupported content type: "unit/testing-invalid"`,
		},

		"application/json": {
			contentType: "application/json",
			in:          []byte(`{"a":"apple","b":"banana"}`),
			want: &map[string]string{
				"a": "apple",
				"b": "banana",
			},
		},
		"application/xml": {
			contentType: "application/xml",
			in:          []byte(`<Example><Sequence>7</Sequence><Message>Hello, Structured Encoding v0.2!</Message></Example>`),
			want:        &Example{Sequence: 7, Message: "Hello, Structured Encoding v0.2!"},
		},

		"custom content type": {
			contentType: "unit/testing",
			in:          []byte("Hello, Testing"),
			decoder: func(in, out interface{}) error {
				if b, ok := in.([]byte); ok {
					if s, k := out.(*map[string]string); k {
						if (*s) == nil {
							(*s) = make(map[string]string)
						}
						(*s)["upper"] = strings.ToUpper(string(b))
						(*s)["lower"] = strings.ToLower(string(b))
					}
				}
				return nil
			},
			want: &map[string]string{
				"upper": "HELLO, TESTING",
				"lower": "hello, testing",
			},
		},
		"custom content type error": {
			contentType: "unit/testing",
			in:          []byte("Hello, Testing"),
			decoder: func(in, out interface{}) error {
				return fmt.Errorf("expecting unit test error")
			},
			wantErr: "expecting unit test error",
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			if tc.decoder != nil {
				datacodec.AddDecoder(tc.contentType, tc.decoder)
			}

			got, _ := types.Allocate(tc.want)

			err := datacodec.Decode(tc.contentType, tc.in, got)

			if tc.wantErr != "" || err != nil {
				if diff := cmp.Diff(tc.wantErr, err.Error()); diff != "" {
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

func TestCodecEncode(t *testing.T) {
	testCases := map[string]struct {
		contentType string
		encoder     datacodec.Encoder
		in          interface{}
		want        []byte
		wantErr     string
	}{
		"empty": {},
		"invalid content type": {
			contentType: "unit/testing-invalid",
			wantErr:     `[encode] unsupported content type: "unit/testing-invalid"`,
		},

		"application/json": {
			contentType: "application/json",
			in: map[string]string{
				"a": "apple",
				"b": "banana",
			},
			want: []byte(`{"a":"apple","b":"banana"}`),
		},
		"application/xml": {
			contentType: "application/xml",
			in:          &Example{Sequence: 7, Message: "Hello, Structured Encoding v0.2!"},
			want:        []byte(`<Example><Sequence>7</Sequence><Message>Hello, Structured Encoding v0.2!</Message></Example>`),
		},

		"custom content type": {
			contentType: "unit/testing",
			in: []string{
				"Hello,",
				"Testing",
			},
			encoder: func(in interface{}) ([]byte, error) {
				if s, ok := in.([]string); ok {
					sb := strings.Builder{}
					for _, v := range s {
						if sb.Len() > 0 {
							sb.WriteString(" ")
						}
						sb.WriteString(v)
					}
					return []byte(sb.String()), nil
				}
				return nil, fmt.Errorf("don't get here")
			},
			want: []byte("Hello, Testing"),
		},
		"custom content type error": {
			contentType: "unit/testing",
			in:          []byte("Hello, Testing"),
			encoder: func(in interface{}) ([]byte, error) {
				return nil, fmt.Errorf("expecting unit test error")
			},
			wantErr: "expecting unit test error",
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			if tc.encoder != nil {
				datacodec.AddEncoder(tc.contentType, tc.encoder)
			}

			got, err := datacodec.Encode(tc.contentType, tc.in)

			if tc.wantErr != "" || err != nil {
				if diff := cmp.Diff(tc.wantErr, err.Error()); diff != "" {
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

//
//func TestCodecRoundTrip(t *testing.T) {
//	testCases := map[string]struct {
//		contentType string
//		decoder     datacodec.Decoder
//		encoder     datacodec.Encoder
//		in          interface{}
//		want        interface{}
//		wantErr     string
//	}{
//		"empty": {},
//	}
//	for n, tc := range testCases {
//		t.Run(n, func(t *testing.T) {
//
//			if tc.decoder != nil {
//				datacodec.AddDecoder(tc.contentType, tc.decoder)
//			}
//
//
//			// TODO
//			got, _ := types.Allocate(tc.want)
//
//			err := datacodec.Decode(tc.contentType, tc.in, got)
//
//			if tc.wantErr != "" || err != nil {
//				if diff := cmp.Diff(tc.wantErr, err.Error()); diff != "" {
//					t.Errorf("unexpected error (-want, +got) = %v", diff)
//				}
//				return
//			}
//
//			if tc.want != nil {
//				if diff := cmp.Diff(tc.want, got); diff != "" {
//					t.Errorf("unexpected data (-want, +got) = %v", diff)
//				}
//			}
//		})
//	}
//}
