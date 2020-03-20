package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	ce "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/client"
	ceamqp "github.com/cloudevents/sdk-go/v2/protocol/amqp"
	amqp "pack.ag/amqp"
)

// Parse AMQP_URL env variable. Return server URL, AMQP node (from path) and SASLPlain
// option if user/pass are present.
func sampleConfig() (server, node string, opts []ceamqp.Option) {
	env := os.Getenv("AMQP_URL")
	if env == "" {
		env = "/test"
	}
	u, err := url.Parse(env)
	if err != nil {
		log.Fatal(err)
	}
	if u.User != nil {
		user := u.User.Username()
		pass, _ := u.User.Password()
		opts = append(opts, ceamqp.WithConnOpt(amqp.ConnSASLPlain(user, pass)))
	}
	return env, strings.TrimPrefix(u.Path, "/"), opts
}

func main() {
	host, node, opts := sampleConfig()
	t, err := ceamqp.NewProtocol(host, node, []amqp.ConnOption{}, []amqp.SessionOption{}, opts...)
	if err != nil {
		log.Fatalf("Failed to create AMQP protocol: %v", err)
	}
	c, err := client.New(t)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	err = c.StartReceiver(context.Background(), func(e ce.Event) {
		fmt.Printf("==== Got CloudEvent\n%+v\n----\n", e)
	})
	if err != nil {
		log.Fatalf("AMQP receiver error: %v", err)
	}
}
