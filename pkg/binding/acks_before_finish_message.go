package binding

import "sync/atomic"

type acksMessage struct {
	Message
	requiredAcks uint32
}

func (m *acksMessage) Finish(err error) error {
	remainingAcks := atomic.AddUint32(&m.requiredAcks, -1)
	if remainingAcks == 0 {
		return m.Message.Finish(err)
	}
	return nil
}

// WithAcksBeforeFinish returns a wrapper for m that calls m.Finish()
// only after the specified number of acks are received.
// Use it when you need to route a Message to more Sender instances
func WithAcksBeforeFinish(m Message, requiredAcks uint) Message {
	return &acksMessage{Message: m, requiredAcks: uint32(requiredAcks)}
}
