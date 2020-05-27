package extensions_test

import (
	"context"
	"encoding/hex"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/cloudevents/sdk-go/v2/binding"
	bindingtest "github.com/cloudevents/sdk-go/v2/binding/test"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/cloudevents/sdk-go/v2/test"

	"github.com/cloudevents/sdk-go/v2/extensions"
	"github.com/cloudevents/sdk-go/v2/types"
	"github.com/google/go-cmp/cmp"
	"go.opencensus.io/trace"
	"go.opencensus.io/trace/tracestate"
)

type Data struct {
	Message string
}

var now = types.Timestamp{Time: time.Now().UTC()}

var sourceUrl, _ = url.Parse("http://example.com/source")
var source = &types.URIRef{URL: *sourceUrl}
var sourceUri = &types.URIRef{URL: *sourceUrl}

var schemaUrl, _ = url.Parse("http://example.com/schema")
var schema = &types.URIRef{URL: *schemaUrl}
var schemaUri = &types.URI{URL: *schemaUrl}

type values struct {
	context interface{}
	want    map[string]interface{}
}

func TestAddTracingAttributes_Scenario1(t *testing.T) {
	var st = extensions.DistributedTracingExtension{
		TraceParent: "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",
		TraceState:  "rojo=00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01,congo=lZWRzIHRoNhcm5hbCBwbGVhc3VyZS4=",
	}

	var eventContextVersions = map[string]values{
		"EventContextV1": {
			context: event.EventContextV1{
				ID:              "ABC-123",
				Time:            &now,
				Type:            "com.example.test",
				DataSchema:      schemaUri,
				DataContentType: event.StringOfApplicationJSON(),
				Source:          *sourceUri,
			},
			want: map[string]interface{}{"traceparent": st.TraceParent, "tracestate": st.TraceState},
		},
		"EventContextV03": {
			context: event.EventContextV03{
				ID:              "ABC-123",
				Time:            &now,
				Type:            "com.example.test",
				SchemaURL:       schema,
				DataContentType: event.StringOfApplicationJSON(),
				Source:          *source,
			},
			want: map[string]interface{}{"traceparent": st.TraceParent, "tracestate": st.TraceState},
		},
	}

	for k, ecv := range eventContextVersions {
		testAddTracingAttributesFunc(t, st, ecv, k)
	}
}

func TestAddTracingAttributes_Scenario2(t *testing.T) {
	var st = extensions.DistributedTracingExtension{
		TraceParent: "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",
	}

	var eventContextVersions = map[string]values{
		"EventContextV1": {
			context: event.EventContextV1{
				ID:              "ABC-123",
				Time:            &now,
				Type:            "com.example.test",
				DataSchema:      schemaUri,
				DataContentType: event.StringOfApplicationJSON(),
				Source:          *sourceUri,
			},
			want: map[string]interface{}{"traceparent": st.TraceParent},
		},
		"EventContextV03": {
			context: event.EventContextV03{
				ID:              "ABC-123",
				Time:            &now,
				Type:            "com.example.test",
				SchemaURL:       schema,
				DataContentType: event.StringOfApplicationJSON(),
				Source:          *source,
			},
			want: map[string]interface{}{"traceparent": st.TraceParent},
		},
	}

	for k, ecv := range eventContextVersions {
		testAddTracingAttributesFunc(t, st, ecv, k)
	}
}

func TestAddTracingAttributes_Scenario3(t *testing.T) {
	var st = extensions.DistributedTracingExtension{}

	var eventContextVersions = map[string]values{
		"EventContextV1": {
			context: event.EventContextV1{
				ID:              "ABC-123",
				Time:            &now,
				Type:            "com.example.test",
				DataSchema:      schemaUri,
				DataContentType: event.StringOfApplicationJSON(),
				Source:          *sourceUri,
			},
			want: map[string]interface{}(nil),
		},
		"EventContextV03": {
			context: event.EventContextV03{
				ID:              "ABC-123",
				Time:            &now,
				Type:            "com.example.test",
				SchemaURL:       schema,
				DataContentType: event.StringOfApplicationJSON(),
				Source:          *source,
			},
			want: map[string]interface{}(nil),
		},
	}

	for k, ecv := range eventContextVersions {
		testAddTracingAttributesFunc(t, st, ecv, k)
	}
}

func TestAddTracingAttributes_Scenario4(t *testing.T) {
	var st = extensions.DistributedTracingExtension{
		TraceState: "rojo=00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01,congo=lZWRzIHRoNhcm5hbCBwbGVhc3VyZS4=",
	}

	var eventContextVersions = map[string]values{
		"EventContextV1": {
			context: event.EventContextV1{
				ID:              "ABC-123",
				Time:            &now,
				Type:            "com.example.test",
				DataSchema:      schemaUri,
				DataContentType: event.StringOfApplicationJSON(),
				Source:          *sourceUri,
			},
			want: map[string]interface{}(nil),
		},
		"EventContextV03": {
			context: event.EventContextV03{
				ID:              "ABC-123",
				Time:            &now,
				Type:            "com.example.test",
				SchemaURL:       schema,
				DataContentType: event.StringOfApplicationJSON(),
				Source:          *source,
			},
			want: map[string]interface{}(nil),
		},
	}

	for k, ecv := range eventContextVersions {
		testAddTracingAttributesFunc(t, st, ecv, k)
	}
}

func testAddTracingAttributesFunc(t *testing.T, st extensions.DistributedTracingExtension, ecv values, ces string) {
	var e event.Event
	switch ces {
	case "EventContextV1":
		e = event.Event{Context: ecv.context.(event.EventContextV1).AsV1()}
		e.SetData(event.ApplicationJSON, &Data{Message: "Hello world"})
	case "EventContextV03":
		e = event.Event{Context: ecv.context.(event.EventContextV03).AsV03()}
		e.SetData(event.ApplicationJSON, &Data{Message: "Hello world"})
	}
	st.AddTracingAttributes(&e)
	got := e.Extensions()

	if diff := cmp.Diff(ecv.want, got); diff != "" {
		t.Errorf("\nunexpected (-want, +got) = %v", diff)
	}
}

func decodeTID(s string) (tid [16]byte, err error) {
	buf, err := hex.DecodeString(s)
	copy(tid[:], buf)
	return
}

func decodeSID(s string) (sid [8]byte, err error) {
	buf, err := hex.DecodeString(s)
	copy(sid[:], buf)
	return
}

func TestConvertSpanContext(t *testing.T) {
	tid, err := decodeTID("4bf92f3577b34da6a3ce929d0e0e4736")
	if err != nil {
		t.Fatalf("failed to decode traceID: %v", err)
	}
	sid, err := decodeSID("00f067aa0ba902b7")
	if err != nil {
		t.Fatalf("failed to decode spanID: %v", err)
	}
	ts, err := tracestate.New(nil,
		tracestate.Entry{
			Key:   "rojo",
			Value: "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",
		},
		tracestate.Entry{
			Key:   "tenant@congo",
			Value: "lZWRzIHRoNhcm5hbCBwbGVhc3VyZS4",
		},
	)
	if err != nil {
		t.Fatalf("failed to make tracestate: %v", err)
	}
	tests := []struct {
		name string
		sc   trace.SpanContext
		ext  extensions.DistributedTracingExtension
	}{{
		name: "with tracestate",
		sc: trace.SpanContext{
			TraceID:      trace.TraceID(tid),
			SpanID:       sid,
			TraceOptions: 1,
			Tracestate:   ts,
		},
		ext: extensions.DistributedTracingExtension{
			TraceParent: "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",
			TraceState:  "rojo=00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01,tenant@congo=lZWRzIHRoNhcm5hbCBwbGVhc3VyZS4",
		},
	}, {
		name: "without tracestate",
		sc: trace.SpanContext{
			TraceID:      trace.TraceID(tid),
			SpanID:       sid,
			TraceOptions: 1,
		},
		ext: extensions.DistributedTracingExtension{
			TraceParent: "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",
		},
	}, {
		name: "unsampled",
		sc: trace.SpanContext{
			TraceID:      trace.TraceID(tid),
			SpanID:       sid,
			TraceOptions: 0,
		},
		ext: extensions.DistributedTracingExtension{
			TraceParent: "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-00",
		},
	}}

	for _, tt := range tests {
		tt := tt
		t.Run("FromSpanContext: "+tt.name, func(t *testing.T) {
			t.Parallel()
			got := extensions.FromSpanContext(tt.sc)
			if diff := cmp.Diff(tt.ext, got); diff != "" {
				t.Errorf("\nunexpected (-want, +got) = %v", diff)
			}
		})
		t.Run("ToSpanContext: "+tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := tt.ext.ToSpanContext()
			if err != nil {
				t.Error(err)
			}
			if diff := cmp.Diff(
				tt.sc, got,
				cmp.Transformer(
					"entries",
					func(ts tracestate.Tracestate) []tracestate.Entry {
						return ts.Entries()
					},
				),
			); diff != "" {
				t.Errorf("\nunexpected (-want, +got) = %v", diff)
			}
		})
	}
}

func TestDistributedTracingExtension_ReadTransformer_empty(t *testing.T) {
	e := test.MinEvent()
	e.Context = e.Context.AsV1()

	tests := []bindingtest.TransformerTestArgs{
		{
			Name:         "Read from Mock Structured message",
			InputMessage: bindingtest.MustCreateMockStructuredMessage(t, e),
			WantEvent:    e,
		},
		{
			Name:         "Read from Mock Binary message",
			InputMessage: bindingtest.MustCreateMockBinaryMessage(e),
			WantEvent:    e,
		},
		{
			Name:       "Read from Event message",
			InputEvent: e,
			WantEvent:  e,
		},
	}
	for _, tt := range tests {
		ext := extensions.DistributedTracingExtension{}
		tt.Transformers = binding.Transformers{ext.ReadTransformer()}
		bindingtest.RunTransformerTests(t, context.TODO(), []bindingtest.TransformerTestArgs{tt})
		require.Zero(t, ext.TraceState)
		require.Zero(t, ext.TraceParent)
	}
}

func TestDistributedTracingExtension_ReadTransformer(t *testing.T) {
	e := test.MinEvent()
	e.Context = e.Context.AsV1()
	wantExt := extensions.DistributedTracingExtension{
		TraceParent: "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",
		TraceState:  "rojo=00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01,congo=lZWRzIHRoNhcm5hbCBwbGVhc3VyZS4=",
	}
	wantExt.AddTracingAttributes(&e)

	tests := []bindingtest.TransformerTestArgs{
		{
			Name:         "Read from Mock Structured message",
			InputMessage: bindingtest.MustCreateMockStructuredMessage(t, e),
			WantEvent:    e,
		},
		{
			Name:         "Read from Mock Binary message",
			InputMessage: bindingtest.MustCreateMockBinaryMessage(e),
			WantEvent:    e,
		},
		{
			Name:       "Read from Event message",
			InputEvent: e,
			WantEvent:  e,
		},
	}
	for _, tt := range tests {
		haveExt := extensions.DistributedTracingExtension{}
		tt.Transformers = binding.Transformers{haveExt.ReadTransformer()}
		bindingtest.RunTransformerTests(t, context.TODO(), []bindingtest.TransformerTestArgs{tt})
		require.Equal(t, wantExt.TraceParent, haveExt.TraceParent)
		require.Equal(t, wantExt.TraceState, haveExt.TraceState)
	}
}

func TestDistributedTracingExtension_WriteTransformer(t *testing.T) {
	e := test.MinEvent()
	e.Context = e.Context.AsV1()

	ext := extensions.DistributedTracingExtension{
		TraceParent: "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",
		TraceState:  "rojo=00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01,congo=lZWRzIHRoNhcm5hbCBwbGVhc3VyZS4=",
	}
	want := e.Clone()
	ext.AddTracingAttributes(&want)

	bindingtest.RunTransformerTests(t, context.TODO(), []bindingtest.TransformerTestArgs{
		{
			Name:         "Write to Mock Structured message",
			InputMessage: bindingtest.MustCreateMockStructuredMessage(t, e),
			WantEvent:    want,
			Transformers: binding.Transformers{ext.WriteTransformer()},
		},
		{
			Name:         "Write to Mock Binary message",
			InputMessage: bindingtest.MustCreateMockBinaryMessage(e),
			WantEvent:    want,
			Transformers: binding.Transformers{ext.WriteTransformer()},
		},
		{
			Name:         "Write to Event message",
			InputEvent:   e,
			WantEvent:    want,
			Transformers: binding.Transformers{ext.WriteTransformer()},
		},
	})
}
