/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package stan

import (
	"context"
	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/nats-io/stan.go"
	"github.com/nats-io/stan.go/pb"
	"reflect"
	"testing"
)

func TestMetadataContextDecorator(t *testing.T) {
	type args struct {
		msg binding.Message
	}
	tests := []struct {
		name  string
		args  args
		want  MsgMetadata
		want1 bool
	}{
		{
			name: "STAN message contains metadata on context",
			args: args{
				msg: newSTANMessage(&stan.Msg{
					MsgProto: pb.MsgProto{
						Sequence:        13,
						Redelivered:     true,
						RedeliveryCount: 42,
					},
				}),
			},
			want: MsgMetadata{
				Sequence:        13,
				Redelivered:     true,
				RedeliveryCount: 42,
			},
			want1: true,
		},
		{
			name: "non-STAN message does not contain metadata on context",
			args: args{
				msg: fakeMessage{},
			},
			want:  MsgMetadata{},
			want1: false,
		},
	}

	decorator := MetadataContextDecorator()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := decorator(context.Background(), tt.args.msg)
			got, got1 := MessageMetadataFrom(ctx)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MessageMetadataFrom() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("MessageMetadataFrom() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

// newMessage wraps NewMessage and ignores any errors
func newSTANMessage(msg *stan.Msg) binding.Message {
	m, _ := NewMessage(msg)
	return m
}

// fakeMessage implements binding.Message for tests
type fakeMessage struct {
}

func (f fakeMessage) ReadEncoding() binding.Encoding {
	panic("implement me")
}

func (f fakeMessage) ReadStructured(context.Context, binding.StructuredWriter) error {
	panic("implement me")
}

func (f fakeMessage) ReadBinary(context.Context, binding.BinaryWriter) error {
	panic("implement me")
}

func (f fakeMessage) Finish(error) error {
	panic("implement me")
}
