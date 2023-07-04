package mqtt_paho

import (
	"context"
	"testing"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/eclipse/paho.golang/paho"
)

func TestReadStructured(t *testing.T) {
	tests := []struct {
		name    string
		msg     *paho.Publish
		wantErr error
	}{
		{
			name: "nil format",
			msg: &paho.Publish{
				Payload: []byte(""),
			},
			wantErr: binding.ErrNotStructured,
		},
		{
			name: "json format",
			msg: &paho.Publish{
				Payload:    []byte(""),
				Properties: &paho.PublishProperties{ContentType: event.ApplicationCloudEventsJSON},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			msg := NewMessage(tc.msg)
			err := msg.ReadStructured(context.Background(), (*pubMessageWriter)(tc.msg))
			if err != tc.wantErr {
				t.Errorf("Error unexpected. got: %v, want: %v", err, tc.wantErr)
			}
		})
	}
}

func TestReadBinary(t *testing.T) {
	// msg := &paho.Publish{
	// 	Payload:    []byte("{hello:world}"),
	// 	Properties: &paho.PublishProperties{ContentType: event.ApplicationCloudEventsJSON},
	// }

	msg := &paho.Publish{
		Payload: []byte("{hello:world}"),
		Properties: &paho.PublishProperties{
			User: []paho.UserProperty{
				{Key: "ce-specversion", Value: "1.0"},
				{Key: "ce-type", Value: "binary.test"},
				{Key: "ce-source", Value: "test-source"},
				{Key: "ce-id", Value: "ABC-123"},
			},
		},
	}

	message := NewMessage(msg)
	err := message.ReadBinary(context.Background(), (*pubMessageWriter)(msg))
	if err != nil {
		t.Errorf("Error unexpected. got: %v", err)
	}
}
