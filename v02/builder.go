package v02

import (
	"net/url"
	"time"

	cloudevents "github.com/cloudevents/sdk-go"
)

type CloudEventBuilder interface {
	SpecVersion(specVersion string) CloudEventBuilder
	Type(t string) CloudEventBuilder
	Source(source url.URL) CloudEventBuilder
	ID(id string) CloudEventBuilder
	Time(t time.Time) CloudEventBuilder
	SchemaURL(schemaURL url.URL) CloudEventBuilder
	ContentType(contentType string) CloudEventBuilder
	Data(data interface{}) CloudEventBuilder
	Extension(key string, value interface{}) CloudEventBuilder
	Build() (Event, error)
}

type cloudEventBuilder struct {
	specVersion string
	eventType   string
	source      url.URL
	id          string
	eventTime   *time.Time
	schemaURL   url.URL
	contentType string
	data        interface{}
	extensions  map[string]interface{}
}

func NewCloudEventBuilder() CloudEventBuilder {
	return &cloudEventBuilder{
		extensions: make(map[string]interface{}),
	}
}

func (b *cloudEventBuilder) SpecVersion(specVersion string) CloudEventBuilder {
	b.specVersion = specVersion
	return b
}

func (b *cloudEventBuilder) Type(t string) CloudEventBuilder {
	b.eventType = t
	return b
}

func (b *cloudEventBuilder) Source(source url.URL) CloudEventBuilder {
	b.source = source
	return b
}

func (b *cloudEventBuilder) ID(id string) CloudEventBuilder {
	b.id = id
	return b
}

func (b *cloudEventBuilder) Time(t time.Time) CloudEventBuilder {
	b.eventTime = &t
	return b
}

func (b *cloudEventBuilder) SchemaURL(schemaURL url.URL) CloudEventBuilder {
	b.schemaURL = schemaURL
	return b
}

func (b *cloudEventBuilder) ContentType(contentType string) CloudEventBuilder {
	b.contentType = contentType
	return b
}

func (b *cloudEventBuilder) Data(data interface{}) CloudEventBuilder {
	b.data = data
	return b
}

func (b *cloudEventBuilder) Extension(key string, value interface{}) CloudEventBuilder {
	b.extensions[key] = value
	return b
}

func (b *cloudEventBuilder) Build() (Event, error) {
	if b.specVersion == "" {
		b.specVersion = cloudevents.Version02
	}

	if (b.eventType == "" || b.source == url.URL{} || b.id == "") {
		return Event{}, cloudevents.IllegalArgumentError("type, source, and id are required fields")
	}

	event := Event{
		SpecVersion: b.specVersion,
		Type:        b.eventType,
		Source:      b.source,
		ID:          b.id,
		Time:        b.eventTime,
		SchemaURL:   b.schemaURL,
		ContentType: b.contentType,
		Data:        b.data,
	}

	return event, nil
}
