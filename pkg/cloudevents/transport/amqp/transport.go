package amqp

import (
	"context"
	"fmt"
	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	cecontext "github.com/cloudevents/sdk-go/pkg/cloudevents/context"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport"
	"go.uber.org/zap"
	"pack.ag/amqp"
)

// Transport adheres to transport.Transport.
var _ transport.Transport = (*Transport)(nil)

// Transport acts as both a http client and a http handler.
type Transport struct {
	connOpts         []amqp.ConnOption
	sessionOpts      []amqp.SessionOption
	senderLinkOpts   []amqp.LinkOption
	receiverLinkOpts []amqp.LinkOption

	// Encoding
	Encoding Encoding
	codec    transport.Codec

	// AMQP
	Client  *amqp.Client
	Session *amqp.Session
	Sender  *amqp.Sender

	Queue string

	// Receiver
	Receiver transport.Receiver
}

// New creates a new amqp transport.
func New(server, queue string, opts ...Option) (*Transport, error) {
	t := &Transport{
		Queue:            queue,
		connOpts:         []amqp.ConnOption(nil),
		sessionOpts:      []amqp.SessionOption(nil),
		senderLinkOpts:   []amqp.LinkOption(nil),
		receiverLinkOpts: []amqp.LinkOption(nil),
	}
	if err := t.applyOptions(opts...); err != nil {
		return nil, err
	}

	client, err := amqp.Dial(server, t.connOpts...)
	if err != nil {
		return nil, err
	}
	t.Client = client

	// Open a session
	session, err := client.NewSession(t.sessionOpts...)
	if err != nil {
		_ = client.Close()
		return nil, err
	}
	t.Session = session

	t.senderLinkOpts = append(t.senderLinkOpts, amqp.LinkTargetAddress(queue))

	// Create a sender
	sender, err := session.NewSender(t.senderLinkOpts...)
	if err != nil {
		_ = client.Close()
		_ = session.Close(context.Background())
		return nil, err
	}
	t.Sender = sender // TODO: in the future we might have more than one sender.

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
		// TODO: no response?
		return nil, t.Sender.Send(ctx, &amqp.Message{
			Properties: &amqp.MessageProperties{
				ContentType: m.ContentType,
			},
			ApplicationProperties: m.ApplicationProperties,
			Data:                  [][]byte{m.Body},
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
	logger := cecontext.LoggerFrom(ctx)

	logger.Info("StartReceiver on ", t.Queue)

	t.receiverLinkOpts = append(t.receiverLinkOpts, amqp.LinkSourceAddress(t.Queue))
	receiver, err := t.Session.NewReceiver(t.receiverLinkOpts...)
	if err != nil {
		return err
	}

	if ok := t.loadCodec(); !ok {
		return fmt.Errorf("unknown encoding set on transport: %d", t.Encoding)
	}

	ctx, cancel := context.WithCancel(ctx)
	go func() {
		for {
			// Receive next message
			msg, err := receiver.Receive(ctx)
			if err != nil {
				logger.Errorw("Failed reading message from AMQP.", zap.Error(err))
				panic(err) // TODO: panic?
			}

			m := &Message{
				Body:                  msg.Data[0], // TODO: omg why is it a list of lists?
				ContentType:           msg.Properties.ContentType,
				ApplicationProperties: msg.ApplicationProperties,
			}
			event, err := t.codec.Decode(m)
			if err != nil {
				logger.Errorw("failed to decode message", zap.Error(err)) // TODO: create an error channel to pass this up
			} else {
				// TODO: I do not know enough about amqp to implement reply.
				// For now, amqp does not support reply.
				if err := t.Receiver.Receive(context.TODO(), *event, nil); err != nil {
					logger.Warnw("amqp receiver return err", zap.Error(err))
				} else {
					if err := msg.Accept(); err != nil {
						logger.Warnw("amqp accept return err", zap.Error(err))
					}
				}
			}
		}
	}()

	<-ctx.Done()

	_ = receiver.Close(ctx)
	cancel()
	return nil
}
