package http

import (
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/binding/format"
	"github.com/cloudevents/sdk-go/pkg/binding/spec"
	"github.com/cloudevents/sdk-go/pkg/types"
)

// Write out to the the provided httpResponseWriter with the message m.
// Using context you can tweak the encoding processing (more details on binding.Encode documentation).
func EncodeHttpResponseWriter(ctx context.Context, m binding.Message, rw http.ResponseWriter, transformers binding.TransformerFactories) error {
	structuredEncoder := &httpResponseWriterEncoder{rw: rw}
	binaryEncoder := &httpResponseWriterEncoder{rw: rw}

	_, err := binding.Encode(
		ctx,
		m,
		structuredEncoder,
		binaryEncoder,
		transformers,
	)
	return err
}

type httpResponseWriterEncoder struct {
	rw     http.ResponseWriter
	status int
}

func (b *httpResponseWriterEncoder) SetStructuredEvent(ctx context.Context, format format.Format, event io.Reader) error {
	b.rw.Header().Set(ContentType, format.MediaType())
	return b.SetData(event)
}

func (b *httpResponseWriterEncoder) Start(ctx context.Context) error {
	return nil
}

func (b *httpResponseWriterEncoder) End() error {
	return nil
}

func (b *httpResponseWriterEncoder) SetData(reader io.Reader) error {
	body, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}
	n, err := b.rw.Write(body)
	b.rw.Header().Set(ContentLength, strconv.Itoa(n))
	return nil
}

func (b *httpResponseWriterEncoder) SetAttribute(attribute spec.Attribute, value interface{}) error {
	// Http headers, everything is a string!
	s, err := types.Format(value)
	if err != nil {
		return err
	}

	if attribute.Kind() == spec.DataContentType {
		b.rw.Header().Add(ContentType, s)
	} else {
		b.rw.Header().Add(prefix+attribute.Name(), s)
	}
	return nil
}

func (b *httpResponseWriterEncoder) SetExtension(name string, value interface{}) error {
	// Http headers, everything is a string!
	s, err := types.Format(value)
	if err != nil {
		return err
	}
	b.rw.Header().Add(prefix+name, s)
	return nil
}

var _ binding.StructuredEncoder = (*httpResponseEncoder)(nil) // Test it conforms to the interface
var _ binding.BinaryEncoder = (*httpResponseEncoder)(nil)     // Test it conforms to the interface
