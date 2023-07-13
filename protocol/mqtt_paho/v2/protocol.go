/*
 Copyright 2023 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package mqtt_paho

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/protocol"
	"github.com/eclipse/paho.golang/paho"

	cecontext "github.com/cloudevents/sdk-go/v2/context"
)

type Protocol struct {
	client         *paho.Client
	connConfig     *paho.Connect
	senderTopic    string
	receiverTopics []string
	qos            byte
	retained       bool

	// receiver
	incoming chan *paho.Publish
	// inOpen
	openerMutex sync.Mutex

	closeChan chan struct{}
}

var (
	_ protocol.Sender   = (*Protocol)(nil)
	_ protocol.Opener   = (*Protocol)(nil)
	_ protocol.Receiver = (*Protocol)(nil)
	_ protocol.Closer   = (*Protocol)(nil)
)

func New(ctx context.Context, clientConfig *paho.ClientConfig, connConfig *paho.Connect, SenderTopic string,
	ReceiverTopics []string, qos byte, retained bool,
) (*Protocol, error) {
	client := paho.NewClient(*clientConfig)
	ca, err := client.Connect(ctx, connConfig)
	if err != nil {
		return nil, err
	}
	if ca.ReasonCode != 0 {
		return nil, fmt.Errorf("failed to connect to %s : %d - %s", client.Conn.RemoteAddr(), ca.ReasonCode,
			ca.Properties.ReasonString)
	}

	return &Protocol{
		client:         client,
		connConfig:     connConfig,
		senderTopic:    SenderTopic,
		receiverTopics: ReceiverTopics,
		qos:            qos,
		retained:       retained,
		incoming:       make(chan *paho.Publish),
		openerMutex:    sync.Mutex{},
		closeChan:      make(chan struct{}),
	}, nil
}

func (t *Protocol) Send(ctx context.Context, m binding.Message, transformers ...binding.Transformer) error {
	var err error
	defer m.Finish(err)

	topic := cecontext.TopicFrom(ctx)
	if topic == "" {
		topic = t.senderTopic
	}

	msg := &paho.Publish{
		QoS:    t.qos,
		Retain: t.retained,
		Topic:  topic,
	}
	err = WritePubMessage(ctx, m, msg, transformers...)
	if err != nil {
		return err
	}

	_, err = t.client.Publish(ctx, msg)
	if err != nil {
		return err
	}
	return err
}

func (t *Protocol) OpenInbound(ctx context.Context) error {
	t.openerMutex.Lock()
	defer t.openerMutex.Unlock()

	logger := cecontext.LoggerFrom(ctx)

	t.client.Router = paho.NewSingleHandlerRouter(func(m *paho.Publish) {
		t.incoming <- m
	})

	subs := make(map[string]paho.SubscribeOptions)
	for _, topic := range t.receiverTopics {
		subs[topic] = paho.SubscribeOptions{
			QoS:               t.qos,
			RetainAsPublished: t.retained,
		}
	}

	logger.Infof("subscribe to topics: %v", t.receiverTopics)
	_, err := t.client.Subscribe(ctx, &paho.Subscribe{
		Subscriptions: subs,
	})
	if err != nil {
		return err
	}

	// Wait until external or internal context done
	select {
	case <-ctx.Done():
	case <-t.closeChan:
	}
	return t.client.Disconnect(&paho.Disconnect{ReasonCode: 0})
}

// Receive implements Receiver.Receive
func (t *Protocol) Receive(ctx context.Context) (binding.Message, error) {
	select {
	case m, ok := <-t.incoming:
		if !ok {
			return nil, io.EOF
		}
		msg := NewMessage(m)
		return msg, nil
	case <-ctx.Done():
		return nil, io.EOF
	}
}

func (p *Protocol) Close(ctx context.Context) error {
	close(p.closeChan)
	return nil
}
