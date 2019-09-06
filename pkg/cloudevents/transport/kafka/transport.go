package kafka

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/confluentinc/confluent-kafka-go-dev/kafka"

	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	cecontext "github.com/cloudevents/sdk-go/pkg/cloudevents/context"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport"
)

// Transport adheres to transport.Transport.
var _ transport.Transport = (*Transport)(nil)

const (
	TransportName = "Kafka"
)

// Transport acts as both a kafka topic and a kafka subscription .
type Transport struct {
	// Encoding
	Encoding Encoding

	// DefaultEncodingSelectionFn allows for other encoding selection strategies to be injected.
	DefaultEncodingSelectionFn EncodingSelector

	codec transport.Codec
	// Codec Mutex
	coMu sync.Mutex

	// Kafka

	// AllowCreateTopic controls if the transport can create a topic if it does
	// not exist.
	AllowCreateTopic bool

	config      *kafka.ConfigMap
	adminClient *kafka.AdminClient
	consumer    *kafka.Consumer
	producer    *kafka.Producer

	topic     string
	topicOnce sync.Once

	// Receiver
	Receiver transport.Receiver

	// Converter is invoked if the incoming transport receives an undecodable
	// message.
	Converter transport.Converter
}

// New creates a new kafka transport.
func New(ctx context.Context, opts ...Option) (*Transport, error) {
	t := &Transport{}
	if err := t.applyOptions(opts...); err != nil {
		return nil, err
	}

	if t.adminClient == nil {
		adminClient, err := kafka.NewAdminClient(t.config)
		if err != nil {
			return nil, err
		}
		t.adminClient = adminClient
	}

	if t.producer == nil {
		producer, err := kafka.NewProducer(t.config)
		if err != nil {
			return nil, err
		}
		t.producer = producer
	}

	if t.consumer == nil {
		consumer, err := kafka.NewConsumer(t.config)
		if err != nil {
			return nil, err
		}
		t.consumer = consumer
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

func (t *Transport) loadCodec(ctx context.Context) bool {
	if t.codec == nil {
		t.coMu.Lock()
		if t.DefaultEncodingSelectionFn != nil && t.Encoding != Default {
			logger := cecontext.LoggerFrom(ctx)
			logger.Warn("transport has a DefaultEncodingSelectionFn set but Encoding is not Default. DefaultEncodingSelectionFn will be ignored.")

			t.codec = &Codec{
				Encoding: t.Encoding,
			}
		} else {
			t.codec = &Codec{
				Encoding:                   t.Encoding,
				DefaultEncodingSelectionFn: t.DefaultEncodingSelectionFn,
			}
		}
		t.coMu.Unlock()
	}
	return true
}

func (t *Transport) createTopicIfNotExists(ctx context.Context) error {
	t.topicOnce.Do(func() {
		//var ok bool
		//// Load the topic.
		//topic := t.client.Topic(t.topicID)
		//ok, err = topic.Exists(ctx)
		//if err != nil {
		//	_ = t.client.Close()
		//	return
		//}
		//// If the topic does not exist, create a new topic with the given name.
		//if !ok {
		//	if !t.AllowCreateTopic {
		//		err = fmt.Errorf("transport not allowed to create topic %q", t.topicID)
		//		return
		//	}
		//	topic, err = t.client.CreateTopic(ctx, t.topicID)
		//	if err != nil {
		//		return
		//	}
		//	t.x = true
		//}
	})

	return nil
}

// Send implements Transport.Send
func (t *Transport) Send(ctx context.Context, event cloudevents.Event) (*cloudevents.Event, error) {
	if ok := t.loadCodec(ctx); !ok {
		return nil, fmt.Errorf("unknown encoding set on transport: %d", t.Encoding)
	}

	msg, err := t.codec.Encode(ctx, event)
	if err != nil {
		return nil, err
	}

	err = t.createTopicIfNotExists(ctx)
	if err != nil {
		return nil, err
	}

	if m, ok := msg.(*Message); ok {

		err := t.producer.Produce(&kafka.Message{
			Headers:        t.toKafkaHeaders(m.Headers),
			Value:          m.Data,
			TopicPartition: kafka.TopicPartition{Topic: &t.topic, Partition: kafka.PartitionAny},
		}, nil)

		if err != nil {
			return nil, err
		}
		return &event, nil
	}

	return nil, fmt.Errorf("failed to encode Event into a Message")
}

// SetReceiver implements Transport.SetReceiver
func (t *Transport) SetReceiver(r transport.Receiver) {
	t.Receiver = r
}

// SetConverter implements Transport.SetConverter
func (t *Transport) SetConverter(c transport.Converter) {
	t.Converter = c
}

// HasConverter implements Transport.HasConverter
func (t *Transport) HasConverter() bool {
	return t.Converter != nil
}

// StartReceiver implements Transport.StartReceiver
// NOTE: This is a blocking call.
func (t *Transport) StartReceiver(ctx context.Context) error {
	logger := cecontext.LoggerFrom(ctx)

	// Load the codec.
	if ok := t.loadCodec(ctx); !ok {
		return fmt.Errorf("unknown encoding set on transport: %d", t.Encoding)
	}

	err := t.consumer.SubscribeTopics([]string{t.topic}, nil)
	if err != nil {
		return err
	}

	run := true

	for run == true {
		select {
		case sig := <-ctx.Done():
			fmt.Printf("Caught signal %v: terminating\n", sig)
			err := t.consumer.Close()
			if err != nil {
				return err
			}
			run = false
		default:
			m, err := t.consumer.ReadMessage(30 * time.Second)
			if err != nil {
				// The client will automatically try to recover from all errors.
				logger.Warnw("Consumer error", zap.Error(err))
				//return err
			}
			fmt.Printf("consumed from topic %s [%d] at offset %v: "+
				string(m.Value), *m.TopicPartition.Topic,
				m.TopicPartition.Partition, m.TopicPartition.Offset)

			msg := &Message{
				Headers: t.fromKafkaHeaders(m.Headers),
				Data:    m.Value,
			}

			event, err := t.codec.Decode(ctx, msg)
			// If codec returns and error, try with the converter if it is set.
			if err != nil && t.HasConverter() {
				event, err = t.Converter.Convert(ctx, msg, err)
			}
			if err != nil {
				logger.Errorw("failed to decode message", zap.Error(err))
				return err
			}

			err = t.Receiver.Receive(ctx, *event, nil)
			if err != nil {
				logger.Warnw("kafka receiver return err", zap.Error(err))
				return err
			}

		}
	}

	return nil
}
func (t *Transport) toKafkaHeaders(h map[string]string) []kafka.Header {
	var kh []kafka.Header
	for k, v := range h {
		kh = append(kh, kafka.Header{Key: k, Value: []byte(v)})
	}
	return kh
}

func (t *Transport) fromKafkaHeaders(h []kafka.Header) map[string]string {
	m := map[string]string{}
	for _, header := range h {
		m[header.Key] = string(header.Value)
	}
	return m
}
