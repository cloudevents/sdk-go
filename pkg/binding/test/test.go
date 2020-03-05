// Package test provides re-usable functions for binding tests.
package test

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/binding/format"
	"github.com/cloudevents/sdk-go/pkg/binding/spec"
	"github.com/cloudevents/sdk-go/pkg/event"
	"github.com/cloudevents/sdk-go/pkg/types"
)

// NameOf generates a string test name from x, esp. for ce.Event and ce.Message.
func NameOf(x interface{}) string {
	switch x := x.(type) {
	case event.Event:
		b, err := json.Marshal(x)
		if err == nil {
			return fmt.Sprintf("Event%s", b)
		}
	case binding.Message:
		return fmt.Sprintf("Message{%s}", reflect.TypeOf(x).String())
	}
	return fmt.Sprintf("%T(%#v)", x, x)
}

// Run f as a test for each event in events
func EachEvent(t *testing.T, events []event.Event, f func(*testing.T, event.Event)) {
	for _, e := range events {
		in := e
		t.Run(NameOf(in), func(t *testing.T) { f(t, in) })
	}
}

// Run f as a test for each message in messages
func EachMessage(t *testing.T, messages []binding.Message, f func(*testing.T, binding.Message)) {
	for _, m := range messages {
		in := m
		t.Run(NameOf(in), func(t *testing.T) { f(t, in) })
	}
}

// SendReceive does s.Send(in), then it receives the message in r.Receive() and executes outAssert
// Halt test on error.
func SendReceive(t *testing.T, ctx context.Context, in binding.Message, s binding.Sender, r binding.Receiver, outAssert func(binding.Message)) {
	t.Helper()
	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		out, recvErr := r.Receive(ctx)
		require.NoError(t, recvErr)
		outAssert(out)
		finishErr := out.Finish(nil)
		require.NoError(t, finishErr)
	}()

	go func() {
		defer wg.Done()
		err := s.Send(ctx, in)
		require.NoError(t, err)
	}()

	wg.Wait()
}

// Assert two event.Event context are equals
func AssertEventContextEquals(t *testing.T, want event.EventContext, have event.EventContext) {
	wantVersion, err := spec.VS.Version(want.GetSpecVersion())
	require.NoError(t, err)
	haveVersion, err := spec.VS.Version(have.GetSpecVersion())
	require.NoError(t, err)
	require.Equal(t, wantVersion, haveVersion)

	for _, a := range wantVersion.Attributes() {
		require.Equal(t, a.Get(want), a.Get(have), "Attribute %s does not match: %v != %v", a.PrefixedName(), a.Get(want), a.Get(have))
	}

	require.Equal(t, want.GetExtensions(), have.GetExtensions(), "Extensions")
}

// Assert two event.Event are equals
func AssertEventEquals(t *testing.T, want event.Event, have event.Event) {
	AssertEventContextEquals(t, want.Context, have.Context)
	wantPayload, err := want.DataBytes()
	assert.NoError(t, err)
	havePayload, err := have.DataBytes()
	assert.NoError(t, err)
	assert.Equal(t, wantPayload, havePayload)
}

// Returns a copy of the event.Event where all extensions are converted to strings. Fails the test if conversion fails
func ExToStr(t *testing.T, e event.Event) event.Event {
	out := CopyEventContext(e)
	for k, v := range e.Extensions() {
		var vParsed interface{}
		var err error

		switch v.(type) {
		case json.RawMessage:
			err = json.Unmarshal(v.(json.RawMessage), &vParsed)
			require.NoError(t, err)
		default:
			vParsed, err = types.Format(v)
			require.NoError(t, err)
		}
		out.SetExtension(k, vParsed)
	}
	return out
}

// Must marshal the event.Event to JSON structured representation
func MustJSON(e event.Event) []byte {
	b, err := format.JSON.Marshal(e)
	if err != nil {
		panic(err)
	}
	return b
}

// Must convert the Message to event.Event
func MustToEvent(ctx context.Context, m binding.Message) (e event.Event, encoding binding.Encoding) {
	var err error
	e, encoding, err = binding.ToEvent(ctx, m, nil)
	if err != nil {
		panic(err)
	}
	return
}

// Returns a copy of the event.Event with only the event.EventContext copied
func CopyEventContext(e event.Event) event.Event {
	newE := event.Event{}
	newE.Context = e.Context.Clone()
	newE.DataEncoded = e.DataEncoded
	newE.Data = e.Data
	newE.DataBinary = e.DataBinary
	return newE
}
