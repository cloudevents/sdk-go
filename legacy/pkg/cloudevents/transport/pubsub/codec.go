package pubsub

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/cloudevents/sdk-go/legacy/pkg/cloudevents"
	cecontext "github.com/cloudevents/sdk-go/legacy/pkg/cloudevents/context"
	"github.com/cloudevents/sdk-go/legacy/pkg/cloudevents/transport"
)

type Codec struct {
	Encoding Encoding

	// DefaultEncodingSelectionFn allows for encoding selection strategies to be injected.
	DefaultEncodingSelectionFn EncodingSelector

	v03 *CodecV03
	v1  *CodecV1

	_v03 sync.Once
	_v1  sync.Once
}

const (
	prefix = "ce-"
)

var _ transport.Codec = (*Codec)(nil)

func (c *Codec) loadCodec(encoding Encoding) (transport.Codec, error) {
	switch encoding {
	case Default:
		fallthrough
	case BinaryV1, StructuredV1:
		c._v1.Do(func() {
			c.v1 = &CodecV1{DefaultEncoding: c.Encoding}
		})
		return c.v1, nil
	case BinaryV03, StructuredV03:
		c._v03.Do(func() {
			c.v03 = &CodecV03{DefaultEncoding: c.Encoding}
		})
		return c.v03, nil
	}

	return nil, fmt.Errorf("unknown encoding: %s", encoding)
}

func (c *Codec) Encode(ctx context.Context, e cloudevents.Event) (transport.Message, error) {
	encoding := c.Encoding
	if encoding == Default && c.DefaultEncodingSelectionFn != nil {
		encoding = c.DefaultEncodingSelectionFn(ctx, e)
	}
	codec, err := c.loadCodec(encoding)
	if err != nil {
		return nil, err
	}
	ctx = cecontext.WithEncoding(ctx, encoding.Name())
	return codec.Encode(ctx, e)
}

func (c *Codec) Decode(ctx context.Context, msg transport.Message) (*cloudevents.Event, error) {
	codec, err := c.loadCodec(c.inspectEncoding(ctx, msg))
	if err != nil {
		return nil, err
	}
	event, err := codec.Decode(ctx, msg)
	if err != nil {
		return nil, err
	}
	return c.convertEvent(event)
}

// Give the context back as the user expects
func (c *Codec) convertEvent(event *cloudevents.Event) (*cloudevents.Event, error) {
	if event == nil {
		return nil, errors.New("event is nil, can not convert")
	}

	switch c.Encoding {
	case Default:
		return event, nil
	case BinaryV03, StructuredV03:
		ca := event.Context.AsV03()
		event.Context = ca
		return event, nil
	case BinaryV1, StructuredV1:
		ca := event.Context.AsV1()
		event.Context = ca
		return event, nil
	default:
		return nil, fmt.Errorf("unknown encoding: %s", c.Encoding)
	}
}

func (c *Codec) inspectEncoding(ctx context.Context, msg transport.Message) Encoding {
	// Try v1.0.
	_, _ = c.loadCodec(BinaryV1)
	encoding := c.v1.inspectEncoding(ctx, msg)
	if encoding != Unknown {
		return encoding
	}

	// Try v0.3.
	_, _ = c.loadCodec(BinaryV03)
	encoding = c.v03.inspectEncoding(ctx, msg)
	if encoding != Unknown {
		return encoding
	}

	// We do not understand the message encoding.
	return Unknown
}
