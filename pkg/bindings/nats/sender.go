package nats

import (
	"context"
	"fmt"
	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/nats-io/nats.go"
)

type Sender struct {
	conn                 *nats.Conn
	subject              string
	transformerFactories binding.TransformerFactories
}

func NewSender(conn *nats.Conn, subject string, options ...SenderOptionFunc) *Sender {
	s := &Sender{
		conn:                 conn,
		subject:              subject,
		transformerFactories: make(binding.TransformerFactories, 0),
	}

	for _, optionFunc := range options {
		optionFunc(s)
	}

	return s
}

func (s *Sender) Send(ctx context.Context, m binding.Message) (err error) {
	defer func() { _ = m.Finish(err) }()

	if s.conn == nil || s.subject == "" {
		return fmt.Errorf("not initialized: %#v", s)
	}

	msg := &nats.Msg{
		Subject: s.subject,
	}

	if err = EncodeNatsMsg(ctx, m, msg, nil); err != nil {
		return err
	}

	return s.conn.PublishMsg(msg)
}

var _ binding.Sender = (*Sender)(nil)
