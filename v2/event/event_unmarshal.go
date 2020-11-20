package event

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/cloudevents/sdk-go/v2/observability"
	"github.com/cloudevents/sdk-go/v2/types"
)

// TODO TBD where to put this stuff

const BeginObject uint8 = 1
const SpecVersionV03Flag uint8 = 1 << 4
const SpecVersionV1Flag uint8 = 1 << 4
const DataContentEncodingFlag uint8 = 1 << 5
const DataContentTypeFlag uint8 = 1 << 6
const EndObject uint8 = 1 << 7

func checkFlag(state uint8, flag uint8) bool {
	return state&flag != 0
}

func appendFlag(state *uint8, flag uint8) {
	*state = (*state) | flag
}

func nextString(dec *json.Decoder, key string) (string, error) {
	token, err := dec.Token()
	if err != nil {
		return "", err
	}

	if val, ok := token.(string); !ok {
		return "", fmt.Errorf("%s should be a string, actual '%v'", key, token)
	} else {
		return val, nil
	}
}

func nextRaw(dec *json.Decoder) (json.RawMessage, error) {
	raw := json.RawMessage{}
	err := dec.Decode(&raw)
	if err != nil {
		return nil, err
	}
	return raw, nil
}

func drainTokenQueue(tokenQueue []entry, event *Event, state *uint8, dataToken **entry) error {
	switch ctx := event.Context.(type) {
	case *EventContextV03:
		for _, e := range tokenQueue {
			switch e.key {
			case "id":
				if err := json.Unmarshal(e.value, &ctx.ID); err != nil {
					return err
				}
			case "type":
				if err := json.Unmarshal(e.value, &ctx.Type); err != nil {
					return err
				}
			case "source":
				if err := json.Unmarshal(e.value, &ctx.Source); err != nil {
					return err
				}
			case "subject":
				ctx.Subject = new(string)
				if err := json.Unmarshal(e.value, ctx.Subject); err != nil {
					return err
				}
			case "time":
				ctx.Time = new(types.Timestamp)
				if err := json.Unmarshal(e.value, ctx.Time); err != nil {
					return err
				}
			case "schemaurl":
				ctx.SchemaURL = new(types.URIRef)
				if err := json.Unmarshal(e.value, ctx.SchemaURL); err != nil {
					return err
				}
			case "datacontenttype":
				ctx.DataContentType = new(string)
				if err := json.Unmarshal(e.value, ctx.DataContentType); err != nil {
					return err
				}
				appendFlag(state, DataContentTypeFlag)
			case "datacontentencoding":
				ctx.DataContentEncoding = new(string)
				if err := json.Unmarshal(e.value, ctx.DataContentEncoding); err != nil {
					return err
				}
				if *ctx.DataContentEncoding != Base64 {
					return fmt.Errorf("invalid datacontentencoding value: '%s'", *ctx.DataContentEncoding)
				}
				appendFlag(state, DataContentEncodingFlag)
			case "data":
				*dataToken = &entry{key: e.key, value: e.value}
			default:
				var val interface{}
				if err := json.Unmarshal(e.value, &val); err != nil {
					return err
				}
				if ctx.Extensions == nil {
					ctx.Extensions = make(map[string]interface{}, 1)
				}
				if err := ctx.SetExtension(e.key, val); err != nil {
					return err
				}
			}
		}
	case *EventContextV1:
		for _, e := range tokenQueue {
			switch e.key {
			case "id":
				if err := json.Unmarshal(e.value, &ctx.ID); err != nil {
					return err
				}
			case "type":
				if err := json.Unmarshal(e.value, &ctx.Type); err != nil {
					return err
				}
			case "source":
				if err := json.Unmarshal(e.value, &ctx.Source); err != nil {
					return err
				}
			case "subject":
				ctx.Subject = new(string)
				if err := json.Unmarshal(e.value, ctx.Subject); err != nil {
					return err
				}
			case "time":
				ctx.Time = new(types.Timestamp)
				if err := json.Unmarshal(e.value, ctx.Time); err != nil {
					return err
				}
			case "dataschema":
				ctx.DataSchema = new(types.URI)
				if err := json.Unmarshal(e.value, ctx.DataSchema); err != nil {
					return err
				}
			case "datacontenttype":
				ctx.DataContentType = new(string)
				if err := json.Unmarshal(e.value, ctx.DataContentType); err != nil {
					return err
				}
				appendFlag(state, DataContentTypeFlag)
			case "data":
				*dataToken = &entry{key: e.key, value: e.value}
			case "data_base64":
				*dataToken = &entry{key: e.key, value: e.value}
			default:
				var val interface{}
				if err := json.Unmarshal(e.value, &val); err != nil {
					return err
				}
				if ctx.Extensions == nil {
					ctx.Extensions = make(map[string]interface{}, 1)
				}
				if err := ctx.SetExtension(e.key, val); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func bestCaseProcessNext(dec *json.Decoder, state uint8, key string, e *Event) error {
	switch ctx := e.Context.(type) {
	case *EventContextV03:
		switch key {
		case "id":
			return dec.Decode(&ctx.ID)
		case "type":
			return dec.Decode(&ctx.Type)
		case "source":
			return dec.Decode(&ctx.Source)
		case "subject":
			ctx.Subject = new(string)
			return dec.Decode(ctx.Subject)
		case "time":
			ctx.Time = new(types.Timestamp)
			return dec.Decode(ctx.Time)
		case "schemaurl":
			ctx.SchemaURL = new(types.URIRef)
			return dec.Decode(ctx.SchemaURL)
		case "data":
			return consumeData(e, state, key, dec)
		default:
			var val interface{}
			if err := dec.Decode(&val); err != nil {
				return err
			}
			if ctx.Extensions == nil {
				ctx.Extensions = make(map[string]interface{}, 1)
			}
			return ctx.SetExtension(key, val)
		}
	case *EventContextV1:
		switch key {
		case "id":
			return dec.Decode(&ctx.ID)
		case "type":
			return dec.Decode(&ctx.Type)
		case "source":
			return dec.Decode(&ctx.Source)
		case "subject":
			ctx.Subject = new(string)
			return dec.Decode(ctx.Subject)
		case "time":
			ctx.Time = new(types.Timestamp)
			return dec.Decode(ctx.Time)
		case "dataschema":
			ctx.DataSchema = new(types.URI)
			return dec.Decode(ctx.DataSchema)
		case "data":
			return consumeData(e, state, key, dec)
		case "data_base64":
			return consumeData(e, state, key, dec)
		default:
			var val interface{}
			if err := dec.Decode(&val); err != nil {
				return err
			}
			if ctx.Extensions == nil {
				ctx.Extensions = make(map[string]interface{}, 1)
			}
			return ctx.SetExtension(key, val)
		}
	}
	return nil
}

type entry struct {
	key   string
	value json.RawMessage
}

//TODO make this public?
func unmarshalJSON(reader io.Reader, e *Event) error {
	dec := json.NewDecoder(reader)

	// Parsing dependency graph:
	//         SpecVersion
	//          ^     ^
	//          |     +--------------+
	//          +                    +
	//  All Attributes           datacontenttype (and datacontentencoding for v0.3)
	//  (except datacontenttype)     ^
	//                               |
	//                               |
	//                               +
	//                              Data

	var state uint8 = 0
	var tokenQueue []entry
	var dataToken *entry

	for {
		token, err := dec.Token()
		if err == io.EOF {
			if !checkFlag(state, EndObject) {
				return errors.New("unexpected EOF, there isn't any '}' symbol in the input")
			}
			// If there is a dataToken cached, we always defer at the end the processing
			// because nor datacontenttype or datacontentencoding are mandatory.
			if dataToken != nil {
				if err := drainData(e, state, dataToken.key, dataToken.value); err != nil {
					return err
				}
			}
			return nil
		}
		if err != nil {
			return err
		}

		// Let's check if we're in the main object
		if !checkFlag(state, BeginObject) {
			if delim, ok := token.(json.Delim); !ok || delim.String() != "{" {
				return errors.New("CloudEvent should be a json object")
			}
			appendFlag(&state, BeginObject)
			continue
		}

		// We need to figure out if this token is a key or the end of the object
		var key string
		switch token.(type) {
		case json.Delim:
			if token.(json.Delim).String() == "}" {
				appendFlag(&state, EndObject)
				continue
			}
		case string:
			key = token.(string)
			break
		default:
			return fmt.Errorf("unexpected token '%v', expecting a string or '}'", token)
		}

		// We have a key, now we need to figure out what to do
		// depending on the parsing state

		// If it's a specversion, trigger state change
		if key == "specversion" {
			if checkFlag(state, SpecVersionV1Flag|SpecVersionV03Flag) {
				return fmt.Errorf("specversion was already provided")
			}
			sv, err := nextString(dec, key)
			if err != nil {
				return err
			}

			// Check proper specversion
			switch sv {
			case CloudEventsVersionV1:
				e.Context = &EventContextV1{}
				appendFlag(&state, SpecVersionV1Flag)
			case CloudEventsVersionV03:
				e.Context = &EventContextV03{}
				appendFlag(&state, SpecVersionV03Flag)
			default:
				return fmt.Errorf("unexpected specversion '%s'", token)
			}

			if err := drainTokenQueue(tokenQueue, e, &state, &dataToken); err != nil {
				return err
			}
			continue
		}

		// If no specversion, enqueue unconditionally
		if !checkFlag(state, SpecVersionV03Flag|SpecVersionV1Flag) {
			raw, err := nextRaw(dec)
			if err != nil {
				return err
			}
			tokenQueue = append(tokenQueue, entry{
				key:   key,
				value: raw,
			})
			continue
		}

		// From this point downward -> we can assume the event has a context pointer non nil

		// If it's a datacontenttype, trigger state change
		if key == "datacontenttype" {
			if checkFlag(state, DataContentTypeFlag) {
				return fmt.Errorf("datacontenttype was already provided")
			}

			dct, err := nextString(dec, key)
			if err != nil {
				return err
			}

			e.SetDataContentType(dct)
			appendFlag(&state, DataContentTypeFlag)
			continue
		}

		// If it's a datacontentencoding and it's v0.3, trigger state change
		if key == "datacontentencoding" && checkFlag(state, SpecVersionV03Flag) {
			if checkFlag(state, DataContentEncodingFlag) {
				return fmt.Errorf("datacontentencoding was already provided")
			}

			dct, err := nextString(dec, key)
			if err != nil {
				return err
			}

			if dct != Base64 {
				return fmt.Errorf("invalid datacontentencoding value: '%s'", dct)
			}

			e.Context.(*EventContextV03).DataContentEncoding = &dct
			appendFlag(&state, DataContentEncodingFlag)
			continue
		}

		// We can parse all attributes, except data.
		// If it's data or data_base64 and we don't have the attributes to process it, then we enqueue
		if ((key == "data" || key == "data_base64") && !checkFlag(state, SpecVersionV1Flag&DataContentTypeFlag)) &&
			key == "data" && checkFlag(state, SpecVersionV03Flag&DataContentTypeFlag&DataContentEncodingFlag) {
			raw, err := nextRaw(dec)
			if err != nil {
				return err
			}
			dataToken = &entry{
				key:   key,
				value: raw,
			}
			continue
		}

		// At this point or this value is an attribute (excluding datacontenttype and datacontentencoding), or this value is data and this condition is valid:
		// (SpecVersionV1Flag & DataContentTypeFlag) || (SpecVersionV03Flag & DataContentTypeFlag & DataContentEncodingFlag)
		if err := bestCaseProcessNext(dec, state, key, e); err != nil {
			return err
		}
	}
}

func drainData(e *Event, state uint8, key string, value json.RawMessage) error {
	if checkFlag(state, DataContentEncodingFlag) || key == "data_base64" {
		e.DataBase64 = true
		return json.Unmarshal(value, &e.DataEncoded)
	}

	ct := e.DataContentType()
	if ct != ApplicationJSON && ct != TextJSON {
		// If not json, then data is encoded as string
		var dataStr string
		if err := json.Unmarshal(value, &dataStr); err != nil {
			return err
		}
		e.DataEncoded = []byte(dataStr)
	}
	e.DataEncoded = value
	return nil
}

func consumeData(e *Event, state uint8, key string, dec *json.Decoder) error {
	if checkFlag(state, DataContentEncodingFlag) || key == "data_base64" {
		return dec.Decode(&e.DataEncoded)
	}

	ct := e.DataContentType()
	if ct != ApplicationJSON && ct != TextJSON {
		// If not json, then data is encoded as string
		var dataStr string
		if err := dec.Decode(&dataStr); err != nil {
			return err
		}
		e.DataEncoded = []byte(dataStr)
	}
	var value json.RawMessage
	if err := dec.Decode(&value); err != nil {
		return err
	}
	e.DataEncoded = value
	return nil
}

// UnmarshalJSON implements the json unmarshal method used when this type is
// unmarshaled using json.Unmarshal.
func (e *Event) UnmarshalJSON(b []byte) error {
	//TODO wrap observability later
	_, r := observability.NewReporter(context.Background(), eventJSONObserved{o: reportUnmarshal})

	err := unmarshalJSON(bytes.NewReader(b), e)

	// Report the observable
	if err != nil {
		r.Error()
	} else {
		r.OK()
	}
	return err

}
