package http

import (
	"io"
	"io/ioutil"
	"net/http"

	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/binding/format"
	"github.com/cloudevents/sdk-go/pkg/binding/spec"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/types"
)

//TODO (slinkydeveloper) this is the public access to http encoders, document it
func EncodeHttpRequest(m binding.Message, req *http.Request, forceStructured bool, forceBinary bool, transformerFactories binding.TransformerFactories) error {
	var structuredEncoder binding.StructuredEncoder
	if !forceBinary {
		structuredEncoder = (*httpRequestEncoder)(req)
	}

	var binaryEncoder binding.BinaryEncoder
	if !forceStructured {
		binaryEncoder = (*httpRequestEncoder)(req)
	}

	var preferredEventEncoding binding.Encoding
	if forceStructured {
		preferredEventEncoding = binding.EncodingStructured
	} else {
		preferredEventEncoding = binding.EncodingBinary
	}

	_, err := binding.Encode(
		m,
		structuredEncoder,
		binaryEncoder,
		transformerFactories,
		preferredEventEncoding,
	)
	return err
}

type httpRequestEncoder http.Request

func (b *httpRequestEncoder) Start() error {
	return nil
}

func (b *httpRequestEncoder) End() error {
	return nil
}

func (b *httpRequestEncoder) SetStructuredEvent(format format.Format, event io.Reader) error {
	b.Header.Set(ContentType, format.MediaType())
	b.Body = ioutil.NopCloser(event)
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

var _ binding.BinaryEncoder = (*httpRequestEncoder)(nil)     // Test it conforms to the interface
var _ binding.StructuredEncoder = (*httpRequestEncoder)(nil) // Test it conforms to the interface
