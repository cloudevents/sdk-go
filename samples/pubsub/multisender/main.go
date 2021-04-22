/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/kelseyhightower/envconfig"

	cepubsub "github.com/cloudevents/sdk-go/protocol/pubsub/v2"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	cecontext "github.com/cloudevents/sdk-go/v2/context"
)

/*
gcloud pubsub topics create ce1
gcloud pubsub topics create ce2
gcloud pubsub topics create ce3
gcloud pubsub subscriptions create ce1_sub --topic=ce1
gcloud pubsub subscriptions create ce2_sub --topic=ce2
gcloud pubsub subscriptions create ce3_sub --topic=ce3
*/

type envConfig struct {
	ProjectID string `envconfig:"GOOGLE_CLOUD_PROJECT" required:"true"`

	TopicIDs string `envconfig:"PUBSUB_TOPICS" default:"demo_cloudevents" required:"true"`
}

// Example is a basic data struct.
type Example struct {
	Sequence int    `json:"id"`
	Message  string `json:"message"`
}

func main() {
	var env envConfig
	if err := envconfig.Process("", &env); err != nil {
		log.Printf("[ERROR] Failed to process env var: %s", err)
		os.Exit(1)
	}

	t, err := cepubsub.New(context.Background(),
		cepubsub.WithProjectID(env.ProjectID))
	if err != nil {
		log.Printf("failed to create pubsub protocol, %s", err.Error())
		os.Exit(1)
	}
	c, err := cloudevents.NewClient(t, cloudevents.WithTimeNow(), cloudevents.WithUUIDs())
	if err != nil {
		log.Printf("failed to create client, %s", err.Error())
		os.Exit(1)
	}

	for i, topic := range strings.Split(env.TopicIDs, ",") {
		ctx := cecontext.WithTopic(context.Background(), topic)
		event := cloudevents.NewEvent()
		event.SetType("com.cloudevents.sample.sent")
		event.SetSource("github.com/cloudevents/sdk-go/samples/pubsub/multisender/")
		_ = event.SetData("application/json", &Example{
			Sequence: i,
			Message:  "HELLO " + topic,
		})

		if result := c.Send(ctx, event); cloudevents.IsUndelivered(result) {
			log.Printf("failed to send: %v", err)
			os.Exit(1)
		} else {
			log.Printf("sent, accepted: %t", cloudevents.IsACK(result))
		}
	}

	os.Exit(0)
}
