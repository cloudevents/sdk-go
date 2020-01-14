package http

import (
	"io"
	"io/ioutil"
	"net/http"

	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/binding/format"
)

type structuredMessageEncoder struct {
	req *http.Request
}

var _ binding.StructuredEncoder = (*structuredMessageEncoder)(nil) // Test it conforms to the interface

func (b *structuredMessageEncoder) SetStructuredEvent(format format.Format, event io.Reader) error {
	b.req.Header.Set(ContentType, format.MediaType())
	b.req.Body = ioutil.NopCloser(event)
	return nil
}
