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

const specVersionV03Flag uint8 = 1 << 4
const specVersionV1Flag uint8 = 1 << 5
const dataBase64Flag uint8 = 1 << 6
const dataContentTypeFlag uint8 = 1 << 7

const preallocQueueSize = 16 // entry = 16 bytes, 16*16 = 256 bytes = 4 cache lines on x86_64

func checkFlag(state uint8, flag uint8) bool {
	return state&flag != 0
}

func appendFlag(state *uint8, flag uint8) {
	*state = (*state) | flag
}

type entry struct {
	key   string
	value jsoniter.Any
}

// ReadJson allows you to read the bytes reader as an event
func ReadJson(out *Event, reader io.Reader) error {
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
			if checkFlag(state, specVersionV1Flag|specVersionV03Flag) {
				return fmt.Errorf("specversion was already provided")
			}
			sv := iterator.ReadString()

			// Check proper specversion
			switch sv {
			case CloudEventsVersionV1:
				out.Context = &EventContextV1{}
				appendFlag(&state, specVersionV1Flag)
			case CloudEventsVersionV03:
				out.Context = &EventContextV03{}
				appendFlag(&state, specVersionV03Flag)
			default:
				return fmt.Errorf("unexpected specversion '%s'", sv)
			}

			// Now we have a specversion, so drain the token queue
			switch ctx := out.Context.(type) {
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
						appendFlag(&state, dataContentTypeFlag)
					case "datacontentencoding":
						ctx.DataContentEncoding, err = toStrPtr(val)
						if *ctx.DataContentEncoding != Base64 {
							err = fmt.Errorf("invalid datacontentencoding value: '%s'", *ctx.DataContentEncoding)
						}
						appendFlag(&state, dataBase64Flag)
					case "data":
						stream := jsoniter.NewStream(jsoniter.ConfigFastest, nil, 1024)
						val.WriteTo(stream)
						cachedData = stream.Buffer()
						err = stream.Error
					default:
						value := val.GetInterface()
						if ctx.Extensions == nil {
							ctx.Extensions = make(map[string]interface{}, 1)
						}
						err = ctx.SetExtension(tokenQueue.slice[i].key, value)
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
						appendFlag(&state, dataContentTypeFlag)
					case "data":
						stream := jsoniter.NewStream(jsoniter.ConfigFastest, nil, 1024)
						val.WriteTo(stream)
						cachedData = stream.Buffer()
						err = stream.Error
					case "data_base64":
						stream := jsoniter.NewStream(jsoniter.ConfigFastest, nil, 1024)
						val.WriteTo(stream)
						cachedData = stream.Buffer()
						err = stream.Error
						appendFlag(&state, dataBase64Flag)
					default:
						value := val.GetInterface()
						if ctx.Extensions == nil {
							ctx.Extensions = make(map[string]interface{}, 1)
						}
						err = ctx.SetExtension(tokenQueue.slice[i].key, value)
					}
					if err != nil {
						return err
					}
				}
			}
			continue
		}

		// If no specversion, enqueue unconditionally
		if !checkFlag(state, specVersionV03Flag|specVersionV1Flag) {
			tokenQueue.push(key, iterator.ReadAny())
			continue
		}

		// From this point downward -> we can assume the event has a context pointer non nil

		// If it's a datacontenttype, trigger state change
		if key == "datacontenttype" {
			if checkFlag(state, dataContentTypeFlag) {
				return fmt.Errorf("datacontenttype was already provided")
			}

			dct := iterator.ReadString()

			switch ctx := out.Context.(type) {
			case *EventContextV03:
				ctx.DataContentType = &dct
			case *EventContextV1:
				ctx.DataContentType = &dct
			}
			appendFlag(&state, dataContentTypeFlag)
			continue
		}

		// If it's a datacontentencoding and it's v0.3, trigger state change
		if checkFlag(state, specVersionV03Flag) && key == "datacontentencoding" {
			if checkFlag(state, dataBase64Flag) {
				return fmt.Errorf("datacontentencoding was already provided")
			}

			dce := iterator.ReadString()

			if dce != Base64 {
				return fmt.Errorf("invalid datacontentencoding value: '%s'", dce)
			}

			out.Context.(*EventContextV03).DataContentEncoding = &dce
			appendFlag(&state, dataBase64Flag)
			continue
		}

		// We can parse all attributes, except data.
		// If it's data or data_base64 and we don't have the attributes to process it, then we cache it
		// The expanded form of this condition is:
		// (checkFlag(state, specVersionV1Flag) && !checkFlag(state, dataContentTypeFlag) && (key == "data" || key == "data_base64")) ||
		// (checkFlag(state, specVersionV03Flag) && !(checkFlag(state, dataContentTypeFlag) && checkFlag(state, dataBase64Flag)) && key == "data")
		if (state&(specVersionV1Flag|dataContentTypeFlag) == specVersionV1Flag && (key == "data" || key == "data_base64")) ||
			((state&specVersionV03Flag == specVersionV03Flag) && (state&(dataContentTypeFlag|dataBase64Flag) != (dataContentTypeFlag | dataBase64Flag)) && key == "data") {
			if key == "data_base64" {
				appendFlag(&state, dataBase64Flag)
			}
			cachedData = iterator.SkipAndReturnBytes()
			continue
		}

		// At this point or this value is an attribute (excluding datacontenttype and datacontentencoding), or this value is data and this condition is valid:
		// (specVersionV1Flag & dataContentTypeFlag) || (specVersionV03Flag & dataContentTypeFlag & dataBase64Flag)
		switch ctx := out.Context.(type) {
		case *EventContextV03:
			switch key {
			case "id":
				ctx.ID = iterator.ReadString()
			case "type":
				ctx.Type = iterator.ReadString()
			case "source":
				ctx.Source = readUriRef(iterator)
			case "subject":
				ctx.Subject = readStrPtr(iterator)
			case "time":
				ctx.Time = readTimestamp(iterator)
			case "schemaurl":
				ctx.SchemaURL = readUriRefPtr(iterator)
			case "data":
				iterator.Error = consumeData(out, checkFlag(state, dataBase64Flag), iterator)
			default:
				if ctx.Extensions == nil {
					ctx.Extensions = make(map[string]interface{}, 1)
				}
				iterator.Error = ctx.SetExtension(key, iterator.ReadAny().GetInterface())
			}
		case *EventContextV1:
			switch key {
			case "id":
				ctx.ID = iterator.ReadString()
			case "type":
				ctx.Type = iterator.ReadString()
			case "source":
				ctx.Source = readUriRef(iterator)
			case "subject":
				ctx.Subject = readStrPtr(iterator)
			case "time":
				ctx.Time = readTimestamp(iterator)
			case "dataschema":
				ctx.DataSchema = readUriPtr(iterator)
			case "data":
				iterator.Error = consumeData(out, false, iterator)
			case "data_base64":
				iterator.Error = consumeData(out, true, iterator)
			default:
				if ctx.Extensions == nil {
					ctx.Extensions = make(map[string]interface{}, 1)
				}
				iterator.Error = ctx.SetExtension(key, iterator.ReadAny().GetInterface())
			}
		}
	}

	if iterator.Error != nil {
		return iterator.Error
	}

	// If there is a dataToken cached, we always defer at the end the processing
	// because nor datacontenttype or datacontentencoding are mandatory.
	if cachedData != nil {
		iter := jsoniter.ParseBytes(jsoniter.ConfigFastest, cachedData)
		return consumeData(out, checkFlag(state, dataBase64Flag), iter)
	}
	return nil
}

func consumeData(e *Event, isBase64 bool, iter *jsoniter.Iterator) error {
	if isBase64 {
		e.DataBase64 = true

		// Allocate payload byte buffer
		base64Encoded := iter.ReadStringAsSlice()
		e.DataEncoded = make([]byte, base64.StdEncoding.DecodedLen(len(base64Encoded)))
		len, err := base64.StdEncoding.Decode(e.DataEncoded, base64Encoded)
		if err != nil {
			return err
		}
		e.DataEncoded = e.DataEncoded[0:len]
		return nil
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
		return types.URIRef{}
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
	_, r := observability.NewReporter(context.Background(), eventJSONObserved{o: reportUnmarshal})
	err := ReadJson(e, bytes.NewReader(b))

	// Report the observable
	if err != nil {
		r.Error()
	} else {
		r.OK()
	}
	return err

}
