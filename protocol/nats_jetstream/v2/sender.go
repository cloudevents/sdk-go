/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package nats_jetstream

import (
	"bytes"
	"context"
	"fmt"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/protocol"
)

type Sender struct {
	// v1 implementation
	Jsm nats.JetStreamContext

	// v2 implementation
	JetStream   jetstream.JetStream
	PublishOpts []jetstream.PublishOpt

	Conn      *nats.Conn
	Subject   string
	Stream    string
	connOwned bool
}

// NewSender creates a new protocol.Sender responsible for opening and closing the NATS connection
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

// NewSenderV2 creates a new protocol.Sender responsible for opening and closing the NATS connection
func NewSenderV2(ctx context.Context, url, subject string, natsOpts []nats.Option, jsOpts []jetstream.JetStreamOpt, opts ...SenderOption) (*Sender, error) {
	conn, err := nats.Connect(url, natsOpts...)
	if err != nil {
		return nil, err
	}

	s, err := NewSenderFromConnV2(ctx, conn, subject, jsOpts, opts...)
	if err != nil {
		conn.Close()
		return nil, err
	}

	s.connOwned = true

	return s, nil
}

// NewSenderFromConn creates a new protocol.Sender which leaves responsibility for opening and closing the NATS
// connection to the caller
func NewSenderFromConn(conn *nats.Conn, stream, subject string, jsmOpts []nats.JSOpt, opts ...SenderOption) (*Sender, error) {
	jsm, err := conn.JetStream(jsmOpts...)
	if err != nil {
		return nil, err
	}

	streamInfo, err := jsm.StreamInfo(stream, jsmOpts...)

	if streamInfo == nil || err != nil && err.Error() == "stream not found" {
		_, err = jsm.AddStream(&nats.StreamConfig{
			Name:     stream,
			Subjects: []string{stream + ".*"},
		})
		if err != nil {
			return nil, err
		}
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

// NewSenderFromConnV2 creates a new protocol.Sender which leaves responsibility for opening and closing the NATS
// connection to the caller
func NewSenderFromConnV2(ctx context.Context, conn *nats.Conn, subject string, jsOpts []jetstream.JetStreamOpt, opts ...SenderOption) (*Sender, error) {
	var js jetstream.JetStream
	var err error
	var stream string
	if js, err = jetstream.New(conn, jsOpts...); err != nil {
		return nil, err
	}

	if stream, err = js.StreamNameBySubject(ctx, subject); err != nil {
		return nil, err
	}

	s := &Sender{
		JetStream: js,
		Conn:      conn,
		Stream:    stream,
		Subject:   subject,
	}

	err = s.applyOptions(opts...)
	if err != nil {
		return nil, err
	}

	return s, nil
}

// Close implements Sender.Sender
// Sender sends messages.
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
	header, err := WriteMsg(ctx, in, writer, transformers...)
	if err != nil {
		return err
	}

	natsMsg := &nats.Msg{
		Subject: s.Subject,
		Data:    writer.Bytes(),
		Header:  header,
	}

	version := s.getVersion()
	switch version {
	case 0:
		return ErrNoJetstream
	case 1:
		_, err = s.Jsm.PublishMsg(natsMsg)
	case 2:
		_, err = s.JetStream.PublishMsg(ctx, natsMsg, s.PublishOpts...)
	}
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

// getVersion returns whether the consumer uses the older jsm or newer jetstream package
// 0 - None, 1 - jsm, 2 - jetstream
func (s *Sender) getVersion() int {
	if s.Jsm != nil {
		return 1
	}
	if s.JetStream != nil {
		return 2
	}
	return 0
}

var _ protocol.Sender = (*Sender)(nil)
var _ protocol.Closer = (*Protocol)(nil)
