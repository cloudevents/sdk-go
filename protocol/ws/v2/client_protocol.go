package v2

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"

	"nhooyr.io/websocket"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/binding/format"
	"github.com/cloudevents/sdk-go/v2/protocol"
)

type ClientProtocol struct {
	conn *websocket.Conn

	format      format.Format
	messageType websocket.MessageType
}

// Dial wraps websocket.Dial and creates the ClientProtocol.
func Dial(ctx context.Context, u string, opts *websocket.DialOptions) (*ClientProtocol, error) {
	if opts == nil {
		opts = &websocket.DialOptions{}
	}
	opts.Subprotocols = SupportedSubprotocols
	c, _, err := websocket.Dial(ctx, u, opts)
	if err != nil {
		return nil, err
	}
	return NewClientProtocol(c)
}

// NewClientProtocol wraps a websocket.Conn in a type that implements protocol.Receiver, protocol.Sender and protocol.Closer.
func NewClientProtocol(c *websocket.Conn) (*ClientProtocol, error) {
	format, messageType, err := resolveFormat(c.Subprotocol())
	if err != nil {
		return nil, err
	}
	return &ClientProtocol{
		conn:        c,
		format:      format,
		messageType: messageType,
	}, nil
}

func (c ClientProtocol) Send(ctx context.Context, m binding.Message, transformers ...binding.Transformer) error {
	writer, err := c.conn.Writer(ctx, c.messageType)
	if err != nil {
		return err
	}
	return WriteWriter(ctx, m, writer, transformers...)
}

func (c ClientProtocol) Receive(ctx context.Context) (binding.Message, error) {
	messageType, reader, err := c.conn.Reader(ctx)
	if err == io.EOF {
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

	return binding.NewStructuredMessage(c.format, reader), nil
}

func consumeStream(reader io.Reader) {
	//TODO is there a less expensive way to consume the stream?
	ioutil.ReadAll(reader)
}

func (c ClientProtocol) Close(ctx context.Context) error {
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

var _ protocol.Receiver = ClientProtocol{}
var _ protocol.Sender = ClientProtocol{}
var _ protocol.Closer = ClientProtocol{}
