/*
 Copyright 2026 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package kafka_franz

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/twmb/franz-go/pkg/kgo"

	"github.com/cloudevents/sdk-go/v2/binding"
	cecontext "github.com/cloudevents/sdk-go/v2/context"
	"github.com/cloudevents/sdk-go/v2/protocol"
	"github.com/cloudevents/sdk-go/v2/test"
)

type fakeProduceResult struct {
	err error
}

func (r fakeProduceResult) FirstErr() error {
	return r.err
}

type fakeFetchResult struct {
	clientClosed bool
	errors       []kgo.FetchError
	records      []*kgo.Record
}

func (r fakeFetchResult) IsClientClosed() bool {
	return r.clientClosed
}

func (r fakeFetchResult) Errors() []kgo.FetchError {
	return r.errors
}

func (r fakeFetchResult) Records() []*kgo.Record {
	return r.records
}

type fakeClient struct {
	produceErr error
	commitErr  error

	producedRecords  []*kgo.Record
	committedRecords []*kgo.Record
	fetches          []fetchResult

	allowRebalanceCalls int
	closeCalls          int
	closeAllowCalls     int
}

func (c *fakeClient) ProduceSync(context.Context, ...*kgo.Record) produceResult {
	panic("not implemented")
}

func (c *fakeClient) PollFetches(context.Context) fetchResult {
	if len(c.fetches) == 0 {
		return fakeFetchResult{clientClosed: true}
	}
	result := c.fetches[0]
	c.fetches = c.fetches[1:]
	return result
}

func (c *fakeClient) CommitRecords(_ context.Context, records ...*kgo.Record) error {
	for _, record := range records {
		c.committedRecords = append(c.committedRecords, cloneRecord(record))
	}
	return c.commitErr
}

func (c *fakeClient) AllowRebalance() {
	c.allowRebalanceCalls++
}

func (c *fakeClient) Close() {
	c.closeCalls++
}

func (c *fakeClient) CloseAllowingRebalance() {
	c.closeAllowCalls++
}

type fakeProtocolClient struct {
	*fakeClient
}

func (c *fakeProtocolClient) ProduceSync(_ context.Context, records ...*kgo.Record) produceResult {
	for _, record := range records {
		c.producedRecords = append(c.producedRecords, cloneRecord(record))
	}
	return fakeProduceResult{err: c.produceErr}
}

func TestNewProtocol(t *testing.T) {
	t.Run("requires client configuration", func(t *testing.T) {
		_, err := New()
		require.EqualError(t, err, "at least one of the following to initialize the protocol must be set: client or kgo client options")
	})

	t.Run("client and client options are mutually exclusive", func(t *testing.T) {
		_, err := New(
			WithClient(&kgo.Client{}),
			WithClientOptions(kgo.SeedBrokers("127.0.0.1:9092")),
		)
		require.EqualError(t, err, "the client and kgo client options must not be set together")
	})
}

func TestSendUsesDefaultTopicContextTopicAndMessageKey(t *testing.T) {
	client := &fakeProtocolClient{fakeClient: &fakeClient{}}
	p := &Protocol{
		client:               client,
		producerDefaultTopic: "default-topic",
		consumerIncoming:     make(chan binding.Message),
	}

	event := test.FullEvent()
	msg := (*binding.EventMessage)(&event)

	ctx := WithMessageKey(cecontext.WithTopic(context.Background(), "override-topic"), []byte("record-key"))
	err := p.Send(ctx, msg)
	require.NoError(t, err)

	require.Len(t, client.producedRecords, 1)
	record := client.producedRecords[0]
	require.Equal(t, "override-topic", record.Topic)
	require.Equal(t, []byte("record-key"), record.Key)
	require.NotEmpty(t, record.Headers)
}

func TestSendRequiresTopic(t *testing.T) {
	client := &fakeProtocolClient{fakeClient: &fakeClient{}}
	p := &Protocol{
		client:           client,
		consumerIncoming: make(chan binding.Message),
	}

	event := test.FullEvent()
	msg := (*binding.EventMessage)(&event)

	err := p.Send(context.Background(), msg)
	require.EqualError(t, err, "the producer topic must be set either by option or context")
	require.Empty(t, client.producedRecords)
}

func TestOpenInboundACKCommitsRecord(t *testing.T) {
	record := binaryRecord("input-topic", 2, 42)
	client := &fakeProtocolClient{fakeClient: &fakeClient{
		fetches: []fetchResult{
			fakeFetchResult{records: []*kgo.Record{record}},
			fakeFetchResult{clientClosed: true},
		},
	}}
	p := &Protocol{
		client:           client,
		consumerIncoming: make(chan binding.Message),
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- p.OpenInbound(context.Background())
	}()

	msg, err := p.Receive(context.Background())
	require.NoError(t, err)
	require.NoError(t, msg.Finish(protocol.ResultACK))
	require.NoError(t, waitForError(t, errCh))

	require.Len(t, client.committedRecords, 1)
	require.Equal(t, record.Topic, client.committedRecords[0].Topic)
	require.Equal(t, record.Offset, client.committedRecords[0].Offset)
	require.Equal(t, 1, client.allowRebalanceCalls)
}

func TestOpenInboundNACKSkipsCommit(t *testing.T) {
	record := binaryRecord("input-topic", 0, 5)
	client := &fakeProtocolClient{fakeClient: &fakeClient{
		fetches: []fetchResult{
			fakeFetchResult{records: []*kgo.Record{record}},
			fakeFetchResult{clientClosed: true},
		},
	}}
	p := &Protocol{
		client:           client,
		consumerIncoming: make(chan binding.Message),
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- p.OpenInbound(context.Background())
	}()

	msg, err := p.Receive(context.Background())
	require.NoError(t, err)
	require.NoError(t, msg.Finish(protocol.ResultNACK))
	require.NoError(t, waitForError(t, errCh))

	require.Empty(t, client.committedRecords)
	require.Equal(t, 1, client.allowRebalanceCalls)
}

func TestOpenInboundFinishReturnsCommitError(t *testing.T) {
	commitErr := errors.New("commit failed")
	client := &fakeProtocolClient{fakeClient: &fakeClient{
		commitErr: commitErr,
		fetches: []fetchResult{
			fakeFetchResult{records: []*kgo.Record{binaryRecord("input-topic", 0, 7)}},
			fakeFetchResult{clientClosed: true},
		},
	}}
	p := &Protocol{
		client:           client,
		consumerIncoming: make(chan binding.Message),
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- p.OpenInbound(context.Background())
	}()

	msg, err := p.Receive(context.Background())
	require.NoError(t, err)
	require.EqualError(t, msg.Finish(protocol.ResultACK), "commit failed")
	require.NoError(t, waitForError(t, errCh))
}

func TestOpenInboundReturnsFetchErrors(t *testing.T) {
	client := &fakeProtocolClient{fakeClient: &fakeClient{
		fetches: []fetchResult{
			fakeFetchResult{
				errors: []kgo.FetchError{{
					Topic:     "input-topic",
					Partition: 3,
					Err:       errors.New("permission denied"),
				}},
			},
		},
	}}
	p := &Protocol{
		client:           client,
		consumerIncoming: make(chan binding.Message),
	}

	err := p.OpenInbound(context.Background())
	require.EqualError(t, err, "fetch error for topic input-topic partition 3: permission denied")
}

func TestCloseOwnClient(t *testing.T) {
	client := &fakeProtocolClient{fakeClient: &fakeClient{}}
	p := &Protocol{
		client:           client,
		ownClient:        true,
		consumerIncoming: make(chan binding.Message),
	}

	require.NoError(t, p.Close(context.Background()))
	require.Equal(t, 1, client.closeAllowCalls)
	require.Zero(t, client.closeCalls)
}

func waitForError(t *testing.T, errCh <-chan error) error {
	t.Helper()
	select {
	case err := <-errCh:
		return err
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for OpenInbound to finish")
		return nil
	}
}

func binaryRecord(topic string, partition int32, offset int64) *kgo.Record {
	return &kgo.Record{
		Topic:     topic,
		Partition: partition,
		Offset:    offset,
		Value:     []byte("hello"),
		Headers: []kgo.RecordHeader{
			{Key: "ce_specversion", Value: []byte("1.0")},
			{Key: "ce_type", Value: []byte("example.type")},
			{Key: "ce_source", Value: []byte("example/source")},
			{Key: "ce_id", Value: []byte("example-id")},
			{Key: "ce_datacontenttype", Value: []byte("text/plain")},
		},
	}
}

func cloneRecord(record *kgo.Record) *kgo.Record {
	if record == nil {
		return nil
	}

	cloned := *record
	cloned.Key = append([]byte(nil), record.Key...)
	cloned.Value = append([]byte(nil), record.Value...)
	cloned.Headers = make([]kgo.RecordHeader, len(record.Headers))
	for i, header := range record.Headers {
		cloned.Headers[i] = kgo.RecordHeader{
			Key:   header.Key,
			Value: append([]byte(nil), header.Value...),
		}
	}
	return &cloned
}
