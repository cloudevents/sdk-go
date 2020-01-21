package main

import (
	"bytes"
	"io"
	"io/ioutil"
	nethttp "net/http"
	"net/http/httptest"
	"net/url"

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/bindings/http"
	cehttp "github.com/cloudevents/sdk-go/pkg/cloudevents/transport/http"
)

type RoundTripFunc func(req *nethttp.Request) *nethttp.Response

func (f RoundTripFunc) RoundTrip(req *nethttp.Request) (*nethttp.Response, error) {
	return f(req), nil
}

func NewTestClient(fn RoundTripFunc) *nethttp.Client {
	return &nethttp.Client{
		Transport: RoundTripFunc(fn),
	}
}

func MockedSender(options ...http.SenderOptionFunc) binding.Sender {
	u, _ := url.Parse("http://localhost")
	return http.NewSender(NewTestClient(func(req *nethttp.Request) *nethttp.Response {
		return &nethttp.Response{
			StatusCode: 202,
			Header:     make(nethttp.Header),
		}
	}), u, options...)
}

func MockedClient() (cloudevents.Client, *cehttp.Transport) {
	t, err := cehttp.New(cehttp.WithTarget("http://localhost"))

	if err != nil {
		panic(err)
	}

	t.Client = NewTestClient(func(req *nethttp.Request) *nethttp.Response {
		return &nethttp.Response{
			StatusCode: 202,
			Header:     make(nethttp.Header),
			Body:       ioutil.NopCloser(bytes.NewReader([]byte{})),
		}
	})

	client, err := cloudevents.NewClient(t)

	if err != nil {
		panic(err)
	}

	return client, t
}

func MockedBinaryRequest(body []byte) *nethttp.Request {
	r := httptest.NewRequest("POST", "http://localhost:8080", bytes.NewBuffer(body))
	r.Header.Add("Ce-id", "0")
	r.Header.Add("Ce-subject", "sub")
	r.Header.Add("Ce-specversion", "1.0")
	r.Header.Add("Ce-type", "t")
	r.Header.Add("Ce-source", "http://localhost")
	r.Header.Add("Content-type", "text/plain")
	return r
}

var (
	eventBegin = []byte("{" +
		"\"id\":0," +
		"\"subject\":\"sub\"," +
		"\"specversion\":\"1.0\"," +
		"\"type\":\"t\"," +
		"\"source\":\"http://localhost\"," +
		"\"datacontenttype\":\"text/plain\"," +
		"\"data\": \"")
	eventEnd = []byte("\"}")
)

func MockedStructuredRequest(body []byte) *nethttp.Request {
	r := httptest.NewRequest(
		"POST",
		"http://localhost:8080",
		io.MultiReader(bytes.NewReader(eventBegin), bytes.NewBuffer(body), bytes.NewReader(eventEnd)),
	)
	r.Header.Add("Content-type", cloudevents.ApplicationCloudEventsJSON)
	return r
}
