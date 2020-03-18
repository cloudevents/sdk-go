package binding_test

import (
	"context"
	"fmt"
	"io"
	"strconv"

	"github.com/cloudevents/sdk-go/v1/cloudevents"
	"github.com/cloudevents/sdk-go/v1/cloudevents/client"
)

const count = 3 // Example ends after this many events.

// The sender uses the cloudevents.Client API, not the transport APIs directly.
func runSender(w io.Writer) error {
	c, err := client.New(NewExTransport(nil, w), client.WithoutTracePropagation())
	if err != nil {
		return err
	}
	for i := 0; i < count; i++ {
		e := cloudevents.New()
		e.SetType("example.com/event")
		e.SetSource("example.com/source")
		e.SetID(strconv.Itoa(i))
		if err := e.SetData(fmt.Sprintf("hello %d", i)); err != nil {
			return err
		}
		if _, _, err := c.Send(context.TODO(), e); err != nil {
			return err
		}
	}
	return nil
}

// The receiver uses the cloudevents.Client API, not the transport APIs directly.
func runReceiver(r io.Reader) error {
	i := 0
	process := func(e cloudevents.Event) error {
		fmt.Printf("%s\n", e)
		i++
		if i == count {
			return io.EOF
		}
		return nil
	}
	c, err := client.New(NewExTransport(r, nil), client.WithoutTracePropagation())
	if err != nil {
		return err
	}
	return c.StartReceiver(context.TODO(), process)
}

// The intermediary receives events and forwards them to another
// process using ExReceiver and ExSender directly.
//
// By forwarding a transport.Message instead of a cloudevents.Event,
// it allows the transports to avoid un-necessary decoding of
// structured events, and to exchange delivery status between reliable
// transports. Even transports using different protocols can ensure
// reliable delivery.
//
func runIntermediary(r io.Reader, w io.WriteCloser) error {
	defer w.Close()
	for {
		receiver := NewExReceiver(r)
		sender := NewExSender(w)
		for i := 0; i < count; i++ {
			if m, err := receiver.Receive(context.TODO()); err != nil {
				return err
			} else if err := sender.Send(context.TODO(), m); err != nil {
				return err
			}
		}
	}
}

// This example shows how to use a transport in sender, receiver,
// and intermediary processes.
//
// The sender and receiver use the client.Client API to send and
// receive messages.  the transport.  Only the intermediary example
// actually uses the transport APIs for efficiency and reliability in
// forwarding events.
func Example_using() {
	r1, w1 := io.Pipe() // The sender-to-intermediary pipe
	r2, w2 := io.Pipe() // The intermediary-to-receiver pipe

	done := make(chan error)
	go func() { done <- runReceiver(r2) }()
	go func() { done <- runIntermediary(r1, w2) }()
	go func() { done <- runSender(w1) }()
	for i := 0; i < 2; i++ {
		if err := <-done; err != nil && err != io.EOF {
			fmt.Println(err)
		}
	}

	// Output:
	// Validation: valid
	// Context Attributes,
	//   specversion: 1.0
	//   type: example.com/event
	//   source: example.com/source
	//   id: 0
	// Data,
	//   "hello 0"
	//
	// Validation: valid
	// Context Attributes,
	//   specversion: 1.0
	//   type: example.com/event
	//   source: example.com/source
	//   id: 1
	// Data,
	//   "hello 1"
	//
	// Validation: valid
	// Context Attributes,
	//   specversion: 1.0
	//   type: example.com/event
	//   source: example.com/source
	//   id: 2
	// Data,
	//   "hello 2"
}
