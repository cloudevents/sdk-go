package nats

import (
	"github.com/cloudevents/sdk-go/pkg/cloudevents/client"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport/nats"
)

func sortOptions(opts ...interface{}) ([]nats.Option, []client.Option) {
	tOpts := []nats.Option(nil)
	cOpts := []client.Option(nil)

	for _, o := range opts {
		if opt, ok := o.(nats.Option); ok {
			tOpts = append(tOpts, opt)
		} else if opt, ok := o.(client.Option); ok {
			cOpts = append(cOpts, opt)
		}
	}
	return tOpts, cOpts
}

// New returns a new NATS client
func New(natsServer, subject string, opts ...interface{}) (client.Client, error) {
	tOpts, cOpts := sortOptions(opts...)

	t, err := nats.New(natsServer, subject, tOpts...)
	if err != nil {
		return nil, err
	}
	c, err := client.New(t, cOpts...)
	if err != nil {
		return nil, err
	}
	return c, nil
}
