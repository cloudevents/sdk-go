/*
 Copyright 2024 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package nats_jetstream

import (
	"context"
	"testing"

	"github.com/cloudevents/sdk-go/v2/test"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

// mockJetStream implements jetstream.JetStream interface for testing
type mockJetStream struct {
	streamNameBySubjectFunc func(ctx context.Context, subject string) (string, error)
	publishMsgFunc          func(ctx context.Context, msg *nats.Msg, opts ...jetstream.PublishOpt) (*jetstream.PubAck, error)
	jetstream.JetStream
}

func (m *mockJetStream) StreamNameBySubject(ctx context.Context, subject string) (string, error) {
	return m.streamNameBySubjectFunc(ctx, subject)
}

func (m *mockJetStream) PublishMsg(ctx context.Context, msg *nats.Msg, opts ...jetstream.PublishOpt) (*jetstream.PubAck, error) {
	return m.publishMsgFunc(ctx, msg, opts...)
}

func TestSend(t *testing.T) {
	tests := []struct {
		name        string
		sendSubject string
		ctx         context.Context
		wantSubject string
	}{
		{
			name:        "using p.sendSubject",
			sendSubject: "test.subject",
			ctx:         context.Background(),
			wantSubject: "test.subject",
		},
		{
			name:        "using WithSubject",
			sendSubject: "",
			ctx:         WithSubject(context.Background(), "test.ctxSubject"),
			wantSubject: "test.ctxSubject",
		},
		{
			name:        "using WithSubject and p.sendSubject",
			sendSubject: "test.subject",
			ctx:         WithSubject(context.Background(), "test.ctxSubject"),
			wantSubject: "test.ctxSubject",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockJS := &mockJetStream{
				streamNameBySubjectFunc: func(ctx context.Context, subject string) (string, error) {
					if subject != tt.wantSubject {
						t.Errorf("unexpected subject: got %s, want %s", subject, tt.wantSubject)
					}
					return "test-stream", nil
				},
				publishMsgFunc: func(ctx context.Context, msg *nats.Msg, opts ...jetstream.PublishOpt) (*jetstream.PubAck, error) {
					if msg.Subject != tt.wantSubject {
						t.Errorf("unexpected subject in publish: got %s, want %s", msg.Subject, tt.wantSubject)
					}
					return nil, nil
				},
			}

			p := &Protocol{
				jetStream:   mockJS,
				sendSubject: tt.sendSubject,
			}

			msg := test.FullMessage()
			_ = p.Send(tt.ctx, msg)
		})
	}
}
