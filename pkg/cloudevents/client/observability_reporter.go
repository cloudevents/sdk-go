package client

import (
	"github.com/cloudevents/sdk-go/pkg/cloudevents/observability"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
)

var (
	ClientLatencyMs = stats.Float64("client/latency", "The latency in milliseconds per REPL loop", "ms")
)

var (
	ClientLatencyView = &view.View{
		Name:        "client/latency",
		Measure:     ClientLatencyMs,
		Description: "The distribution of latency inside of client.",
		Aggregation: view.Distribution(0, .01, .1, 1, 10, 100, 1000, 10000),
		TagKeys:     []tag.Key{observability.KeyMethod, observability.KeyResult}}
)

type Observed int32

const (
	ReportSend Observed = iota
	ReportReceive
	ReportReceiveFn
)

func (o Observed) TraceName() string {
	switch o {
	case ReportSend:
		return "client/send"
	case ReportReceive:
		return "client/receive"
	case ReportReceiveFn:
		return "client/receive/fn"
	default:
		return "client/unknown"
	}
}

func (o Observed) MethodName() string {
	switch o {
	case ReportSend:
		return "send"
	case ReportReceive:
		return "receive"
	case ReportReceiveFn:
		return "fn"
	default:
		return "unknown"
	}
}

func (oO Observed) LatencyMs() *stats.Float64Measure {
	return ClientLatencyMs
}
