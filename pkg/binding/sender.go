package binding

import (
	"context"

	"github.com/cloudevents/sdk-go/pkg/binding/format"
	"github.com/cloudevents/sdk-go/pkg/binding/spec"
	ce "github.com/cloudevents/sdk-go/pkg/cloudevents"
)

type messageWrapper struct{ original, modified Message }

func wrapMessage(original, modified Message) Message {
	if w, ok := original.(*messageWrapper); ok {
		w.modified = modified
		return w
	}
	return &messageWrapper{original: original, modified: modified}
}
func (w *messageWrapper) Finish(err error) error       { return w.original.Finish(err) }
func (w *messageWrapper) Event() (ce.Event, error)     { return w.modified.Event() }
func (w *messageWrapper) Structured() (string, []byte) { return w.modified.Structured() }
func (w *messageWrapper) Received(settle func(error)) {
	if once, ok := w.original.(ExactlyOnceMessage); ok {
		once.Received(settle)
	} else {
		settle(nil)
	}
}

type senderWrapper struct {
	Sender      Sender
	Version     spec.Version
	Format      format.Format
	ForceBinary bool
}

func (s *senderWrapper) transform(m Message) (Message, error) {
	// Version first, we have to parse to check the version.
	if s.Version != nil {
		e, err := m.Event()
		if err != nil {
			return nil, err
		}
		if e.SpecVersion() != s.Version.String() {
			e.Context = s.Version.Convert(e.Context)
			m = wrapMessage(m, EventMessage(e))
		}
	}
	if s.ForceBinary {
		if f, _ := m.Structured(); f != "" { // Not already binary
			e, err := m.Event()
			return wrapMessage(m, EventMessage(e)), err
		}
	} else if s.Format != nil {
		if f, _ := m.Structured(); f != s.Format.MediaType() { // Not already formatted
			e, err := m.Event()
			if err != nil {
				return nil, err
			}
			sm, err := StructEncoder{Format: s.Format}.Encode(e)
			return wrapMessage(m, sm), err
		}
	}
	return m, nil
}

func (s *senderWrapper) Send(ctx context.Context, m Message) error {
	m, err := s.transform(m)
	if err != nil {
		return err
	}
	return s.Sender.Send(ctx, m)
}

func (s *senderWrapper) Close(ctx context.Context) error {
	if c, ok := s.Sender.(Closer); ok {
		return c.Close(ctx)
	}
	return nil
}

func wrapSender(s Sender) *senderWrapper {
	if sw, ok := s.(*senderWrapper); ok {
		return sw
	}
	return &senderWrapper{Sender: s}
}

// VersionSender returns a Sender that transforms messages to spec-version v,
// then calls s.Send().
//
// By default VersionSender creates binary-mode messages. If combined with
// StructSender it will create structured messages of version v.
func VersionSender(s Sender, v spec.Version) Sender {
	sw := wrapSender(s)
	sw.Version = v
	return sw
}

// BinarySender returns a Sender that transforms messages to structured mode
// with format f, then calls s.Send().
func StructSender(s Sender, f format.Format) Sender {
	sw := wrapSender(s)
	sw.ForceBinary = false
	sw.Format = f
	return sw
}

// BinarySender returns a Sender that transforms messages to binary mode,
// then calls s.Send().
func BinarySender(s Sender) Sender {
	sw := wrapSender(s)
	sw.ForceBinary = true
	return sw
}
