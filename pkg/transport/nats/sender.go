package nats

import (
	"context"
	"github.com/cloudevents/sdk-go/pkg/binding"
	bindings "github.com/cloudevents/sdk-go/pkg/transport"
	"github.com/nats-io/nats.go"
)

// sender implements binding.Sender
type sender struct {
	conn         *nats.Conn
	subject      string
	transformers binding.TransformerFactories
}

func (s *sender) Send(ctx context.Context, in binding.Message) error {
	msg := &nats.Msg{}
	if err := WriteNATSMessage(ctx, in, msg, nil); err != nil {
		return err
	}
	msg.Subject = s.subject // TODO: allow for overwriting this.
	return s.conn.PublishMsg(msg)
}

func (s *sender) Close(ctx context.Context) error {
	s.conn.Close()
	return nil
}

// Create a new nats Sender, implements binding.Sender
func NewSender(conn *nats.Conn, subject string, options ...SenderOptionFunc) bindings.Sender {
	s := &sender{conn: conn, subject: subject, transformers: make(binding.TransformerFactories, 0)}
	for _, o := range options {
		o(s)
	}
	return s
}
