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

func TestWithCustomHeaderField(t *testing.T) {
	type args struct {
		ctx   context.Context
		name  string
		value string
	}
	tests := []struct {
		name string
		args args
		want http.Header
	}{
		{
			name: "no headers in context before",
			args: args{
				ctx:   context.TODO(),
				name:  "Header",
				value: "value",
			},
			want: map[string][]string{
				"Header": {"value"},
			},
		},
		{
			name: "headers in context before",
			args: args{
				ctx: WithCustomHeader(context.TODO(), map[string][]string{
					"Header": {"value"},
				}),
				name:  "Header-2",
				value: "value-2",
			},
			want: map[string][]string{
				"Header":   {"value"},
				"Header-2": {"value-2"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := WithCustomHeaderField(tt.args.ctx, tt.args.name, tt.args.value)
			got := HeaderFrom(ctx)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithCustomHeaderField() = %v, want %v", got, tt.want)
			}
		})
	}
}
