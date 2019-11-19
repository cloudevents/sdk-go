// Package test provides re-usable functions for binding tests.
package test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/binding/format"
	ce "github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/types"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		if f, b := x.Structured(); f != "" {
			return fmt.Sprintf("Message{MediaType:%s, Bytes:%s}", f, b)
		} else if e, err := x.Event(); err == nil {
			return fmt.Sprintf("Message{%s}", NameOf(e))
		}
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

func EncodeAll(t *testing.T, in []ce.Event, enc binding.Encoder) (out []binding.Message) {
	t.Helper()
	for _, e := range in {
		m, err := enc.Encode(e)
		require.NoError(t, err)
		out = append(out, m)
	}
	return out
}

// EncodeDecode enc.Encode(); m.Decode(); return result. Halt test on error.
func EncodeDecode(t *testing.T, in ce.Event, enc binding.Encoder) (out ce.Event) {
	t.Helper()
	m, err := enc.Encode(in)
	require.NoError(t, err)
	out, err = m.Event()
	require.NoError(t, err)
	return out
}

// SendReceive does, s.Send(in) and returns r.Receive().
// Halt test on error.
func SendReceive(t *testing.T, in binding.Message, s binding.Sender, r binding.Receiver) binding.Message {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var out binding.Message
	var recvErr error
	done := make(chan struct{})
	go func() {
		defer close(done)
		if out, recvErr = r.Receive(ctx); recvErr == nil {
			recvErr = out.Finish(nil)
		}
	}()
	require.NoError(t, s.Send(ctx, in), "Send error")
	<-done
	require.NoError(t, recvErr, "Receive error")
	return out
}

// DiffMessageStruct compares x.Structured() and y.Structured()
func DiffMessageStruct(x, y binding.Message) string {
	var sx, sy binding.StructMessage
	sx.Format, sx.Bytes = x.Structured()
	sy.Format, sy.Bytes = x.Structured()
	return cmp.Diff(sx, sy)
}

// AssertMessageEventEqual compares x.Event() and y.Event()
func AssertMessageEventEqual(t testing.TB, x, y binding.Message) {
	t.Helper()
	ex, errx := x.Event()
	ey, erry := y.Event()
	assert.True(t, errx == nil && erry == nil, "Error comparing events: %q, %q", errx, erry)
	assert.Equal(t, ex, ey)
}

// AssertMessageEqual asserts that x and y are both structured or both binary and equal.
func AssertMessageEqual(t testing.TB, x, y binding.Message) {
	t.Helper()
	var sx, sy binding.StructMessage
	sx.Format, sx.Bytes = x.Structured()
	sy.Format, sy.Bytes = x.Structured()
	if sx.Format != "" || sy.Format != "" {
		assert.Equal(t, sx, sy)
	} else {
		AssertMessageEventEqual(t, x, y)
	}
}

func MustJSON(e ce.Event) []byte {
	b, err := format.JSON.Marshal(e)
	if err != nil {
		panic(err)
	}
	return b
}
