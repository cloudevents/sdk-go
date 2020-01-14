package http

import (
	"io/ioutil"
	"net/http"

	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/binding/spec"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/types"
)

type binaryMessageEncoder struct {
	req *http.Request
}

var _ binding.BinaryEncoder = (*binaryMessageEncoder)(nil) // Test it conforms to the interface

func (b *binaryMessageEncoder) SetData(payload binding.MessagePayloadReader) error {
	if !payload.IsEmpty() {
		b.req.Body = ioutil.NopCloser(payload.Reader())
	}
	return nil
}

func (b *binaryMessageEncoder) SetAttribute(attribute spec.Attribute, value interface{}) error {
	// Http headers, everything is a string!
	s, err := types.Format(value)
	if err != nil {
		return err
	}

	if attribute.Kind() == spec.DataContentType {
		b.req.Header.Add(ContentType, s)
	} else {
		b.req.Header.Add(prefix+attribute.Name(), s)
	}
	return nil
}

func (b *binaryMessageEncoder) SetExtension(name string, value interface{}) error {
	// Http headers, everything is a string!
	s, err := types.Format(value)
	if err != nil {
		return err
	}
	b.req.Header.Add(prefix+name, s)
	return nil
}
