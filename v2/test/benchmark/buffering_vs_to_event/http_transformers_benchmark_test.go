package buffering_vs_to_event_test

import (
	"context"
	nethttp "net/http"
	"strings"
	"testing"
	"time"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/binding/buffering"
	"github.com/cloudevents/sdk-go/v2/binding/test"
	"github.com/cloudevents/sdk-go/v2/binding/transformer"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/cloudevents/sdk-go/v2/protocol/http"
	"github.com/cloudevents/sdk-go/v2/types"
)

var (
	binaryHttpRequest       *nethttp.Request
	binaryHttpRequestNoData *nethttp.Request

	transformers binding.TransformerFactories

	ctx = context.TODO()
)

func init() {
	initialEvent := test.FullEvent()
	initialEvent.SetExtension("key", "aaa")

	binaryHttpRequest, _ = nethttp.NewRequest("POST", "http://localhost", nil)
	Err = http.WriteRequest(context.TODO(), binding.ToMessage(&initialEvent), binaryHttpRequest)
	if Err != nil {
		panic(Err)
	}

	initialEventNoData := test.FullEvent()
	initialEventNoData.DataEncoded = nil
	initialEventNoData.SetDataContentType("")
	initialEventNoData.SetExtension("key", "aaa")

	binaryHttpRequestNoData, _ = nethttp.NewRequest("POST", "http://localhost", nil)
	Err = http.WriteRequest(context.TODO(), binding.ToMessage(&initialEventNoData), binaryHttpRequestNoData)
	if Err != nil {
		panic(Err)
	}

	transformers = append(binding.TransformerFactories{},
		transformer.SetExtension("aaa", "AAAA", func(i2 interface{}) (interface{}, error) {
			vStr, err := types.Format(i2)
			if err != nil {
				return nil, err
			}
			return strings.ToUpper(vStr), nil
		})...,
	)
	transformers = append(transformers,
		transformer.SetExtension("aTime", time.Now(), func(i2 interface{}) (interface{}, error) {
			vTime, err := types.ToTime(i2)
			if err != nil {
				return nil, err
			}
			return vTime.Add(3 * time.Hour), nil
		})...,
	)
}

var Req *nethttp.Request

func BenchmarkHttpWithToEvent(b *testing.B) {
	for i := 0; i < b.N; i++ {
		M = http.NewMessageFromHttpRequest(binaryHttpRequest)
		E, Err = binding.ToEvent(ctx, M)
		if Err != nil {
			panic(Err)
		}
		transformEvent(E)
		Req, Err = nethttp.NewRequest("POST", "http://localhost", nil)
		if Err != nil {
			panic(Err)
		}
		Err = http.WriteRequest(ctx, binding.ToMessage(E), Req)
	}
}

func BenchmarkNoDataHttpWithToEvent(b *testing.B) {
	for i := 0; i < b.N; i++ {
		M = http.NewMessageFromHttpRequest(binaryHttpRequestNoData)
		E, Err = binding.ToEvent(ctx, M)
		if Err != nil {
			panic(Err)
		}
		transformEvent(E)
		Req, Err = nethttp.NewRequest("POST", "http://localhost", nil)
		if Err != nil {
			panic(Err)
		}
		Err = http.WriteRequest(ctx, binding.ToMessage(E), Req)
	}
}

func BenchmarkHttpWithBuffering(b *testing.B) {
	for i := 0; i < b.N; i++ {
		M = http.NewMessageFromHttpRequest(binaryHttpRequest)
		M, Err = buffering.CopyMessage(ctx, M, transformers)
		if Err != nil {
			panic(Err)
		}
		Req, Err = nethttp.NewRequest("POST", "http://localhost", nil)
		if Err != nil {
			panic(Err)
		}
		Err = http.WriteRequest(ctx, M, Req)
	}
}

func BenchmarkHttpWithDirect(b *testing.B) {
	for i := 0; i < b.N; i++ {
		M = http.NewMessageFromHttpRequest(binaryHttpRequest)
		Req, Err = nethttp.NewRequest("POST", "http://localhost", nil)
		if Err != nil {
			panic(Err)
		}
		Err = http.WriteRequest(ctx, M, Req, transformers)
	}
}

func BenchmarkNoDataHttpWithBuffering(b *testing.B) {
	for i := 0; i < b.N; i++ {
		M = http.NewMessageFromHttpRequest(binaryHttpRequestNoData)
		M, Err = buffering.CopyMessage(ctx, M, transformers)
		if Err != nil {
			panic(Err)
		}
		Req, Err = nethttp.NewRequest("POST", "http://localhost", nil)
		if Err != nil {
			panic(Err)
		}
		Err = http.WriteRequest(ctx, M, Req)
	}
}

func BenchmarkNoDataHttpWithDirect(b *testing.B) {
	for i := 0; i < b.N; i++ {
		M = http.NewMessageFromHttpRequest(binaryHttpRequestNoData)
		Req, Err = nethttp.NewRequest("POST", "http://localhost", nil)
		if Err != nil {
			panic(Err)
		}
		Err = http.WriteRequest(ctx, M, Req, transformers)
	}
}

func transformEvent(e *event.Event) {
	if v, ok := e.Extensions()["aaa"]; ok {
		vStr, err := types.Format(v)
		if err != nil {
			panic(err)
		}
		e.SetExtension("aaa", strings.ToUpper(vStr))
	} else {
		e.SetExtension("aaa", strings.ToUpper("AAA"))
	}
	if v, ok := e.Extensions()["aTime"]; ok {
		vTime, err := types.ToTime(v)
		if err != nil {
			panic(err)
		}
		e.SetExtension("aTime", vTime.Add(3*time.Hour))
	} else {
		e.SetExtension("aTime", time.Now().UTC().Round(0))
	}
}
