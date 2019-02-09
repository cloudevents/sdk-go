package transport

type Message interface {
	// CloudEventVersion returns the version of the CloudEvent.
	CloudEventVersion() string

	// TODO maybe get encoding
}
