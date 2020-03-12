package buffering_vs_to_event_test

import (
	"context"
	nethttp "net/http"
	"strings"
	"testing"
	"time"

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/binding/buffering"
	"github.com/cloudevents/sdk-go/pkg/binding/test"
	"github.com/cloudevents/sdk-go/pkg/binding/transformer"
	"github.com/cloudevents/sdk-go/pkg/transport/http"
	"github.com/cloudevents/sdk-go/pkg/types"
)

var (
	initialEvent      cloudevents.Event
	binaryHttpRequest *nethttp.Request

	ctx = context.TODO()
)

func init() {
	initialEvent = test.FullEvent()
	initialEvent.SetExtension("key", "aaa")

	binaryHttpRequest, _ = nethttp.NewRequest("POST", "http://localhost", nil)
	Err = http.WriteRequest(context.TODO(), binding.ToEventMessage(&initialEvent), binaryHttpRequest, nil)
	if Err != nil {
		panic(Err)
	}

}

var Req *nethttp.Request

func BenchmarkHttpToEventAndUpdateExtensionsAndToHttp(b *testing.B) {
	initialEvent := test.FullEvent()
	initialEvent.SetExtension("aaa", "bbb")
	for i := 0; i < b.N; i++ {
		M = http.NewMessageFromHttpRequest(binaryHttpRequest)
		E, Err = binding.ToEvent(ctx, M, nil)
		if Err != nil {
			panic(Err)
		}
		if v, ok := E.Extensions()["aaa"]; ok {
			vStr, err := types.Format(v)
			if err != nil {
				panic(err)
			}
			E.SetExtension("aaa", strings.ToUpper(vStr))
		} else {
			E.SetExtension("aaa", strings.ToUpper("AAA"))
		}
		if v, ok := E.Extensions()["aTime"]; ok {
			vTime, err := types.ToTime(v)
			if err != nil {
				panic(err)
			}
			E.SetExtension("aTime", vTime.Add(3*time.Hour))
		} else {
			E.SetExtension("aTime", time.Now().UTC().Round(0))
		}
		Req, Err = nethttp.NewRequest("POST", "http://localhost", nil)
		if Err != nil {
			panic(Err)
		}
		Err = http.WriteRequest(ctx, binding.ToEventMessage(E), Req, nil)
	}
}

func BenchmarkHttpToBufferingAndUpdateExtensionsAndToHttp(b *testing.B) {
	initialEvent := test.FullEvent()
	ctx := context.TODO()
	initialEvent.SetExtension("aaa", "bbb")

	transformers := binding.TransformerFactories{}
	transformers = append(transformers,
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

	for i := 0; i < b.N; i++ {
		M = test.MustCreateMockBinaryMessage(initialEvent)
		M, _ = buffering.CopyMessage(ctx, M, transformers)
		Req, Err = nethttp.NewRequest("POST", "http://localhost", nil)
		if Err != nil {
			panic(Err)
		}
		Err = http.WriteRequest(ctx, M, Req, nil)
	}
}
