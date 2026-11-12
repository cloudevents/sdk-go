/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package http

import (
	"context"
	"fmt"
	"github.com/cloudevents/sdk-go/v2/binding"
	"net/http"
	"net/textproto"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/cloudevents/sdk-go/v2/binding/spec"
)

var attributeHeadersMapping map[string]string

type customHeaderKey int

const (
	headerKey customHeaderKey = iota
)

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

func extNameToHeaderName(name string) string {
	var b strings.Builder
	b.Grow(len(name) + len(prefix))
	b.WriteString(prefix)
	b.WriteRune(unicode.ToUpper(rune(name[0])))
	b.WriteString(name[1:])
	return b.String()
}

func HeaderFrom(ctx context.Context) http.Header {
	return binding.GetOrDefaultFromCtx(ctx, headerKey, make(http.Header)).(http.Header)
}

func WithCustomHeader(ctx context.Context, header http.Header) context.Context {
	return context.WithValue(ctx, headerKey, header)
}

const upperHex = "0123456789ABCDEF"

// encodeHeaderValue percent-encodes a CloudEvents header value. Bytes outside
// U+0021-U+007E are encoded, along with '%' and '"'. Spaces are only encoded
// when leading or trailing, since interior spaces are preserved by HTTP.
func encodeHeaderValue(v string) string {
	first, last := 0, len(v)
	for first < last && v[first] == ' ' {
		first++
	}
	for last > first && v[last-1] == ' ' {
		last--
	}
	var b strings.Builder
	for i := 0; i < len(v); i++ {
		c := v[i]
		if c == ' ' && i >= first && i < last {
			b.WriteByte(c)
		} else if c > 0x20 && c < 0x7F && c != '%' && c != '"' {
			b.WriteByte(c)
		} else {
			b.WriteByte('%')
			b.WriteByte(upperHex[c>>4])
			b.WriteByte(upperHex[c&0xF])
		}
	}
	return b.String()
}

func unhex(c byte) (byte, bool) {
	switch {
	case c >= '0' && c <= '9':
		return c - '0', true
	case c >= 'a' && c <= 'f':
		return c - 'a' + 10, true
	case c >= 'A' && c <= 'F':
		return c - 'A' + 10, true
	}
	return 0, false
}

// decodeHeaderValue decodes a CloudEvents header value per HTTP rules: it trims
// leading/trailing whitespace, removes non-encoded double-quotes (which group
// text), honors backslash escapes within quotes, and percent-decodes. A '%'
// not followed by two hex digits is kept literally. Decoded output that is
// not valid UTF-8 is rejected.
func decodeHeaderValue(v string) (string, error) {
	v = strings.Trim(v, " \t")
	if !strings.ContainsAny(v, "%\"") {
		return v, nil
	}
	b := make([]byte, 0, len(v))
	inQuote := false
	for i := 0; i < len(v); i++ {
		c := v[i]
		switch {
		case c == '"':
			inQuote = !inQuote
		case c == '\\' && inQuote && i+1 < len(v):
			i++
			b = append(b, v[i])
		case c == '%' && i+2 < len(v):
			hi, ok1 := unhex(v[i+1])
			lo, ok2 := unhex(v[i+2])
			if ok1 && ok2 {
				b = append(b, hi<<4|lo)
				i += 2
			} else {
				b = append(b, c)
			}
		default:
			b = append(b, c)
		}
	}
	if !utf8.Valid(b) {
		return "", fmt.Errorf("invalid UTF-8 in header value %q", v)
	}
	return string(b), nil
}
