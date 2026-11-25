/*
 Copyright 2024 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package nats_jetstream

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"reflect"
	"testing"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/test"
	"github.com/nats-io/nats.go"
)

func Test_WriteMsg(t *testing.T) {
	type args struct {
		in binding.Message
	}
	type wants struct {
		err       error
		expHeader nats.Header
	}
	tests := []struct {
		name  string
		args  args
		wants wants
	}{
		{
			name: "valid protocol with URL",
			args: args{
				in: (*binding.EventMessage)(&testEvent),
			},
			wants: wants{
				err: nil,
				expHeader: nats.Header{
					"ce-type":       []string{"com.example.FullEvent"},
					"ce-source":     []string{test.Source.String()},
					"ce-id":         []string{"full-event"},
					"ce-time":       []string{test.Timestamp.String()},
					"ce-dataschema": []string{test.Schema.String()},
					"ce-subject":    []string{"topic"},
					"ce-exbool":     []string{fmt.Sprint(true)},
					"ce-exint":      []string{fmt.Sprint(42)},
					"ce-exstring":   []string{"exstring"},
					"ce-exbinary":   []string{base64.StdEncoding.EncodeToString([]byte{0, 1, 2, 3})},
					"ce-exurl":      []string{test.Source.String()},
					"ce-extime":     []string{test.Timestamp.String()},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writer := new(bytes.Buffer)
			gotHeader, gotErr := WriteMsg(context.Background(), tt.args.in, writer)
			if gotErr != tt.wants.err {
				t.Errorf("WriteMsg() = %v, want %v", gotErr, tt.wants.err)
			}
			for key, value := range tt.wants.expHeader {
				gotValue := gotHeader[key]
				if !reflect.DeepEqual(gotValue, value) {
					t.Errorf("WriteMsg() key %v got = %v, want %v", key, gotValue, value)
				}
			}
		})
	}
}

// Forwarding a message that arrived in binary mode hands the writer a string for
// every attribute, including time.
func Test_WriteMsg_ForwardBinaryMessage(t *testing.T) {
	gotHeader, err := WriteMsg(context.Background(), NewMessage(binaryReceiverMessage), new(bytes.Buffer))
	if err != nil {
		t.Fatalf("WriteMsg() error = %v", err)
	}
	if got, want := gotHeader.Get("ce-time"), test.Timestamp.String(); got != want {
		t.Errorf("WriteMsg() ce-time got = %v, want %v", got, want)
	}
	if got, want := gotHeader.Get("ce-id"), testEvent.ID(); got != want {
		t.Errorf("WriteMsg() ce-id got = %v, want %v", got, want)
	}
}
