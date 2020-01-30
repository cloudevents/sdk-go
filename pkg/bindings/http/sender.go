package http

import (
	"context"
	"fmt"
	"net/http"
	nethttp "net/http"
	"net/url"

	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/binding/format"
)

type Sender struct {
	// Client is the HTTP client used to send events as HTTP requests
	Client *http.Client
	// Target is the URL to send event requests to.
	Target *url.URL

	transformerFactories binding.TransformerFactories

	forceBinary     bool
	forceStructured bool
}

func (s *Sender) Send(ctx context.Context, m binding.Message) (err error) {
	defer func() { _ = m.Finish(err) }()
	if s.Client == nil || s.Target == nil {
		return fmt.Errorf("not initialized: %#v", s)
	}

	var req *http.Request
	req, err = http.NewRequest("POST", s.Target.String(), nil)
	if err != nil {
		return
	}
	req = req.WithContext(ctx)

	if err = s.fillHttpRequest(req, m); err != nil {
		return
	}
	resp, err := s.Client.Do(req)
	if err != nil {
		return
	}
	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("%d %s", resp.StatusCode, nethttp.StatusText(resp.StatusCode))
	}
	return
}

// This function tries:
// 1. Translate from structured
// 2. Translate from binary
// 3. Translate to Event and then re-encode back to Http Request
func (s *Sender) fillHttpRequest(req *http.Request, m binding.Message) error {
	createStructured := func() binding.StructuredEncoder {
		return &structuredMessageEncoder{req}
	}
	if s.forceBinary {
		createStructured = nil
	}

	createBinary := func() binding.BinaryEncoder {
		return &binaryMessageEncoder{req}
	}
	if s.forceStructured {
		createBinary = nil
	}

	createEvent := func() binding.EventEncoder {
		if s.forceStructured {
			return &eventToStructuredMessageEncoder{format: format.JSON, req: req}
		}
		return &eventToBinaryMessageEncoder{req}
	}

	_, _, err := binding.Translate(m, createStructured, createBinary, createEvent, s.transformerFactories)
	return err
}

func NewSender(client *http.Client, target *url.URL, options ...SenderOptionFunc) binding.Sender {
	s := &Sender{Client: client, Target: target, transformerFactories: make(binding.TransformerFactories, 0)}
	for _, o := range options {
		o(s)
	}
	return s
}
