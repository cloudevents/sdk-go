package http

import (
	"bytes"
	"io/ioutil"
	"net/http"

	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/binding/format"
	"github.com/cloudevents/sdk-go/pkg/binding/spec"
	ce "github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/types"
)

type eventToBinaryMessageBuilder struct {
	req *http.Request
}

var _ binding.EventMessageBuilder = (*eventToBinaryMessageBuilder)(nil) // Test it conforms to the interface

func (b *eventToBinaryMessageBuilder) Encode(e ce.Event) error {
	version, err := specs.Version(e.SpecVersion())
	if err != nil {
		return err
	}
	attrs := version.Attributes()
	ext := e.Extensions()
	for _, a := range attrs {
		if a.Kind() == spec.DataContentType {
			b.req.Header.Set(ContentType, e.DataContentType()) // Special header name.
		} else {
			if v := a.Get(e.Context); v != nil {
				s, err := types.Format(v)
				if err != nil {
					return err
				}
				b.req.Header.Set(a.Name(), s)
			}
		}
	}
	for k, v := range ext { // Extension attributes
		s, err := types.Format(v)
		if err != nil {
			return err
		}
		b.req.Header.Set(prefix+k, s)
	}

	data, err := e.DataBytes()
	if err != nil {
		return err
	}
	b.req.Body = ioutil.NopCloser(bytes.NewReader(data))

	return nil
}

type eventToStructuredMessageBuilder struct {
	format format.Format
	req    *http.Request
}

var _ binding.EventMessageBuilder = (*eventToStructuredMessageBuilder)(nil) // Test it conforms to the interface

func (b *eventToStructuredMessageBuilder) Encode(event ce.Event) error {
	data, err := b.format.Marshal(event)
	if err != nil {
		return err
	}
	b.req.Body = ioutil.NopCloser(bytes.NewReader(data))
	b.req.Header.Set(ContentType, b.format.MediaType())

	return nil
}
