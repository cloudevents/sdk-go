package client_test

import (
	"context"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/client"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"testing"
)

func TestClientContextLoop(t *testing.T) {
	testCases := map[string]struct {
		ctx    context.Context
		client *client.Client
		want   *client.Client
	}{
		"nil": {
			want: nil,
		},
		"empty": {
			ctx:    context.TODO(),
			client: &client.Client{},
			want:   &client.Client{},
		},
		"nil context": {
			ctx:    nil,
			client: &client.Client{},
			want:   &client.Client{},
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			ctx := client.ContextWithClient(tc.ctx, tc.want)

			got := client.ClientFromContext(ctx)

			if diff := cmp.Diff(tc.want, got, cmpopts.IgnoreUnexported(client.Client{})); diff != "" {
				t.Errorf("unexpected (-want, +got) = %v", diff)
			}
		})
	}
}
