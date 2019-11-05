package http

import (
	"net/http"

	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/binding/format"
	"github.com/cloudevents/sdk-go/pkg/binding/spec"
	ce "github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/types"
)

func NewBinary(e ce.Event) (*Message, error) {
	data, err := e.DataBytes()
	if err != nil {
		return nil, err
	}
	version, err := specs.Version(e.SpecVersion())
	if err != nil {
		return nil, err
	}
	attrs := version.Attributes()
	ext := e.Extensions()
	header := make(http.Header, len(attrs)+len(ext))
	for _, a := range attrs {
		if a.Kind() == spec.DataContentType {
			header.Set(ContentType, e.DataContentType()) // Special header name.
		} else {
			if v := a.Get(e.Context); v != nil {
				s, err := types.Format(v)
				if err != nil {
					return nil, err
				}
				header.Set(a.Name(), s)
			}
		}
	}
	for k, v := range ext { // Extension attributes
		s, err := types.Format(v)
		if err != nil {
			return nil, err
		}
		header.Set(prefix+k, s)
	}
	return &Message{Header: header, Body: data}, nil
}

func NewStruct(mediaType string, data []byte) *Message {
	return &Message{Header: http.Header{ContentType: []string{mediaType}}, Body: data}
}

type BinaryEncoder struct{}

func (BinaryEncoder) Encode(e ce.Event) (binding.Message, error) { return NewBinary(e) }

type StructEncoder struct{ Format format.Format }

func (enc StructEncoder) Encode(e ce.Event) (binding.Message, error) {
	data, err := enc.Format.Marshal(e)
	if err != nil {
		return nil, err
	}
	return NewStruct(enc.Format.MediaType(), data), nil
}
