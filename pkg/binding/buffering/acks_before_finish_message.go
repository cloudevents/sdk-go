package buffering

import (
	"sync/atomic"

	"github.com/cloudevents/sdk-go/pkg/binding"
)

type acksMessage struct {
	binding.Message
	requiredAcks int32
}

func (m *acksMessage) GetParent() binding.Message {
	return m.Message
}

func (m *acksMessage) Finish(err error) error {
	remainingAcks := atomic.AddInt32(&m.requiredAcks, -1)
	if remainingAcks == 0 {
		return m.Message.Finish(err)
	}
	return nil
}

// WithAcksBeforeFinish returns a wrapper for m that calls m.Finish()
// only after the specified number of acks are received.
// Use it when you need to route a Message to more Sender instances
func WithAcksBeforeFinish(m binding.Message, requiredAcks int) binding.Message {
	return &acksMessage{Message: m, requiredAcks: int32(requiredAcks)}
}
