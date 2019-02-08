package canonical

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
