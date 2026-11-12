/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package http

import (
	"context"
	"net/http"
	"reflect"
	"testing"
)

func TestHeaderFrom(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name string
		args args
		want http.Header
	}{
		{
			name: "empty header",
			args: args{
				ctx: context.TODO(),
			},
			want: make(http.Header),
		},
		{
			name: "header with value",
			args: args{
				ctx: WithCustomHeader(context.TODO(), map[string][]string{"header": {"value"}}),
			},
			want: map[string][]string{"header": {"value"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HeaderFrom(tt.args.ctx); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HeaderFrom() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEncodeHeaderValue(t *testing.T) {
	tests := map[string]string{
		"plain":             "plain",
		"two words":         "two words",
		"  lead and trail ": "%20%20lead and trail%20",
		"   ":               "%20%20%20",
		`say "hi"`:          "say %22hi%22",
		"100%":              "100%25",
		"Euro €😀":           "Euro %E2%82%AC%F0%9F%98%80",
		"tab\there":         "tab%09here",
		`back\slash`:        `back\slash`,
	}
	for in, want := range tests {
		if got := encodeHeaderValue(in); got != want {
			t.Errorf("encode(%q) = %q, want %q", in, got, want)
		}
		if back, err := decodeHeaderValue(encodeHeaderValue(in)); err != nil || back != in {
			t.Errorf("round trip(%q) = %q, %v", in, back, err)
		}
	}
}

func TestDecodeHeaderValue(t *testing.T) {
	tests := []struct {
		in, want string
		wantErr  bool
	}{
		{in: "  plain  ", want: "plain"},
		{in: "a   b", want: "a   b"},
		{in: `"  word1    word2  "`, want: "  word1    word2  "},
		{in: `  "a, b"  `, want: "a, b"},
		{in: `"say \"hi\" \\ x"`, want: `say "hi" \ x`},
		{in: `a\b`, want: `a\b`},
		{in: `say %22hi%22`, want: `say "hi"`},
		{in: `"%22q%22"`, want: `"q"`},
		{in: "%20%20x%20", want: "  x "},
		{in: "%e2%82%ac", want: "€"},
		{in: "%41", want: "A"},
		{in: "100%", want: "100%"},
		{in: "%zz", want: "%zz"},
		{in: "%C0%A0", wantErr: true},
		{in: "%FF", wantErr: true},
	}
	for _, tc := range tests {
		got, err := decodeHeaderValue(tc.in)
		if (err != nil) != tc.wantErr || got != tc.want {
			t.Errorf("decode(%q) = %q, %v; want %q (err=%v)", tc.in, got, err, tc.want, tc.wantErr)
		}
	}
}
