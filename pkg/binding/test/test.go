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

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/binding/format"
	ce "github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/types"
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
func SendReceive(t *testing.T, in binding.Message, s binding.Sender, r binding.Receiver, outAssert func(binding.Message)) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
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

func AssertEventEquals(t *testing.T, want cloudevents.Event, have cloudevents.Event) {
	assert.Equal(t, want.Context, have.Context)
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

func MustToEvent(m binding.Message) (e ce.Event, encoding binding.Encoding) {
	var err error
	e, encoding, err = binding.ToEvent(m)
	if err != nil {
		panic(err)
	}
	return
}
