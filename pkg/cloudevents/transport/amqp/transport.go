package amqp

import (
	"context"
	"fmt"

	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	cecontext "github.com/cloudevents/sdk-go/pkg/cloudevents/context"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport"
	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

// Transport adheres to transport.Transport.
var _ transport.Transport = (*Transport)(nil)

// Transport acts as both a http client and a http handler.
type Transport struct {
	Encoding Encoding
	Conn     *amqp.Connection
	Exchange string
	Ch       *amqp.Channel
	Queue    amqp.Queue

	Receiver transport.Receiver

	codec transport.Codec
}

// New creates a new amqp transport.
func New(server, exchange, key string, opts ...Option) (*Transport, error) {
	conn, err := amqp.Dial(server)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	queue, err := ch.QueueDeclare(
		key,   // name
		false, // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return nil, err
	}

	t := &Transport{
		Conn:  conn,
		Ch:    ch,
		Queue: queue,
	}
	if err := t.applyOptions(opts...); err != nil {
		return nil, err
	}

	return t, nil
}

func (t *Transport) applyOptions(opts ...Option) error {
	for _, fn := range opts {
		if err := fn(t); err != nil {
			return err
		}
	}
	return nil
}

func (t *Transport) loadCodec() bool {
	if t.codec == nil {
		switch t.Encoding {
		case Default, BinaryV02, StructuredV02, BinaryV03, StructuredV03:
			t.codec = &Codec{Encoding: t.Encoding}
		//case BinaryV02, StructuredV02:
		//	t.codec = &CodecV02{Encoding: t.Encoding}
		//case BinaryV03, StructuredV03:
		//	t.codec = &CodecV03{Encoding: t.Encoding}
		default:
			return false
		}
	}
	return true
}

// Send implements Transport.Send
func (t *Transport) Send(ctx context.Context, event cloudevents.Event) (*cloudevents.Event, error) {
	if ok := t.loadCodec(); !ok {
		return nil, fmt.Errorf("unknown encoding set on transport: %d", t.Encoding)
	}

	msg, err := t.codec.Encode(event)
	if err != nil {
		return nil, err
	}

	if m, ok := msg.(*Message); ok {
		return nil, t.Ch.Publish(
			t.Exchange,
			t.Queue.Name,
			false, // mandatory
			false, // immediate
			amqp.Publishing{
				Headers:     m.Headers,
				ContentType: m.ContentType,
				Body:        m.Body,
			})
	}

	return nil, fmt.Errorf("failed to encode Event into a Message")
}

// SetReceiver implements Transport.SetReceiver
func (t *Transport) SetReceiver(r transport.Receiver) {
	t.Receiver = r
}

// StartReceiver implements Transport.StartReceiver
// NOTE: This is a blocking call.
func (t *Transport) StartReceiver(ctx context.Context) error {
	if t.Conn == nil {
		return fmt.Errorf("no active amqp connection")
	}

	msgs, err := t.Ch.Consume(
		t.Queue.Name, // queue
		"",           // consumer
		true,         // auto-ack
		false,        // exclusive
		false,        // no-local
		false,        // no-wait
		nil,          // args
	)
	if err != nil {
		return fmt.Errorf("failed to start consume, %e", err)
	}

	if ok := t.loadCodec(); !ok {
		return fmt.Errorf("unknown encoding set on transport: %d", t.Encoding)
	}

	logger := cecontext.LoggerFrom(ctx)

	go func() {
		for d := range msgs {
			msg := &Message{
				Body:        d.Body,
				ContentType: d.ContentType,
				Headers:     d.Headers,
			}
			event, err := t.codec.Decode(msg)
			if err != nil {
				logger.Errorw("failed to decode message", zap.Error(err)) // TODO: create an error channel to pass this up
			}
			// TODO: I do not know enough about amqp to implement reply.
			// For now, amqp does not support reply.
			if err := t.Receiver.Receive(context.TODO(), *event, nil); err != nil {
				logger.Warnw("amqp receiver return err", zap.Error(err))
			}
		}
	}()

	<-ctx.Done()
	return err
}
