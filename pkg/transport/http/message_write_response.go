package http

import (
	"context"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/binding/format"
	"github.com/cloudevents/sdk-go/pkg/binding/spec"
	"github.com/cloudevents/sdk-go/pkg/types"
)

// Fill the provided httpResponse with the message m.
// Using context you can tweak the encoding processing (more details on binding.Write documentation).
func WriteResponse(ctx context.Context, m binding.Message, httpResponse *http.Response, transformers binding.TransformerFactories) error {
	structuredWriter := (*httpResponseWriter)(httpResponse)
	binaryWriter := (*httpResponseWriter)(httpResponse)

	_, err := binding.Write(
		ctx,
		m,
		structuredWriter,
		binaryWriter,
		transformers,
	)
	return err
}

type httpResponseWriter http.Response

func (b *httpResponseWriter) SetStructuredEvent(ctx context.Context, format format.Format, event io.Reader) error {
	b.Header.Set(ContentType, format.MediaType())
	b.Body = ioutil.NopCloser(event)
	return nil
}

func (b *httpResponseWriter) Start(ctx context.Context) error {
	return nil
}

func (b *httpResponseWriter) End() error {
	return nil
}

func (b *httpResponseWriter) SetData(reader io.Reader) error {
	b.Body = ioutil.NopCloser(reader)
	return nil
}

func (b *httpResponseWriter) SetAttribute(attribute spec.Attribute, value interface{}) error {
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

func (b *httpResponseWriter) SetExtension(name string, value interface{}) error {
	// Http headers, everything is a string!
	s, err := types.Format(value)
	if err != nil {
		return err
	}
	b.Header.Add(prefix+name, s)
	return nil
}

var _ binding.StructuredWriter = (*httpResponseWriter)(nil) // Test it conforms to the interface
var _ binding.BinaryWriter = (*httpResponseWriter)(nil)     // Test it conforms to the interface
