package event

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cloudevents/sdk-go/v2/observability"
)

// MarshalJSON implements a custom json marshal method used when this type is
// marshaled using json.Marshal.
func (e Event) MarshalJSON() ([]byte, error) {
	_, r := observability.NewReporter(context.Background(), eventJSONObserved{o: reportMarshal, v: e.SpecVersion()})

	if err := e.Validate(); err != nil {
		r.Error()
		return nil, err
	}

	var b []byte
	var err error

	switch e.SpecVersion() {
	case CloudEventsVersionV03:
		b, err = JsonEncodeLegacy(e)
	case CloudEventsVersionV1:
		b, err = JsonEncode(e)
	default:
		return nil, ValidationError{"specversion": fmt.Errorf("unknown : %q", e.SpecVersion())}
	}

	// Report the observable
	if err != nil {
		r.Error()
		return nil, err
	} else {
		r.OK()
	}

	return b, nil
}

// JsonEncode encodes an event to JSON
func JsonEncode(e Event) ([]byte, error) {
	return jsonEncode(e.Context, e.DataEncoded, e.DataBase64)
}

// JsonEncodeLegacy performs legacy JSON encoding
func JsonEncodeLegacy(e Event) ([]byte, error) {
	isBase64 := e.Context.DeprecatedGetDataContentEncoding() == Base64
	return jsonEncode(e.Context, e.DataEncoded, isBase64)
}

func jsonEncode(ctx EventContextReader, data []byte, shouldEncodeToBase64 bool) ([]byte, error) {
	var b map[string]json.RawMessage
	var err error

	b, err = marshalEvent(ctx, ctx.GetExtensions())
	if err != nil {
		return nil, err
	}

	if data != nil {
		// data here is a serialized version of whatever payload.
		// If we need to write the payload as base64, shouldEncodeToBase64 is true.
		mediaType, err := ctx.GetDataMediaType()
		if err != nil {
			return nil, err
		}
		isJson := mediaType == "" || mediaType == ApplicationJSON || mediaType == TextJSON
		// If isJson and no encoding to base64, we don't need to perform additional steps
		if isJson && !shouldEncodeToBase64 {
			b["data"] = data
		} else {
			var dataKey = "data"
			if ctx.GetSpecVersion() == CloudEventsVersionV1 && shouldEncodeToBase64 {
				dataKey = "data_base64"
			}
			var dataPointer []byte
			if shouldEncodeToBase64 {
				dataPointer, err = json.Marshal(data)
			} else {
				dataPointer, err = json.Marshal(string(data))
			}
			if err != nil {
				return nil, err
			}

			b[dataKey] = dataPointer
		}
	}

	body, err := json.Marshal(b)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func marshalEvent(eventCtx EventContextReader, extensions map[string]interface{}) (map[string]json.RawMessage, error) {
	b, err := json.Marshal(eventCtx)
	if err != nil {
		return nil, err
	}

	brm := map[string]json.RawMessage{}
	if err := json.Unmarshal(b, &brm); err != nil {
		return nil, err
	}

	sv, err := json.Marshal(eventCtx.GetSpecVersion())
	if err != nil {
		return nil, err
	}

	brm["specversion"] = sv

	for k, v := range extensions {
		k = strings.ToLower(k)
		vb, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		// Don't overwrite spec keys.
		if _, ok := brm[k]; !ok {
			brm[k] = vb
		}
	}

	return brm, nil
}
