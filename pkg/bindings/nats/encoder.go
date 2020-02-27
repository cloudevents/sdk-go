package nats

import (
	"context"
	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/binding/format"
	"github.com/cloudevents/sdk-go/pkg/binding/spec"
	"github.com/nats-io/nats.go"
	"io"
	"io/ioutil"
)

type natsMsgEncoder nats.Msg

func EncodeNatsMsg(ctx context.Context, m binding.Message, msg *nats.Msg, factories binding.TransformerFactories) error {
	structuredEncoder := (*natsMsgEncoder)(msg)

	_, err := binding.Encode(
		ctx,
		m,
		structuredEncoder,
		nil,
		factories,
	)
	return err
}

func (e *natsMsgEncoder) SetStructuredEvent(ctx context.Context, format format.Format, event io.Reader) error {
	bytes, err := ioutil.ReadAll(event)
	if err != nil {
		return err
	}

	e.Data = bytes

	return nil
}

func (e *natsMsgEncoder) Start(ctx context.Context) error {
	return BinaryEncodingNotSupported
}

func (e *natsMsgEncoder) SetAttribute(attribute spec.Attribute, value interface{}) error {
	return BinaryEncodingNotSupported
}

func (e *natsMsgEncoder) SetExtension(name string, value interface{}) error {
	return BinaryEncodingNotSupported
}

func (e *natsMsgEncoder) SetData(data io.Reader) error {
	return BinaryEncodingNotSupported
}

func (e *natsMsgEncoder) End() error {
	return BinaryEncodingNotSupported
}

var _ binding.StructuredEncoder = (*natsMsgEncoder)(nil)
var _ binding.BinaryEncoder = (*natsMsgEncoder)(nil)
