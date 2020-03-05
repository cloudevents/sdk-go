package kafka_sarama_binding

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
	"github.com/cloudevents/sdk-go/pkg/event"
)

const (
	TEST_GROUP_ID = "test_group_id"
)

func TestSendStructuredMessageToStructuredWithKey(t *testing.T) {
	close, s, r := testSenderReceiver(t)
	defer close()
	test.EachEvent(t, test.Events(), func(t *testing.T, eventIn event.Event) {
		eventIn = test.ExToStr(t, eventIn)
		require.NoError(t, eventIn.Context.SetExtension("key", "aaa"))

		in := test.MustCreateMockStructuredMessage(eventIn)
		test.SendReceive(t, binding.WithPreferredEventEncoding(context.TODO(), binding.EncodingStructured), in, s, r, func(out binding.Message) {
			eventOut, encoding := test.MustToEvent(context.Background(), out)
			assert.Equal(t, binding.EncodingEvent, encoding)
			test.AssertEventEquals(t, eventIn, test.ExToStr(t, eventOut))
		})
	})
}

func TestSendStructuredMessageToStructuredWithoutKey(t *testing.T) {
	close, s, r := testSenderReceiver(t)
	defer close()
	test.EachEvent(t, test.Events(), func(t *testing.T, eventIn event.Event) {
		eventIn = test.ExToStr(t, eventIn)

		in := test.MustCreateMockStructuredMessage(eventIn)
		test.SendReceive(t, binding.WithPreferredEventEncoding(context.TODO(), binding.EncodingStructured), in, s, r, func(out binding.Message) {
			eventOut, encoding := test.MustToEvent(context.Background(), out)
			assert.Equal(t, binding.EncodingStructured, encoding)
			test.AssertEventEquals(t, eventIn, test.ExToStr(t, eventOut))
		})
	})
}

func TestSendBinaryMessageToBinaryWithKey(t *testing.T) {
	close, s, r := testSenderReceiver(t)
	defer close()
	test.EachEvent(t, test.Events(), func(t *testing.T, eventIn event.Event) {
		eventIn = test.ExToStr(t, eventIn)
		require.NoError(t, eventIn.Context.SetExtension("key", "aaa"))

		in := test.MustCreateMockBinaryMessage(eventIn)
		test.SendReceive(t, binding.WithPreferredEventEncoding(context.TODO(), binding.EncodingBinary), in, s, r, func(out binding.Message) {
			eventOut, encoding := test.MustToEvent(context.Background(), out)
			assert.Equal(t, binding.EncodingBinary, encoding)
			test.AssertEventEquals(t, eventIn, test.ExToStr(t, eventOut))
		})
	})
}

func TestSendBinaryMessageToBinaryWithoutKey(t *testing.T) {
	close, s, r := testSenderReceiver(t)
	defer close()
	test.EachEvent(t, test.Events(), func(t *testing.T, eventIn event.Event) {
		eventIn = test.ExToStr(t, eventIn)

		in := test.MustCreateMockBinaryMessage(eventIn)
		test.SendReceive(t, binding.WithPreferredEventEncoding(context.TODO(), binding.EncodingBinary), in, s, r, func(out binding.Message) {
			eventOut, encoding := test.MustToEvent(context.Background(), out)
			assert.Equal(t, binding.EncodingBinary, encoding)
			test.AssertEventEquals(t, eventIn, test.ExToStr(t, eventOut))
		})
	})
}

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
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	client, err := sarama.NewClient(strings.Split(s, ","), config)
	if err != nil {
		t.Skipf("Cannot create sarama client to servers [%s]: %v", s, err)
	}

	return client
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
		err = s.Close(context.TODO())
		require.NoError(t, err)
		err = client.Close()
		require.NoError(t, err)
	}, s, r
}

func BenchmarkSendReceive(b *testing.B) {
	c, s, r := testSenderReceiver(b)
	defer c() // Cleanup
	test.BenchmarkSendReceive(b, s, r)
}
