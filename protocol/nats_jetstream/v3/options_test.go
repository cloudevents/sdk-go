/*
 Copyright 2024 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package nats_jetstream

import (
	"crypto/tls"
	"reflect"
	"testing"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

func TestWithURL(t *testing.T) {
	expectedURL := "host.com"
	type args struct {
		protocol *Protocol
		url      string
	}
	type wants struct {
		err      error
		protocol *Protocol
	}
	tests := []struct {
		name  string
		args  args
		wants wants
	}{
		{
			name: "URL given",
			args: args{
				protocol: &Protocol{},
				url:      expectedURL,
			},
			wants: wants{
				err: nil,
				protocol: &Protocol{
					url: expectedURL,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := tt.args.protocol.applyOptions(WithURL(tt.args.url))
			if gotErr != tt.wants.err {
				t.Errorf("applyOptions(WithURL()) = %v, want %v", gotErr, tt.wants.err)
			}

			if tt.args.protocol.url != tt.wants.protocol.url {
				t.Errorf("p = %v, want %v", tt.args.protocol.url, tt.wants.protocol.url)
			}
			if len(tt.args.protocol.natsOpts) != len(tt.wants.protocol.natsOpts) {
				t.Errorf("p = %v, want %v", tt.args.protocol.natsOpts, tt.wants.protocol.natsOpts)
			}
		})
	}
}

func TestWithNatsOptions(t *testing.T) {
	userJWTAndSeed := nats.UserJWTAndSeed("jwt", "seed")
	secure := nats.Secure(&tls.Config{})
	type args struct {
		protocol *Protocol
		options  []nats.Option
	}
	type wants struct {
		err      error
		protocol *Protocol
	}
	tests := []struct {
		name  string
		args  args
		wants wants
	}{
		{
			name: "two NATS options given",
			args: args{
				protocol: &Protocol{},
				options:  []nats.Option{userJWTAndSeed, secure},
			},
			wants: wants{
				err: nil,
				protocol: &Protocol{
					natsOpts: []nats.Option{userJWTAndSeed, secure},
				},
			},
		},
		{
			name: "empty NATS options given",
			args: args{
				protocol: &Protocol{},
				options:  []nats.Option{},
			},
			wants: wants{
				err: nil,
				protocol: &Protocol{
					natsOpts: []nats.Option{},
				},
			},
		},
		{
			name: "nil NATS options given",
			args: args{
				protocol: &Protocol{},
				options:  nil,
			},
			wants: wants{
				err: nil,
				protocol: &Protocol{
					natsOpts: nil,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := tt.args.protocol.applyOptions(WithNatsOptions(tt.args.options))
			if gotErr != tt.wants.err {
				t.Errorf("applyOptions(WithNatsOptions()) = %v, want %v", gotErr, tt.wants.err)
			}

			if tt.args.protocol.url != tt.wants.protocol.url {
				t.Errorf("p = %v, want %v", tt.args.protocol.url, tt.wants.protocol.url)
			}
			if len(tt.args.protocol.natsOpts) != len(tt.wants.protocol.natsOpts) {
				t.Errorf("p = %v, want %v", tt.args.protocol.natsOpts, tt.wants.protocol.natsOpts)
			}
		})
	}
}

func TestWithConnection(t *testing.T) {
	natsConn := &nats.Conn{}
	type args struct {
		protocol *Protocol
		conn     *nats.Conn
	}
	type wants struct {
		err      error
		protocol *Protocol
	}
	tests := []struct {
		name  string
		args  args
		wants wants
	}{
		{
			name: "nats connection given",
			args: args{
				protocol: &Protocol{},
				conn:     natsConn,
			},
			wants: wants{
				err: nil,
				protocol: &Protocol{
					conn: natsConn,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := tt.args.protocol.applyOptions(WithConnection(tt.args.conn))
			if gotErr != tt.wants.err {
				t.Errorf("applyOptions(WithConnection()) = %v, want %v", gotErr, tt.wants.err)
			}

			if !reflect.DeepEqual(tt.args.protocol, tt.wants.protocol) {
				t.Errorf("p = %v, want %v", tt.args.protocol, tt.wants.protocol)
			}
		})
	}
}

func TestWithConsumerConfig(t *testing.T) {
	filterSubjects := []string{"normal"}
	type args struct {
		protocol *Protocol
		config   *jetstream.ConsumerConfig
	}
	type wants struct {
		err      error
		protocol *Protocol
	}
	tests := []struct {
		name  string
		args  args
		wants wants
	}{
		{
			name: "consumer config given",
			args: args{
				protocol: &Protocol{},
				config:   &jetstream.ConsumerConfig{FilterSubjects: filterSubjects},
			},
			wants: wants{
				err: nil,
				protocol: &Protocol{
					consumerConfig: &jetstream.ConsumerConfig{FilterSubjects: filterSubjects},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := tt.args.protocol.applyOptions(WithConsumerConfig(tt.args.config))
			if gotErr != tt.wants.err {
				t.Errorf("applyOptions(WithConsumerConfig()) = %v, want %v", gotErr, tt.wants.err)
			}

			if !reflect.DeepEqual(tt.args.protocol, tt.wants.protocol) {
				t.Errorf("p = %v, want %v", tt.args.protocol, tt.wants.protocol)
			}
		})
	}
}

func TestWithOrderedConsumerConfig(t *testing.T) {
	filterSubjects := []string{"ordered"}
	type args struct {
		protocol *Protocol
		config   *jetstream.OrderedConsumerConfig
	}
	type wants struct {
		err      error
		protocol *Protocol
	}
	tests := []struct {
		name  string
		args  args
		wants wants
	}{
		{
			name: "ordered consumer given",
			args: args{
				protocol: &Protocol{},
				config:   &jetstream.OrderedConsumerConfig{FilterSubjects: filterSubjects},
			},
			wants: wants{
				err: nil,
				protocol: &Protocol{
					orderedConsumerConfig: &jetstream.OrderedConsumerConfig{FilterSubjects: filterSubjects},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := tt.args.protocol.applyOptions(WithOrderedConsumerConfig(tt.args.config))
			if gotErr != tt.wants.err {
				t.Errorf("applyOptions(WithOrderedConsumerConfig()) = %v, want %v", gotErr, tt.wants.err)
			}

			if !reflect.DeepEqual(tt.args.protocol, tt.wants.protocol) {
				t.Errorf("p = %v, want %v", tt.args.protocol, tt.wants.protocol)
			}
		})
	}
}

func TestWithPullConsumeOptions(t *testing.T) {
	maxMessages := jetstream.PullMaxMessages(1)
	maxBytes := jetstream.PullMaxBytes(0)
	type args struct {
		protocol *Protocol
		config   []jetstream.PullConsumeOpt
	}
	type wants struct {
		err      error
		protocol *Protocol
	}
	tests := []struct {
		name  string
		args  args
		wants wants
	}{
		{
			name: "pull consumer option given",
			args: args{
				protocol: &Protocol{},
				config:   []jetstream.PullConsumeOpt{maxMessages, maxBytes},
			},
			wants: wants{
				err: nil,
				protocol: &Protocol{
					pullConsumeOpts: []jetstream.PullConsumeOpt{maxMessages, maxBytes},
				},
			},
		},
		{
			name: "empty pull consumer option given",
			args: args{
				protocol: &Protocol{},
				config:   []jetstream.PullConsumeOpt{},
			},
			wants: wants{
				err:      nil,
				protocol: &Protocol{pullConsumeOpts: []jetstream.PullConsumeOpt{}},
			},
		},
		{
			name: "nil pull consumer option given",
			args: args{
				protocol: &Protocol{},
				config:   nil,
			},
			wants: wants{
				err:      nil,
				protocol: &Protocol{pullConsumeOpts: nil},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := tt.args.protocol.applyOptions(WithPullConsumerOptions(tt.args.config))
			if gotErr != tt.wants.err {
				t.Errorf("applyOptions(WithPullConsumerOptions()) = %v, want %v", gotErr, tt.wants.err)
			}

			if !reflect.DeepEqual(tt.args.protocol, tt.wants.protocol) {
				t.Errorf("p = %v, want %v", tt.args.protocol, tt.wants.protocol)
			}
		})
	}
}

func TestWithSendSubject(t *testing.T) {
	type args struct {
		protocol    *Protocol
		sendSubject string
	}
	type wants struct {
		err      error
		protocol *Protocol
	}
	tests := []struct {
		name  string
		args  args
		wants wants
	}{
		{
			name: "sendSubject given",
			args: args{
				protocol:    &Protocol{},
				sendSubject: "validSubject",
			},
			wants: wants{
				err: nil,
				protocol: &Protocol{
					sendSubject: "validSubject",
				},
			},
		},
		{
			name: "no send subject given",
			args: args{
				protocol:    &Protocol{},
				sendSubject: "",
			},
			wants: wants{
				err:      nil,
				protocol: &Protocol{sendSubject: ""},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := tt.args.protocol.applyOptions(WithSendSubject(tt.args.sendSubject))
			if gotErr != tt.wants.err {
				t.Errorf("applyOptions(WithSendSubject()) = %v, want %v", gotErr, tt.wants.err)
			}

			if !reflect.DeepEqual(tt.args.protocol, tt.wants.protocol) {
				t.Errorf("p = %v, want %v", tt.args.protocol, tt.wants.protocol)
			}
		})
	}
}

func TestWithPublishOptions(t *testing.T) {
	withMsgID := jetstream.WithMsgID("")
	withRetryAttempts := jetstream.WithRetryAttempts(1)
	publishOptions := []jetstream.PublishOpt{withMsgID, withRetryAttempts}
	type args struct {
		protocol *Protocol
		options  []jetstream.PublishOpt
	}
	type wants struct {
		err      error
		protocol *Protocol
	}
	tests := []struct {
		name  string
		args  args
		wants wants
	}{
		{
			name: "publish options given",
			args: args{
				protocol: &Protocol{},
				options:  publishOptions,
			},
			wants: wants{
				err: nil,
				protocol: &Protocol{
					publishOpts: publishOptions,
				},
			},
		},
		{
			name: "empty publish options given",
			args: args{
				protocol: &Protocol{},
				options:  []jetstream.PublishOpt{},
			},
			wants: wants{
				err:      nil,
				protocol: &Protocol{publishOpts: []jetstream.PublishOpt{}},
			},
		},
		{
			name: "nil publish options given",
			args: args{
				protocol: &Protocol{},
				options:  nil,
			},
			wants: wants{
				err:      nil,
				protocol: &Protocol{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := tt.args.protocol.applyOptions(WithPublishOptions(tt.args.options))
			if gotErr != tt.wants.err {
				t.Errorf("applyOptions(WithPublishOptions()) = %v, want %v", gotErr, tt.wants.err)
			}

			if !reflect.DeepEqual(tt.args.protocol, tt.wants.protocol) {
				t.Errorf("p = %v, want %v", tt.args.protocol, tt.wants.protocol)
			}
		})
	}
}

func TestWithJetStreamOptions(t *testing.T) {
	withClientTrace := jetstream.WithClientTrace(nil)
	withPublishAsyncMaxPending := jetstream.WithPublishAsyncMaxPending(1)
	jetStreamOpts := []jetstream.JetStreamOpt{withClientTrace, withPublishAsyncMaxPending}
	type args struct {
		protocol *Protocol
		options  []jetstream.JetStreamOpt
	}
	type wants struct {
		err      error
		protocol *Protocol
	}
	tests := []struct {
		name  string
		args  args
		wants wants
	}{
		{
			name: "jetstream options given",
			args: args{
				protocol: &Protocol{},
				options:  jetStreamOpts,
			},
			wants: wants{
				err: nil,
				protocol: &Protocol{
					jetStreamOpts: jetStreamOpts,
				},
			},
		},
		{
			name: "no jetstream options given",
			args: args{
				protocol: &Protocol{},
				options:  nil,
			},
			wants: wants{
				err:      nil,
				protocol: &Protocol{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := tt.args.protocol.applyOptions(WithJetStreamOptions(tt.args.options))
			if gotErr != tt.wants.err {
				t.Errorf("applyOptions(WithJetStreamOptions()) = %v, want %v", gotErr, tt.wants.err)
			}

			if !reflect.DeepEqual(tt.args.protocol, tt.wants.protocol) {
				t.Errorf("p = %v, want %v", tt.args.protocol, tt.wants.protocol)
			}
		})
	}
}
