package v2

import cloudevents "github.com/cloudevents/sdk-go/v2"

type Expression interface {
	Evaluate(event cloudevents.Event) (interface{}, error)
}
