package client

import (
	"github.com/cloudevents/sdk-go/v2/protocol/http"
)

// NewHTTP provides the good defaults for the common case using an HTTP
// Protocol client.
// The WithTimeNow, and WithUUIDs client options are also applied to the
// client, all outbound events will have a time and id set if not already
// present.
func NewHTTP(opts ...http.Option) (Client, error) {
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

// NewDefault has been replaced by NewHTTP
// Deprecated. To get the same as NewDefault provided, please use the following:
// TODO
var NewDefault = NewHTTP
