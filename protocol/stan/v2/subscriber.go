package stan

import "github.com/nats-io/stan.go"

// The Subscriber interface allows us to configure how the subscription is created
type Subscriber interface {
	Subscribe(conn stan.Conn, subject string, cb stan.MsgHandler,
		opts ...stan.SubscriptionOption) (stan.Subscription, error)
}

// RegularSubscriber creates regular subscriptions
type RegularSubscriber struct {
}

// Subscribe implements Subscriber.Subscribe
func (s *RegularSubscriber) Subscribe(conn stan.Conn, subject string, cb stan.MsgHandler,
	opts ...stan.SubscriptionOption) (stan.Subscription, error) {
	return conn.Subscribe(subject, cb, opts...)
}

var _ Subscriber = (*RegularSubscriber)(nil)

// QueueSubscriber creates queue subscriptions
type QueueSubscriber struct {
	QueueGroup string
}

// Subscribe implements Subscriber.Subscribe
func (s *QueueSubscriber) Subscribe(conn stan.Conn, subject string, cb stan.MsgHandler,
	opts ...stan.SubscriptionOption) (stan.Subscription, error) {
	return conn.QueueSubscribe(subject, s.QueueGroup, cb, opts...)
}

var _ Subscriber = (*QueueSubscriber)(nil)
