package client

import (
	"fmt"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport/http"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport/nats"
	nethttp "net/http"
	"net/url"
)

type ClientOption func(*Client) error

// WithTarget sets the outbound recipient of cloudevents when using a http request.
func WithTarget(targetUrl string) ClientOption {
	return func(c *Client) error {
		if t, ok := c.transport.(*http.Transport); ok {
			if targetUrl != "" {
				var err error
				var target *url.URL
				target, err = url.Parse(targetUrl)
				if err != nil {
					return fmt.Errorf("client option failed to parse target url: %s", err.Error())
				}

				if t.Req == nil {
					t.Req = &nethttp.Request{
						Method: nethttp.MethodPost,
					}
				}
				t.Req.URL = target
				return nil
			} else {
				return fmt.Errorf("target option was empty string")
			}
		}
		return fmt.Errorf("invalid client option recieved for given transport type")
	}
}

// WithTarget sets the outbound recipient of cloudevents when using a http request.
func WithHttpMethod(method string) ClientOption {
	return func(c *Client) error {
		if t, ok := c.transport.(*http.Transport); ok {
			if method != "" {
				if t.Req == nil {
					t.Req = &nethttp.Request{}
				}
				t.Req.Method = method
				return nil
			} else {
				return fmt.Errorf("context client option was nil")
			}
		}
		return fmt.Errorf("invalid client option recieved for given transport type")
	}
}

// WithHttpEncoding sets the encoding for clients with HTTP transports.
func WithHttpEncoding(encoding http.Encoding) ClientOption {
	return func(c *Client) error {
		if t, ok := c.transport.(*http.Transport); ok {
			t.Encoding = encoding
			return nil
		}
		return fmt.Errorf("invalid client option recieved for given client type")
	}
}

// WithHttpPort sets the port for accepting requests using HTTP transport.
func WithHttpPort(port int) ClientOption {
	return func(c *Client) error {
		if t, ok := c.transport.(*http.Transport); ok {
			if port == 0 {
				return fmt.Errorf("client option was given an invalid port: %d", port)
			}
			t.Port = port
			return nil
		}
		return fmt.Errorf("invalid client option recieved for given client type")
	}
}

// WithNatsEncoding sets the encoding for clients with NATS transport.
func WithNatsEncoding(encoding nats.Encoding) ClientOption {
	return func(c *Client) error {
		if t, ok := c.transport.(*nats.Transport); ok {
			t.Encoding = encoding
			return nil
		}
		return fmt.Errorf("invalid client option recieved for given client type")
	}
}
