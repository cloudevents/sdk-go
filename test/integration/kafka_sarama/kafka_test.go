/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package kafka_sarama_binding

import (
	"os"
	"strings"
	"testing"

	"github.com/Shopify/sarama"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/cloudevents/sdk-go/protocol/kafka_sarama/v2"
	clienttest "github.com/cloudevents/sdk-go/v2/client/test"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/cloudevents/sdk-go/v2/test"
)

const (
	TEST_GROUP_ID = "test_group_id"
)

func TestSendEvent(t *testing.T) {
	test.EachEvent(t, test.Events(), func(t *testing.T, eventIn event.Event) {
		eventIn = test.ConvertEventExtensionsToString(t, eventIn)
		clienttest.SendReceive(t, func() interface{} {
			return protocolFactory(t)
		}, eventIn, func(e event.Event) {
			test.AssertEventEquals(t, eventIn, test.ConvertEventExtensionsToString(t, e))
		})
	})
}

// To start a local environment for testing:
// docker run --rm --net=host -e ADV_HOST=localhost -e SAMPLEDATA=0 lensesio/fast-data-dev
func testClient(t testing.TB) sarama.Client {
	t.Helper()
	s := os.Getenv("TEST_KAFKA_BOOTSTRAP_SERVER")
	if s == "" {
		s = "localhost:9092"
	}

	config := sarama.NewConfig()
	config.Version = sarama.V2_0_0_0
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	client, err := sarama.NewClient(strings.Split(s, ","), config)
	if err != nil {
		t.Skipf("Cannot create sarama client to servers [%s]: %v", s, err)
	}

	return client
}

func protocolFactory(t testing.TB) *kafka_sarama.Protocol {
	client := testClient(t)

	topicName := "test-ce-client-" + uuid.New().String()
	options := []kafka_sarama.ProtocolOptionFunc{
		kafka_sarama.WithReceiverGroupId(TEST_GROUP_ID),
	}
	p, err := kafka_sarama.NewProtocolFromClient(client, topicName, topicName, options...)
	require.NoError(t, err)

	return p
}
