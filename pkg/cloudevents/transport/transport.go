package transport

// Sender is the interface for transport sender to send the converted Message
// over the underlying transport.
type Sender interface {
	Send(Message) error
}

// Receiver TODO not sure yet.
type Receiver interface {
	Receive(Message)
}
