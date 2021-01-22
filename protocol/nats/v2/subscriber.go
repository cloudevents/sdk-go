package nats

import (
	"github.com/nats-io/jsm.go"
	"github.com/nats-io/nats.go"
	"time"
)

// The Subscriber interface allows us to configure how the subscription is created
type Subscriber interface {
	Subscribe(conn *nats.Conn, subject string, cb nats.MsgHandler) (Subscription, error)
}

// The Subscription interface allows us to drain a subscription when closed
type Subscription interface {
	Drain() error
}

// RegularSubscriber creates regular subscriptions
type RegularSubscriber struct {
}

// Subscribe implements Subscriber.Subscribe
func (s *RegularSubscriber) Subscribe(conn *nats.Conn, subject string, cb nats.MsgHandler) (Subscription, error) {
	return conn.Subscribe(subject, cb)
}

var _ Subscriber = (*RegularSubscriber)(nil)

// QueueSubscriber creates queue subscriptions
type QueueSubscriber struct {
	Queue string
}

// Subscribe implements Subscriber.Subscribe
func (s *QueueSubscriber) Subscribe(conn *nats.Conn, subject string, cb nats.MsgHandler) (Subscription, error) {
	return conn.QueueSubscribe(subject, s.Queue, cb)
}

var _ Subscriber = (*QueueSubscriber)(nil)

// PullConsumer creates queue subscriptions
type PullConsumer struct {
	stopCh chan struct{}
	Stream string
}

func NewPullConsumer(stream string) Subscriber {
	return &PullConsumer{
		stopCh: make(chan struct{}, 1),
		Stream: stream,
	}
}

const defaultTimeout = 10 * time.Second

// Subscribe implements Subscriber.Subscribe
func (s *PullConsumer) Subscribe(conn *nats.Conn, consumer string, cb nats.MsgHandler) (Subscription, error) {
	mgr, err := jsm.New(conn, jsm.WithTimeout(defaultTimeout))
	if err != nil {
		return nil, err
	}

	c, err := mgr.LoadConsumer(s.Stream, consumer)
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			select {
			case <-s.stopCh:
				break
			default:
				msg, err := c.NextMsg()
				if err == nats.ErrTimeout {
					continue
				}
				if err == nil {
					cb(msg)
					_ = msg.Respond(nil)
				}
			}
		}
	}()
	return s, nil
}

func (s *PullConsumer) Drain() error {
	close(s.stopCh)
	return nil
}

var _ Subscriber = (*PullConsumer)(nil)
