package format

import (
	"fmt"
	"io"
	"strings"

	"github.com/cloudevents/sdk-go/v2/event"
)

// Format marshals and unmarshals structured events to bytes.
type Format interface {
	// MediaType identifies the format
	MediaType() string
	// Marshal event to the provided writer
	Marshal(io.Writer, *event.Event) error
	// Unmarshal bytes from the provided reader
	Unmarshal(io.Reader, *event.Event) error
}

// Prefix for event-format media types.
const Prefix = "application/cloudevents"

// IsFormat returns true if mediaType begins with "application/cloudevents"
func IsFormat(mediaType string) bool { return strings.HasPrefix(mediaType, Prefix) }

type jsonFmt struct{}

func (jsonFmt) MediaType() string { return event.ApplicationCloudEventsJSON }

func (jsonFmt) Marshal(w io.Writer, e *event.Event) error {
	return event.WriteJson(w, e)
}

func (jsonFmt) Unmarshal(r io.Reader, e *event.Event) error {
	return event.ReadJson(e, r)
}

// JSON is the built-in "application/cloudevents+json" format.
var JSON Format = jsonFmt{}

// built-in formats
var formats map[string]Format

func init() {
	formats = map[string]Format{}
	Add(JSON)
}

// Lookup returns the format for contentType, or nil if not found.
func Lookup(contentType string) Format {
	i := strings.IndexRune(contentType, ';')
	if i == -1 {
		i = len(contentType)
	}
	contentType = strings.TrimSpace(strings.ToLower(contentType[0:i]))
	return formats[contentType]
}

func unknown(mediaType string) error {
	return fmt.Errorf("unknown event format media-type %#v", mediaType)
}

// Add a new Format. It can be retrieved by Lookup(f.MediaType())
func Add(f Format) { formats[f.MediaType()] = f }

// Marshal an event to bytes using the mediaType event format.
func Marshal(mediaType string, writer io.Writer, e *event.Event) error {
	if f := formats[mediaType]; f != nil {
		return f.Marshal(writer, e)
	}
	return unknown(mediaType)
}

// Unmarshal bytes to an event using the mediaType event format.
func Unmarshal(mediaType string, reader io.Reader, e *event.Event) error {
	if f := formats[mediaType]; f != nil {
		return f.Unmarshal(reader, e)
	}
	return unknown(mediaType)
}
