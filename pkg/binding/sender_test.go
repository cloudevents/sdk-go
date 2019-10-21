package binding_test

import (
	"context"
	"testing"

	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/binding/format"
	"github.com/cloudevents/sdk-go/pkg/binding/spec"
	"github.com/stretchr/testify/assert"
)

func TestVersionSender(t *testing.T) {
	ch := make(chan binding.Message, 1)
	s := binding.VersionSender(binding.ChanSender(ch), spec.V01)
	want := testEvent
	want.Context = want.Context.AsV01()
	assert.Equal(t, "1.0", testEvent.SpecVersion())

	_ = s.Send(context.Background(), binding.EventMessage(testEvent))
	got, err := (<-ch).Event()
	assert.NoError(t, err)
	assert.Equal(t, want, got)

	sm := binding.StructMessage{Format: format.JSON.MediaType(), Bytes: []byte(testJSON)}
	_ = s.Send(context.Background(), sm)
	got, err = (<-ch).Event()
	assert.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestStructSender(t *testing.T) {
	ch := make(chan binding.Message, 1)
	s := binding.StructSender(binding.ChanSender(ch), format.JSON)

	_ = s.Send(context.Background(), binding.EventMessage(testEvent))
	f, b := (<-ch).Structured()
	assert.Equal(t, f, format.JSON.MediaType())
	assert.Equal(t, b, []byte(testJSON))

	sm := &binding.StructMessage{Format: format.JSON.MediaType(), Bytes: []byte(testJSON)}
	_ = s.Send(context.Background(), sm)
	m := <-ch
	assert.Equal(t, sm, m.(*binding.StructMessage)) // Already structured, same message.
}

func TestBinarySender(t *testing.T) {
	ch := make(chan binding.Message, 1)
	s := binding.BinarySender(binding.ChanSender(ch))

	sm := &binding.StructMessage{Format: format.JSON.MediaType(), Bytes: []byte(testJSON)}
	_ = s.Send(context.Background(), sm)
	m := <-ch
	f, b := m.Structured()
	assert.Equal(t, f, "")
	assert.Equal(t, b, []byte(nil))
	got, err := m.Event()
	assert.NoError(t, err)
	assert.Equal(t, testEvent, got)

	em := binding.EventMessage(testEvent)
	_ = s.Send(context.Background(), em)
	m = <-ch
	assert.Equal(t, em, m.(binding.EventMessage)) // Already structured, same message.
}

func TestWithFinish(t *testing.T) {
	done := make(chan error, 1)
	m := binding.WithFinish(binding.EventMessage(testEvent), func(err error) {
		done <- err
	})
	select {
	case <-done:
		assert.Fail(t, "done early")
	default:
	}
	ch := make(chan binding.Message, 1)
	assert.NoError(t, binding.ChanSender(ch).Send(context.Background(), m))
	assert.NoError(t, <-done)
}
