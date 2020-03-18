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

	cloudevents "github.com/cloudevents/sdk-go/v1"
	"github.com/cloudevents/sdk-go/v1/binding"
	"github.com/cloudevents/sdk-go/v1/binding/format"
	"github.com/cloudevents/sdk-go/v1/binding/spec"
	ce "github.com/cloudevents/sdk-go/v1/cloudevents"
	"github.com/cloudevents/sdk-go/v1/cloudevents/types"
)

// NameOf generates a string test name from x, esp. for ce.Event and ce.Message.
func NameOf(x interface{}) string {
	switch x := x.(type) {
	case ce.Event:
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
func EachEvent(t *testing.T, events []ce.Event, f func(*testing.T, ce.Event)) {
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

// Canonical converts all attributes to canonical string form for comparisons.
func Canonical(t *testing.T, c ce.EventContext) {
	t.Helper()
	for k, v := range c.GetExtensions() {
		s, err := types.Format(v)
		require.NoError(t, err, "extension[%q]=%#v: %v", k, v, err)
		assert.NoError(t, c.SetExtension(k, s))
	}
}

// SendReceive does, s.Send(in) and returns r.Receive().
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

func AssertEventContextEquals(t *testing.T, want cloudevents.EventContext, have cloudevents.EventContext) {
	wantVersion, err := spec.VS.Version(want.GetSpecVersion())
	require.NoError(t, err)
	haveVersion, err := spec.VS.Version(have.GetSpecVersion())
	require.NoError(t, err)
	require.Equal(t, wantVersion, haveVersion)

	for _, a := range wantVersion.Attributes() {
		require.Equal(t, a.Get(want), a.Get(have), "Attribute %s does not match: %v != %v", a.Name(), a.Get(want), a.Get(have))
	}

	require.Equal(t, want.GetExtensions(), have.GetExtensions(), "Extensions")
}

func AssertEventEquals(t *testing.T, want cloudevents.Event, have cloudevents.Event) {
	AssertEventContextEquals(t, want.Context, have.Context)
	wantPayload, err := want.DataBytes()
	assert.NoError(t, err)
	havePayload, err := have.DataBytes()
	assert.NoError(t, err)
	assert.Equal(t, wantPayload, havePayload)
}

func ExToStr(t *testing.T, e ce.Event) ce.Event {
	for k, v := range e.Extensions() {
		var vParsed interface{}
		var err error

		switch v.(type) {
		case json.RawMessage:
			err = json.Unmarshal(v.(json.RawMessage), &vParsed)
			assert.NoError(t, err)
		default:
			vParsed, err = types.Format(v)
			require.NoError(t, err)
		}
		e.SetExtension(k, vParsed)
	}
	return e
}

func MustJSON(e ce.Event) []byte {
	b, err := format.JSON.Marshal(e)
	if err != nil {
		panic(err)
	}
	return b
}

func MustToEvent(ctx context.Context, m binding.Message) (e ce.Event, encoding binding.Encoding) {
	var err error
	e, encoding, err = binding.ToEvent(ctx, m)
	if err != nil {
		panic(err)
	}
	return
}

func CopyEventContext(e ce.Event) ce.Event {
	newE := ce.Event{}
	newE.Context = e.Context.Clone()
	newE.DataEncoded = e.DataEncoded
	newE.Data = e.Data
	newE.DataBinary = e.DataBinary
	return newE
}
