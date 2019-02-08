package transport

import (
	"net/url"
	"time"
)

type Message interface {
	// CloudEventVersion returns the version of the CloudEvent.
	CloudEventVersion() string

	// ContextAttributes returns a list of context attributes that exist for the
	// message.
	//ContextAttributes() []string TODO not sure on this yet.

	// ExtensionAttributes returns a list of context attributes that exist for the
	// message.
	//ExtensionAttributes() []string TODO not sure on this yet.

	// Get takes an attribute name and, if it exists, returns the value of that
	// attribute. The ok return value can be used to verify if the attribute
	// exists.
	Get(key string) (value interface{}, ok bool)

	// Set sets the attribute value
	Set(key string, value interface{}) error

	// GetInt is a convenience method that wraps Get to provide a type checked
	// return value. Ok will be false if the attribute does not exist or the
	// value cannot be converted to an int32.
	GetInt(key string) (value int32, ok bool)

	// SetInt is a convenience method that wraps Set to provide a type checking
	// when setting an int.
	SetInt(key string, value int32) error

	// GetString is a convenience method that wraps Get to provide a type
	// checked return value. Ok will be false if the attribute does not exist or
	// the value cannot be converted to a string.
	GetString(key string) (value string, ok bool)

	// SetString is a convenience method that wraps Set to provide a type
	// checking when setting a string.
	SetString(key string, value string) error

	// GetBinary is a convenience method that wraps Get to provide a type
	// checked return value. Ok will be false if the attribute does not exist or
	// the value cannot be converted to a binary array.
	GetBinary(key string) (value []byte, ok bool)

	// SetBinary is a convenience method that wraps Set to provide a type
	// checking when setting a binary value.
	SetBinary(key string, value string) error

	// GetMap is a convenience method that wraps Get to provide a type checked
	// return value. Ok will be false if the attribute does not exist or the
	// value cannot be converted to a map.
	GetMap(key string) (value map[string]interface{}, ok bool)

	// SetMap is a convenience method that wraps Set to provide a type
	// checking when setting a map value.
	SetMap(key string, value map[string]interface{}) error

	// GetTime is a convenience method that wraps Get to provide a type checked
	// return value. Ok will be false if the attribute does not exist or the
	// value cannot be converted or parsed into a time.Time.
	GetTime(key string) (value time.Time, ok bool)

	// SetTime is a convenience method that wraps Set to provide a type
	// checking when setting a time value.
	SetTime(key string, value time.Time) error

	// GetURL is a convenience method that wraps Get to provide a type checked
	// return value. Ok will be false if the attribute does not exist or the
	// value cannot be converted or parsed into a url.URL.
	GetURL(key string) (value url.URL, ok bool)

	// SetURL is a convenience method that wraps Set to provide a type
	// checking when setting a URL value.
	SetURL(key string, value url.URL) error
}
