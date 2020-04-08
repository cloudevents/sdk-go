package stan

import (
	"bytes"
	"context"
	"fmt"
	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/nats-io/stan.go"
)

type Sender struct {
	Conn         stan.Conn
	Subject      string
	Transformers binding.Transformers

	connOwned bool
}

// NewSender creates a new protocol.Sender responsible for opening and closing the STAN connection
func NewSender(clusterID, clientID, subject string, stanOpts []stan.Option, opts ...SenderOption) (*Sender, error) {
	conn, err := stan.Connect(clusterID, clientID, stanOpts...)
	if err != nil {
		return nil, err
	}

	s, err := NewSenderFromConn(conn, subject, opts...)
	if err != nil {
		if err2 := conn.Close(); err2 != nil {
			return nil, fmt.Errorf("failed to close conn: %s, when recovering from err: %w", err2, err)
		}
		return nil, err
	}

	s.connOwned = true

	return s, nil
}

// NewSenderFromConn creates a new protocol.Sender which leaves responsibility for opening and closing the STAN
// connection to the caller
func NewSenderFromConn(conn stan.Conn, subject string, opts ...SenderOption) (*Sender, error) {
	s := &Sender{
		Conn:         conn,
		Subject:      subject,
		Transformers: make(binding.Transformers, 0),
	}

	err := s.applyOptions(opts...)
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Sender) Send(ctx context.Context, in binding.Message) (err error) {
	defer func() {
		if err2 := in.Finish(err); err2 != nil {
			if err == nil {
				err = err2
			} else {
				err = fmt.Errorf("failed to call in.Finish() when error already occurred: %s: %w", err2.Error(), err)
			}
		}
	}()

	writer := new(bytes.Buffer)
	if err = WriteMsg(ctx, in, writer, s.Transformers...); err != nil {
		return err
	}
	return s.Conn.Publish(s.Subject, writer.Bytes())
}

// Close implements Closer.Close
// This method only closes the connection if the Sender opened it
func (s *Sender) Close(_ context.Context) error {
	if s.connOwned {
		return s.Conn.Close()
	}

	return nil
}

func (s *Sender) applyOptions(opts ...SenderOption) error {
	for _, fn := range opts {
		if err := fn(s); err != nil {
			return err
		}
	}
	return nil
}
