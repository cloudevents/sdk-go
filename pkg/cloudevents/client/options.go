package client

import (
	"context"
	"fmt"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport/http"
	"net/url"
)

type ClientOption func(*Client) error

// WithContext sets the context on the client.
func WithContext(ctx context.Context) ClientOption {
	return func(c *Client) error {
		if ctx == nil {
			return fmt.Errorf("context client option was nil")
		}
		c.ctx = ctx
		return nil
	}
}

// WithTarget sets the outbound recipient of cloudevents when using a http request.
func WithTarget(targetUrl string) ClientOption {
	return func(c *Client) error {
		if targetUrl != "" {
			var err error
			var target *url.URL
			target, err = url.Parse(targetUrl)
			if err != nil {
				return fmt.Errorf("client option failed to parse target url: %s", err.Error())
			}

			// Load the existing request and update it.
			req := http.RequestFromContext(c.ctx)
			req.URL = target

			// Store the request back into the context and client.
			c.ctx = http.ContextWithRequest(c.ctx, req)
		} else {
			return fmt.Errorf("target option was empty string")
		}
		return nil
	}
}

// WithTarget sets the outbound recipient of cloudevents when using a http request.
func WithHttpMethod(method string) ClientOption {
	return func(c *Client) error {
		if method != "" {
			// Load the existing request and update it.
			req := http.RequestFromContext(c.ctx)
			req.Method = method

			// Store the request back into the context and client.
			c.ctx = http.ContextWithRequest(c.ctx, req)
		} else {
			return fmt.Errorf("context client option was nil")
		}
		return nil
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
