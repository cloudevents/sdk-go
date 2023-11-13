/*
Copyright 2023 The CloudEvents Authors
SPDX-License-Identifier: Apache-2.0
*/

package grpc

import (
	"context"
	"testing"

	"github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/event"
)

func TestReadStructured(t *testing.T) {
	tests := []struct {
		name    string
		msg     *pb.CloudEvent
		wantErr error
	}{
		{
			name:    "nil format",
			msg:     &pb.CloudEvent{},
			wantErr: binding.ErrNotStructured,
		},
		{
			name: "json format",
			msg: &pb.CloudEvent{
				Attributes: map[string]*pb.CloudEventAttributeValue{
					contenttype: &pb.CloudEventAttributeValue{
						Attr: &pb.CloudEventAttributeValue_CeString{
							CeString: event.ApplicationCloudEventsJSON,
						},
					},
				},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			msg := NewMessage(tc.msg)
			err := msg.ReadStructured(context.Background(), (*pbEventWriter)(tc.msg))
			if err != tc.wantErr {
				t.Errorf("Error unexpected. got: %v, want: %v", err, tc.wantErr)
			}
		})
	}
}

func TestReadBinary(t *testing.T) {
	msg := &pb.CloudEvent{
		SpecVersion: "1.0",
		Id:          "ABC-123",
		Source:      "test-source",
		Type:        "binary.test",
		Attributes:  map[string]*pb.CloudEventAttributeValue{},
		Data: &pb.CloudEvent_BinaryData{
			BinaryData: []byte("{hello:world}"),
		},
	}

	message := NewMessage(msg)
	err := message.ReadBinary(context.Background(), (*pbEventWriter)(msg))
	if err != nil {
		t.Errorf("Error unexpected. got: %v", err)
	}
}
