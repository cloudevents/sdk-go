/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

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

type failPattern struct {
	// Error to return.  nil for no injected error
	ErrReturn error
	// Times to return this error (or non-error).  0 for infinite.
	Count int
	// Duration to block prior to making the call
	Delay time.Duration
}

// Create a pubsub client.  If failureMap is provided, it gives a set of failures to induce in specific methods.
// failureMap is modified by the event processor and should not be read or modified after calling New()
func (pc *testPubsubClient) New(ctx context.Context, projectID string, failureMap map[string][]failPattern) (*pubsub.Client, error) {
	pc.srv = pstest.NewServer()
	var err error
	var conn *grpc.ClientConn
	if len(failureMap) == 0 {
		conn, err = grpc.Dial(pc.srv.Addr, grpc.WithInsecure())
	} else {
		conn, err = grpc.Dial(pc.srv.Addr, grpc.WithInsecure(), grpc.WithUnaryInterceptor(makeFailureIntercept(failureMap)))

	}
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

// Make a grpc failure injector that failes the specified methods with the
// specified rates.
func makeFailureIntercept(failureMap map[string][]failPattern) grpc.UnaryClientInterceptor {
	var lock sync.Mutex
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		var injectedErr error
		var delay time.Duration

		lock.Lock()
		if failureMap != nil {
			fpArr := failureMap[method]
			if len(fpArr) != 0 {
				injectedErr = fpArr[0].ErrReturn
				delay = fpArr[0].Delay
				if fpArr[0].Count != 0 {
					fpArr[0].Count--
					if fpArr[0].Count == 0 {
						failureMap[method] = fpArr[1:]
					}
				}
			}
		}
		lock.Unlock()
		if delay != 0 {
			time.Sleep(delay)
		}
		if injectedErr != nil {
			return injectedErr
		} else {
			return invoker(ctx, method, req, reply, cc, opts...)
		}
	}

}

// Verify that the topic exists prior to the call, deleting it via psconn succeeds, and
// the topic does not exist after the call.
func verifyTopicDeleteWorks(t *testing.T, client *pubsub.Client, psconn *Connection, topicID string) {
	ctx := context.Background()

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

// Verify that the topic exists before and after the call, and that deleting it via psconn fails
func verifyTopicDeleteFails(t *testing.T, client *pubsub.Client, psconn *Connection, topicID string) {
	ctx := context.Background()

	if ok, err := client.Topic(topicID).Exists(ctx); err != nil || !ok {
		t.Errorf("topic id=%s got exists=%v want=true, err=%v", topicID, ok, err)
	}

	if err := psconn.DeleteTopic(ctx); err == nil {
		t.Errorf("delete topic succeeded unexpectedly")
	}

	if ok, err := client.Topic(topicID).Exists(ctx); err != nil || !ok {
		t.Errorf("topic id=%s after delete got exists=%v want=true, err=%v", topicID, ok, err)
	}
}

// Test that publishing creates a topic
func TestPublishCreateTopic(t *testing.T) {
	ctx := context.Background()
	pc := &testPubsubClient{}
	defer pc.Close()

	projectID, topicID, subID := "test-project", "test-topic", "test-sub"

	client, err := pc.New(ctx, projectID, nil)
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

	verifyTopicDeleteWorks(t, client, psconn, topicID)
}

// Test that publishing to a topic with non default publish settings
func TestPublishWithCustomPublishSettings(t *testing.T) {
	t.Run("create topic and publish to it with custom settings", func(t *testing.T) {
		ctx := context.Background()
		pc := &testPubsubClient{}
		defer pc.Close()

		projectID, topicID, subID := "test-project", "test-topic", "test-sub"

		client, err := pc.New(ctx, projectID, nil)
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
			PublishSettings: &pubsub.PublishSettings{
				DelayThreshold:    100 * time.Millisecond,
				CountThreshold:    00,
				ByteThreshold:     2e6,
				Timeout:           120 * time.Second,
				BufferedByteLimit: 20 * pubsub.MaxPublishRequestBytes,
				FlowControlSettings: pubsub.FlowControlSettings{
					MaxOutstandingMessages: 10,
					MaxOutstandingBytes:    0,
					LimitExceededBehavior:  pubsub.FlowControlBlock,
				},
			},
		}

		topic, err := client.CreateTopic(ctx, topicID)
		if err != nil {
			t.Fatalf("failed to pre-create topic: %v", err)
		}
		topic.Stop()

		msg := &pubsub.Message{
			ID:   "msg-id-1",
			Data: []byte("msg-data-1"),
		}
		if _, err := psconn.Publish(ctx, msg); err != nil {
			t.Errorf("failed to publish message: %v", err)
		}
	})
}

// Test that publishing to an already created topic works and doesn't allow topic deletion
func TestPublishExistingTopic(t *testing.T) {
	for _, allowCreate := range []bool{true, false} {
		t.Run(fmt.Sprintf("allowCreate_%v", allowCreate), func(t *testing.T) {
			ctx := context.Background()
			pc := &testPubsubClient{}
			defer pc.Close()

			projectID, topicID, subID := "test-project", "test-topic", "test-sub"

			client, err := pc.New(ctx, projectID, nil)
			if err != nil {
				t.Fatalf("failed to create pubsub client: %v", err)
			}
			defer client.Close()

			psconn := &Connection{
				AllowCreateSubscription: true,
				AllowCreateTopic:        allowCreate,
				Client:                  client,
				ProjectID:               projectID,
				TopicID:                 topicID,
				SubscriptionID:          subID,
			}

			topic, err := client.CreateTopic(ctx, topicID)
			if err != nil {
				t.Fatalf("failed to pre-create topic: %v", err)
			}
			topic.Stop()

			msg := &pubsub.Message{
				ID:   "msg-id-1",
				Data: []byte("msg-data-1"),
			}
			if _, err := psconn.Publish(ctx, msg); err != nil {
				t.Errorf("failed to publish message: %v", err)
			}

			verifyTopicDeleteFails(t, client, psconn, topicID)
		})
	}
}

// Make sure that Publishing works if the original publish failed due to an
// error in one of the pubsub calls.
func TestPublishAfterPublishFailure(t *testing.T) {
	for _, failureMethod := range []string{
		"/google.pubsub.v1.Publisher/GetTopic",
		"/google.pubsub.v1.Publisher/CreateTopic",
		"/google.pubsub.v1.Publisher/Publish"} {
		t.Run(failureMethod, func(t *testing.T) {
			ctx := context.Background()
			pc := &testPubsubClient{}
			defer pc.Close()

			projectID, topicID, subID := "test-project", "test-topic", "test-sub"

			failureMap := make(map[string][]failPattern)
			failureMap[failureMethod] = []failPattern{{
				ErrReturn: fmt.Errorf("Injected error"),
				Count:     1,
				Delay:     0}}
			client, err := pc.New(ctx, projectID, failureMap)
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
			// Fails due to injected failure
			if _, err := psconn.Publish(ctx, msg); err == nil {
				t.Errorf("Expected publish failure, but didn't see it: %v", err)
			}
			// Succeeds
			if _, err := psconn.Publish(ctx, msg); err != nil {
				t.Errorf("failed to publish message: %v", err)
			}
			verifyTopicDeleteWorks(t, client, psconn, topicID)
		})
	}
}

// Test Publishing after Deleting a first version of a topic
func TestPublishCreateTopicAfterDelete(t *testing.T) {
	ctx := context.Background()
	pc := &testPubsubClient{}
	defer pc.Close()

	projectID, topicID, subID := "test-project", "test-topic", "test-sub"

	client, err := pc.New(ctx, projectID, nil)
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

	verifyTopicDeleteWorks(t, client, psconn, topicID)

	if _, err := psconn.Publish(ctx, msg); err != nil {
		t.Errorf("failed to publish message: %v", err)
	}

	verifyTopicDeleteWorks(t, client, psconn, topicID)
}

// Test that publishing fails if a topic doesn't exist and topic creation isn't allowed
func TestPublishCreateTopicNotAllowedFails(t *testing.T) {
	ctx := context.Background()
	pc := &testPubsubClient{}
	defer pc.Close()

	projectID, topicID, subID := "test-project", "test-topic", "test-sub"

	client, err := pc.New(ctx, projectID, nil)
	if err != nil {
		t.Fatalf("failed to create pubsub client: %v", err)
	}
	defer client.Close()

	psconn := &Connection{
		AllowCreateSubscription: true,
		AllowCreateTopic:        false,
		Client:                  client,
		ProjectID:               projectID,
		TopicID:                 topicID,
		SubscriptionID:          subID,
	}

	msg := &pubsub.Message{
		ID:   "msg-id-1",
		Data: []byte("msg-data-1"),
	}
	if _, err := psconn.Publish(ctx, msg); err == nil {
		t.Errorf("publish succeeded unexpectedly")
	}

	if ok, err := client.Topic(topicID).Exists(ctx); err == nil && ok {
		t.Errorf("topic id=%s got exists=%v want=false, err=%v", topicID, ok, err)
	}
}

// Test that failures of racing topic opens are reported out
func TestPublishParallelFailure(t *testing.T) {
	// This test is racy since it relies on a delay on one goroutine to
	// ensure a second hits a sync.Once while the other is still processing
	// it.  Optimistically try with a short delay, but retry with longer
	// ones so a failure is almost certainly a real failure, not a race.
	var overallError error
	for _, delay := range []time.Duration{time.Second / 4, 2 * time.Second, 10 * time.Second, 40 * time.Second} {
		overallError = func() error {
			failureMethod := "/google.pubsub.v1.Publisher/GetTopic"
			ctx := context.Background()
			pc := &testPubsubClient{}
			defer pc.Close()

			projectID, topicID, subID := "test-project", "test-topic", "test-sub"

			// Inject a failure, but also add a delay to the call sees the error
			failureMap := make(map[string][]failPattern)
			failureMap[failureMethod] = []failPattern{{
				ErrReturn: fmt.Errorf("Injected error"),
				Count:     1,
				Delay:     delay}}
			client, err := pc.New(ctx, projectID, failureMap)
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
			resChan := make(chan error)
			// Try a publish.  We want this to be the first to try to create the channel
			go func() {
				_, err := psconn.Publish(ctx, msg)
				resChan <- err
			}()

			// Try a second publish.  We hope the above has hit it's critical section before
			// this starts so that this reports out the error returned above.
			_, errPub1 := psconn.Publish(ctx, msg)
			errPub2 := <-resChan
			if errPub1 == nil || errPub2 == nil {
				return fmt.Errorf("expected dual expected failure, saw (%v) (%v) last run", errPub1, errPub2)
			} else if errPub1 == nil && errPub2 == nil {
				t.Fatalf("Dual success when expecting at least one failure, delay %v", delay)
			}
			return nil
		}()
		// Saw a successfull run, no retry needed.
		if overallError == nil {
			break
		}
		// Failure.  The loop will bump the delay and try again(unless we've hit the max reasonable delay)
	}
	if overallError != nil {
		t.Errorf(overallError.Error())
	}
}

// Test that creating a subscription also creates the topic and subscription
func TestReceiveCreateTopicAndSubscription(t *testing.T) {
	ctx := context.Background()
	pc := &testPubsubClient{}
	defer pc.Close()

	projectID, topicID, subID := "test-project", "test-topic", "test-sub"
	client, err := pc.New(ctx, projectID, nil)
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
	// Sleep waiting for the goroutine to create the topic and subscription
	// If it takes over a minute, run the test anyway to get failure logging
	for _, delay := range []time.Duration{time.Second / 4, time.Second, 20 * time.Second, 40 * time.Second} {
		time.Sleep(delay)
		ok, err := client.Subscription(subID).Exists(ctx)
		if ok == true && err == nil {
			break
		}
	}

	if ok, err := client.Topic(topicID).Exists(ctx); err != nil || !ok {
		t.Errorf("topic id=%s got exists=%v want=true, err=%v", topicID, ok, err)
	}

	if ok, err := client.Subscription(subID).Exists(ctx); err != nil || !ok {
		t.Errorf("subscription id=%s got exists=%v want=true, err=%v", subID, ok, err)
	}

	si, err := psconn.getOrCreateSubscriptionInfo(context.Background(), true)
	if err != nil {
		t.Errorf("error getting subscription info %v", err)
	}
	if si.sub.ReceiveSettings.NumGoroutines != DefaultReceiveSettings.NumGoroutines {
		t.Errorf("subscription receive settings have NumGoroutines=%d, want %d",
			si.sub.ReceiveSettings.NumGoroutines, DefaultReceiveSettings.NumGoroutines)
	}

	cancel()

	if err := psconn.DeleteSubscription(ctx); err != nil {
		t.Errorf("delete subscription failed: %v", err)
	}

	if ok, err := client.Subscription(subID).Exists(ctx); err != nil || ok {
		t.Errorf("subscription id=%s got exists=%v want=false, err=%v", subID, ok, err)
	}

	verifyTopicDeleteWorks(t, client, psconn, topicID)
}

// Test receive on an existing topic and subscription also works.
func TestReceiveExistingTopic(t *testing.T) {
	for _, allow := range [](struct{ Sub, Topic bool }){{true, true}, {true, false}, {false, true}, {false, false}} {
		t.Run(fmt.Sprintf("sub_%v__topic_%v", allow.Sub, allow.Topic), func(t *testing.T) {

			ctx := context.Background()
			pc := &testPubsubClient{}
			defer pc.Close()

			projectID, topicID, subID := "test-project", "test-topic", "test-sub"
			client, err := pc.New(ctx, projectID, nil)
			if err != nil {
				t.Fatalf("failed to create pubsub client: %v", err)
			}
			defer client.Close()

			psconn := &Connection{
				AllowCreateSubscription: allow.Sub,
				AllowCreateTopic:        allow.Topic,
				Client:                  client,
				ProjectID:               projectID,
				TopicID:                 topicID,
				SubscriptionID:          subID,
			}

			topic, err := client.CreateTopic(ctx, topicID)
			if err != nil {
				pc.Close()
				t.Fatalf("failed to pre-create topic: %v", err)
			}

			_, err = client.CreateSubscription(ctx, subID, pubsub.SubscriptionConfig{
				Topic:             topic,
				AckDeadline:       DefaultAckDeadline,
				RetentionDuration: DefaultRetentionDuration,
			})
			topic.Stop()
			if err != nil {
				pc.Close()
				t.Fatalf("failed to pre-createsubscription: %v", err)
			}

			ctx2, cancel := context.WithCancel(ctx)
			go psconn.Receive(ctx2, func(_ context.Context, msg *pubsub.Message) {
				msg.Ack()
			})
			// Block waiting for receive to succeed
			si, err := psconn.getOrCreateSubscriptionInfo(context.Background(), false)
			if err != nil {
				t.Errorf("error getting subscription info %v", err)
			}
			if si.sub.ReceiveSettings.NumGoroutines != DefaultReceiveSettings.NumGoroutines {
				t.Errorf("subscription receive settings have NumGoroutines=%d, want %d",
					si.sub.ReceiveSettings.NumGoroutines, DefaultReceiveSettings.NumGoroutines)
			}

			cancel()

			if err := psconn.DeleteSubscription(ctx); err == nil {
				t.Errorf("delete subscription unexpectedly succeeded")
			}

			if ok, err := client.Subscription(subID).Exists(ctx); err != nil || !ok {
				t.Errorf("subscription id=%s got exists=%v want=true, err=%v", subID, ok, err)
			}

			verifyTopicDeleteFails(t, client, psconn, topicID)
		})
	}
}

// Test that creating a subscription after a failed attempt to create a subsciption works
func TestReceiveCreateSubscriptionAfterFailure(t *testing.T) {
	for _, failureMethod := range []string{
		"/google.pubsub.v1.Publisher/GetTopic",
		"/google.pubsub.v1.Publisher/CreateTopic",
		"/google.pubsub.v1.Subscriber/GetSubscription",
		"/google.pubsub.v1.Subscriber/CreateSubscription"} {
		t.Run(failureMethod, func(t *testing.T) {

			ctx := context.Background()
			pc := &testPubsubClient{}
			defer pc.Close()

			projectID, topicID, subID := "test-project", "test-topic", "test-sub"
			failureMap := make(map[string][]failPattern)
			failureMap[failureMethod] = []failPattern{{
				ErrReturn: fmt.Errorf("Injected error"),
				Count:     1,
				Delay:     0}}
			client, err := pc.New(ctx, projectID, failureMap)
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

			// We expect this receive to fail due to the injected error
			ctx2, cancel := context.WithCancel(ctx)
			errRet := make(chan error)
			go func() {
				errRet <- psconn.Receive(ctx2, func(_ context.Context, msg *pubsub.Message) {
					msg.Ack()
				})
			}()

			select {
			case err := <-errRet:
				if err == nil {
					t.Fatalf("unexpected nil error from Receive")
				}
			case <-time.After(time.Minute):
				cancel()
				t.Fatalf("timeout waiting for receive error")
			}

			// We expect this receive to succeed
			errRet2 := make(chan error)
			ctx2, cancel = context.WithCancel(context.Background())
			go func() {
				errRet2 <- psconn.Receive(ctx2, func(_ context.Context, msg *pubsub.Message) {
					msg.Ack()
				})
			}()
			// Sleep waiting for the goroutine to create the topic and subscription
			// If it takes over a minute, run the test anyway to get failure logging
			for _, delay := range []time.Duration{time.Second / 4, time.Second, 20 * time.Second, 40 * time.Second} {
				time.Sleep(delay)
				ok, err := client.Subscription(subID).Exists(ctx)
				if ok == true && err == nil {
					break
				}
			}
			select {
			case err := <-errRet2:
				t.Errorf("unexpected error from Receive: %v", err)
			default:
			}
			if ok, err := client.Topic(topicID).Exists(ctx); err != nil || !ok {
				t.Errorf("topic id=%s got exists=%v want=true, err=%v", topicID, ok, err)
			}

			if ok, err := client.Subscription(subID).Exists(ctx); err != nil || !ok {
				t.Errorf("subscription id=%s got exists=%v want=true, err=%v", subID, ok, err)
			}

			cancel()
		})
	}
}

// Test that lack of create privileges for topic or subscription causes a receive to fail for
// a non-existing subscription and topic
func TestReceiveCreateDisallowedFail(t *testing.T) {
	ctx := context.Background()
	pc := &testPubsubClient{}
	defer pc.Close()

	for _, allow := range [](struct{ Sub, Topic bool }){{false, true}, {true, false}, {false, false}} {
		t.Run(fmt.Sprintf("sub_%v__topic_%v", allow.Sub, allow.Topic), func(t *testing.T) {

			projectID, topicID, subID := "test-project", "test-topic", "test-sub"
			client, err := pc.New(ctx, projectID, nil)
			if err != nil {
				t.Fatalf("failed to create pubsub client: %v", err)
			}
			defer client.Close()

			psconn := &Connection{
				AllowCreateSubscription: allow.Sub,
				AllowCreateTopic:        allow.Topic,
				Client:                  client,
				ProjectID:               projectID,
				TopicID:                 topicID,
				SubscriptionID:          subID,
			}

			ctx2, cancel := context.WithCancel(ctx)
			errRet := make(chan error)
			go func() {
				errRet <- psconn.Receive(ctx2, func(_ context.Context, msg *pubsub.Message) {
					msg.Ack()
				})
			}()

			select {
			case err := <-errRet:
				if err == nil {
					t.Fatalf("unexpected nil error from Receive")
				}
			case <-time.After(time.Minute):
				cancel()
				t.Fatalf("timeout waiting for receive error")
			}
			cancel()
		})
	}
}

// Test a full round trip of a message
func TestPublishReceiveRoundtrip(t *testing.T) {
	ctx := context.Background()
	pc := &testPubsubClient{}
	defer pc.Close()

	projectID, topicID, subID := "test-project", "test-topic", "test-sub"
	client, err := pc.New(ctx, projectID, nil)
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

// Test a full round trip of a message with ordering key
func TestPublishReceiveRoundtripWithOrderingKey(t *testing.T) {
	ctx := context.Background()
	pc := &testPubsubClient{}
	defer pc.Close()

	projectID, topicID, subID := "test-project", "test-topic", "test-sub"
	client, err := pc.New(ctx, projectID, nil)
	if err != nil {
		t.Fatalf("failed to create pubsub client: %v", err)
	}
	defer client.Close()

	psconn := &Connection{
		MessageOrdering:         true,
		AllowCreateSubscription: true,
		AllowCreateTopic:        true,
		Client:                  client,
		ProjectID:               projectID,
		TopicID:                 topicID,
		SubscriptionID:          subID,
	}

	wantMsgs := make(map[string]string)
	wantMsgsOrdering := make(map[string]string)
	gotMsgs := make(map[string]string)
	gotMsgsOrdering := make(map[string]string)

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
		gotMsgsOrdering[string(msg.Data)] = string(msg.OrderingKey)
		msg.Ack()
		wg.Done()
	})
	// Wait a little bit for the subscription creation to complete.
	time.Sleep(time.Second)

	for i := 0; i < 10; i++ {
		data := fmt.Sprintf("data-%d", i)
		wantMsgs[data] = data

		order := fmt.Sprintf("order-%d", i)
		wantMsgsOrdering[data] = order

		if _, err := psconn.Publish(ctx, &pubsub.Message{Data: []byte(data), OrderingKey: order}); err != nil {
			t.Errorf("failed to publish message: %v", err)
		}
		wg.Add(1)
	}

	wg.Wait()
	cancel()

	if diff := cmp.Diff(gotMsgs, wantMsgs); diff != "" {
		t.Errorf("received unexpected messages (-want +got):\n%s", diff)
	}

	if diff := cmp.Diff(gotMsgsOrdering, wantMsgsOrdering); diff != "" {
		t.Errorf("received unexpected message ordering keys (-want +got):\n%s", diff)
	}
}
