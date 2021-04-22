/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/kelseyhightower/envconfig"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	cehttp "github.com/cloudevents/sdk-go/v2/protocol/http"
)

const (
	count = 1000
)

type envConfig struct {
	// Target URL where to send cloudevents
	Target string `envconfig:"TARGET" default:"https://localhost:8080" required:"true"`
	// Path for the client certificate used to publish to an HTTPS endpoint
	ClientCert string `envconfig:"CLIENT_CERT" default:"client.crt" required:"true"`
	// Path for the client key used to publish to an HTTPS endpoint
	ClientKey string `envconfig:"CLIENT_KEY" default:"client.key" required:"true"`
}

// Example is a basic data struct.
type Example struct {
	Sequence int    `json:"id"`
	Message  string `json:"message"`
}

func main() {
	var env envConfig
	if err := envconfig.Process("", &env); err != nil {
		log.Fatalf("Failed to process env var: %s", err)
	}

	// Configure a new http.Transport with TLS
	cert, err := tls.LoadX509KeyPair(env.ClientCert, env.ClientKey)
	if err != nil {
		log.Fatalln("unable to load certs", err)
	}
	clientCACert, err := ioutil.ReadFile(env.ClientCert)
	if err != nil {
		log.Fatal("unable to open cert", err)
	}

	clientCertPool := x509.NewCertPool()
	clientCertPool.AppendCertsFromPEM(clientCACert)

	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{cert},
		RootCAs:            clientCertPool,
		InsecureSkipVerify: true,
	}

	httpTransport := &http.Transport{TLSClientConfig: tlsConfig}

	// Create protocol and client
	p, err := cloudevents.NewHTTP(cloudevents.WithTarget(env.Target), cloudevents.WithRoundTripper(httpTransport))
	if err != nil {
		log.Fatalf("Failed to create protocol, %v", err)
	}

	c, err := cloudevents.NewClient(p,
		cloudevents.WithTimeNow(),
	)
	if err != nil {
		log.Fatalf("Failed to create client, %v", err)
	}

	// Send events
	for i := 0; i < count; i++ {
		// Create a new event
		event := cloudevents.NewEvent()
		event.SetID(uuid.New().String())
		event.SetType("com.cloudevents.sample.sent")
		event.SetSource("https://github.com/cloudevents/sdk-go/v2/samples/requester")

		// Set data
		_ = event.SetData("application/json", &Example{
			Sequence: i,
			Message:  "Hello world!",
		})

		if resp, res := c.Request(context.TODO(), event); cloudevents.IsUndelivered(res) {
			log.Printf("Failed to request: %v", res)
		} else if resp != nil {
			fmt.Printf("Response:\n%s\n", resp)
			fmt.Printf("Got Event Response Context: %+v\n", resp.Context)
			data := &Example{}
			if err := resp.DataAs(data); err != nil {
				fmt.Printf("Got Data Error: %s\n", err.Error())
			}
			fmt.Printf("Got Response Data: %+v\n", data)
			fmt.Printf("----------------------------\n")
		} else {
			// Parse result
			var httpResult *cehttp.Result
			cloudevents.ResultAs(res, &httpResult)
			log.Printf("Event sent at %s", time.Now())
			log.Printf("Response status code %d", httpResult.StatusCode)
		}

		time.Sleep(500 * time.Millisecond)
	}
}
