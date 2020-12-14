package v2

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"sync"

	"nhooyr.io/websocket"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/binding/format"
	"github.com/cloudevents/sdk-go/v2/binding/utils"
	"github.com/cloudevents/sdk-go/v2/protocol"
)

// Protocol implements protocol.Receiver, protocol.Sender and protocol.Closer.
// Note: when you use client.StartReceiver with this protocol, you can use just one
// goroutine to poll this protocol, because the protocol itself cannot handle multiple
// received messages at same time (WS has no multiplexing!)
type Protocol struct {
	conn *websocket.Conn

	format      format.Format
	messageType websocket.MessageType

	receiverLock sync.Mutex
	connOwned    bool // whether this protocol created the connection
}

// Dial wraps websocket.Dial and creates the Protocol.
func Accept(ctx context.Context, w http.ResponseWriter, r *http.Request, opts *websocket.AcceptOptions) (*Protocol, error) {
	if opts == nil {
		opts = &websocket.AcceptOptions{}
	}
	opts.Subprotocols = SupportedSubprotocols
	c, err := websocket.Accept(w, r, opts)
	if err != nil {
		return nil, err
	}
	p, err := NewProtocol(c)
	if err != nil {
		return nil, err
	}
	p.connOwned = true
	return p, nil
}

// Dial wraps websocket.Dial and creates the Protocol.
func Dial(ctx context.Context, u string, opts *websocket.DialOptions) (*Protocol, error) {
	if opts == nil {
		opts = &websocket.DialOptions{}
	}
	opts.Subprotocols = SupportedSubprotocols
	c, _, err := websocket.Dial(ctx, u, opts)
	if err != nil {
		return nil, err
	}
	p, err := NewProtocol(c)
	if err != nil {
		return nil, err
	}
	p.connOwned = true
	return p, nil
}

// NewProtocol wraps a websocket.Conn in a type that implements protocol.Receiver, protocol.Sender and protocol.Closer.
// Look at Protocol for more details.
func NewProtocol(c *websocket.Conn) (*Protocol, error) {
	f, messageType, err := resolveFormat(c.Subprotocol())
	if err != nil {
		return nil, err
	}
	return &Protocol{
		conn:        c,
		format:      f,
		messageType: messageType,
		connOwned:   false,
	}, nil
}

func (c *Protocol) Send(ctx context.Context, m binding.Message, transformers ...binding.Transformer) error {
	writer, err := c.conn.Writer(ctx, c.messageType)
	if err != nil {
		return err
	}
	return utils.WriteStructured(ctx, m, writer, transformers...)
}

func (c *Protocol) Receive(ctx context.Context) (binding.Message, error) {
	c.receiverLock.Lock()
	m, err := c.UnsafeReceive(ctx)
	if m != nil {
		m = binding.WithFinish(m, func(err error) {
			c.receiverLock.Unlock()
		})
	} else {
		c.receiverLock.Unlock()
	}
	return m, err
}

// UnsafeReceive is like Receive, except it doesn't guard from multiple invocations
// from different goroutines.
func (c *Protocol) UnsafeReceive(ctx context.Context) (binding.Message, error) {
	messageType, reader, err := c.conn.Reader(ctx)
	if errors.Is(err, io.EOF) || errors.As(err, &websocket.CloseError{}) || (ctx.Err() != nil && errors.Is(err, ctx.Err())) {
		return nil, io.EOF
	}
	if err != nil {
		return nil, err
	}

	if messageType != c.messageType {
		// We need to consume the stream, otherwise it won't be possible to consume the stream
		consumeStream(reader)
		return nil, fmt.Errorf("wrong message type: %s, expected %s", messageType, c.messageType)
	}

	return utils.NewStructuredMessage(c.format, reader), nil
}

func consumeStream(reader io.Reader) {
	//TODO is there a less expensive way to consume the stream?
	ioutil.ReadAll(reader)
}

func (c *Protocol) Close(ctx context.Context) error {
	if c.connOwned {
		statusCode := websocket.StatusNormalClosure
		if val := ctx.Value(codeKey{}); val != nil {
			statusCode = val.(websocket.StatusCode)
		}

		reason := ""
		if val := ctx.Value(reasonKey{}); val != nil {
			reason = val.(string)
		}

		return c.conn.Close(statusCode, reason)
	}
	return nil
}

var _ protocol.Receiver = (*Protocol)(nil)
var _ protocol.Sender = (*Protocol)(nil)
var _ protocol.Closer = (*Protocol)(nil)
