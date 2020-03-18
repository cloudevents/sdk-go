package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"contrib.go.opencensus.io/exporter/prometheus"
	"go.opencensus.io/examples/exporter"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
	"go.opencensus.io/zpages"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/client"
	cecontext "github.com/cloudevents/sdk-go/v2/context"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/cloudevents/sdk-go/v2/event/datacodec"
	"github.com/cloudevents/sdk-go/v2/event/datacodec/json"
	"github.com/cloudevents/sdk-go/v2/event/datacodec/xml"
	"github.com/cloudevents/sdk-go/v2/types"
)

func main() {
	ctx := context.Background()

	c, err := client.NewDefault()
	if err != nil {
		log.Fatalf("failed to create client, %v", err)
	}

	go mainSender()
	go mainMetrics()

	log.Printf("will listen on :8080\n")
	log.Fatalf("failed to start receiver: %s", c.StartReceiver(ctx, gotEvent))
}

// Example is the expected incoming event.
type Example struct {
	Sequence int    `json:"id"`
	Message  string `json:"message"`
}

func gotEvent(event event.Event) {
	data := &Example{}
	if err := event.DataAs(data); err != nil {
		fmt.Printf("failed to get data as Example: %s\n", err.Error())
		return
	}

	time.Sleep(time.Duration(rand.Intn(3000)) * time.Millisecond)

	fmt.Printf("%s: %d - %q\n", event.Context.GetType(), data.Sequence, data.Message)
}

var source = types.ParseURIRef("https://github.com/cloudevents/sdk-go/v2/cmd/samples/sender")

func mainSender() {
	ctx := cecontext.WithTarget(context.Background(), "http://localhost:8181/")

	c, err := client.NewDefault()
	if err != nil {
		log.Fatalf("failed to create client, %v", err)
	}

	for { //ever
		for i := 0; i < 1000; i++ {
			data := &Example{
				Sequence: i,
				Message:  "Hello, World!",
			}
			e := event.Event{
				Context: event.EventContextV1{
					Type:   "com.cloudevents.sample.sent",
					Source: *source,
				}.AsV1(),
			}
			_ = e.SetData(cloudevents.ApplicationJSON, data)

			if resp, err := c.Request(ctx, e); err != nil {
				log.Printf("failed to send: %v", err)
			} else if resp != nil {
				fmt.Printf("got back a response event of type %s", resp.Context.GetType())
			} else {
				log.Printf("%s: %d - %s", e.Context.GetType(), data.Sequence, data.Message)
			}
			time.Sleep(2000 * time.Millisecond)
		}
	}
}

func mainMetrics() {

	printExporter := &exporter.PrintExporter{}

	exporter, err := prometheus.NewExporter(prometheus.Options{})
	if err != nil {
		log.Fatalf("Failed to create the Stackdriver stats exporter: %v", err)
	}

	h := http.NewServeMux()

	h.Handle("/metrics", exporter)
	zpages.Handle(h, "/debug")

	// Register the stats exporter
	view.RegisterExporter(exporter)

	trace.RegisterExporter(printExporter)
	// Always trace for this demo. In a production application, you should
	// configure this to a trace.ProbabilitySampler set at the desired
	// probability.
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})

	// Register the views
	if err := view.Register(
		client.LatencyView,
		//transporthttp.LatencyView, // TODO: add back http metrics.
		event.EventMarshalLatencyView,
		json.LatencyView,
		xml.LatencyView,
		datacodec.LatencyView,
	); err != nil {
		log.Fatalf("Failed to register views: %v", err)
	}

	view.SetReportingPeriod(2 * time.Second)

	log.Fatal("failed metrics ListenAndServe ", http.ListenAndServe("localhost:9090", h))
}
