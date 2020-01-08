package internal

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/pubsub/pstest"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
)

type testPubsubClient struct {
	srv  *pstest.Server
	conn *grpc.ClientConn
}

func (pc *testPubsubClient) New(ctx context.Context, projectID string) (*pubsub.Client, error) {
	pc.srv = pstest.NewServer()
	conn, err := grpc.Dial(pc.srv.Addr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	pc.conn = conn
	return pubsub.NewClient(ctx, projectID, option.WithGRPCConn(conn))
}

func (pc *testPubsubClient) Close() {
	pc.srv.Close()
	pc.conn.Close()
}

func TestPublishCreateTopic(t *testing.T) {
	ctx := context.Background()
	pc := &testPubsubClient{}
	defer pc.Close()

	projectID, topicID, subID := "test-project", "test-topic", "test-sub"
	client, err := pc.New(ctx, projectID)
	if err != nil {
		t.Fatalf("failed to create pubsub client: %v", err)
	}
	defer client.Close()

	psconn := &Connection{
		AllowCreateSubscription: true,
		AllowCreateTopic:        true,
		Client:                  client,
		ProjectID:               projectID,
		TopicID:                 topicID,
		SubscriptionID:          subID,
	}

	msg := &pubsub.Message{
		ID:   "msg-id-1",
		Data: []byte("msg-data-1"),
	}
	if _, err := psconn.Publish(ctx, msg); err != nil {
		t.Errorf("failed to publish message: %v", err)
	}

	if ok, err := client.Topic(topicID).Exists(ctx); err != nil || !ok {
		t.Errorf("topic id=%s got exists=%v want=true, err=%v", topicID, ok, err)
	}

	if err := psconn.DeleteTopic(ctx); err != nil {
		t.Errorf("delete topic failed: %v", err)
	}

	if ok, err := client.Topic(topicID).Exists(ctx); err != nil || ok {
		t.Errorf("topic id=%s got exists=%v want=false, err=%v", topicID, ok, err)
	}
}

func TestReceiveCreateTopicAndSubscription(t *testing.T) {
	ctx := context.Background()
	pc := &testPubsubClient{}
	defer pc.Close()

	projectID, topicID, subID := "test-project", "test-topic", "test-sub"
	client, err := pc.New(ctx, projectID)
	if err != nil {
		t.Fatalf("failed to create pubsub client: %v", err)
	}
	defer client.Close()

	psconn := &Connection{
		AllowCreateSubscription: true,
		AllowCreateTopic:        true,
		Client:                  client,
		ProjectID:               projectID,
		TopicID:                 topicID,
		SubscriptionID:          subID,
	}

	ctx2, cancel := context.WithCancel(ctx)
	go psconn.Receive(ctx2, func(_ context.Context, msg *pubsub.Message) {
		msg.Ack()
	})
	// Sleep one sec for the goroutine to create the topic and subscription.
	time.Sleep(time.Second)

	if ok, err := client.Topic(topicID).Exists(ctx); err != nil || !ok {
		t.Errorf("topic id=%s got exists=%v want=true, err=%v", topicID, ok, err)
	}

	if ok, err := client.Subscription(subID).Exists(ctx); err != nil || !ok {
		t.Errorf("subscription id=%s got exists=%v want=true, err=%v", subID, ok, err)
	}

	if psconn.sub.ReceiveSettings.NumGoroutines != DefaultReceiveSettings.NumGoroutines {
		t.Errorf("subscription receive settings have NumGoroutines=%d, want %d",
			psconn.sub.ReceiveSettings.NumGoroutines, DefaultReceiveSettings.NumGoroutines)
	}

	cancel()

	if err := psconn.DeleteSubscription(ctx); err != nil {
		t.Errorf("delete subscription failed: %v", err)
	}

	if ok, err := client.Subscription(subID).Exists(ctx); err != nil || ok {
		t.Errorf("subscription id=%s got exists=%v want=false, err=%v", subID, ok, err)
	}

	if err := psconn.DeleteTopic(ctx); err != nil {
		t.Errorf("delete topic failed: %v", err)
	}

	if ok, err := client.Topic(topicID).Exists(ctx); err != nil || ok {
		t.Errorf("topic id=%s got exists=%v want=false, err=%v", topicID, ok, err)
	}
}

func TestPublishReceiveRoundtrip(t *testing.T) {
	ctx := context.Background()
	pc := &testPubsubClient{}
	defer pc.Close()

	projectID, topicID, subID := "test-project", "test-topic", "test-sub"
	client, err := pc.New(ctx, projectID)
	if err != nil {
		t.Fatalf("failed to create pubsub client: %v", err)
	}
	defer client.Close()

	psconn := &Connection{
		AllowCreateSubscription: true,
		AllowCreateTopic:        true,
		Client:                  client,
		ProjectID:               projectID,
		TopicID:                 topicID,
		SubscriptionID:          subID,
	}

	wantMsgs := make(map[string]string)
	gotMsgs := make(map[string]string)
	wg := &sync.WaitGroup{}

	ctx2, cancel := context.WithCancel(ctx)
	mux := &sync.Mutex{}
	// Pubsub will drop all messages if there is no subscription.
	// Call Receive first so that subscription can be created before
	// we publish any message.
	go psconn.Receive(ctx2, func(_ context.Context, msg *pubsub.Message) {
		mux.Lock()
		defer mux.Unlock()
		gotMsgs[string(msg.Data)] = string(msg.Data)
		msg.Ack()
		wg.Done()
	})
	// Wait a little bit for the subscription creation to complete.
	time.Sleep(time.Second)

	for i := 0; i < 10; i++ {
		data := fmt.Sprintf("data-%d", i)
		wantMsgs[data] = data

		if _, err := psconn.Publish(ctx, &pubsub.Message{Data: []byte(data)}); err != nil {
			t.Errorf("failed to publish message: %v", err)
		}
		wg.Add(1)
	}

	wg.Wait()
	cancel()

	if diff := cmp.Diff(gotMsgs, wantMsgs); diff != "" {
		t.Errorf("received unexpected messages (-want +got):\n%s", diff)
	}
}
