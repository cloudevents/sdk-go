package http_test

import (
	"context"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport/http"
	"github.com/google/go-cmp/cmp"
	"testing"
)

func TestTransportContext(t *testing.T) {
	testCases := map[string]struct {
		t    http.TransportContext
		ctx  context.Context
		want http.TransportContext
	}{
		"nil context": {},
		"nil context, set transport context": {
			t: http.TransportContext{
				Host:   "unit test",
				Method: "unit test",
			},
			want: http.TransportContext{
				Host:   "unit test",
				Method: "unit test",
			},
		},
		"todo context, set transport context": {
			ctx: context.TODO(),
			t: http.TransportContext{
				Host:   "unit test",
				Method: "unit test",
			},
			want: http.TransportContext{
				Host:   "unit test",
				Method: "unit test",
			},
		},
		"bad transport context": {
			ctx: context.TODO(),
		},
		"already set transport context": {
			ctx: http.WithTransportContext(context.TODO(),
				http.TransportContext{
					Host:   "existing test",
					Method: "exiting test",
				}),
			t: http.TransportContext{
				Host:   "unit test",
				Method: "unit test",
			},
			want: http.TransportContext{
				Host:   "unit test",
				Method: "unit test",
			},
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			ctx := http.WithTransportContext(tc.ctx, tc.t)

			got := http.TransportContextFrom(ctx)

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("unexpected (-want, +got) = %v", diff)
			}
		})
	}
}
