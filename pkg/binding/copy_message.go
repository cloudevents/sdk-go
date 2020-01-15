package binding

// CopyMessage reads m once and creates an in-memory copy using EventMessage or
// StructMessage depending on the encoding of m.  The returned copy is not
// dependent on any transport and can be read many times.
//
// NOTE: Calling Finish() on the returned message calls m.Finish()
func CopyMessage(m Message) (Message, error) {
	// Try structured first, it's cheaper.
	sm := StructMessage{}
	err := m.Structured(&sm)
	switch err {
	case nil:
		return &sm, nil
	case ErrNotStructured:
		break
	default:
		return nil, err
	}
	em := EventMessage{}
	err = m.Event(&em)
	if err != nil {
		return nil, err
	}
	return WithFinish(em, func(err error) { _ = m.Finish(err) }), nil
}
