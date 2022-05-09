/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package http

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"go.uber.org/atomic"

	"github.com/cloudevents/sdk-go/v2/client"

	"github.com/google/uuid"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	cehttp "github.com/cloudevents/sdk-go/v2/protocol/http"
)

// Blocking Test:

//         Obj -> Send -> Wire Format -> Receive -> Got[n]
// Given:   ^                                           ^==Want
// Obj is an event of a version.
// Client is a set to binary or

type BlockingSenderReceiverTest struct {
	now          time.Time
	event        *cloudevents.Event
	receiverWait time.Duration
	timeout      time.Duration
	want         int
}

type BlockingSenderReceiverTestOutput struct {
	duration time.Duration
	got      int
}

type BlockingSenderReceiverTestCases map[string]BlockingSenderReceiverTest

func TestNonBlockingSenderReceiver(t *testing.T) {
	t.Parallel()
	now := time.Now()

	testCases := BlockingSenderReceiverTestCases{
		"10 at 1 second": {
			now: now,
			event: &cloudevents.Event{
				Context: cloudevents.EventContextV1{
					Type:            "unit.test.client.sent.10.1",
					Source:          *cloudevents.ParseURIRef("/unit/test/client"),
					Subject:         strptr("resource"),
					DataContentType: cloudevents.StringOfApplicationJSON(),
				}.AsV1(),
				DataEncoded: toBytes(map[string]interface{}{"hello": "unittest"}),
			},
			receiverWait: 1 * time.Second,
			want:         10,
			timeout:      5 * time.Second,
		},
		"50 at 5 second": {
			now: now,
			event: &cloudevents.Event{
				Context: cloudevents.EventContextV1{
					Type:            "unit.test.client.sent.50.5",
					Source:          *cloudevents.ParseURIRef("/unit/test/client"),
					Subject:         strptr("resource"),
					DataContentType: cloudevents.StringOfApplicationJSON(),
				}.AsV1(),
				DataEncoded: toBytes(map[string]interface{}{"hello": "unittest"}),
			},
			receiverWait: 5 * time.Second,
			want:         50,
			timeout:      15 * time.Second,
		},
		"100 at 10 seconds": {
			now: now,
			event: &cloudevents.Event{
				Context: cloudevents.EventContextV1{
					Type:            "unit.test.client.sent.100.10",
					Source:          *cloudevents.ParseURIRef("/unit/test/client"),
					Subject:         strptr("resource"),
					DataContentType: cloudevents.StringOfApplicationJSON(),
				}.AsV1(),
				DataEncoded: toBytes(map[string]interface{}{"hello": "unittest"}),
			},
			receiverWait: 10 * time.Second,
			want:         100,
			timeout:      30 * time.Second,
		},
	}

	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			ReceiverNonBlocking(t, tc)
		})
	}
}

func TestBlockingSenderReceiver(t *testing.T) {
	t.Parallel()
	now := time.Now()

	testCases := BlockingSenderReceiverTestCases{
		"10 at 100 milisecond": {
			now: now,
			event: &cloudevents.Event{
				Context: cloudevents.EventContextV1{
					Type:            "unit.test.client.sent.10.1",
					Source:          *cloudevents.ParseURIRef("/unit/test/client"),
					Subject:         strptr("resource"),
					DataContentType: cloudevents.StringOfApplicationJSON(),
				}.AsV1(),
				DataEncoded: toBytes(map[string]interface{}{"hello": "unittest"}),
			},
			receiverWait: 100 * time.Millisecond,
			want:         10,
			timeout:      5 * time.Second,
		},
		"50 at 20 milisecond": {
			now: now,
			event: &cloudevents.Event{
				Context: cloudevents.EventContextV1{
					Type:            "unit.test.client.sent.50.5",
					Source:          *cloudevents.ParseURIRef("/unit/test/client"),
					Subject:         strptr("resource"),
					DataContentType: cloudevents.StringOfApplicationJSON(),
				}.AsV1(),
				DataEncoded: toBytes(map[string]interface{}{"hello": "unittest"}),
			},
			receiverWait: 20 * time.Millisecond,
			want:         50,
			timeout:      5 * time.Second,
		},
		"100 at 10 milisecond": {
			now: now,
			event: &cloudevents.Event{
				Context: cloudevents.EventContextV1{
					Type:            "unit.test.client.sent.100.10",
					Source:          *cloudevents.ParseURIRef("/unit/test/client"),
					Subject:         strptr("resource"),
					DataContentType: cloudevents.StringOfApplicationJSON(),
				}.AsV1(),
				DataEncoded: toBytes(map[string]interface{}{"hello": "unittest"}),
			},
			receiverWait: 10 * time.Millisecond,
			want:         100,
			timeout:      5 * time.Second,
		},
	}

	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			ReceiverBlocking(t, tc)
		})
	}
}

const verbose = false

func ReceiverNonBlocking(t *testing.T, tc BlockingSenderReceiverTest) {
	output := receive(t, tc, client.WithPollGoroutines(1))

	if tc.want != output.got {
		t.Errorf("expected %d, got %d", tc.want, output)
	}

	// Look at how long the test took.
	dm := output.duration.Milliseconds()
	tw := tc.receiverWait.Milliseconds() * 110 / 100 // 110% budget.
	if dm > tw {
		t.Errorf("expected test duration to be around ~%d ms, actual %d ms", tc.receiverWait.Milliseconds(), dm)
	}
}

func ReceiverBlocking(t *testing.T, tc BlockingSenderReceiverTest) {
	output := receive(t, tc, client.WithPollGoroutines(1), client.WithBlockingCallback())

	if tc.want != output.got {
		t.Errorf("expected %d, got %d", tc.want, output)
	}

	// Look at how long the test took.
	dm := output.duration.Milliseconds()
	tw := tc.receiverWait.Milliseconds() * int64(tc.want) // no concurrent processing
	if dm < tw {
		t.Errorf("expected test duration to be over %d ms, actual %d ms", tw, dm)
	}
}

func receive(t *testing.T, tc BlockingSenderReceiverTest, copts ...client.Option) *BlockingSenderReceiverTestOutput {
	opts := make([]cehttp.Option, 0)
	opts = append(opts, cloudevents.WithPort(0)) // random port

	protocol, err := cloudevents.NewHTTP(opts...)
	if err != nil {
		t.Fatal(err)
	}

	copts = append(copts, cloudevents.WithUUIDs(), cloudevents.WithEventDefaulter(AlwaysThen(tc.now)))

	ce, err := cloudevents.NewClient(protocol, copts...)
	if err != nil {
		t.Fatal(err)
	}

	testID := uuid.New().String()
	tc.event.SetExtension(unitTestIDKey, testID)

	recvCtx, recvCancel := context.WithTimeout(context.Background(), tc.timeout)
	defer recvCancel()

	wg := new(sync.WaitGroup)
	wg.Add(tc.want)

	got := new(atomic.Int32)

	go func() {
		if err := ce.StartReceiver(recvCtx, func(event cloudevents.Event) {
			if verbose {
				t.Logf("%s - sleep", event.ID())
			}
			got.Inc()
			time.Sleep(tc.receiverWait)
			wg.Done()
			if verbose {
				t.Logf("%s - done", event.ID())
			}
		}); err != nil {
			t.Errorf("[receiver] unexpected error: %s", err)
		}
	}()

	time.Sleep(time.Second) // wait for the receiver to start.

	then := time.Now()

	sendCtx, sendCancel := context.WithTimeout(context.Background(), tc.timeout)
	defer sendCancel()
	sendCtx = cloudevents.ContextWithTarget(sendCtx, fmt.Sprintf("http://localhost:%d", protocol.GetListeningPort()))

	for i := 0; i < tc.want; i++ {
		go func() {
			if result := ce.Send(sendCtx, *tc.event); !cloudevents.IsACK(result) {
				t.Errorf("[sender] unexpected result: %s", result)
			}
		}()
	}

	wg.Wait()

	duration := time.Since(then)

	time.Sleep(tc.receiverWait) // cool off just in case we have some more sleepers.

	return &BlockingSenderReceiverTestOutput{
		duration: duration,
		got:      int(got.Load()),
	}
}
