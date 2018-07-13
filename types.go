package cloudevents

// CloudEvent interface is a generic abstraction over all possible versions and implementations of CloudEvents.
type CloudEvent interface {
	// CloudEventVersion returns the version of CloudEvent specification followed by the underlying implementation.
	CloudEventVersion() string
	// Context returns the event context, which a set of properties associated with the event.
	Context() map[string]interface{}
}

// VersionMismatchError is returned when expected CloudEvent version does not match the actual one, e.g.
// when using GetEventV01 or when using transport bindings.
type VersionMismatchError string

func (e VersionMismatchError) Error() string {
	return "provided event is not CloudEvent or does not implement expected version: " + string(e)
}

// VersionNotSupportedError is returned when provided version is not supported by this library. This is returned
// by different parsing components, and by validation.
type VersionNotSupportedError string

func (e VersionNotSupportedError) Error() string {
	return "provided version: " + string(e) + " is not supported"
}
