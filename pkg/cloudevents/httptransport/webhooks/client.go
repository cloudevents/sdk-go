package webhooks

import (
	"fmt"

	"github.com/cloudevents/sdk-go/pkg/cloudevents"
)

// Deliver delivers the event to the endpoint
func Deliver(event cloudevents.Event) (string, error) {
	return "", fmt.Errorf("not implemented")
}
