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

func TestPortContextLoop(t *testing.T) {
	testCases := map[string]struct {
		ctx  context.Context
		port int
		want int
	}{
		"nil": {
			want: 8080,
		},
		"default": {
			ctx:  context.TODO(),
			want: 8080,
		},
		"loop": {
			ctx:  context.TODO(),
			port: 1337,
			want: 1337,
		},
		"loop, nil context": {
			port: 1337,
			want: 1337,
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			ctx := client.ContextWithPort(tc.ctx, tc.want)

			got := client.PortFromContext(ctx)

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("unexpected (-want, +got) = %v", diff)
			}
		})
	}
}
