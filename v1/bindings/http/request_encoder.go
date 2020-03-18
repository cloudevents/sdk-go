package http

import (
	"context"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/cloudevents/sdk-go/v1/binding"
	"github.com/cloudevents/sdk-go/v1/binding/format"
	"github.com/cloudevents/sdk-go/v1/binding/spec"
	"github.com/cloudevents/sdk-go/v1/cloudevents/types"
)

// Fill the provided req with the message m.
// Using context you can tweak the encoding processing (more details on binding.Translate documentation).
func EncodeHttpRequest(ctx context.Context, m binding.Message, req *http.Request, transformerFactories binding.TransformerFactories) error {
	structuredEncoder := (*httpRequestEncoder)(req)
	binaryEncoder := (*httpRequestEncoder)(req)

	_, err := binding.Encode(
		ctx,
		m,
		structuredEncoder,
		binaryEncoder,
		transformerFactories,
	)
	return err
}

type httpRequestEncoder http.Request

func (b *httpRequestEncoder) SetStructuredEvent(ctx context.Context, format format.Format, event io.Reader) error {
	b.Header.Set(ContentType, format.MediaType())
	b.Body = ioutil.NopCloser(event)
	return nil
}

func (b *httpRequestEncoder) Start(ctx context.Context) error {
	return nil
}

func (b *httpRequestEncoder) End() error {
	return nil
}

func (b *httpRequestEncoder) SetData(reader io.Reader) error {
	b.Body = ioutil.NopCloser(reader)
	return nil
}

func (b *httpRequestEncoder) SetAttribute(attribute spec.Attribute, value interface{}) error {
	// Http headers, everything is a string!
	s, err := types.Format(value)
	if err != nil {
		return err
	}

	if attribute.Kind() == spec.DataContentType {
		b.Header.Add(ContentType, s)
	} else {
		b.Header.Add(prefix+attribute.Name(), s)
	}
	return nil
}

func (b *httpRequestEncoder) SetExtension(name string, value interface{}) error {
	// Http headers, everything is a string!
	s, err := types.Format(value)
	if err != nil {
		return err
	}
	b.Header.Add(prefix+name, s)
	return nil
}

var _ binding.StructuredEncoder = (*httpRequestEncoder)(nil) // Test it conforms to the interface
var _ binding.BinaryEncoder = (*httpRequestEncoder)(nil)     // Test it conforms to the interface
