package event

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"

	"github.com/cloudevents/sdk-go/pkg/event/datacodec"
)

// Data is special. Break it out into it's own file.

// SetData implements EventWriter.SetData
func (e *Event) SetData(obj interface{}, contentType string) error {
	e.SetDataContentType(contentType)

	if e.SpecVersion() != CloudEventsVersionV1 {
		return e.legacySetData(obj)
	}

	// Version 1.0 and above.
	switch obj := obj.(type) {
	case []byte:
		e.DataEncoded = obj
		e.DataBinary = true
	default:
		data, err := datacodec.Encode(context.Background(), e.DataMediaType(), obj)
		if err != nil {
			return err
		}
		e.DataEncoded = data
		e.DataBinary = false
	}

	return nil
}

// Deprecated: Delete when we do not have to support Spec v0.3.
func (e *Event) legacySetData(obj interface{}) error {
	data, err := datacodec.Encode(context.Background(), e.DataMediaType(), obj)
	if err != nil {
		return err
	}
	if e.DeprecatedDataContentEncoding() == Base64 {
		buf := make([]byte, base64.StdEncoding.EncodedLen(len(data)))
		base64.StdEncoding.Encode(buf, data)
		e.DataEncoded = buf
		e.DataBinary = false
	} else {
		data, err := datacodec.Encode(context.Background(), e.DataMediaType(), obj)
		if err != nil {
			return err
		}
		e.DataEncoded = data
		e.DataBinary = false
	}
	return nil
}

const (
	quotes = `"'`
)

func (e Event) Data() []byte {
	return e.DataEncoded
}

// DataAs attempts to populate the provided data object with the event payload.
// data should be a pointer type.
func (e Event) DataAs(obj interface{}) error {
	data := e.Data()

	if len(data) == 0 {
		// No data.
		return nil
	}

	if e.SpecVersion() != CloudEventsVersionV1 {
		var err error
		if data, err = e.legacyConvertData(data); err != nil {
			return err
		}
	}

	return datacodec.Decode(context.Background(), e.DataMediaType(), data, obj)
}

func (e Event) legacyConvertData(data []byte) ([]byte, error) {
	if e.Context.DeprecatedGetDataContentEncoding() == Base64 {
		var bs []byte
		// test to see if we need to unquote the data.
		if data[0] == quotes[0] || data[0] == quotes[1] {
			str, err := strconv.Unquote(string(data))
			if err != nil {
				return nil, err
			}
			bs = []byte(str)
		} else {
			bs = data
		}

		buf := make([]byte, base64.StdEncoding.DecodedLen(len(bs)))
		n, err := base64.StdEncoding.Decode(buf, bs)
		if err != nil {
			return nil, fmt.Errorf("failed to decode data from base64: %s", err.Error())
		}
		data = buf[:n]
	}

	return data, nil
}
