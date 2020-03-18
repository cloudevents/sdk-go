package http

import (
	"context"
	"fmt"
	"net/http"
	nethttp "net/http"
	"net/url"

	"github.com/cloudevents/sdk-go/v1/binding"
)

type Sender struct {
	// Client is the HTTP client used to send events as HTTP requests
	Client *http.Client
	// Target is the URL to send event requests to.
	Target *url.URL

	transformerFactories binding.TransformerFactories
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

	if err = EncodeHttpRequest(ctx, m, req, s.transformerFactories); err != nil {
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

func NewSender(client *http.Client, target *url.URL, options ...SenderOptionFunc) binding.Sender {
	s := &Sender{Client: client, Target: target, transformerFactories: make(binding.TransformerFactories, 0)}
	for _, o := range options {
		o(s)
	}
	return s
}
