package http

import (
	"context"
	"fmt"
	"net/http"
	nethttp "net/http"
	"net/url"

	"github.com/cloudevents/sdk-go/pkg/binding"
)

type Requester struct {
	// Client is the HTTP client used to send events as HTTP requests
	Client *http.Client
	// Target is the URL to send event requests to.
	Target *url.URL

	transformerFactories binding.TransformerFactories
}

func NewRequester(client *http.Client, target *url.URL, options ...RequesterOptionFunc) binding.Requester {
	r := &Requester{Client: client, Target: target, transformerFactories: make(binding.TransformerFactories, 0)}
	for _, o := range options {
		o(r)
	}
	return r
}

func (r *Requester) Request(ctx context.Context, m binding.Message) (binding.Receiver, error) {
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

	if err = EncodeHttpRequest(ctx, m, req, r.transformerFactories); err != nil {
		return nil, err
	}
	resp, err := r.Client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("%d %s", resp.StatusCode, nethttp.StatusText(resp.StatusCode))
	}

	msg, err := NewMessage(resp.Header, resp.Body)
	return &cachedReceiver{Message: msg, Error: err}, err
}

type cachedReceiver struct {
	Message binding.Message
	Error   error
}

func (r *cachedReceiver) Receive(ctx context.Context) (binding.Message, error) {
	return r.Message, r.Error
}
