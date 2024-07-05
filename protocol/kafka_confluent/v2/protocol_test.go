/*
 Copyright 2024 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package kafka_confluent

import (
	"context"
	"testing"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/stretchr/testify/assert"
)

func TestNewProtocol(t *testing.T) {
	tests := []struct {
		name         string
		ctx          context.Context
		options      []Option
		errorMessage string
	}{
		{
			name:         "invalidated parameters",
			options:      nil,
			errorMessage: "at least one of the following to initialize the protocol must be set: config, producer, or consumer",
		},
		{
			name: "Insufficient parameters",
			options: []Option{
				WithConfigMap(&kafka.ConfigMap{
					"bootstrap.servers": "127.0.0.1:9092",
				})},
			errorMessage: "at least receiver or sender topic must be set",
		},
		{
			name: "Insufficient consumer parameters - group.id",
			options: []Option{
				WithConfigMap(&kafka.ConfigMap{
					"bootstrap.servers": "127.0.0.1:9092",
				}),
				WithReceiverTopics([]string{"topic1", "topic2"}),
			},
			errorMessage: "Required property group.id not set",
		},
		{
			name: "Insufficient consumer parameters - configmap or consumer",
			options: []Option{
				WithReceiverTopics([]string{"topic1", "topic2"}),
			},
			errorMessage: "at least configmap or consumer must be set for the receiver topics: [topic1 topic2]",
		},
		{
			name: "Insufficient producer parameters",
			options: []Option{
				WithSenderTopic("topic3"),
			},
			errorMessage: "at least configmap or producer must be set for the sender topic: topic3",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(tt.options...)
			if err != nil {
				assert.Equal(t, tt.errorMessage, err.Error())
			}
		})
	}
}
