/*
 Copyright 2024 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package nats_jetstream

import (
	"testing"

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
			name: "invalid protocol without connection",
			args: args{
				protocol: &Protocol{},
			},
			wants: wants{
				err: ErrNoConnection,
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
				err: ErrMoreThanOneConsumerConfig,
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
				err: ErrReceiverOptionsWithoutConfig,
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
				err: ErrSenderOptionsWithoutSubject,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := tt.args.protocol.validateOptions()
			if gotErr != tt.wants.err {
				t.Errorf("validateOptions() = %v, want %v", gotErr, tt.wants.err)
			}
		})
	}
}
