package test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/errgroup"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/protocol"
	"github.com/cloudevents/sdk-go/v2/test"
)

// BenchmarkSendReceive implements a simple send/receive benchmark.
// Requires a sender and receiver that are connected to each other.
func BenchmarkSendReceive(b *testing.B, s protocol.Sender, r protocol.Receiver) {
	e := test.FullEvent()
	m := (*binding.EventMessage)(&e)
	ctx := context.Background()
	b.ResetTimer() // Don't count setup.
	for i := 0; i < b.N; i++ {
		n := 10 // Messages to send async.
		g := errgroup.Group{}
		g.Go(func() error {
			for j := 0; j < n; j++ {
				if err := s.Send(ctx, m); err != nil {
					return err
				}
			}
			return nil
		})
		g.Go(func() error {
			for j := 0; j < n; j++ {
				m, err := r.Receive(ctx)
				if err != nil {
					return err
				}
				if err := m.Finish(nil); err != nil {
					return err
				}
			}
			return nil
		})
		assert.NoError(b, g.Wait())
	}
}
