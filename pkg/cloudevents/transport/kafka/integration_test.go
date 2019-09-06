package kafka_test

import (
	"context"
	"fmt"
	tkafka "github.com/cloudevents/sdk-go/pkg/cloudevents/transport/kafka"
	"testing"

	"gotest.tools/assert"

	"github.com/confluentinc/confluent-kafka-go-dev/kafka"

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/docker/docker/client"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// Basic data struct.
type Example struct {
	Sequence int    `json:"id"`
	Message  string `json:"message"`
}

func TestSendCloudEventRoundTrip(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	zr := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "confluentinc/cp-zookeeper:4.0.0",
			ExposedPorts: []string{"2181/tcp"},
			Env:          map[string]string{"ZOOKEEPER_CLIENT_PORT": "2181"},
			WaitingFor:   wait.ForLog("INFO binding to port 0.0.0.0/0.0.0.0:2181 (org.apache.zookeeper.server.NIOServerCnxnFactory)"),
			Name:         "zookeeper",
		},
		Started: true,
	}
	zC, err := testcontainers.GenericContainer(ctx, zr)
	if err != nil {
		t.Error(err)
	}

	//err = zC.Start(ctx)
	//if err != nil {
	//	t.Error(err)
	//}
	//zl, err := zC.Logs(ctx)
	//if err != nil {
	//	t.Error(err)
	//}
	//zlb, err := ioutil.ReadAll(zl)
	//if err != nil {
	//	t.Error(err)
	//}
	//
	//fmt.Println(string(zlb))
	defer zC.Terminate(ctx)

	ip, err := getContainerIP(ctx, zC.GetContainerID())
	if err != nil {
		t.Error(err)
	}

	zkStr := fmt.Sprintf("%s:%s", ip, "2181")

	fmt.Println(zkStr)
	kr := testcontainers.ContainerRequest{
		Image:        "confluentinc/cp-kafka:5.3.0",
		ExposedPorts: []string{"9092:9092"},
		Env: map[string]string{
			"KAFKA_ZOOKEEPER_CONNECT":                zkStr,
			"KAFKA_ADVERTISED_LISTENERS":             "PLAINTEXT://localhost:9092",
			"KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR": "1",
		},
		WaitingFor: wait.ForLog("started (kafka.server.KafkaServer"),
		Name:       "kafka",
	}
	kC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: kr,
		Started:          true,
	})

	//err = kC.Start(ctx)
	if err != nil {
		t.Error(err)
	}
	//kl, err := kC.Logs(ctx)
	//if err != nil {
	//	t.Error(err)
	//}
	//klb, err := ioutil.ReadAll(kl)
	//if err != nil {
	//	t.Error(err)
	//}
	//
	//fmt.Println(string(klb))

	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	defer kC.Terminate(ctx)

	//if resp.StatusCode != http.StatusOK {
	//	t.Fatal("Expected status code %d. Got %d.", http.StatusOK, resp.StatusCode)
	//}

	port, err := kC.MappedPort(ctx, "9092/tcp")
	if err != nil {
		t.Error(err)
	}
	kb := fmt.Sprintf("%s:%s", "0.0.0.0", port.Port())

	kc := &kafka.ConfigMap{
		"bootstrap.servers":       kb,
		"group.id":                "testgroup",
		"auto.offset.reset":       "earliest",
		"broker.version.fallback": "0.10.0.0",
		"api.version.fallback.ms": 0,
	}

	var p *kafka.Producer
	if p, err = kafka.NewProducer(kc); err != nil {
		t.Fatal("failed to create kafka transport", err.Error())
	}

	fmt.Println(kc)
	kt, err := tkafka.New(context.Background(),
		tkafka.WithTopic("test"),
		tkafka.WithKafkaConfig(kc),
		tkafka.WithProducer(p))
	if err != nil {
		t.Fatal("failed to create kafka transport", err.Error())

	}
	c, err := cloudevents.NewClient(kt, cloudevents.WithTimeNow(), cloudevents.WithUUIDs())
	if err != nil {
		t.Fatal("failed to create client", err.Error())
	}

	sample := &Example{
		Sequence: 0,
		Message:  "HELLO",
	}
	event := cloudevents.NewEvent(cloudevents.VersionV03)
	event.SetType("com.cloudevents.sample.sent")
	event.SetSource("github.com/cloudevents/sdk-go/cmd/samples/kafka/sender/")
	_ = event.SetData(sample)
	sentEvent, err := c.Send(context.Background(), event)
	fmt.Printf("sentEvent: %+v\n", sentEvent)

	if err != nil {
		t.Fatal("failed to send", err)
	}

	p.Flush(1 * 1000)

	err = c.StartReceiver(ctx, func(ctx context.Context, ce cloudevents.Event, resp *cloudevents.EventResponse) error {
		fmt.Printf("Event Context: %+v\n", ce.Context)

		data := &Example{}
		if err := ce.DataAs(data); err != nil {
			t.Fatal("Got Data Error", err)
		}

		fmt.Printf("Data: %+v\n", data)

		assert.Equal(t, ce.Context.AsV03().SpecVersion, sentEvent.Context.AsV03().SpecVersion)
		assert.Equal(t, ce.Context.AsV03().ID, sentEvent.Context.AsV03().ID)
		assert.Equal(t, ce.Context.AsV03().Source, sentEvent.Context.AsV03().Source)
		assert.Equal(t, ce.Context.AsV03().Type, sentEvent.Context.AsV03().Type)
		assert.Equal(t, ce.Context.AsV03().Time.String(), sentEvent.Context.AsV03().Time.String())
		//assert.Equal(t, ce.Data, sentEvent.Data)
		assert.Equal(t, data.Sequence, sample.Sequence)
		assert.Equal(t, data.Message, sample.Message)

		cancel()
		return nil
	})

	if err != nil {
		t.Fatal("failed to recieve", err)
	}

}

func getContainerIP(ctx context.Context, id string) (string, error) {
	client, err := client.NewEnvClient()
	if err != nil {
		return "", err
	}
	client.NegotiateAPIVersion(ctx)

	inspect, err := client.ContainerInspect(ctx, id)
	if err != nil {
		return "", err
	}
	return inspect.NetworkSettings.IPAddress, nil
}
