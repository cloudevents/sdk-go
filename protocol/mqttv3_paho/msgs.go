package mqttv3_paho

import (
	"bytes"
	"context"
	"io"

	ce "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/binding/format"
)

type message []byte

var (
	_ binding.StructuredWriter = (*message)(nil)
	_ binding.Message          = (*message)(nil)
)

func NewMessage(b []byte) binding.Message {
	m := message(b)
	return &m
}

func WritePubMessage(ctx context.Context, m binding.Message, transformers ...binding.Transformer) ([]byte, error) {
	ctx = binding.WithForceStructured(ctx)

	var msg message

	_, err := binding.Write(
		ctx,
		m,
		&msg,
		nil,
		transformers...,
	)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

func (m *message) SetStructuredEvent(_ context.Context, _ format.Format, event io.Reader) error {
	res, err := io.ReadAll(event)
	if err != nil {
		return err
	}

	*m = res
	return nil
}

func (m *message) ReadEncoding() binding.Encoding {
	var ev ce.Event
	if err := ev.UnmarshalJSON(*m); err != nil {
		return binding.EncodingUnknown
	}

	if err := ev.Validate(); err != nil {
		return binding.EncodingUnknown
	}

	return binding.EncodingStructured
}

func (m *message) ReadStructured(ctx context.Context, encoder binding.StructuredWriter) error {
	return encoder.SetStructuredEvent(ctx, format.JSON, bytes.NewReader(*m))
}

func (m *message) ReadBinary(context.Context, binding.BinaryWriter) error {
	return binding.ErrNotBinary
}

func (m *message) Finish(error) error {
	return nil
}
