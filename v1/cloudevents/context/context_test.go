package context_test

import (
	"context"
	"net/url"
	"testing"

	cecontext "github.com/cloudevents/sdk-go/v1/cloudevents/context"
	"github.com/google/go-cmp/cmp"
)

func TestTargetContext(t *testing.T) {
	exampleDotCom, _ := url.Parse("http://example.com")

	testCases := map[string]struct {
		target string
		ctx    context.Context
		want   *url.URL
	}{
		"nil context": {},
		"nil context, set url": {
			target: "http://example.com",
			want:   exampleDotCom,
		},
		"todo context, set url": {
			ctx:    context.TODO(),
			target: "http://example.com",
			want:   exampleDotCom,
		},
		"bad url": {
			ctx:    context.TODO(),
			target: "%",
		},
		"already set target": {
			ctx:    cecontext.WithTarget(context.TODO(), "http://example2.com"),
			target: "http://example.com",
			want:   exampleDotCom,
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			ctx := cecontext.WithTarget(tc.ctx, tc.target)

			got := cecontext.TargetFrom(ctx)

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("unexpected (-want, +got) = %v", diff)
			}
		})
	}
}

func TestEncodingContext(t *testing.T) {
	testCases := map[string]struct {
		encoding string
		ctx      context.Context
		want     string
	}{
		"nil context": {},
		"nil context, set encoding": {
			encoding: "foo",
			want:     "foo",
		},
		"todo context, set encoding": {
			ctx:      context.TODO(),
			encoding: "foo",
			want:     "foo",
		},
		"already set encoding": {
			ctx:      cecontext.WithTarget(context.TODO(), "foo"),
			encoding: "bar",
			want:     "bar",
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			ctx := cecontext.WithEncoding(tc.ctx, tc.encoding)

			got := cecontext.EncodingFrom(ctx)

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("unexpected (-want, +got) = %v", diff)
			}
		})
	}
}
