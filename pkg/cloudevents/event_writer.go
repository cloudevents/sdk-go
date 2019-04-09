package cloudevents

import (
	"fmt"
	"time"
)

var _ EventWriter = (*Event)(nil)

// SetSpecVersion implements EventWriter.SetSpecVersion
func (e *Event) SetSpecVersion(v string) error {
	if e.Context == nil {
		switch v {
		case CloudEventsVersionV01:
			e.Context = EventContextV01{}.AsV01()
		case CloudEventsVersionV02:
			e.Context = EventContextV02{}.AsV02()
		case CloudEventsVersionV03:
			e.Context = EventContextV03{}.AsV03()
		default:
			return fmt.Errorf("a valid spec version is required: [%s, %s, %s]",
				CloudEventsVersionV01, CloudEventsVersionV02, CloudEventsVersionV03)
		}
		return nil
	}
	return e.Context.SetSpecVersion(v)
}

// SetType implements EventWriter.SetType
func (e *Event) SetType(t string) error {
	return e.Context.SetType(t)
}

// SetSource implements EventWriter.SetSource
func (e *Event) SetSource(s string) error {
	return e.Context.SetSource(s)
}

// SetSubject implements EventWriter.SetSubject
func (e *Event) SetSubject(s string) error {
	return e.Context.SetSubject(s)
}

// SetID implements EventWriter.SetID
func (e *Event) SetID(id string) error {
	return e.Context.SetID(id)
}

// SetTime implements EventWriter.SetTime
func (e *Event) SetTime(t time.Time) error {
	return e.Context.SetTime(t)
}

// SetSchemaURL implements EventWriter.SetSchemaURL
func (e *Event) SetSchemaURL(s string) error {
	return e.Context.SetSchemaURL(s)
}

// SetDataContentType implements EventWriter.SetDataContentType
func (e *Event) SetDataContentType(ct string) error {
	return e.Context.SetDataContentType(ct)
}

// SetDataContentEncoding implements EventWriter.SetDataContentEncoding
func (e *Event) SetDataContentEncoding(enc string) error {
	return e.Context.SetDataContentEncoding(enc)
}

// SetDataContentEncoding implements EventWriter.SetDataContentEncoding
func (e *Event) SetExtension(name string, obj interface{}) error {
	return e.Context.SetExtension(name, obj)
}

// SetData implements EventWriter.SetData
func (e *Event) SetData(obj interface{}) error {
	e.Data = obj
	return nil
}
