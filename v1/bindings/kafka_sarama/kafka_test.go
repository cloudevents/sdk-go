// +build kafka

package kafka_sarama

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Shopify/sarama"

	cloudevents "github.com/cloudevents/sdk-go/v1"
	"github.com/cloudevents/sdk-go/v1/binding"
	"github.com/cloudevents/sdk-go/v1/binding/test"
)

const (
	TEST_GROUP_ID = "test_group_id"
)

var (
	e                                   = test.FullEvent()
	StructuredConsumerMessageWithoutKey = &sarama.ConsumerMessage{
		Value: test.MustJSON(e),
		Headers: []*sarama.RecordHeader{{
			Key:   []byte(ContentType),
			Value: []byte(cloudevents.ApplicationCloudEventsJSON),
		}},
	}
	StructuredConsumerMessageWithKey = &sarama.ConsumerMessage{
		Key:   []byte("aaa"),
		Value: test.MustJSON(e),
		Headers: []*sarama.RecordHeader{{
			Key:   []byte(ContentType),
			Value: []byte(cloudevents.ApplicationCloudEventsJSON),
		}},
	}
	BinaryConsumerMessageWithoutKey = &sarama.ConsumerMessage{
		Value: []byte("hello world!"),
		Headers: mustToSaramaConsumerHeaders(map[string]string{
			"ce_type":            e.Type(),
			"ce_source":          e.Source(),
			"ce_id":              e.ID(),
			"ce_time":            test.Timestamp.String(),
			"ce_specversion":     "1.0",
			"ce_dataschema":      test.Schema.String(),
			"ce_datacontenttype": "text/json",
			"ce_subject":         "topic",
			"ce_exta":            "someext",
		}),
	}
	BinaryConsumerMessageWithKey = &sarama.ConsumerMessage{
		Key:   []byte("akey"),
		Value: []byte("hello world!"),
		Headers: mustToSaramaConsumerHeaders(map[string]string{
			"ce_type":            e.Type(),
			"ce_source":          e.Source(),
			"ce_id":              e.ID(),
			"ce_time":            test.Timestamp.String(),
			"ce_specversion":     "1.0",
			"ce_dataschema":      test.Schema.String(),
			"ce_datacontenttype": "text/json",
			"ce_subject":         "topic",
			"ce_exta":            "someext",
		}),
	}
)

func TestSendStructuredMessageToStructuredWithKey(t *testing.T) {
	close, s, r := testSenderReceiver(t)
	defer close()
	test.EachEvent(t, test.Events(), func(t *testing.T, eventIn cloudevents.Event) {
		eventIn = test.ExToStr(t, eventIn)
		require.NoError(t, eventIn.Context.SetExtension("key", "aaa"))

		in := test.NewMockStructuredMessage(eventIn)
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
	test.EachEvent(t, test.Events(), func(t *testing.T, eventIn cloudevents.Event) {
		eventIn = test.ExToStr(t, eventIn)

		in := test.NewMockStructuredMessage(eventIn)
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
	test.EachEvent(t, test.Events(), func(t *testing.T, eventIn cloudevents.Event) {
		eventIn = test.ExToStr(t, eventIn)
		require.NoError(t, eventIn.Context.SetExtension("key", "aaa"))

		in := test.NewMockBinaryMessage(eventIn)
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
	test.EachEvent(t, test.Events(), func(t *testing.T, eventIn cloudevents.Event) {
		eventIn = test.ExToStr(t, eventIn)

		in := test.NewMockBinaryMessage(eventIn)
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

func testSenderReceiver(t testing.TB, options ...SenderOptionFunc) (func(), binding.Sender, binding.Receiver) {
	client := testClient(t)

	topicName := "test-ce-client-" + uuid.New().String()
	r := NewReceiver(client, TEST_GROUP_ID, topicName)
	s, err := NewSender(client, topicName, options...)
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

func mustToSaramaConsumerHeaders(m map[string]string) []*sarama.RecordHeader {
	res := make([]*sarama.RecordHeader, len(m))
	i := 0
	for k, v := range m {
		res[i] = &sarama.RecordHeader{Key: []byte(k), Value: []byte(v)}
		i++
	}
	return res
}
