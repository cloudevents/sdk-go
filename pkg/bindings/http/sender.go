package http

import (
	"context"
	"fmt"
	"net/http"
	nethttp "net/http"
	"net/url"

	"github.com/cloudevents/sdk-go/pkg/binding"
)

// Sender implements binding.Sender wrapping a nethttp.Client and a target URL
type Sender struct {
	// Client is the HTTP client used to send events as HTTP requests
	Client *http.Client
	// Target is the URL to send event requests to.
	Target *url.URL

	transformers binding.TransformerFactories
}

func NewSender(client *http.Client, target *url.URL, options ...SenderOptionFunc) binding.Sender {
	s := &Sender{Client: client, Target: target, transformers: make(binding.TransformerFactories, 0)}
	for _, o := range options {
		o(s)
	}
	return s
}

// Confirm Sender implements binding.Requester
var _ binding.Requester = (*Sender)(nil)

// Send implements binding.Sender
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

	if err = EncodeHttpRequest(ctx, m, req, s.transformers); err != nil {
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

// Request implements binding.Requester
func (r *Sender) Request(ctx context.Context, m binding.Message) (binding.Message, error) {
	var err error
	defer func() { _ = m.Finish(err) }()
	if r.Client == nil || r.Target == nil {
		return nil, fmt.Errorf("not initialized: %#v", r)
	}

	var req *http.Request
	req, err = http.NewRequest("POST", r.Target.String(), nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)

	if err = EncodeHttpRequest(ctx, m, req, r.transformers); err != nil {
		return nil, err
	}
	resp, err := r.Client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("%d %s", resp.StatusCode, nethttp.StatusText(resp.StatusCode))
	}

	return NewMessageFromHttpResponse(resp), nil
}
