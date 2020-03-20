package main

import (
	"bytes"
	"io"
	"io/ioutil"
	nethttp "net/http"
	"net/http/httptest"
	"net/url"

	bindings "github.com/cloudevents/sdk-go/v2/protocol"
	http "github.com/cloudevents/sdk-go/v2/protocol/http"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	cehttp "github.com/cloudevents/sdk-go/v2/protocol/http"
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

func MockedSender(options ...http.Option) bindings.Sender {
	u, _ := url.Parse("http://localhost")

	p, err := http.New(options...)
	if err != nil {
		panic(err)
	}
	p.Client = NewTestClient(func(req *nethttp.Request) *nethttp.Response {
		return &nethttp.Response{
			StatusCode: 202,
			Header:     make(nethttp.Header),
		}
	})
	p.Target = u
	return p
}

func MockedClient() (cloudevents.Client, *cehttp.Protocol) {
	mockTransport := RoundTripFunc(func(req *nethttp.Request) *nethttp.Response {
		return &nethttp.Response{
			StatusCode: 202,
			Header:     make(nethttp.Header),
			Body:       ioutil.NopCloser(bytes.NewReader([]byte{})),
		}
	})

	p, err := cehttp.New(cehttp.WithTarget("http://localhost"), cehttp.WithRoundTripper(mockTransport))
	if err != nil {
		panic(err)
	}

	client, err := cloudevents.NewClient(p)
	if err != nil {
		panic(err)
	}

	return client, p
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
		"\"id\":\"0\"," +
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
