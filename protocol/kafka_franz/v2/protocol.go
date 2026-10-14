/*
 Copyright 2026 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package kafka_franz

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/twmb/franz-go/pkg/kgo"

	"github.com/cloudevents/sdk-go/v2/binding"
	cecontext "github.com/cloudevents/sdk-go/v2/context"
	"github.com/cloudevents/sdk-go/v2/protocol"
)

var (
	_ protocol.Sender   = (*Protocol)(nil)
	_ protocol.Opener   = (*Protocol)(nil)
	_ protocol.Receiver = (*Protocol)(nil)
	_ protocol.Closer   = (*Protocol)(nil)
)

type produceResult interface {
	FirstErr() error
}

type fetchResult interface {
	IsClientClosed() bool
	Errors() []kgo.FetchError
	Records() []*kgo.Record
}

type kafkaClient interface {
	ProduceSync(context.Context, ...*kgo.Record) produceResult
	PollFetches(context.Context) fetchResult
	CommitRecords(context.Context, ...*kgo.Record) error
	AllowRebalance()
	Close()
	CloseAllowingRebalance()
}

type kgoClient struct {
	client *kgo.Client
}

type kgoProduceResult struct {
	result kgo.ProduceResults
}

type kgoFetchResult struct {
	result kgo.Fetches
}

func newKgoClient(client *kgo.Client) kafkaClient {
	return &kgoClient{client: client}
}

func (c *kgoClient) ProduceSync(ctx context.Context, records ...*kgo.Record) produceResult {
	return kgoProduceResult{result: c.client.ProduceSync(ctx, records...)}
}

func (c *kgoClient) PollFetches(ctx context.Context) fetchResult {
	return kgoFetchResult{result: c.client.PollFetches(ctx)}
}

func (c *kgoClient) CommitRecords(ctx context.Context, records ...*kgo.Record) error {
	return c.client.CommitRecords(ctx, records...)
}

func (c *kgoClient) AllowRebalance() {
	c.client.AllowRebalance()
}

func (c *kgoClient) Close() {
	c.client.Close()
}

func (c *kgoClient) CloseAllowingRebalance() {
	c.client.CloseAllowingRebalance()
}

func (r kgoProduceResult) FirstErr() error {
	return r.result.FirstErr()
}

func (r kgoFetchResult) IsClientClosed() bool {
	return r.result.IsClientClosed()
}

func (r kgoFetchResult) Errors() []kgo.FetchError {
	return r.result.Errors()
}

func (r kgoFetchResult) Records() []*kgo.Record {
	return r.result.Records()
}

// Protocol implements a CloudEvents Kafka transport backed by franz-go.
type Protocol struct {
	client               kafkaClient
	clientOptions        []kgo.Opt
	ownClient            bool
	producerDefaultTopic string

	consumerIncoming chan binding.Message

	receiverMux    sync.Mutex
	receiverCtx    context.Context
	receiverCancel context.CancelFunc
	receiverOpened bool
}

// New creates a new franz-go Kafka transport.
func New(opts ...Option) (*Protocol, error) {
	p := &Protocol{
		consumerIncoming: make(chan binding.Message),
	}
	if err := p.applyOptions(opts...); err != nil {
		return nil, err
	}

	if p.client != nil && len(p.clientOptions) > 0 {
		return nil, errors.New("the client and kgo client options must not be set together")
	}

	if p.client == nil {
		if len(p.clientOptions) == 0 {
			return nil, errors.New("at least one of the following to initialize the protocol must be set: client or kgo client options")
		}
		clientOptions := append([]kgo.Opt{}, p.clientOptions...)
		clientOptions = append(clientOptions, kgo.DisableAutoCommit(), kgo.BlockRebalanceOnPoll())

		client, err := kgo.NewClient(clientOptions...)
		if err != nil {
			return nil, err
		}
		p.client = newKgoClient(client)
		p.ownClient = true
	}

	return p, nil
}

func (p *Protocol) applyOptions(opts ...Option) error {
	for _, fn := range opts {
		if err := fn(p); err != nil {
			return err
		}
	}
	return nil
}

// Send transmits a CloudEvent using franz-go's synchronous produce API.
func (p *Protocol) Send(ctx context.Context, in binding.Message, transformers ...binding.Transformer) (err error) {
	if p.client == nil {
		return errors.New("producer client must be set")
	}
	defer in.Finish(err)

	topic := cecontext.TopicFrom(ctx)
	if topic == "" {
		topic = p.producerDefaultTopic
	}
	if topic == "" {
		return errors.New("the producer topic must be set either by option or context")
	}

	record := &kgo.Record{Topic: topic}
	if key := MessageKeyFrom(ctx); key != nil {
		record.Key = key
	}

	if err = WriteProducerMessage(ctx, in, record, transformers...); err != nil {
		return fmt.Errorf("create producer record: %w", err)
	}

	if err = p.client.ProduceSync(ctx, record).FirstErr(); err != nil {
		return fmt.Errorf("produce record: %w", err)
	}
	return nil
}

// OpenInbound starts the receive loop. This call blocks until the receiver stops.
func (p *Protocol) OpenInbound(ctx context.Context) error {
	if p.client == nil {
		return errors.New("consumer client must be set")
	}

	p.receiverMux.Lock()
	if p.receiverOpened {
		p.receiverMux.Unlock()
		return errors.New("receiver already open")
	}
	p.receiverOpened = true
	p.receiverCtx, p.receiverCancel = context.WithCancel(ctx)
	receiveCtx := p.receiverCtx
	p.receiverMux.Unlock()

	defer func() {
		p.receiverMux.Lock()
		if p.receiverCancel != nil {
			p.receiverCancel()
			p.receiverCancel = nil
		}
		p.receiverCtx = nil
		p.receiverMux.Unlock()
		close(p.consumerIncoming)
	}()

	logger := cecontext.LoggerFrom(ctx)

	for {
		fetches := p.client.PollFetches(receiveCtx)
		if fetches.IsClientClosed() {
			return nil
		}

		records := fetches.Records()
		if len(records) > 0 {
			batch := receiveBatch{ctx: receiveCtx, client: p.client}
			for _, record := range records {
				msg := batch.wrap(record)
				select {
				case p.consumerIncoming <- msg:
				case <-receiveCtx.Done():
					batch.drop()
					p.client.AllowRebalance()
					return receiveCtx.Err()
				}
			}
			batch.wait()
			p.client.AllowRebalance()
		}

		if err := joinFetchErrors(fetches.Errors()); err != nil {
			logger.Warnw("franz-go fetch error", "error", err)
			return err
		}

		if err := receiveCtx.Err(); err != nil {
			return err
		}
	}
}

// Receive implements protocol.Receiver.
func (p *Protocol) Receive(ctx context.Context) (binding.Message, error) {
	select {
	case msg, ok := <-p.consumerIncoming:
		if !ok {
			return nil, io.EOF
		}
		return msg, nil
	case <-ctx.Done():
		return nil, io.EOF
	}
}

// Close cleans up resources after use.
func (p *Protocol) Close(context.Context) error {
	p.receiverMux.Lock()
	cancel := p.receiverCancel
	p.receiverCancel = nil
	p.receiverCtx = nil
	p.receiverMux.Unlock()

	if cancel != nil {
		cancel()
	}
	if p.ownClient && p.client != nil {
		p.client.CloseAllowingRebalance()
	}
	return nil
}

type receiveBatch struct {
	ctx    context.Context
	client kafkaClient
	wg     sync.WaitGroup
}

func (b *receiveBatch) wrap(record *kgo.Record) *receivedMessage {
	b.wg.Add(1)
	return &receivedMessage{
		Message: NewMessage(record),
		finish: func(err error) error {
			defer b.wg.Done()
			if !protocol.IsACK(err) || b.ctx.Err() != nil {
				return nil
			}
			return b.client.CommitRecords(b.ctx, record)
		},
	}
}

func (b *receiveBatch) drop() {
	b.wg.Done()
}

func (b *receiveBatch) wait() {
	b.wg.Wait()
}

func joinFetchErrors(fetchErrors []kgo.FetchError) error {
	if len(fetchErrors) == 0 {
		return nil
	}

	errs := make([]error, 0, len(fetchErrors))
	for _, fetchErr := range fetchErrors {
		errs = append(errs, fmt.Errorf("fetch error for topic %s partition %d: %w", fetchErr.Topic, fetchErr.Partition, fetchErr.Err))
	}
	return errors.Join(errs...)
}
