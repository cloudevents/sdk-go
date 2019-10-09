package binding_test

import (
	"context"
	"testing"

	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/binding/format"
	"github.com/cloudevents/sdk-go/pkg/binding/spec"
	"github.com/stretchr/testify/assert"
)

type chanSR chan binding.Message

func (ch chanSR) Send(_ context.Context, m binding.Message) error    { ch <- m; return nil }
func (ch chanSR) Receive(_ context.Context) (binding.Message, error) { return <-ch, nil }

func TestVersionSender(t *testing.T) {
	ch := make(chanSR, 1)
	s := binding.VersionSender(ch, spec.V01)
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
	ch := make(chanSR, 1)
	s := binding.StructSender(ch, format.JSON)

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
	ch := make(chanSR, 1)
	s := binding.BinarySender(ch)

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

// FIXME(alanconway) verify callback to original message Finish.
