package client

import (
	"github.com/cloudevents/sdk-go/v2/protocol/http"
	obshttp "github.com/cloudevents/sdk-go/observability/opencensus/v2/http"
	"github.com/cloudevents/sdk-go/v2/client"
)

func NewClientHTTP(topt []http.Option, copt []client.Option) (client.Client, error) {
	t, err := obshttp.NewObservedHTTP(topt...)
	if err != nil {
		return nil, err
	}

	copt = append(copt, client.WithTimeNow(), client.WithUUIDs(), client.WithObservabilityService(New()))

	c, err := client.New(t, copt...)
	if err != nil {
		return nil, err
	}

	return c, nil
}
