package http

import (
	"github.com/cloudevents/sdk-go/pkg/cloudevents/observability"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
)

var (
	TransportHttpLatencyMs = stats.Float64(
		"transport/http/latency",
		"The latency in milliseconds for the http transport methods for CloudEvents.",
		"ms")
)

var (
	TransportHttpLatencyView = &view.View{
		Name:        "transport/http/latency",
		Measure:     TransportHttpLatencyMs,
		Description: "The distribution of latency inside of http transport for CloudEvents.",
		Aggregation: view.Distribution(0, .01, .1, 1, 10, 100, 1000, 10000),
		TagKeys:     observability.LatencyTags(),
	}
)

type Observed int32

const (
	ReportSend Observed = iota
	ReportReceive
	ReportServeHTTP
)

func (o Observed) TraceName() string {
	switch o {
	case ReportSend:
		return "transport/http/send"
	case ReportReceive:
		return "transport/http/receive"
	case ReportServeHTTP:
		return "transport/http/servehttp"
	default:
		return "transport/http/unknown"
	}
}

func (o Observed) MethodName() string {
	switch o {
	case ReportSend:
		return "send"
	case ReportReceive:
		return "receive"
	case ReportServeHTTP:
		return "servehttp"
	default:
		return "unknown"
	}
}

func (o Observed) LatencyMs() *stats.Float64Measure {
	return TransportHttpLatencyMs
}
