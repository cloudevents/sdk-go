/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package pubsub

import (
	"context"
	"fmt"
	"testing"

	"cloud.google.com/go/pubsub/v2"
	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/binding/spec"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/cloudevents/sdk-go/v2/protocol"
	"github.com/cloudevents/sdk-go/v2/test"
	"github.com/stretchr/testify/require"
)

func TestReadBinaryDataContentType(t *testing.T) {
	tests := []struct {
		name       string
		attributes map[string]string
		want       string
	}{
		{
			name:       "prefixed attribute",
			attributes: map[string]string{"ce-datacontenttype": "application/json; charset=utf-8"},
			want:       "application/json; charset=utf-8",
		},
		{
			name:       "legacy attribute",
			attributes: map[string]string{"Content-Type": "application/json; charset=utf-8"},
			want:       "application/json; charset=utf-8",
		},
		{
			name:       "lowercase content type",
			attributes: map[string]string{"content-type": "application/json; charset=utf-8"},
			want:       "application/json; charset=utf-8",
		},
		{
			name: "lowercase content type takes precedence over legacy attribute",
			attributes: map[string]string{
				"content-type": "application/json; charset=utf-8",
				"Content-Type": "text/plain",
			},
			want: "application/json; charset=utf-8",
		},
		{
			name: "matching lowercase and prefixed attributes",
			attributes: map[string]string{
				"ce-datacontenttype": "application/json; charset=utf-8",
				"content-type":       "application/json; charset=utf-8",
			},
			want: "application/json; charset=utf-8",
		},
		{
			name: "matching lowercase and prefixed attributes override legacy attribute",
			attributes: map[string]string{
				"ce-datacontenttype": "application/json; charset=utf-8",
				"content-type":       "application/json; charset=utf-8",
				"Content-Type":       "text/plain",
			},
			want: "application/json; charset=utf-8",
		},
		{
			name: "all content type attributes match",
			attributes: map[string]string{
				"ce-datacontenttype": "application/json; charset=utf-8",
				"content-type":       "application/json; charset=utf-8",
				"Content-Type":       "application/json; charset=utf-8",
			},
			want: "application/json; charset=utf-8",
		},
		{
			name: "matching empty lowercase and prefixed attributes",
			attributes: map[string]string{
				"ce-datacontenttype": "",
				"content-type":       "",
			},
		},
		{
			name: "empty lowercase content type takes precedence over legacy attribute",
			attributes: map[string]string{
				"content-type": "",
				"Content-Type": "text/plain",
			},
		},
		{
			name: "matching prefixed and legacy attributes",
			attributes: map[string]string{
				"ce-datacontenttype": "application/json; charset=utf-8",
				"Content-Type":       "application/json; charset=utf-8",
			},
			want: "application/json; charset=utf-8",
		},
		{
			name: "prefixed attribute takes precedence over legacy attribute",
			attributes: map[string]string{
				"ce-datacontenttype": "application/json; charset=utf-8",
				"Content-Type":       "text/plain",
			},
			want: "application/json; charset=utf-8",
		},
		{
			name: "empty prefixed attribute takes precedence over legacy attribute",
			attributes: map[string]string{
				"ce-datacontenttype": "",
				"Content-Type":       "text/plain",
			},
		},
		{
			name:       "empty legacy attribute",
			attributes: map[string]string{"Content-Type": ""},
		},
		{name: "no content type"},
	}
	for _, version := range []string{event.CloudEventsVersionV03, event.CloudEventsVersionV1} {
		t.Run(version, func(t *testing.T) {
			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					attributes := map[string]string{
						"ce-specversion": version,
						"ce-id":          "testid",
						"ce-source":      "/test",
						"ce-type":        "test.type",
					}
					for name, value := range tt.attributes {
						attributes[name] = value
					}
					data := []byte(`{"hello":"world"}`)
					msg := NewMessage(&pubsub.Message{
						Attributes: attributes,
						Data:       data,
					})

					require.Equal(t, binding.EncodingBinary, msg.ReadEncoding())
					attr, value := msg.GetAttribute(spec.DataContentType)
					require.Equal(t, spec.DataContentType, attr.Kind())
					require.Equal(t, tt.want, value)
					got, err := binding.ToEvent(context.Background(), msg)
					require.NoError(t, err)
					require.Equal(t, tt.want, got.DataContentType())
					require.Equal(t, data, got.Data())
				})
			}
		})
	}
}

func TestReadBinaryConflictingDataContentTypes(t *testing.T) {
	tests := []struct {
		name       string
		attributes map[string]string
	}{
		{
			name: "conflicting lowercase and prefixed attributes",
			attributes: map[string]string{
				"content-type":       "application/json",
				"ce-datacontenttype": "text/plain",
			},
		},
		{
			name: "empty lowercase attribute conflicts with prefixed attribute",
			attributes: map[string]string{
				"content-type":       "",
				"ce-datacontenttype": "text/plain",
			},
		},
		{
			name: "empty prefixed attribute conflicts with lowercase attribute",
			attributes: map[string]string{
				"content-type":       "application/json",
				"ce-datacontenttype": "",
			},
		},
		{
			name: "legacy attribute does not resolve a conflict",
			attributes: map[string]string{
				"content-type":       "application/json",
				"ce-datacontenttype": "text/plain",
				"Content-Type":       "application/json",
			},
		},
	}
	for _, version := range []string{event.CloudEventsVersionV03, event.CloudEventsVersionV1} {
		t.Run(version, func(t *testing.T) {
			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					attributes := map[string]string{
						"ce-specversion": version,
						"ce-id":          "testid",
						"ce-source":      "/test",
						"ce-type":        "test.type",
					}
					for name, value := range tt.attributes {
						attributes[name] = value
					}
					msg := NewMessage(&pubsub.Message{Attributes: attributes, Data: []byte(`{"hello":"world"}`)})
					require.Equal(t, binding.EncodingBinary, msg.ReadEncoding())

					writer := &pubsubMessagePublisher{Attributes: make(map[string]string)}
					err := msg.ReadBinary(context.Background(), writer)
					require.ErrorContains(t, err, `conflicting "content-type" and "ce-datacontenttype"`)
					require.Empty(t, writer.Attributes)
					require.Nil(t, writer.Data)
					attr, value := msg.GetAttribute(spec.DataContentType)
					require.Equal(t, spec.DataContentType, attr.Kind())
					require.Nil(t, value)

					got, conversionErr := binding.ToEvent(context.Background(), msg)
					require.EqualError(t, conversionErr, err.Error())
					require.Nil(t, got)
					for _, encoding := range []binding.Encoding{binding.EncodingBinary, binding.EncodingStructured} {
						t.Run(encoding.String(), func(t *testing.T) {
							ctx := binding.WithForceBinary(context.Background())
							if encoding == binding.EncodingStructured {
								ctx = binding.WithForceStructured(context.Background())
							}
							pm := &pubsub.Message{Attributes: make(map[string]string)}
							require.EqualError(t, WritePubSubMessage(ctx, msg, pm), err.Error())
							require.Empty(t, pm.Attributes)
							require.Nil(t, pm.Data)
						})
					}
				})
			}
		})
	}
}

func TestNewMessageContentType(t *testing.T) {
	tests := []struct {
		name       string
		attributes map[string]string
		encoding   binding.Encoding
	}{
		{
			name:       "lowercase structured content type",
			attributes: map[string]string{"content-type": "application/cloudevents+json; charset=utf-8"},
			encoding:   binding.EncodingStructured,
		},
		{
			name:       "legacy structured content type",
			attributes: map[string]string{"Content-Type": "application/cloudevents+json"},
			encoding:   binding.EncodingStructured,
		},
		{
			name: "matching structured content types",
			attributes: map[string]string{
				"content-type": "application/cloudevents+json",
				"Content-Type": "application/cloudevents+json",
			},
			encoding: binding.EncodingStructured,
		},
		{
			name: "lowercase structured content type takes precedence over legacy content type",
			attributes: map[string]string{
				"content-type": "application/cloudevents+json",
				"Content-Type": "application/json",
			},
			encoding: binding.EncodingStructured,
		},
		{
			name: "lowercase structured content type with ce-specversion",
			attributes: map[string]string{
				"content-type":   "application/cloudevents+json",
				"ce-specversion": "1.0",
			},
			encoding: binding.EncodingStructured,
		},
		{
			name: "lowercase binary content type takes precedence over legacy content type",
			attributes: map[string]string{
				"content-type":   "application/json",
				"Content-Type":   "application/cloudevents+json",
				"ce-specversion": "1.0",
			},
			encoding: binding.EncodingBinary,
		},
		{
			name: "empty lowercase content type takes precedence over legacy content type",
			attributes: map[string]string{
				"content-type":   "",
				"Content-Type":   "application/cloudevents+json",
				"ce-specversion": "1.0",
			},
			encoding: binding.EncodingBinary,
		},
		{
			name: "legacy structured content type with ce-specversion",
			attributes: map[string]string{
				"Content-Type":   "application/cloudevents+json",
				"ce-specversion": "1.0",
			},
			encoding: binding.EncodingStructured,
		},
		{
			name: "unsupported lowercase format does not fall back to legacy format",
			attributes: map[string]string{
				"content-type":   "application/cloudevents+unknown",
				"Content-Type":   "application/cloudevents+json",
				"ce-specversion": "1.0",
			},
			encoding: binding.EncodingBinary,
		},
		{
			name: "lowercase content type without binary metadata",
			attributes: map[string]string{
				"content-type": "application/json",
				"Content-Type": "application/cloudevents+json",
			},
			encoding: binding.EncodingUnknown,
		},
		{name: "no attributes", encoding: binding.EncodingUnknown},
		{
			name:       "unsupported spec version",
			attributes: map[string]string{"ce-specversion": "unknown"},
			encoding:   binding.EncodingUnknown,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := NewMessage(&pubsub.Message{Attributes: tt.attributes})
			require.Equal(t, tt.encoding, msg.ReadEncoding())
			writer := &pubsubMessagePublisher{Attributes: make(map[string]string)}
			err := msg.ReadStructured(context.Background(), writer)
			if tt.encoding == binding.EncodingStructured {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, binding.ErrNotStructured)
			}
			err = msg.ReadBinary(context.Background(), writer)
			if tt.encoding == binding.EncodingBinary {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, binding.ErrNotBinary)
			}
		})
	}
}

func TestReadStructuredWithCloudEventAttributes(t *testing.T) {
	tests := []struct {
		name         string
		contentTypes map[string]string
	}{
		{
			name: "lowercase content type only",
			contentTypes: map[string]string{
				"content-type": "application/cloudevents+json; charset=utf-8",
			},
		},
		{
			name: "legacy content type only",
			contentTypes: map[string]string{
				"Content-Type": "application/cloudevents+json; charset=utf-8",
			},
		},
		{
			name: "matching lowercase and legacy content types",
			contentTypes: map[string]string{
				"content-type": "application/cloudevents+json; charset=utf-8",
				"Content-Type": "application/cloudevents+json; charset=utf-8",
			},
		},
		{
			name: "lowercase content type overrides legacy content type",
			contentTypes: map[string]string{
				"content-type": "application/cloudevents+json; charset=utf-8",
				"Content-Type": "application/protobuf",
			},
		},
	}
	test.EachEvent(t, test.Events(), func(t *testing.T, eventIn event.Event) {
		eventIn = test.ConvertEventExtensionsToString(t, eventIn)
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				attributes := map[string]string{
					"ce-specversion":     eventIn.SpecVersion(),
					"ce-id":              "ignored-id",
					"ce-datacontenttype": "application/protobuf",
				}
				for name, value := range tt.contentTypes {
					attributes[name] = value
				}
				msg := NewMessage(&pubsub.Message{
					Attributes: attributes,
					Data:       test.MustJSON(t, eventIn),
				})
				require.Equal(t, binding.EncodingStructured, msg.ReadEncoding())
				eventOut, err := binding.ToEvent(context.Background(), msg)
				require.NoError(t, err)
				test.AssertEventEquals(t, eventIn, *eventOut)
			})
		}
	})
}

func TestReadStructuredMissingSpecVersion(t *testing.T) {
	for _, name := range []string{"content-type", "Content-Type"} {
		t.Run(name, func(t *testing.T) {
			msg := NewMessage(&pubsub.Message{
				Attributes: map[string]string{
					name:             event.ApplicationCloudEventsJSON,
					"ce-specversion": event.CloudEventsVersionV1,
				},
				Data: []byte(`{"id":"testid","source":"/test","type":"test.type"}`),
			})
			require.Equal(t, binding.EncodingStructured, msg.ReadEncoding())
			eventOut, err := binding.ToEvent(context.Background(), msg)
			require.ErrorContains(t, err, "no specversion")
			require.Nil(t, eventOut)
		})
	}
}

func TestReadStructured(t *testing.T) {
	tests := []struct {
		name    string
		pm      *pubsub.Message
		wantErr error
	}{
		{
			name: "nil format",
			pm: &pubsub.Message{
				ID: "testid",
			},
			wantErr: binding.ErrNotStructured,
		},
		{
			name: "json format",
			pm: &pubsub.Message{
				ID:         "testid",
				Attributes: map[string]string{"Content-Type": event.ApplicationCloudEventsJSON},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			msg := NewMessage(tc.pm)
			err := msg.ReadStructured(context.Background(), (*pubsubMessagePublisher)(tc.pm))
			if err != tc.wantErr {
				t.Errorf("Error unexpected. got: %v, want: %v", err, tc.wantErr)
			}
		})
	}
}

func TestFinish(t *testing.T) {
	tests := []struct {
		name    string
		pm      *pubsub.Message
		err     error
		wantErr bool
	}{
		{
			name: "return error",
			pm: &pubsub.Message{
				ID: "testid",
			},
			err:     fmt.Errorf("error"),
			wantErr: true,
		},
		{
			name: "result not acked",
			pm: &pubsub.Message{
				ID: "testid",
			},
			err:     protocol.NewReceipt(false, "error"),
			wantErr: true,
		},
		{
			name: "result acked",
			pm: &pubsub.Message{
				ID: "testid",
			},
			err:     protocol.NewReceipt(true, "error"),
			wantErr: true,
		},
		{
			name: "no errors",
			pm: &pubsub.Message{
				ID: "testid",
			},
			wantErr: false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			msg := NewMessage(tc.pm)
			err := msg.Finish(tc.err)
			if tc.wantErr {
				if err != tc.err {
					t.Errorf("Error mismatch. got: %v, want: %v", err, tc.err)
				}
			}
			if !tc.wantErr && err != nil {
				t.Errorf("Should not error but got: %v", err)
			}
		})
	}
}
