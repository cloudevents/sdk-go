package main

import (
	"bytes"
	"context"
	"encoding/csv"
	"flag"
	"io"
	"io/ioutil"
	nethttp "net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"runtime"
	"runtime/pprof"
	"strconv"
	"testing"

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/bindings/http"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport"
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

func generateRandomValue(kb int, value byte) []byte {
	length := 1024 * kb
	b := make([]byte, length)
	for i := 0; i < length; i++ {
		b[i] = value
	}
	return b
}

func MockedSender() *http.Sender {
	u, _ := url.Parse("http://localhost")
	return http.NewSender(NewTestClient(func(req *nethttp.Request) *nethttp.Response {
		return &nethttp.Response{
			StatusCode: 202,
			Header:     make(nethttp.Header),
		}
	}), u)
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

	t.SetReceiver(transport.ReceiveFunc(func(ctx context.Context, e cloudevents.Event, er *cloudevents.EventResponse) error {
		_, _, _ = client.Send(ctx, e)
		er.RespondWith(202, nil)
		return nil
	}))

	return client, t
}

func MockedRequest(body []byte) *nethttp.Request {
	r := httptest.NewRequest("POST", "http://localhost:8080", bytes.NewBuffer(body))
	r.Header.Add("Ce-id", "0")
	r.Header.Add("Ce-subject", "sub")
	r.Header.Add("Ce-specversion", "1.0")
	r.Header.Add("Ce-type", "t")
	r.Header.Add("Ce-source", "http://localhost")
	r.Header.Add("Content-type", "text/plain")
	return r
}

// Avoid DCE
var W *httptest.ResponseRecorder
var R *nethttp.Request

type BenchResult struct {
	parallelism   int
	payloadSizeKb int
	testing.BenchmarkResult
}

func runBench(do func(body []byte)) []BenchResult {
	results := make([]BenchResult, 0)
	for p := 1; p <= runtime.NumCPU(); p++ {
		for k := 1; k <= 32; k *= 2 {
			body := generateRandomValue(k, byte('a'))
			r := testing.Benchmark(func(b *testing.B) {
				b.SetParallelism(p)
				b.RunParallel(func(pb *testing.PB) {
					for pb.Next() {
						do(body)
					}
				})
			})
			results = append(results, BenchResult{p, k, r})
		}
	}
	return results
}

func benchmarkBaseline() []BenchResult {
	return runBench(func(body []byte) {
		W = httptest.NewRecorder()
		R = MockedRequest(body)
	})
}

func benchmarkReceiverSender() []BenchResult {
	r := http.NewReceiver()

	results := make([]BenchResult, 0)
	for p := 1; p <= runtime.NumCPU(); p++ {
		ctx, cancel := context.WithCancel(context.TODO())

		// Spawn dispatchers
		for i := 0; i < p; i++ {
			go func(r *http.Receiver) {
				s := MockedSender()
				var err error
				var m binding.Message
				messageCtx := context.Background()
				for err != io.EOF {
					select {
					case _, ok := <-ctx.Done():
						if !ok {
							return
						}
					default:
						m, err = r.Receive(messageCtx)
						if err != nil {
							continue
						}
						_ = s.Send(messageCtx, m)
					}
				}
			}(r)
		}

		for k := 1; k <= 32; k *= 2 {
			body := generateRandomValue(k, byte('a'))
			r := testing.Benchmark(func(b *testing.B) {
				b.SetParallelism(p)
				b.RunParallel(func(pb *testing.PB) {
					for pb.Next() {
						w := httptest.NewRecorder()
						r.ServeHTTP(w, MockedRequest(body))
					}
				})
			})
			results = append(results, BenchResult{p, k, r})
		}

		cancel()
	}
	return results
}

func benchmarkClient() []BenchResult {
	_, mockedTransport := MockedClient()

	return runBench(func(body []byte) {
		w := httptest.NewRecorder()
		mockedTransport.ServeHTTP(w, MockedRequest(body))
	})
}

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
var memprofile = flag.String("memprofile", "", "write memory profile to `file`")
var bench = flag.String("bench", "baseline", "[baseline, receiver-sender, client]")

func main() {
	flag.Parse()

	if *cpuprofile != "" {
		f, _ := os.Create(*cpuprofile)
		_ = pprof.StartCPUProfile(f)
		defer f.Close()
		defer pprof.StopCPUProfile()
	}

	var results []BenchResult

	switch *bench {
	case "baseline":
		results = benchmarkBaseline()
		break
	case "receiver-sender":
		results = benchmarkReceiverSender()
		break
	case "client":
		results = benchmarkClient()
		break
	default:
		panic("Wrong bench flag")
	}

	pprof.StopCPUProfile()

	if *memprofile != "" {
		f, _ := os.Create(*memprofile)
		defer f.Close()
		runtime.GC() // get up-to-date statistics
		_ = pprof.WriteHeapProfile(f)
	}

	writer := csv.NewWriter(os.Stdout)
	defer writer.Flush()

	for _, res := range results {
		_ = writer.Write([]string{
			strconv.Itoa(res.parallelism),
			strconv.Itoa(res.payloadSizeKb),
			strconv.FormatInt(res.NsPerOp(), 10),
			strconv.FormatInt(res.AllocedBytesPerOp(), 10),
		})
	}

}
