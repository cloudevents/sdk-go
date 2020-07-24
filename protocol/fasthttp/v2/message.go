package fasthttp

import (
	"bytes"
	"context"
	"github.com/valyala/fasthttp"
	"net/textproto"
	"strings"
	"unicode"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/binding/format"
	"github.com/cloudevents/sdk-go/v2/binding/spec"
)

const prefix = "Ce-"

var specs = spec.WithPrefixMatchExact(
	func(s string) string {
		if s == "datacontenttype" {
			return "Content-Type"
		} else {
			return textproto.CanonicalMIMEHeaderKey("Ce-" + s)
		}
	},
	"Ce-",
)

const ContentType = "Content-Type"
const ContentLength = "Content-Length"

// Message holds the Header and Body of a HTTP Request or Response.
// The Message instance *must* be constructed from NewMessage function.
// This message *cannot* be read several times. In order to read it more times, buffer it using binding/buffering methods
type Message struct {
	OnFinish func(error) error

	rtx *fasthttp.RequestCtx

	format  format.Format
	version spec.Version
}

// Check if http.Message implements binding.Message
var _ binding.Message = (*Message)(nil)
var _ binding.MessageMetadataReader = (*Message)(nil)

// NewMessage returns a binding.Message with header and data.
// The returned binding.Message *cannot* be read several times. In order to read it more times, buffer it using binding/buffering methods
func NewMessage(rtx *fasthttp.RequestCtx) *Message {
	if rtx == nil {
		return nil
	}

	m := Message{rtx: rtx}

	if m.format = format.Lookup(string(rtx.Request.Header.ContentType())); m.format == nil {
		version := rtx.Request.Header.Peek(specs.PrefixedSpecVersionName())
		m.version = specs.Version(string(version))
	}
	return &m
}

func (m *Message) ReadEncoding() binding.Encoding {
	if m.version != nil {
		return binding.EncodingBinary
	}
	if m.format != nil {
		return binding.EncodingStructured
	}
	return binding.EncodingUnknown
}

func (m *Message) ReadStructured(ctx context.Context, encoder binding.StructuredWriter) error {
	if m.format == nil {
		return binding.ErrNotStructured
	} else {
		reader := bytes.NewReader(m.rtx.Request.Body())
		return encoder.SetStructuredEvent(ctx, m.format, reader)
	}
}

func (m *Message) ReadBinary(ctx context.Context, encoder binding.BinaryWriter) (err error) {
	if m.version == nil {
		return binding.ErrNotBinary
	}

	if m.rtx.Request.Header.VisitAll(func(key, value []byte) {
		k := string(key)
		v := string(value)
		attr := m.version.Attribute(k)
		if attr != nil {
			err = encoder.SetAttribute(attr, v)
		} else if strings.HasPrefix(k, prefix) {
			// Trim Prefix + To lower
			var b strings.Builder
			b.Grow(len(k) - len(prefix))
			b.WriteRune(unicode.ToLower(rune(k[len(prefix)])))
			b.WriteString(k[len(prefix)+1:])
			err = encoder.SetExtension(b.String(), v)
		}
	}); err != nil {
		return err
	}

	reader := bytes.NewReader(m.rtx.Request.Body())
	if reader.Len() > 0 {
		err = encoder.SetData(reader)
		if err != nil {
			return err
		}
	}

	return
}

func (m *Message) GetAttribute(k spec.Kind) (spec.Attribute, interface{}) {
	attr := m.version.AttributeFromKind(k)
	if attr != nil {
		h := m.rtx.Request.Header.Peek(attributeHeadersMapping[attr.Name()])
		if h != nil {
			return attr, string(h)
		}
		return attr, nil
	}
	return nil, nil
}

func (m *Message) GetExtension(name string) interface{} {
	h := m.rtx.Request.Header.Peek(extNameToHeaderName(name))
	if h != nil {
		return string(h)
	}
	return nil
}

func (m *Message) Finish(err error) error {
	if m.OnFinish != nil {
		return m.OnFinish(err)
	}
	return nil
}
