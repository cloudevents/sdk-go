package context_test

import (
	"context"
	cecontext "github.com/cloudevents/sdk-go/pkg/cloudevents/context"
	"github.com/google/go-cmp/cmp"
	"net/url"
	"testing"
)

func TestContext(t *testing.T) {
	// TODO: add a test. This makes coverage count this dir.
}

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

func TestTransportContext(t *testing.T) {
	testCases := map[string]struct {
		transport interface{}
		ctx       context.Context
		want      interface{}
	}{
		"nil context": {},
		"nil context, set transport context": {
			transport: map[string]string{"hi": "unit test"},
			want:      map[string]string{"hi": "unit test"},
		},
		"todo context, set transport context": {
			ctx:       context.TODO(),
			transport: map[string]string{"hi": "unit test"},
			want:      map[string]string{"hi": "unit test"},
		},
		"bad transport context": {
			ctx: context.TODO(),
		},
		"already set transport context": {
			ctx:       cecontext.WithTransportContext(context.TODO(), map[string]string{"bye": "unit test"}),
			transport: map[string]string{"hi": "unit test"},
			want:      map[string]string{"hi": "unit test"},
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			ctx := cecontext.WithTransportContext(tc.ctx, tc.transport)

			got := cecontext.TransportContextFrom(ctx)

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("unexpected (-want, +got) = %v", diff)
			}
		})
	}
}
