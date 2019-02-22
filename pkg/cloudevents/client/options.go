package client

import (
	"fmt"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport/http"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport/nats"
	nethttp "net/http"
	"net/url"
)

type ClientOption func(*Client) error

// WithTarget sets the outbound recipient of cloudevents when using an HTTP request.
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
		return fmt.Errorf("invalid target client option recieved for transport type")
	}
}

// WithHTTPMethod sets the outbound recipient of cloudevents when using an HTTP request.
func WithHTTPMethod(method string) ClientOption {
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
		return fmt.Errorf("invalid HTTP method client option recieved for transport type")
	}
}

// WithHTTPEncoding sets the encoding for clients with HTTP transports.
func WithHTTPEncoding(encoding http.Encoding) ClientOption {
	return func(c *Client) error {
		if t, ok := c.transport.(*http.Transport); ok {
			t.Encoding = encoding
			return nil
		}
		return fmt.Errorf("invalid HTTP encoding client option recieved for transport type")
	}
}

// WithHTTPPort sets the port for for clients with HTTP transports.
func WithHTTPPort(port int) ClientOption {
	return func(c *Client) error {
		if t, ok := c.transport.(*http.Transport); ok {
			if port == 0 {
				return fmt.Errorf("client option was given an invalid port: %d", port)
			}
			t.Port = port
			return nil
		}
		return fmt.Errorf("invalid HTTP port client option recieved for transport type")
	}
}

// WithNATSEncoding sets the encoding for clients with NATS transport.
func WithNATSEncoding(encoding nats.Encoding) ClientOption {
	return func(c *Client) error {
		if t, ok := c.transport.(*nats.Transport); ok {
			t.Encoding = encoding
			return nil
		}
		return fmt.Errorf("invalid NATS encoding client option recieved for transport type")
	}
}
