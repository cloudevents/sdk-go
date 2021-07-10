/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package jsm

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/protocol"

	"github.com/nats-io/jsm.go"
	"github.com/nats-io/nats.go"
)

type Sender struct {
	Jsm       nats.JetStreamContext
	Conn      *nats.Conn
	Subject   string
	Stream    string
	connOwned bool
}

// NewSender creates a new protocol.Sender responsible for opening and closing the STAN connection
func NewSender(url, stream, subject string, natsOpts []nats.Option, jsmOpts []nats.JSOpt, opts ...SenderOption) (*Sender, error) {
	conn, err := nats.Connect(url, natsOpts...)
	if err != nil {
		return nil, err
	}

	s, err := NewSenderFromConn(conn, stream, subject, jsmOpts, opts...)
	if err != nil {
		conn.Close()
		return nil, err
	}

	s.connOwned = true

	return s, nil
}

// NewSenderFromConn creates a new protocol.Sender which leaves responsibility for opening and closing the STAN
// connection to the caller
func NewSenderFromConn(conn *nats.Conn, stream, subject string, jsmOpts []nats.JSOpt, opts ...SenderOption) (*Sender, error) {
	mgr, err := jsm.New(conn)
	if err != nil {
		return nil, err
	}

	template, err := jsm.NewStreamConfiguration(jsm.DefaultStream, jsm.MaxAge(24*365*time.Hour), jsm.FileStorage())
	if err != nil {
		return nil, err
	}

	_, err = mgr.NewStreamFromDefault(stream, *template, jsm.Subjects(stream+".*"))
	if err != nil {
		return nil, err
	}

	jsm, err := conn.JetStream(jsmOpts...)
	if err != nil {
		return nil, err
	}

	s := &Sender{
		Jsm:     jsm,
		Conn:    conn,
		Stream:  stream,
		Subject: subject,
	}

	err = s.applyOptions(opts...)
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Sender) Send(ctx context.Context, in binding.Message, transformers ...binding.Transformer) (err error) {
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
	if err = WriteMsg(ctx, in, writer, transformers...); err != nil {
		return err
	}
	_, err = s.Jsm.Publish(s.Subject, writer.Bytes())

	return err
}

// Close implements Closer.Close
// This method only closes the connection if the Sender opened it
func (s *Sender) Close(_ context.Context) error {
	if s.connOwned {
		s.Conn.Close()
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

var _ protocol.Sender = (*Sender)(nil)
var _ protocol.Closer = (*Protocol)(nil)
