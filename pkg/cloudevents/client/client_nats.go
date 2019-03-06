package client

import (
	cloudeventsnats "github.com/cloudevents/sdk-go/pkg/cloudevents/transport/nats"
	"github.com/nats-io/go-nats"
)

func NewNATSClient(natsServer, subject string, opts ...Option) (Client, error) {
	conn, err := nats.Connect(natsServer)
	if err != nil {
		return nil, err
	}
	transport := cloudeventsnats.Transport{
		Conn:    conn,
		Subject: subject,
	}
	c := &ceClient{
		transport: &transport,
	}

	if err := c.applyOptions(opts...); err != nil {
		return nil, err
	}

	return c, nil
}
