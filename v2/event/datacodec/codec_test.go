/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package datacodec_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/cloudevents/sdk-go/v2/event/datacodec"
	"github.com/cloudevents/sdk-go/v2/types"
)

func strptr(s string) *string { return &s }

type Example struct {
	Sequence int    `json:"id"`
	Message  string `json:"message"`
}

func TestCodecDecode(t *testing.T) {
	testCases := map[string]struct {
		contentType      string
		decoder          datacodec.Decoder
		structuredSuffix string
		in               []byte
		want             interface{}
		wantErr          string
	}{
		"empty": {},
		"invalid content type": {
			contentType: "unit/testing-invalid",
			wantErr:     `[decode] unsupported content type: "unit/testing-invalid"`,
		},

		"text/plain": {
			contentType: "text/plain",
			in:          []byte("hello😀"),
			want:        strptr("hello😀"), // Test  unicode outiside UTF-8
		},
		"application/json": {
			contentType: "application/json",
			in:          []byte(`{"a":"apple","b":"banana"}`),
			want: &map[string]string{
				"a": "apple",
				"b": "banana",
			},
		},
		"application/vnd.custom-type+json": {
			contentType: "application/vnd.custom-type+json",
			in:          []byte(`{"a":"apple","b":"banana"}`),
			want: &map[string]string{
				"a": "apple",
				"b": "banana",
			},
		},
		"application/xml": {
			contentType: "application/xml",
			in:          []byte(`<Example><Sequence>7</Sequence><Message>Hello, Structured Encoding v1.0!</Message></Example>`),
			want:        &Example{Sequence: 7, Message: "Hello, Structured Encoding v1.0!"},
		},
		"application/vnd.custom-type+xml": {
			contentType: "application/vnd.custom-type+xml",
			in:          []byte(`<Example><Sequence>7</Sequence><Message>Hello, Structured Encoding v1.0!</Message></Example>`),
			want:        &Example{Sequence: 7, Message: "Hello, Structured Encoding v1.0!"},
		},
		"custom content type": {
			contentType: "unit/testing",
			in:          []byte("Hello, Testing"),
			decoder: func(ctx context.Context, in []byte, out interface{}) error {
				if s, k := out.(*map[string]string); k {
					if (*s) == nil {
						(*s) = make(map[string]string)
					}
					(*s)["upper"] = strings.ToUpper(string(in))
					(*s)["lower"] = strings.ToLower(string(in))
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
			decoder: func(ctx context.Context, in []byte, out interface{}) error {
				return fmt.Errorf("expecting unit test error")
			},
			wantErr: "expecting unit test error",
		},
		"custom structured suffix": {
			contentType:      "unit/testing+custom",
			structuredSuffix: "custom",
			in:               []byte("Hello, Testing"),
			decoder: func(ctx context.Context, in []byte, out interface{}) error {
				if s, k := out.(*map[string]string); k {
					if (*s) == nil {
						(*s) = make(map[string]string)
					}
					(*s)["upper"] = strings.ToUpper(string(in))
					(*s)["lower"] = strings.ToLower(string(in))
				}
				return nil
			},
			want: &map[string]string{
				"upper": "HELLO, TESTING",
				"lower": "hello, testing",
			},
		},
		"custom structured suffix error": {
			contentType:      "unit/testing+custom",
			structuredSuffix: "custom",
			in:               []byte("Hello, Testing"),
			decoder: func(ctx context.Context, in []byte, out interface{}) error {
				return fmt.Errorf("expecting unit test error")
			},
			wantErr: "expecting unit test error",
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			if tc.decoder != nil {
				if tc.structuredSuffix == "" {
					datacodec.AddDecoder(tc.contentType, tc.decoder)
				} else {
					datacodec.AddStructuredSuffixDecoder(tc.structuredSuffix, tc.decoder)
				}
			}

			got, _ := types.Allocate(tc.want)
			err := datacodec.Decode(context.TODO(), tc.contentType, tc.in, got)

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
		contentType      string
		structuredSuffix string
		encoder          datacodec.Encoder
		in               interface{}
		want             []byte
		wantErr          string
	}{
		"empty": {},
		"invalid content type": {
			contentType: "unit/testing-invalid",
			wantErr:     `[encode] unsupported content type: "unit/testing-invalid"`,
		},
		"blank": {
			contentType: "",
			in: map[string]string{
				"a": "apple",
				"b": "banana",
			},
			want: []byte(`{"a":"apple","b":"banana"}`),
		},
		"application/json": {
			contentType: "application/json",
			in: map[string]string{
				"a": "apple",
				"b": "banana",
			},
			want: []byte(`{"a":"apple","b":"banana"}`),
		},
		"application/vnd.custom-type+json": {
			contentType: "application/vnd.custom-type+json",
			in: map[string]string{
				"a": "apple",
				"b": "banana",
			},
			want: []byte(`{"a":"apple","b":"banana"}`),
		},
		"application/xml": {
			contentType: "application/xml",
			in:          &Example{Sequence: 7, Message: "Hello, Structured Encoding v1.0!"},
			want:        []byte(`<Example><Sequence>7</Sequence><Message>Hello, Structured Encoding v1.0!</Message></Example>`),
		},
		"application/vnd.custom-type+xml": {
			contentType: "application/vnd.custom-type+xml",
			in:          &Example{Sequence: 7, Message: "Hello, Structured Encoding v1.0!"},
			want:        []byte(`<Example><Sequence>7</Sequence><Message>Hello, Structured Encoding v1.0!</Message></Example>`),
		},

		"custom content type": {
			contentType: "unit/testing",
			in: []string{
				"Hello,",
				"Testing",
			},
			encoder: func(ctx context.Context, in interface{}) ([]byte, error) {
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
			encoder: func(ctx context.Context, in interface{}) ([]byte, error) {
				return nil, fmt.Errorf("expecting unit test error")
			},
			wantErr: "expecting unit test error",
		},
		"custom structured suffix": {
			contentType:      "unit/testing+custom",
			structuredSuffix: "custom",
			in: []string{
				"Hello,",
				"Testing",
			},
			encoder: func(ctx context.Context, in interface{}) ([]byte, error) {
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
		"custom structured suffix error": {
			contentType:      "unit/testing+custom",
			structuredSuffix: "custom",
			in:               []byte("Hello, Testing"),
			encoder: func(ctx context.Context, in interface{}) ([]byte, error) {
				return nil, fmt.Errorf("expecting unit test error")
			},
			wantErr: "expecting unit test error",
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			if tc.encoder != nil {
				if tc.structuredSuffix == "" {
					datacodec.AddEncoder(tc.contentType, tc.encoder)
				} else {
					datacodec.AddStructuredSuffixEncoder(tc.structuredSuffix, tc.encoder)
				}
			}

			got, err := datacodec.Encode(context.TODO(), tc.contentType, tc.in)

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
