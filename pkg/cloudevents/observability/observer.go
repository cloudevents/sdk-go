package observability

import (
	"context"
	"time"

	"go.opencensus.io/stats"
	"go.opencensus.io/tag"
	"go.opencensus.io/trace"
)

type Observable interface {
	TraceName() string
	MethodName() string
	LatencyMs() *stats.Float64Measure
}

type Reporter interface {
	Error()
	OK()
}

type reporter struct {
	ctx     context.Context
	span    *trace.Span
	on      Observable
	start   time.Time
	measure stats.Measure
}

func NewReporter(ctx context.Context, on Observable) (context.Context, Reporter) {
	ctx, span := trace.StartSpan(ctx, on.TraceName())
	r := &reporter{
		ctx:   ctx,
		on:    on,
		span:  span,
		start: time.Now(),
	}
	r.tagMethod()
	return ctx, r
}

func (r *reporter) tagMethod() {
	var err error
	r.ctx, err = tag.New(r.ctx, tag.Insert(KeyMethod, r.on.MethodName()))
	if err != nil {
		panic(err)
	}
}

func (r *reporter) record() {
	ms := float64(time.Since(r.start) / time.Millisecond)
	stats.Record(r.ctx, r.on.LatencyMs().M(ms))
	r.span.End()
}

func (r *reporter) Error() {
	var err error
	r.ctx, err = tag.New(r.ctx, tag.Insert(KeyResult, ResultError))
	if err != nil {
		panic(err)
	}
	r.record()
}

func (r *reporter) OK() {
	var err error
	r.ctx, err = tag.New(r.ctx, tag.Insert(KeyResult, ResultOK))
	if err != nil {
		panic(err)
	}
	r.record()
}
