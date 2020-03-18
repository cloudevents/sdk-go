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

	"github.com/cloudevents/sdk-go/v2/binding"
	. "github.com/cloudevents/sdk-go/v2/binding/test"
	"github.com/cloudevents/sdk-go/v2/event"
	bindings "github.com/cloudevents/sdk-go/v2/protocol"
	"github.com/cloudevents/sdk-go/v2/protocol/kafka_sarama"
	"github.com/cloudevents/sdk-go/v2/protocol/test"
)

const (
	TEST_GROUP_ID = "test_group_id"
)

func TestSendStructuredMessageToStructured(t *testing.T) {
	close, s, r := testSenderReceiver(t)
	defer close()
	EachEvent(t, Events(), func(t *testing.T, eventIn event.Event) {
		eventIn = ExToStr(t, eventIn)

		in := MustCreateMockStructuredMessage(eventIn)
		test.SendReceive(t, binding.WithPreferredEventEncoding(context.TODO(), binding.EncodingStructured), in, s, r, func(out binding.Message) {
			eventOut := MustToEvent(t, context.Background(), out)
			assert.Equal(t, binding.EncodingStructured, out.ReadEncoding())
			AssertEventEquals(t, eventIn, ExToStr(t, eventOut))
		})
	})
}

func TestSendBinaryMessageToBinary(t *testing.T) {
	close, s, r := testSenderReceiver(t)
	defer close()
	EachEvent(t, Events(), func(t *testing.T, eventIn event.Event) {
		eventIn = ExToStr(t, eventIn)

		in := MustCreateMockBinaryMessage(eventIn)
		test.SendReceive(t, binding.WithPreferredEventEncoding(context.TODO(), binding.EncodingBinary), in, s, r, func(out binding.Message) {
			eventOut := MustToEvent(t, context.Background(), out)
			assert.Equal(t, binding.EncodingBinary, out.ReadEncoding())
			AssertEventEquals(t, eventIn, ExToStr(t, eventOut))
		})
	})
}

// To start a local environment for testing:
// docker run --rm --net=host -e ADV_HOST=localhost lensesio/fast-data-dev
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

func testSenderReceiver(t testing.TB, options ...kafka_sarama.SenderOptionFunc) (func(), bindings.Sender, bindings.Receiver) {
	client := testClient(t)

	topicName := "test-ce-client-" + uuid.New().String()
	r := kafka_sarama.NewConsumerFromClient(client, TEST_GROUP_ID, topicName)
	s, err := kafka_sarama.NewSenderFromClient(client, topicName, options...)
	require.NoError(t, err)

	go func() {
		require.NoError(t, r.OpenInbound(context.TODO()))
	}()

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
