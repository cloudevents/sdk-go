package eventbridge

import (
	"context"
	"fmt"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/protocol"

	sdkeb "github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
)

const (
	defaultEventBusName = "default"
)

type Protocol struct {
	client       *sdkeb.Client
	eventBusName string
}

// New creates a new EventBridge protocol.
func New(opts ...Option) (*Protocol, error) {
	p := &Protocol{
		eventBusName: defaultEventBusName,
	}
	if err := p.applyOptions(opts...); err != nil {
		return nil, err
	}
	if p.client == nil {
		return nil, fmt.Errorf("eventbridge client is nil")
	}
	return p, nil
}

func (p *Protocol) applyOptions(opts ...Option) error {
	for _, fn := range opts {
		if err := fn(p); err != nil {
			return err
		}
	}
	return nil
}

// Send sends messages. Send implements Sender.Sender
func (p *Protocol) Send(ctx context.Context, in binding.Message, transformers ...binding.Transformer) (err error) {
	defer func() { _ = in.Finish(err) }()
	entryInput := types.PutEventsRequestEntry{
		EventBusName: &p.eventBusName,
	}
	err = WriteMsgInput(ctx, in, &entryInput, transformers...)
	if err != nil {
		return err
	}

	input := &sdkeb.PutEventsInput{
		Entries: []types.PutEventsRequestEntry{entryInput},
	}

	_, err = p.client.PutEvents(ctx, input)
	return err
}

var _ protocol.Sender = (*Protocol)(nil)
