/*
Copyright 2023 The CloudEvents Authors
SPDX-License-Identifier: Apache-2.0
*/

package sqs

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

// Option provides a way to configure the protocol
type Option func(*Protocol) error

func WithNewClientFromConfig(cfg aws.Config, optFns ...func(*sqs.Options)) Option {
	return func(p *Protocol) error {
		p.client = sqs.NewFromConfig(cfg, optFns...)
		return nil
	}
}

func WithClient(client *sqs.Client) Option {
	return func(p *Protocol) error {
		if client == nil {
			return fmt.Errorf("client cannot be nil")
		}
		p.client = client
		return nil
	}
}

func WithVisibilityTimeout(visibilityTimeout int32) Option {
	return func(p *Protocol) error {
		if visibilityTimeout <= 0 {
			return fmt.Errorf("visibilityTimeout must be greater than 0")
		}
		p.visibilityTimeout = visibilityTimeout
		return nil
	}
}

func WithMaxMessages(maxMessages int32) Option {
	return func(p *Protocol) error {
		if maxMessages <= 0 {
			return fmt.Errorf("maxMessages must be greater than 0")
		}
		if maxMessages > 10 {
			return fmt.Errorf("maxMessages must be less than 10")
		}
		p.maxMessages = maxMessages
		return nil
	}
}

func WithWaitTimeSeconds(waitTimeSeconds int32) Option {
	return func(p *Protocol) error {
		if waitTimeSeconds <= 0 {
			return fmt.Errorf("waitTimeSeconds must be greater than 0 second")
		}
		if waitTimeSeconds > 20 {
			return fmt.Errorf("waitTimeSeconds must be less than 20 seconds")
		}
		p.waitTimeSeconds = waitTimeSeconds
		return nil
	}
}
