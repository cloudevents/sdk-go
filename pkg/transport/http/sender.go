package http

import (
	"context"
	"fmt"
	"net/http"
	nethttp "net/http"
	"net/url"

	cecontext "github.com/cloudevents/sdk-go/pkg/context"
	bindings "github.com/cloudevents/sdk-go/pkg/transport"

	"github.com/cloudevents/sdk-go/pkg/binding"
)

// Sender implements binding.Sender wrapping a nethttp.Client and a target URL
type Sender struct {
	// Client is the HTTP client used to send events as HTTP requests
	Client *http.Client

	// RequestTemplate is the base http request that is used for http.Do.
	// Only .Method, .URL, .Close, and .Header is considered.
	// If not set, Req.Method defaults to POST.
	// Req.URL or context.WithTarget(url) are required for sending.
	RequestTemplate *http.Request

	transformers binding.TransformerFactories
}

func NewRequester(client *http.Client, target *url.URL, options ...SenderOptionFunc) bindings.Requester {
	s := &Sender{
		Client:          client,
		RequestTemplate: &http.Request{Method: http.MethodPost, URL: target},
		transformers:    make(binding.TransformerFactories, 0),
	}
	for _, o := range options {
		o(s)
	}
	return s
}

func NewSender(client *http.Client, target *url.URL, options ...SenderOptionFunc) bindings.Sender {
	return NewRequester(client, target, options...)
}

// Confirm Sender implements binding.Requester
var _ bindings.Requester = (*Sender)(nil)

// Send implements binding.Sender
func (s *Sender) Send(ctx context.Context, m binding.Message) error {
	_, err := s.Request(ctx, m)
	return err
}

// Request implements binding.Requester
func (s *Sender) Request(ctx context.Context, m binding.Message) (binding.Message, error) {
	var err error
	defer func() { _ = m.Finish(err) }()

	req := s.makeRequest(ctx)

	if s.Client == nil || req == nil || req.URL == nil {
		return nil, fmt.Errorf("not initialized: %#v", s)
	}

	if err = WriteHttpRequest(ctx, m, req, s.transformers); err != nil {
		return nil, err
	}
	resp, err := s.Client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("%d %s", resp.StatusCode, nethttp.StatusText(resp.StatusCode))
	}

	return NewMessage(resp.Header, resp.Body), nil
}

func (s *Sender) makeRequest(ctx context.Context) *http.Request {
	// TODO: support custom headers from context?
	req := &http.Request{
		Header: make(http.Header),
		// TODO: HeaderFrom(ctx),
	}

	if s.RequestTemplate != nil {
		req.Method = s.RequestTemplate.Method
		req.URL = s.RequestTemplate.URL
		req.Close = s.RequestTemplate.Close
		req.Host = s.RequestTemplate.Host
		copyHeadersEnsure(s.RequestTemplate.Header, &req.Header)
	}

	// Override the default request with target from context.
	if target := cecontext.TargetFrom(ctx); target != nil {
		req.URL = target
	}
	return req.WithContext(ctx)
}

// Ensure to is a non-nil map before copying
func copyHeadersEnsure(from http.Header, to *http.Header) {
	if len(from) > 0 {
		if *to == nil {
			*to = http.Header{}
		}
		copyHeaders(from, *to)
	}
}

func copyHeaders(from, to http.Header) {
	if from == nil || to == nil {
		return
	}
	for header, values := range from {
		for _, value := range values {
			to.Add(header, value)
		}
	}
}
