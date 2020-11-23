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

const SpecVersionV03Flag uint8 = 1 << 4
const SpecVersionV1Flag uint8 = 1 << 5
const DataBase64Flag uint8 = 1 << 6
const DataContentTypeFlag uint8 = 1 << 7

const preallocQueueSize = 8 // entry = 16 bytes, 8*16 = 128 bytes = 2 cache lines on x86_64

func checkFlag(state uint8, flag uint8) bool {
	return state&flag != 0
}

func appendFlag(state *uint8, flag uint8) {
	*state = (*state) | flag
}

func nextString(dec *json.Decoder, key string) (string, error) {
	var str string
	if err := dec.Decode(&str); err != nil {
		return "", fmt.Errorf("%s should be a valid json string: %w", key, err)
	}
	return str, nil
}

func nextRaw(dec *json.Decoder) (json.RawMessage, error) {
	raw := json.RawMessage{}
	err := dec.Decode(&raw)
	if err != nil {
		return nil, err
	}
	return raw, nil
}

func drainTokenQueue(tokenQueue *tokenQueue, event *Event, state *uint8, cachedDataPointer *json.RawMessage) error {
	switch ctx := event.Context.(type) {
	case *EventContextV03:
		for i := 0; i < tokenQueue.i; i++ {
			switch tokenQueue.slice[i].key {
			case "id":
				if err := json.Unmarshal(tokenQueue.slice[i].value, &ctx.ID); err != nil {
					return err
				}
			case "type":
				if err := json.Unmarshal(tokenQueue.slice[i].value, &ctx.Type); err != nil {
					return err
				}
			case "source":
				if err := json.Unmarshal(tokenQueue.slice[i].value, &ctx.Source); err != nil {
					return err
				}
			case "subject":
				ctx.Subject = new(string)
				if err := json.Unmarshal(tokenQueue.slice[i].value, ctx.Subject); err != nil {
					return err
				}
			case "time":
				ctx.Time = new(types.Timestamp)
				if err := json.Unmarshal(tokenQueue.slice[i].value, ctx.Time); err != nil {
					return err
				}
			case "schemaurl":
				ctx.SchemaURL = new(types.URIRef)
				if err := json.Unmarshal(tokenQueue.slice[i].value, ctx.SchemaURL); err != nil {
					return err
				}
			case "datacontenttype":
				ctx.DataContentType = new(string)
				if err := json.Unmarshal(tokenQueue.slice[i].value, ctx.DataContentType); err != nil {
					return err
				}
				appendFlag(state, DataContentTypeFlag)
			case "datacontentencoding":
				ctx.DataContentEncoding = new(string)
				if err := json.Unmarshal(tokenQueue.slice[i].value, ctx.DataContentEncoding); err != nil {
					return err
				}
				if *ctx.DataContentEncoding != Base64 {
					return fmt.Errorf("invalid datacontentencoding value: '%s'", *ctx.DataContentEncoding)
				}
				appendFlag(state, DataBase64Flag)
			case "data":
				*cachedDataPointer = tokenQueue.slice[i].value
			default:
				var val interface{}
				if err := json.Unmarshal(tokenQueue.slice[i].value, &val); err != nil {
					return err
				}
				if ctx.Extensions == nil {
					ctx.Extensions = make(map[string]interface{}, 1)
				}
				if err := ctx.SetExtension(tokenQueue.slice[i].key, val); err != nil {
					return err
				}
			}
		}
	case *EventContextV1:
		for i := 0; i < tokenQueue.i; i++ {
			switch tokenQueue.slice[i].key {
			case "id":
				if err := json.Unmarshal(tokenQueue.slice[i].value, &ctx.ID); err != nil {
					return err
				}
			case "type":
				if err := json.Unmarshal(tokenQueue.slice[i].value, &ctx.Type); err != nil {
					return err
				}
			case "source":
				if err := json.Unmarshal(tokenQueue.slice[i].value, &ctx.Source); err != nil {
					return err
				}
			case "subject":
				ctx.Subject = new(string)
				if err := json.Unmarshal(tokenQueue.slice[i].value, ctx.Subject); err != nil {
					return err
				}
			case "time":
				ctx.Time = new(types.Timestamp)
				if err := json.Unmarshal(tokenQueue.slice[i].value, ctx.Time); err != nil {
					return err
				}
			case "dataschema":
				ctx.DataSchema = new(types.URI)
				if err := json.Unmarshal(tokenQueue.slice[i].value, ctx.DataSchema); err != nil {
					return err
				}
			case "datacontenttype":
				ctx.DataContentType = new(string)
				if err := json.Unmarshal(tokenQueue.slice[i].value, ctx.DataContentType); err != nil {
					return err
				}
				appendFlag(state, DataContentTypeFlag)
			case "data":
				*cachedDataPointer = tokenQueue.slice[i].value
			case "data_base64":
				appendFlag(state, DataBase64Flag)
				*cachedDataPointer = tokenQueue.slice[i].value
			default:
				var val interface{}
				if err := json.Unmarshal(tokenQueue.slice[i].value, &val); err != nil {
					return err
				}
				if ctx.Extensions == nil {
					ctx.Extensions = make(map[string]interface{}, 1)
				}
				if err := ctx.SetExtension(tokenQueue.slice[i].key, val); err != nil {
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
			return consumeData(e, checkFlag(state, DataBase64Flag), dec)
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
			return consumeData(e, false, dec)
		case "data_base64":
			return consumeData(e, true, dec)
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
	tokenQueue := &tokenQueue{
		slice: make([]entry, preallocQueueSize),
		i:     0,
	}
	var cachedData json.RawMessage

	// Get first token, which should be '{'
	token, err := dec.Token()
	if err != nil {
		return err
	}
	if delim, ok := token.(json.Delim); !ok || delim.String() != "{" {
		return errors.New("CloudEvent should be a json object")
	}

	for dec.More() {
		token, err := dec.Token()
		if err != nil {
			return err
		}

		var key string
		if str, ok := token.(string); ok {
			key = str
		} else {
			return fmt.Errorf("unexpected token '%v', expecting a string", token)
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

			if err := drainTokenQueue(tokenQueue, e, &state, &cachedData); err != nil {
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
			tokenQueue.push(key, raw)
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
		if checkFlag(state, SpecVersionV03Flag) && key == "datacontentencoding" {
			if checkFlag(state, DataBase64Flag) {
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
			appendFlag(&state, DataBase64Flag)
			continue
		}

		// We can parse all attributes, except data.
		// If it's data or data_base64 and we don't have the attributes to process it, then we enqueue
		if (!checkFlag(state, SpecVersionV1Flag&DataContentTypeFlag) && (key == "data" || key == "data_base64")) &&
			(!checkFlag(state, SpecVersionV03Flag&DataContentTypeFlag&DataBase64Flag) && key == "data") {
			raw, err := nextRaw(dec)
			if err != nil {
				return err
			}
			if key == "data_base64" {
				appendFlag(&state, DataBase64Flag)
			}
			cachedData = raw
			continue
		}

		// At this point or this value is an attribute (excluding datacontenttype and datacontentencoding), or this value is data and this condition is valid:
		// (SpecVersionV1Flag & DataContentTypeFlag) || (SpecVersionV03Flag & DataContentTypeFlag & DataBase64Flag)
		if err := bestCaseProcessNext(dec, state, key, e); err != nil {
			return err
		}
	}

	// Get what I expect to be the last token
	token, err = dec.Token()
	if token.(json.Delim).String() != "}" {
		return fmt.Errorf("unexpected token '%v', expecting '}'", token)
	}

	token, err = dec.Token()
	if err != io.EOF {
		return fmt.Errorf("unexpected token '%v', expecting EOF", token)
	}

	// If there is a dataToken cached, we always defer at the end the processing
	// because nor datacontenttype or datacontentencoding are mandatory.
	if cachedData != nil {
		if err := drainData(e, checkFlag(state, DataBase64Flag), cachedData); err != nil {
			return err
		}
	}

	return nil
}

func drainData(e *Event, isBase64 bool, value json.RawMessage) error {
	if isBase64 {
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

func consumeData(e *Event, isBase64 bool, dec *json.Decoder) error {
	if isBase64 {
		e.DataBase64 = true
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

type tokenQueue struct {
	slice []entry
	i     int
}

func (q *tokenQueue) push(key string, raw json.RawMessage) {
	if q.i < len(q.slice) {
		q.slice[q.i].key = key
		q.slice[q.i].value = raw
	} else {
		e := entry{key: key, value: raw}
		q.slice = append(q.slice, e)
	}

	q.i++
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
