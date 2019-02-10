package canonical

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"
)

type Event struct {
	Context EventContext
	Data    interface{}
}

type EventContext interface {
	// AsV01 provides a translation from whatever the "native" encoding of the
	// CloudEvent was to the equivalent in v0.1 field names, moving fields to or
	// from extensions as necessary.
	AsV01() EventContextV01

	// AsV02 provides a translation from whatever the "native" encoding of the
	// CloudEvent was to the equivalent in v0.2 field names, moving fields to or
	// from extensions as necessary.
	AsV02() EventContextV02

	// DataContentType returns the MIME content type for encoding data, which is
	// needed by both encoding and decoding.
	DataContentType() string
}

type Timestamp struct {
	time.Time
}

func ParseTimestamp(t string) *Timestamp {
	if t == "" {
		return nil
	}
	timestamp, err := time.Parse(time.RFC3339Nano, t)
	if err != nil {
		return nil
	}
	return &Timestamp{Time: timestamp}
}

// This allows json marshaling to always be in RFC3339Nano format.
func (t Timestamp) MarshalJSON() ([]byte, error) {
	if t.IsZero() {
		return []byte(`""`), nil
	}
	rfc3339 := fmt.Sprintf("%q", t.Format(time.RFC3339Nano))
	return []byte(rfc3339), nil
}

func (t *Timestamp) UnmarshalJSON(b []byte) error {
	var timestamp string
	if err := json.Unmarshal(b, &timestamp); err != nil {
		return err
	}
	*t = *ParseTimestamp(timestamp)
	return nil
}

func (t Timestamp) String() string {
	return t.Format(time.RFC3339Nano)
}

type URLRef struct {
	url.URL
}

func ParseURLRef(u string) *URLRef {
	if u == "" {
		return nil
	}
	pu, err := url.Parse(u)
	if err != nil {
		return nil
	}
	return &URLRef{URL: *pu}
}

// This allows json marshaling to always be in RFC3339Nano format.
func (u URLRef) MarshalJSON() ([]byte, error) {
	rfc3339 := fmt.Sprintf("%q", u.String())
	return []byte(rfc3339), nil
}

func (u *URLRef) UnmarshalJSON(b []byte) error {
	var ref string
	if err := json.Unmarshal(b, &ref); err != nil {
		return err
	}
	*u = *ParseURLRef(ref)
	return nil
}
