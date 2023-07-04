package mqtt_paho

import (
	"context"
	"fmt"
	"sync"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/protocol"
	"github.com/eclipse/paho.golang/paho"

	cecontext "github.com/cloudevents/sdk-go/v2/context"
)

type Protocol struct {
	ctx            context.Context
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
}

// MQTT protocol implements Sender, Receiver
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
	return &Protocol{
		client:         client,
		connConfig:     connConfig,
		senderTopic:    SenderTopic,
		receiverTopics: ReceiverTopics,
		qos:            qos,
		retained:       retained,
		incoming:       make(chan *paho.Publish),
		openerMutex:    sync.Mutex{},
	}, nil
}

func (t *Protocol) Send(ctx context.Context, m binding.Message, transformers ...binding.Transformer) error {
	var err error
	defer m.Finish(err)

	topic := cecontext.TopicFrom(ctx)
	if topic == "" {
		topic = t.senderTopic
	}

	err = t.connect(ctx)
	if err != nil {
		return err
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
	return err
}

func (t *Protocol) OpenInbound(ctx context.Context) error {
	t.openerMutex.Lock()
	defer t.openerMutex.Unlock()

	logger := cecontext.LoggerFrom(ctx)
	if err := t.connect(ctx); err != nil {
		return err
	}

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

	logger.Infof("Subscribing to topics: %v", t.receiverTopics)
	_, err := t.client.Subscribe(ctx, &paho.Subscribe{
		Subscriptions: subs,
	})
	if err != nil {
		return err
	}

	<-ctx.Done()
	logger.Infof("context done: %v", ctx.Err())
	return ctx.Err()
}

func (t *Protocol) connect(ctx context.Context) error {
	ca, err := t.client.Connect(ctx, t.connConfig)
	if err != nil {
		return err
	}
	if ca.ReasonCode != 0 {
		return fmt.Errorf("failed to connect to %s : %d - %s", t.client.Conn.RemoteAddr(), ca.ReasonCode,
			ca.Properties.ReasonString)
	}
	return nil
}
