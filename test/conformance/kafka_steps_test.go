/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package conformance

import (
	"context"

	"github.com/IBM/sarama"
	"github.com/cloudevents/sdk-go/protocol/kafka_sarama/v2"
	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cucumber/godog"
)

var consumerMessage *sarama.ConsumerMessage

func KafkaFeatureContext(ctx *godog.ScenarioContext) {
	ctx.BeforeScenario(func(*godog.Scenario) {
		consumerMessage = nil
	})
	ctx.Step(`^Kafka Protocol Binding is supported$`, func() error {
		return nil
	})

	ctx.Step(`^a Kafka message with payload:$`, func(payload *godog.DocString) error {
		consumerMessage = &sarama.ConsumerMessage{
			Value: []byte(payload.Content),
		}

		return nil
	})

	ctx.Step(`^Kafka headers:$`, func(headers *godog.Table) error {
		consumerMessage.Headers = make([]*sarama.RecordHeader, len(headers.Rows))

		for i, row := range headers.Rows {
			var key = row.Cells[0].Value
			var value = row.Cells[1].Value
			consumerMessage.Headers[i] = &sarama.RecordHeader{Key: []byte(key), Value: []byte(value)}
		}

		return nil
	})

	ctx.Step(`^parsed as Kafka message$`, func() error {
		message := kafka_sarama.NewMessageFromConsumerMessage(consumerMessage)

		event, err := binding.ToEvent(context.TODO(), message)

		if err != nil {
			return err
		}

		currentEvent = event
		return nil
	})
}
