package mqttv3_paho

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"

	"github.com/cloudevents/sdk-go/v2/binding"
	cecontext "github.com/cloudevents/sdk-go/v2/context"
	"github.com/cloudevents/sdk-go/v2/protocol"
)

const defaultQuiesce = 250 // milliseconds

type Protocol struct {
	client  mqtt.Client
	quiesce uint

	incoming    chan mqtt.Message
	openerMutex sync.Mutex
	closeChan   chan struct{}

	subscriptions map[string]byte

	topic    string
	qos      byte
	retained bool
}

var (
	_ protocol.Opener   = (*Protocol)(nil)
	_ protocol.Sender   = (*Protocol)(nil)
	_ protocol.Receiver = (*Protocol)(nil)
	_ protocol.Closer   = (*Protocol)(nil)
)

func New(ctx context.Context, clientOptions *mqtt.ClientOptions, opts ...Option) (*Protocol, error) {
	p := &Protocol{
		client:    mqtt.NewClient(clientOptions),
		quiesce:   defaultQuiesce,
		incoming:  make(chan mqtt.Message),
		closeChan: make(chan struct{}),
	}

	for _, opt := range opts {
		if err := opt(p); err != nil {
			return nil, err
		}
	}

	token := p.client.Connect()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-token.Done():
		if token.Error() != nil {
			return nil, token.Error()
		}

		return p, nil
	case <-p.closeChan:
		return nil, errors.New("client closed")
	}
}

func (p *Protocol) Receive(ctx context.Context) (binding.Message, error) {
	select {
	case m, ok := <-p.incoming:
		if !ok {
			return nil, io.EOF
		}
		msg := NewMessage(m.Payload())
		return msg, nil
	case <-ctx.Done():
		return nil, io.EOF
	}
}

func (p *Protocol) Send(ctx context.Context, m binding.Message, transformers ...binding.Transformer) error {
	logger := cecontext.LoggerFrom(ctx)

	var err error
	defer func() {
		if fErr := m.Finish(err); fErr != nil {
			logger.Warnf("failed to finish message: %v", fErr)
		}
	}()

	topic := p.topic
	if cecontext.TopicFrom(ctx) != "" {
		topic = cecontext.TopicFrom(ctx)
		cecontext.WithTopic(ctx, "")
	}

	payload, err := WritePubMessage(ctx, m, transformers...)
	if err != nil {
		return err
	}

	token := p.client.Publish(topic, p.qos, p.retained, payload)
	if !token.WaitTimeout(time.Minute) {
		err = fmt.Errorf("publish to %q: timeout", topic)
		logger.Error(err)
		return err
	}

	return token.Error()
}

func (p *Protocol) OpenInbound(ctx context.Context) error {
	if len(p.subscriptions) == 0 {
		return errors.New("no subscriptions available")
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	p.openerMutex.Lock()
	defer p.openerMutex.Unlock()

	logger := cecontext.LoggerFrom(ctx)
	logger.Infof("subscribing to topics: %v", p.subscriptions)

	token := p.client.SubscribeMultiple(p.subscriptions, func(_ mqtt.Client, message mqtt.Message) {
		p.incoming <- message
	})
	<-token.Done()

	if err := token.Error(); err != nil {
		err = fmt.Errorf("subscribe to %v failed: %w", p.subscriptions, err)
		logger.Error(err)
		return err
	}

	select {
	case <-ctx.Done():
	case <-p.closeChan:
		cancel()
	}

	p.client.Disconnect(p.quiesce)
	return nil
}

func (p *Protocol) Close(context.Context) error {
	close(p.closeChan)
	return nil
}
