package event

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"

	jsoniter "github.com/json-iterator/go"

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

func drainTokenQueue(tokenQueue *tokenQueue, event *Event, state *uint8, cachedDataPointer *[]byte) error {
	switch ctx := event.Context.(type) {
	case *EventContextV03:
		for i := 0; i < tokenQueue.i; i++ {
			val := tokenQueue.slice[i].value
			var err error
			switch tokenQueue.slice[i].key {
			case "id":
				ctx.ID = val.ToString()
			case "type":
				ctx.Type = val.ToString()
			case "source":
				ctx.Source, err = toUriRef(val)
			case "subject":
				ctx.Subject, err = toStrPtr(val)
			case "time":
				ctx.Time, err = toTimestamp(val)
			case "schemaurl":
				ctx.SchemaURL, err = toUriRefPtr(val)
			case "datacontenttype":
				ctx.DataContentType, err = toStrPtr(val)
				appendFlag(state, DataContentTypeFlag)
			case "datacontentencoding":
				ctx.DataContentEncoding, err = toStrPtr(val)
				if *ctx.DataContentEncoding != Base64 {
					err = fmt.Errorf("invalid datacontentencoding value: '%s'", *ctx.DataContentEncoding)
				}
				appendFlag(state, DataBase64Flag)
			case "data":
				stream := jsoniter.NewStream(jsoniter.ConfigFastest, nil, 1024)
				val.WriteTo(stream)
				*cachedDataPointer = stream.Buffer()
				err = stream.Error
			default:
				value := val.GetInterface()
				if ctx.Extensions == nil {
					ctx.Extensions = make(map[string]interface{}, 1)
				}
				if err := ctx.SetExtension(tokenQueue.slice[i].key, value); err != nil {
					return err
				}
			}
			if err != nil {
				return err
			}
		}
	case *EventContextV1:
		for i := 0; i < tokenQueue.i; i++ {
			val := tokenQueue.slice[i].value
			var err error
			switch tokenQueue.slice[i].key {
			case "id":
				ctx.ID = val.ToString()
			case "type":
				ctx.Type = val.ToString()
			case "source":
				ctx.Source, err = toUriRef(val)
			case "subject":
				ctx.Subject, err = toStrPtr(val)
			case "time":
				ctx.Time, err = toTimestamp(val)
			case "dataschema":
				ctx.DataSchema, err = toUriPtr(val)
			case "datacontenttype":
				ctx.DataContentType, err = toStrPtr(val)
				appendFlag(state, DataContentTypeFlag)
			case "data":
				stream := jsoniter.NewStream(jsoniter.ConfigFastest, nil, 1024)
				val.WriteTo(stream)
				*cachedDataPointer = stream.Buffer()
				err = stream.Error
			case "data_base64":
				stream := jsoniter.NewStream(jsoniter.ConfigFastest, nil, 1024)
				val.WriteTo(stream)
				*cachedDataPointer = stream.Buffer()
				err = stream.Error
				appendFlag(state, DataBase64Flag)
			default:
				value := val.GetInterface()
				if ctx.Extensions == nil {
					ctx.Extensions = make(map[string]interface{}, 1)
				}
				if err := ctx.SetExtension(tokenQueue.slice[i].key, value); err != nil {
					return err
				}
			}
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func bestCaseProcessNext(iter *jsoniter.Iterator, state uint8, key string, e *Event) error {
	switch ctx := e.Context.(type) {
	case *EventContextV03:
		switch key {
		case "id":
			ctx.ID = iter.ReadString()
		case "type":
			ctx.Type = iter.ReadString()
		case "source":
			ctx.Source = readUriRef(iter)
		case "subject":
			ctx.Subject = readStrPtr(iter)
		case "time":
			ctx.Time = readTimestamp(iter)
		case "schemaurl":
			ctx.SchemaURL = readUriRefPtr(iter)
		case "data":
			return consumeData(e, checkFlag(state, DataBase64Flag), iter)
		default:
			if ctx.Extensions == nil {
				ctx.Extensions = make(map[string]interface{}, 1)
			}
			return ctx.SetExtension(key, iter.ReadAny())
		}
	case *EventContextV1:
		switch key {
		case "id":
			ctx.ID = iter.ReadString()
		case "type":
			ctx.Type = iter.ReadString()
		case "source":
			ctx.Source = readUriRef(iter)
		case "subject":
			ctx.Subject = readStrPtr(iter)
		case "time":
			ctx.Time = readTimestamp(iter)
		case "dataschema":
			ctx.DataSchema = readUriPtr(iter)
		case "data":
			return consumeData(e, false, iter)
		case "data_base64":
			return consumeData(e, true, iter)
		default:
			if ctx.Extensions == nil {
				ctx.Extensions = make(map[string]interface{}, 1)
			}
			return ctx.SetExtension(key, iter.ReadAny())
		}
	}
	return nil
}

type entry struct {
	key   string
	value jsoniter.Any
}

//TODO make this public?
func unmarshalJSON(reader io.Reader, e *Event) error {
	iterator := jsoniter.Parse(jsoniter.ConfigFastest, reader, 1024)

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
	var cachedData []byte

	for key := iterator.ReadObject(); key != ""; key = iterator.ReadObject() {
		// Check if we have some error in our error cache
		if iterator.Error != nil {
			return iterator.Error
		}

		// We have a key, now we need to figure out what to do
		// depending on the parsing state

		// If it's a specversion, trigger state change
		if key == "specversion" {
			if checkFlag(state, SpecVersionV1Flag|SpecVersionV03Flag) {
				return fmt.Errorf("specversion was already provided")
			}
			sv := iterator.ReadString()

			// Check proper specversion
			switch sv {
			case CloudEventsVersionV1:
				e.Context = &EventContextV1{}
				appendFlag(&state, SpecVersionV1Flag)
			case CloudEventsVersionV03:
				e.Context = &EventContextV03{}
				appendFlag(&state, SpecVersionV03Flag)
			default:
				return fmt.Errorf("unexpected specversion '%s'", sv)
			}

			if err := drainTokenQueue(tokenQueue, e, &state, &cachedData); err != nil {
				return err
			}
			continue
		}

		// If no specversion, enqueue unconditionally
		if !checkFlag(state, SpecVersionV03Flag|SpecVersionV1Flag) {
			tokenQueue.push(key, iterator.ReadAny())
			continue
		}

		// From this point downward -> we can assume the event has a context pointer non nil

		// If it's a datacontenttype, trigger state change
		if key == "datacontenttype" {
			if checkFlag(state, DataContentTypeFlag) {
				return fmt.Errorf("datacontenttype was already provided")
			}

			dct := iterator.ReadString()

			e.SetDataContentType(dct)
			appendFlag(&state, DataContentTypeFlag)
			continue
		}

		// If it's a datacontentencoding and it's v0.3, trigger state change
		if checkFlag(state, SpecVersionV03Flag) && key == "datacontentencoding" {
			if checkFlag(state, DataBase64Flag) {
				return fmt.Errorf("datacontentencoding was already provided")
			}

			dce := iterator.ReadString()

			if dce != Base64 {
				return fmt.Errorf("invalid datacontentencoding value: '%s'", dce)
			}

			e.Context.(*EventContextV03).DataContentEncoding = &dce
			appendFlag(&state, DataBase64Flag)
			continue
		}

		// We can parse all attributes, except data.
		// If it's data or data_base64 and we don't have the attributes to process it, then we enqueue
		if (!checkFlag(state, SpecVersionV1Flag&DataContentTypeFlag) && (key == "data" || key == "data_base64")) &&
			(!checkFlag(state, SpecVersionV03Flag&DataContentTypeFlag&DataBase64Flag) && key == "data") {
			if key == "data_base64" {
				appendFlag(&state, DataBase64Flag)
			}
			cachedData = iterator.SkipAndReturnBytes()
			continue
		}

		// At this point or this value is an attribute (excluding datacontenttype and datacontentencoding), or this value is data and this condition is valid:
		// (SpecVersionV1Flag & DataContentTypeFlag) || (SpecVersionV03Flag & DataContentTypeFlag & DataBase64Flag)
		if err := bestCaseProcessNext(iterator, state, key, e); err != nil {
			return err
		}
	}

	if iterator.Error != nil {
		return iterator.Error
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

func drainData(e *Event, isBase64 bool, value []byte) error {
	iter := jsoniter.ParseBytes(jsoniter.ConfigFastest, value)
	return consumeData(e, isBase64, iter)
}

func consumeData(e *Event, isBase64 bool, iter *jsoniter.Iterator) error {
	if isBase64 {
		e.DataBase64 = true

		// Allocate payload byte buffer
		base64Encoded := iter.ReadStringAsSlice()
		e.DataEncoded = make([]byte, base64.StdEncoding.DecodedLen(len(base64Encoded)))
		_, err := base64.StdEncoding.Decode(e.DataEncoded, base64Encoded)
		return err
	}

	ct := e.DataContentType()
	if ct != ApplicationJSON && ct != TextJSON {
		// If not json, then data is encoded as string
		src := iter.ReadStringAsSlice()
		e.DataEncoded = make([]byte, len(src))
		copy(e.DataEncoded, src)
		return nil
	}
	e.DataEncoded = iter.SkipAndReturnBytes()
	return nil
}

type tokenQueue struct {
	slice []entry
	i     int
}

func (q *tokenQueue) push(key string, raw jsoniter.Any) {
	if q.i < len(q.slice) {
		q.slice[q.i].key = key
		q.slice[q.i].value = raw
	} else {
		e := entry{key: key, value: raw}
		q.slice = append(q.slice, e)
	}

	q.i++
}

func readUriRef(iter *jsoniter.Iterator) types.URIRef {
	str := iter.ReadString()
	uriRef := types.ParseURIRef(str)
	if uriRef == nil {
		iter.Error = fmt.Errorf("cannot parse uri ref: %v", str)
	}
	return *uriRef
}

func readStrPtr(iter *jsoniter.Iterator) *string {
	str := iter.ReadString()
	if str == "" {
		return nil
	}
	return &str
}

func readUriRefPtr(iter *jsoniter.Iterator) *types.URIRef {
	return types.ParseURIRef(iter.ReadString())
}

func readUriPtr(iter *jsoniter.Iterator) *types.URI {
	return types.ParseURI(iter.ReadString())
}

func readTimestamp(iter *jsoniter.Iterator) *types.Timestamp {
	t, err := types.ParseTimestamp(iter.ReadString())
	if err != nil {
		iter.Error = err
	}
	return t
}

func toUriRef(val jsoniter.Any) (types.URIRef, error) {
	str := val.ToString()
	if val.LastError() != nil {
		return types.URIRef{}, val.LastError()
	}
	uriRef := types.ParseURIRef(str)
	if uriRef == nil {
		return types.URIRef{}, fmt.Errorf("cannot parse uri ref: %v", str)
	}
	return *uriRef, nil
}

func toStrPtr(val jsoniter.Any) (*string, error) {
	str := val.ToString()
	if val.LastError() != nil {
		return nil, val.LastError()
	}
	if str == "" {
		return nil, nil
	}
	return &str, nil
}

func toUriRefPtr(val jsoniter.Any) (*types.URIRef, error) {
	str := val.ToString()
	if val.LastError() != nil {
		return nil, val.LastError()
	}
	return types.ParseURIRef(str), nil
}

func toUriPtr(val jsoniter.Any) (*types.URI, error) {
	str := val.ToString()
	if val.LastError() != nil {
		return nil, val.LastError()
	}
	return types.ParseURI(str), nil
}

func toTimestamp(val jsoniter.Any) (*types.Timestamp, error) {
	str := val.ToString()
	if val.LastError() != nil {
		return nil, val.LastError()
	}
	t, err := types.ParseTimestamp(str)
	if err != nil {
		return nil, err
	}
	return t, nil
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
