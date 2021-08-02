/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package kafka_segmentio

import (
	"testing"
)

func TestSenderWithKey(t *testing.T) {
	// topic := "aaa"
	// writer := kafka.NewWriter(kafka.WriterConfig{
	// 	Brokers: []string{"localhost:9092"},
	// })
	// sender := &Sender{receiverTopic: topic, writer: writer}
	// require.NoError(t, sender.Send(
	// 	WithMessageKey(context.TODO(), []byte("hello")),
	// 	test.FullMessage(),
	// ))

	// require.Len(t, writer.Stats().Messages, 1)
}
