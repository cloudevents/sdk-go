package cloudevents

// Version01 holds a version string for CloudEvents specification version 0.1. See also EventV01 interface
// https://github.com/cloudevents/spec/blob/v0.1/spec.md
const Version01 = "0.1"

// Event interface is a generic abstraction over all possible versions and implementations of CloudEvents.
type Event interface {
	// CloudEventVersion returns the version of Event specification followed by the underlying implementation.
	CloudEventVersion() string
	// Get takes a property name and, if it exists, returns the value of that property. The ok return value can
	// be used to verify if the property exists.
	Get(property string) (value interface{}, ok bool)
	// Set sets the property value
	Set(property string, value interface{})
	// Properties returns a map of all event properties as keys and their mandatory status as values
	Properties() map[string]bool
}
