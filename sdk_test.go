package cloudevents_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dispatchframework/cloudevents-go-sdk"

	"github.com/dispatchframework/cloudevents-go-sdk/v01"
)

func TestProduceCloudEventSuccess(t *testing.T) {

	converters := []cloudevents.HttpCloudEventConverter{v01.NewJsonHttpCloudEventConverter(), v01.NewBinaryHttpCloudEventConverter()}
	processor := v01.NewDefaultHttpRequestExtractor(converters)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		processor.Extract(r)
	}))
	defer server.Close()

	http.Post(server.URL, "application/cloudevent+json", nil)
}
