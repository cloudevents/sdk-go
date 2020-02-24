// +build kafka

package kafka_sarama_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/Shopify/sarama"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/binding/test"
	"github.com/cloudevents/sdk-go/pkg/bindings/kafka_sarama"
	ce "github.com/cloudevents/sdk-go/pkg/cloudevents"
)

const (
	TEST_GROUP_ID = "test_group_id"
)

func testClient(t testing.TB) sarama.Client {
	t.Helper()
	s := os.Getenv("TEST_KAFKA_BOOTSTRAP_SERVER")
	if s == "" {
		s = "localhost:9092"
	}

	config := sarama.NewConfig()
	config.Version = sarama.V2_0_0_0
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	client, err := sarama.NewClient(strings.Split(s, ","), config)
	if err != nil {
		t.Skipf("Cannot create sarama client to servers [%s]: %v", s, err)
	}

	return client
}

func TestSendSkipBinary(t *testing.T) {
	close, s, r := testSenderReceiver(t)
	defer close()
	test.EachEvent(t, test.Events(), func(t *testing.T, eventIn ce.Event) {
		eventIn = test.ExToStr(t, eventIn)
		in := test.NewMockBinaryMessage(eventIn)
		test.SendReceive(t, kafka_sarama.WithSkipKeyExtension(binding.WithSkipDirectBinaryEncoding(binding.WithPreferredEventEncoding(context.Background(), binding.EncodingStructured), true)), in, s, r, func(out binding.Message) {
			eventOut, encoding := test.MustToEvent(context.Background(), out)
			assert.Equal(t, encoding, binding.EncodingStructured)
			test.AssertEventEquals(t, eventIn, test.ExToStr(t, eventOut))
		})
	})
}

func TestSendSkipStructured(t *testing.T) {
	close, s, r := testSenderReceiver(t)
	defer close()
	test.EachEvent(t, test.Events(), func(t *testing.T, eventIn ce.Event) {
		eventIn = test.ExToStr(t, eventIn)
		in := test.NewMockStructuredMessage(eventIn)
		test.SendReceive(t, binding.WithSkipDirectStructuredEncoding(context.Background(), true), in, s, r, func(out binding.Message) {
			eventOut, encoding := test.MustToEvent(context.Background(), out)
			assert.Equal(t, encoding, binding.EncodingBinary)
			test.AssertEventEquals(t, eventIn, test.ExToStr(t, eventOut))
		})
	})
}

func TestSendBinaryReceiveBinary(t *testing.T) {
	close, s, r := testSenderReceiver(t)
	defer close()
	test.EachEvent(t, test.Events(), func(t *testing.T, eventIn ce.Event) {
		eventIn = test.ExToStr(t, eventIn)
		in := test.NewMockBinaryMessage(eventIn)
		test.SendReceive(t, context.Background(), in, s, r, func(out binding.Message) {
			eventOut, encoding := test.MustToEvent(context.Background(), out)
			assert.Equal(t, encoding, binding.EncodingBinary)
			test.AssertEventEquals(t, eventIn, test.ExToStr(t, eventOut))
		})
	})
}

func TestSendStructReceiveStruct(t *testing.T) {
	close, s, r := testSenderReceiver(t)
	defer close()
	test.EachEvent(t, test.Events(), func(t *testing.T, eventIn ce.Event) {
		eventIn = test.ExToStr(t, eventIn)
		in := test.NewMockStructuredMessage(eventIn)
		test.SendReceive(t, context.Background(), in, s, r, func(out binding.Message) {
			eventOut, encoding := test.MustToEvent(context.Background(), out)
			assert.Equal(t, encoding, binding.EncodingStructured)
			test.AssertEventEquals(t, eventIn, test.ExToStr(t, eventOut))
		})
	})
}

func TestSendEventReceiveBinary(t *testing.T) {
	close, s, r := testSenderReceiver(t)
	defer close()
	test.EachEvent(t, test.Events(), func(t *testing.T, eventIn ce.Event) {
		eventIn = test.ExToStr(t, eventIn)
		in := binding.EventMessage(eventIn)
		test.SendReceive(t, context.Background(), in, s, r, func(out binding.Message) {
			eventOut, encoding := test.MustToEvent(context.Background(), out)
			assert.Equal(t, encoding, binding.EncodingBinary)
			test.AssertEventEquals(t, eventIn, test.ExToStr(t, eventOut))
		})
	})
}

func testSenderReceiver(t testing.TB, options ...kafka_sarama.SenderOptionFunc) (func(), binding.Sender, binding.Receiver) {
	client := testClient(t)

	topicName := "test-ce-client-" + uuid.New().String()
	r := kafka_sarama.NewReceiver(client, TEST_GROUP_ID, topicName)
	s, err := kafka_sarama.NewSender(client, topicName, options...)
	require.NoError(t, err)

	return func() {
		err = r.Close(context.TODO())
		require.NoError(t, err)
		_ = s.Close(context.TODO())
		require.NoError(t, err)
	}, s, r
}

func BenchmarkSendReceive(b *testing.B) {
	c, s, r := testSenderReceiver(b)
	defer c() // Cleanup
	test.BenchmarkSendReceive(b, s, r)
}
