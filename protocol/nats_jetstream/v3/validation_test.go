/*
 Copyright 2024 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package nats_jetstream

import (
	"testing"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

func Test_validateOptions(t *testing.T) {
	url := "host.com"
	type args struct {
		protocol *Protocol
	}
	type wants struct {
		err error
	}
	tests := []struct {
		name  string
		args  args
		wants wants
	}{
		{
			name: "valid protocol with URL",
			args: args{
				protocol: &Protocol{
					url: url,
				},
			},
			wants: wants{
				err: nil,
			},
		},
		{
			name: "invalid protocol with URL and Connection both provided",
			args: args{
				protocol: &Protocol{
					url:  url,
					conn: &nats.Conn{},
				},
			},
			wants: wants{
				err: newValidationError(fieldURL, messageConflictingConnection),
			},
		},
		{
			name: "invalid protocol without URL or connection",
			args: args{
				protocol: &Protocol{},
			},
			wants: wants{
				err: newValidationError(fieldURL, messageNoConnection),
			},
		},
		{
			name: "invalid protocol too many consumer options",
			args: args{
				protocol: &Protocol{
					url:                   url,
					consumerConfig:        &jetstream.ConsumerConfig{},
					orderedConsumerConfig: &jetstream.OrderedConsumerConfig{},
				},
			},
			wants: wants{
				err: newValidationError(fieldConsumerConfig, messageMoreThanOneConsumerConfig),
			},
		},
		{
			name: "invalid protocol receiver options without config",
			args: args{
				protocol: &Protocol{
					url: url,
					pullConsumeOpts: []jetstream.PullConsumeOpt{
						jetstream.PullMaxMessages(1),
						jetstream.PullMaxBytes(0),
					},
				},
			},
			wants: wants{
				err: newValidationError(fieldPullConsumerOpts, messageReceiverOptionsWithoutConfig),
			},
		},
		{
			name: "invalid protocol sender options without send subject",
			args: args{
				protocol: &Protocol{
					url: url,
					publishOpts: []jetstream.PublishOpt{
						jetstream.WithMsgID(""),
						jetstream.WithRetryAttempts(1),
					},
				},
			},
			wants: wants{
				err: newValidationError(fieldPublishOptions, messageSenderOptionsWithoutSubject),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := validateOptions(tt.args.protocol)
			if gotErr != tt.wants.err {
				t.Errorf("validateOptions() = %v, want %v", gotErr, tt.wants.err)
			}
		})
	}
}
