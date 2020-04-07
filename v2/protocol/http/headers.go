package http

import (
	"net/textproto"

	"github.com/cloudevents/sdk-go/v2/binding/spec"
)

var attributeHeadersMapping map[string]string

func init() {
	attributeHeadersMapping = make(map[string]string)
	for _, v := range specs.Versions() {
		for _, a := range v.Attributes() {
			if a.Kind() == spec.DataContentType {
				attributeHeadersMapping[a.Name()] = ContentType
			} else {
				attributeHeadersMapping[a.Name()] = textproto.CanonicalMIMEHeaderKey(prefix + a.Name())
			}
		}
	}
}
