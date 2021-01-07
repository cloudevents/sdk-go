package client

import (
	"github.com/cloudevents/sdk-go/v2/protocol/http"
)

// NewDefault provides the good defaults for the common case using an HTTP
// Protocol client.
// The WithTimeNow, and WithUUIDs client options are also applied to the
// client, all outbound events will have a time and id set if not already
// present.
func NewDefault(opts ...http.Option) (Client, error) {
	p, err := http.New(opts...)
	if err != nil {
		return nil, err
	}

	c, err := New(p, WithTimeNow(), WithUUIDs())
	if err != nil {
		return nil, err
	}

	return c, nil
}
