package event

// EventResponse represents the canonical representation of a Result to a
// CloudEvent from a receiver. Result implementation is Transport dependent.
// Deprecated: handle the protocol directly.
type EventResponse struct {
	// Deprecated: handle the protocol directly.
	Status int
	// Deprecated: handle the protocol directly.
	Event *Event
	// Deprecated: handle the protocol directly.
	Reason string
	// Context is transport specific struct to allow for controlling transport
	// response details.
	// For example, see http.TransportResponseContext.
	// Deprecated: handle the protocol directly.
	Context interface{}
}

// RespondWith sets up the instance of EventResponse to be set with status and
// an event. Result implementation is Transport dependent.
func (e *EventResponse) RespondWith(status int, event *Event) {
	if e == nil {
		// if nil, response not supported
		return
	}
	e.Status = status
	if event != nil {
		e.Event = event
	}
}

// Error sets the instance of EventResponse to be set with an error code and
// reason string. Result implementation is Transport dependent.
func (e *EventResponse) Error(status int, reason string) {
	if e == nil {
		// if nil, response not supported
		return
	}
	e.Status = status
	e.Reason = reason
}
