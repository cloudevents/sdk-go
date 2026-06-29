package eventbridge

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
)

// Option provides a way to configure the protocol
type Option func(*Protocol) error

func WithNewClientFromConfig(cfg aws.Config, optFns ...func(*eventbridge.Options)) Option {
	return func(p *Protocol) error {
		p.client = eventbridge.NewFromConfig(cfg, optFns...)
		return nil
	}
}

func WithClient(client *eventbridge.Client) Option {
	return func(p *Protocol) error {
		if client == nil {
			return fmt.Errorf("client cannot be nil")
		}
		p.client = client
		return nil
	}
}

func WithEventBusName(eventBusName string) Option {
	return func(p *Protocol) error {
		if eventBusName == "" {
			return fmt.Errorf("event bus name cannot be empty")
		}
		p.eventBusName = eventBusName
		return nil
	}
}
