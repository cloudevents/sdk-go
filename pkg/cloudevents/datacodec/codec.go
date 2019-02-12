package datacodec

import (
	"fmt"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/datacodec/json"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/datacodec/xml"
)

type Decoder func(in, out interface{}) error
type Encoder func(in, out interface{}) error

var decoder map[string]Decoder
var encoder map[string]Encoder

func init() {
	decoder = make(map[string]Decoder, 10)

	decoder[""] = json.Decode
	decoder["application/json"] = json.Decode
	decoder["application/xml"] = xml.Decode
}

func Decode(contentType string, in, out interface{}) error {
	if decode, ok := decoder[contentType]; ok {
		return decode(in, out)
	}
	return fmt.Errorf("[decode] unsupported content type: %q", contentType)
}

func Encode(contentType string, in, out interface{}) error {
	if decode, ok := encoder[contentType]; ok {
		return decode(in, out)
	}
	return fmt.Errorf("[encode] unsupported content type: %q", contentType)
}
