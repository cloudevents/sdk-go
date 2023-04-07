/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package kafka_sarama

import (
	"context"
	"sync"
	"testing"

	"github.com/Shopify/sarama"
	"github.com/stretchr/testify/require"

	"github.com/cloudevents/sdk-go/v2/test"
)

type syncProducerMock struct {
	lock            sync.Mutex
	sent            []*sarama.ProducerMessage
	isTransactional bool
	status          sarama.ProducerTxnStatusFlag
}

func (s *syncProducerMock) SendMessage(msg *sarama.ProducerMessage) (partition int32, offset int64, err error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.sent = append(s.sent, msg)
	return 0, int64(len(s.sent) - 1), err
}

func (s *syncProducerMock) SendMessages(msgs []*sarama.ProducerMessage) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.sent = append(s.sent, msgs...)
	return nil
}

func (s *syncProducerMock) Close() error {
	s.lock.Lock()
	defer s.lock.Unlock()
	return nil
}

func (s *syncProducerMock) IsTransactional() bool {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.isTransactional
}

func (s *syncProducerMock) BeginTxn() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.status = sarama.ProducerTxnFlagInTransaction
	return nil
}

func (s *syncProducerMock) CommitTxn() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.status = sarama.ProducerTxnFlagReady
	return nil
}

func (s *syncProducerMock) AbortTxn() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.status = sarama.ProducerTxnFlagReady
	return nil
}

func (s *syncProducerMock) TxnStatus() sarama.ProducerTxnStatusFlag {
	return s.status
}

func (s *syncProducerMock) AddOffsetsToTxn(offsets map[string][]*sarama.PartitionOffsetMetadata, groupId string) error {
	return nil
}

func (s *syncProducerMock) AddMessageToTxn(msg *sarama.ConsumerMessage, groupId string, metadata *string) error {
	return nil
}

func TestSenderWithKey(t *testing.T) {
	syncProducerMock := &syncProducerMock{
		isTransactional: true,
		status:          sarama.ProducerTxnFlagReady,
	}
	topic := "aaa"

	sender := &Sender{topic: topic, syncProducer: syncProducerMock}
	require.NoError(t, sender.Send(
		WithMessageKey(context.TODO(), sarama.StringEncoder("hello")),
		test.FullMessage(),
	))

	require.Len(t, syncProducerMock.sent, 1)
	kafkaMsg := syncProducerMock.sent[0]
	require.Equal(t, kafkaMsg.Topic, topic)
	require.Equal(t, kafkaMsg.Key, sarama.StringEncoder("hello"))
}
