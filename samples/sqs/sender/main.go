/*
Copyright 2023 The CloudEvents Authors
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"context"
	"log"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/cloudevents/sdk-go/protocol/sqs/v2"
	"github.com/kelseyhightower/envconfig"
)

const (
	count = 10
)

type envConfig struct {
	QueueURL string `envconfig:"AWS_SQS_QUEUE_URL" required:"true"`
}

func main() {
	var env envConfig
	if err := envconfig.Process("", &env); err != nil {
		log.Fatalf("[ERROR] Failed to process env var: %s", err)
	}
	ctx := context.Background()
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("failed to load AWS SDK configuration: %s", err.Error())
	}

	// set a default topic with test-topic1
	p, err := sqs.New(env.QueueURL, sqs.WithNewClientFromConfig(awsCfg))
	if err != nil {
		log.Fatalf("failed to create protocol: %v", err)
	}

	c, err := cloudevents.NewClient(p, cloudevents.WithTimeNow(), cloudevents.WithUUIDs())
	if err != nil {
		log.Fatalf("failed to create client, %v", err)
	}

	for i := 0; i < count; i++ {
		e := cloudevents.NewEvent()
		e.SetID(uuid.New().String())
		e.SetType("com.cloudevents.sample.sent")
		e.SetSource("https://github.com/cloudevents/sdk-go/samples/sqs/sender")
		err = e.SetData(cloudevents.ApplicationJSON, map[string]interface{}{
			"id":      i,
			"message": "Hello, World!",
		})
		if err != nil {
			log.Printf("failed to set data: %v", err)
		}
		if result := c.Send(
			ctx,
			// binding.WithForceStructured(ctx),
			// binding.WithForceBinary(ctx),
			e,
		); cloudevents.IsUndelivered(result) {
			log.Printf("failed to send: %v", result)
		} else {
			log.Printf("sent: %d, accepted: %t", i, cloudevents.IsACK(result))
		}
		time.Sleep(1 * time.Second)
	}
}
