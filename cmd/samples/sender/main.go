package main

import (
	"fmt"
	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/context"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport"
	cloudeventshttp "github.com/cloudevents/sdk-go/pkg/cloudevents/transport/http"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/types"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/google/uuid"

	"github.com/kelseyhightower/envconfig"
)

type envConfig struct {
	// Target URL where to send cloudevents
	Target string `envconfig:"TARGET" default:"http://localhost:8080" required:"true"`
}

func main() {
	var env envConfig
	if err := envconfig.Process("", &env); err != nil {
		log.Printf("[ERROR] Failed to process env var: %s", err)
		os.Exit(1)
	}
	os.Exit(_main(os.Args[1:], env))
}

// Simple holder for the sending sample.
type Demo struct {
	Message string
	Source  url.URL
	Target  url.URL

	Sender transport.Sender
}

// Basic data struct.
type Example struct {
	Sequence int    `json:"id"`
	Message  string `json:"message"`
}

func (d *Demo) Send(i int) error {

	data := &Example{
		Sequence: i,
		Message:  d.Message,
	}

	now := time.Now()
	event := cloudevents.Event{
		Context: context.EventContextV01{
			EventID:   uuid.New().String(),
			EventType: "com.cloudevents.sample.sent",
			EventTime: &types.Timestamp{Time: now},
			Source:    types.URLRef{URL: d.Source},
		},
		Data: data,
	}
	req := &http.Request{
		Method: http.MethodPost,
		URL:    &d.Target,
	}
	resp, err := d.Sender.Send(event, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if accepted(resp) {
		return nil
	}
	return fmt.Errorf("error sending cloudevent: %s", status(resp))
}

// accepted is a helper method to understand if the response from the target
// accepted the CloudEvent.
func accepted(resp *http.Response) bool {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return true
	}
	return false
}

// status is a helper method to read the response of the target.
func status(resp *http.Response) string {
	status := resp.Status
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Sprintf("Status[%s] error reading response body: %v", status, err)
	}
	return fmt.Sprintf("Status[%s] %s", status, body)
}

func _main(args []string, env envConfig) int {
	source, err := url.Parse("https://github.com/cloudevents/sdk-go/cmd/samples/sender")
	if err != nil {
		log.Printf("failed to parse source url, %v", err)
		return 1
	}
	target, err := url.Parse(env.Target)
	if err != nil {
		log.Printf("failed to parse target url, %v", err)
		return 1
	}

	d := &Demo{
		Message: "Hello, World!",
		Source:  *source,
		Target:  *target,
		Sender:  &cloudeventshttp.Transport{Encoding: cloudeventshttp.StructuredV02},
	}

	for i := 0; i < 10; i++ {
		if err := d.Send(i); err != nil {
			log.Printf("failed to send: %v", err)
			return 1
		}
		time.Sleep(1 * time.Second)
	}

	return 0
}
