package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cloudevents/sdk-go/v2/binding/spec"
	"github.com/cloudevents/sdk-go/v2/event"
)

type EventMatcher func(t testing.TB, have event.Event)

// ContainsContextAttributes checks if the event contains at least the provided context attributes
func ContainsContextAttributes(attrs ...string) EventMatcher {
	return func(t testing.TB, have event.Event) {
		haveVersion := spec.VS.Version(have.SpecVersion())
		for _, k := range attrs {
			attr := haveVersion.Attribute(k)
			require.NotNil(t, attr, "Attribute name '%s' unrecognized", k)
			require.NotNil(t, attr.Get(have.Context))
		}
	}
}

// HasContextAttributes checks if the event contains at least the provided context attributes and their values
func HasContextAttributes(m map[string]interface{}) EventMatcher {
	return func(t testing.TB, have event.Event) {
		haveVersion := spec.VS.Version(have.SpecVersion())
		for k, v := range m {
			attr := haveVersion.Attribute(k)
			require.NotNil(t, attr, "Attribute name '%s' unrecognized", k)
			require.Equal(t, v, attr.Get(have.Context))
		}
	}
}

// ContainsExtensions checks if the event contains at least the provided extension names
func ContainsExtensions(exts ...string) EventMatcher {
	return func(t testing.TB, have event.Event) {
		for _, ext := range exts {
			require.NotNil(t, have.Extensions()[ext], "Expecting extension %s not to be nil", ext)
		}
	}
}

// HasExactlyExtensions checks if the event contains exactly the provided extensions
func HasExactlyExtensions(ext map[string]interface{}) EventMatcher {
	return func(t testing.TB, have event.Event) {
		require.Equal(t, ext, have.Extensions())
	}
}

// HasExtensions checks if the event contains at least the provided extensions
func HasExtensions(ext map[string]interface{}) EventMatcher {
	return func(t testing.TB, have event.Event) {
		for k, v := range ext {
			require.Equal(t, v, have.Extensions()[k])
		}
	}
}

// HasExtension checks if the event contains the provided extension
func HasExtension(key string, value interface{}) EventMatcher {
	return HasExtensions(map[string]interface{}{key: value})
}

// HasData checks if the event contains the provided data
func HasData(want []byte) EventMatcher {
	return func(t testing.TB, have event.Event) {
		require.Equal(t, want, have.Data())
	}
}

// HasData checks if the event doesn't contain data
func HasNoData() EventMatcher {
	return func(t testing.TB, have event.Event) {
		require.Nil(t, have.Data())
	}
}

// IsEqualTo performs a semantic equality check of the event (like AssertEventEquals)
func IsEqualTo(want event.Event) EventMatcher {
	return func(t testing.TB, have event.Event) {
		AssertEventEquals(t, want, have)
	}
}

// IsContextEqualTo performs a semantic equality check of the event context (like AssertEventContextEquals)
func IsContextEqualTo(want event.Event) EventMatcher {
	return func(t testing.TB, have event.Event) {
		AssertEventContextEquals(t, want.Context, have.Context)
	}
}

// IsValid checks if the event is valid
func IsValid() EventMatcher {
	return func(t testing.TB, have event.Event) {
		require.NoError(t, have.Validate())
	}
}

// IsValid checks if the event is invalid
func IsInvalid() EventMatcher {
	return func(t testing.TB, have event.Event) {
		require.Error(t, have.Validate())
	}
}

// AssertEvent is a "matcher like" assertion method to test the properties of an event
func AssertEvent(t testing.TB, have event.Event, assertions ...EventMatcher) {
	for _, a := range assertions {
		a(t, have)
	}
}

// AssertEventContextEquals asserts that two event.Event contexts are equals
func AssertEventContextEquals(t testing.TB, want event.EventContext, have event.EventContext) {
	wantVersion := spec.VS.Version(want.GetSpecVersion())
	require.NotNil(t, wantVersion)
	haveVersion := spec.VS.Version(have.GetSpecVersion())
	require.NotNil(t, haveVersion)
	require.Equal(t, wantVersion, haveVersion)

	for _, a := range wantVersion.Attributes() {
		require.Equal(t, a.Get(want), a.Get(have), "Attribute %s does not match: %v != %v", a.PrefixedName(), a.Get(want), a.Get(have))
	}

	require.Equal(t, want.GetExtensions(), have.GetExtensions(), "Extensions")
}

// AssertEventEquals asserts that two event.Event are equals
func AssertEventEquals(t testing.TB, want event.Event, have event.Event) {
	AssertEventContextEquals(t, want.Context, have.Context)
	wantPayload := want.Data()
	havePayload := have.Data()
	assert.Equal(t, wantPayload, havePayload)
}
