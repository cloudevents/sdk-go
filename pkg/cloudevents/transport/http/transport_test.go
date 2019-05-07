package http_test

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	cehttp "github.com/cloudevents/sdk-go/pkg/cloudevents/transport/http"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/types"
	"golang.org/x/sync/errgroup"
)

type task func() error

// We can't use net/http/httptest.Server here because it's connection
// tracking logic interferes with the connection lifecycle under test
func startTestServer(handler http.Handler) (*http.Server, error) {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return nil, err
	}
	server := &http.Server{
		Addr:    listener.Addr().String(),
		Handler: handler,
	}
	go server.Serve(listener)
	return server, nil
}

func doConcurrently(concurrency int, duration time.Duration, fn task) error {
	var group errgroup.Group
	for i := 0; i < concurrency; i++ {
		group.Go(func() error {
			done := time.After(duration)
			for {
				select {
				case <-done:
					return nil
				default:
					if err := fn(); err != nil {
						return err
					}
				}
			}
		})
	}

	if err := group.Wait(); err != nil {
		return err
	}
	return nil
}

// An example of how to make a stable client under sustained
// concurrency sending to a single host
func makeStableClient(addr string) (*cehttp.Transport, error) {
	ceClient, err := cehttp.New(cehttp.WithTarget(addr))
	if err != nil {
		return nil, err
	}
	netHTTPTransport := &http.Transport{
		MaxIdleConnsPerHost: 1000,
		MaxConnsPerHost:     5000,
	}
	netHTTPClient := &http.Client{
		Transport: netHTTPTransport,
	}
	ceClient.Client = netHTTPClient
	return ceClient, nil
}

func TestStableConnectionsToSingleHost(t *testing.T) {
	// Start a dummy HTTP server
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
		fmt.Fprintf(w, `{"success": true}`)
	})
	sinkServer, err := startTestServer(handler)
	if err != nil {
		t.Fatalf("unexpected error starting test http server %v", err.Error())
	}
	defer sinkServer.Close()

	// Keep track of all new connections to that dummy HTTP server
	var newConnectionCount uint64
	sinkServer.ConnState = func(connection net.Conn, state http.ConnState) {
		if state == http.StateNew {
			atomic.AddUint64(&newConnectionCount, 1)
		}
	}

	ceClient, err := makeStableClient("http://" + sinkServer.Addr)
	if err != nil {
		t.Fatalf("unexpected error creating CloudEvents client %v", err.Error())
	}
	event := cloudevents.Event{
		Context: &cloudevents.EventContextV02{
			SpecVersion: cloudevents.CloudEventsVersionV02,
			Type:        "test.event",
			Source:      *types.ParseURLRef("test"),
		},
	}

	ctx := context.TODO()
	concurrency := 64
	duration := 1 * time.Second
	var sent uint64
	err = doConcurrently(concurrency, duration, func() error {
		_, err := ceClient.Send(ctx, event)
		if err != nil {
			return fmt.Errorf("unexpected error sending CloudEvent %v", err.Error())
		}
		atomic.AddUint64(&sent, 1)
		return nil
	})
	if err != nil {
		t.Errorf("error sending concurrent CloudEvents: %v", err)
	}

	// newConnectionCount usually equals concurrency, but give some
	// leeway. When this fails, it fails by a lot
	if newConnectionCount > uint64(concurrency*2) {
		t.Errorf("too many new connections opened: expected %d, got %d", concurrency, newConnectionCount)
	}
	t.Log("sent ", sent)
}

func TestMiddleware(t *testing.T) {
	testCases := map[string]struct {
		middleware []string
		want string
	} {
		"none": {},
		"one": {
			middleware: []string{ "Foo" },
		},
		"nested": {
			middleware: []string{ "Foo", "Bar", "Qux" },
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			m := make([]cehttp.Option, 0, len(tc.middleware) + 2)
			m = append(m, cehttp.WithPort(0), cehttp.WithShutdownTimeout(time.Nanosecond))
			for _, ms := range tc.middleware {
				ms := ms
				m = append(m, cehttp.WithMiddleware(func(next http.Handler) http.Handler {
					return &namedHandler{
						name: ms,
						next: next,
					}
				}))
			}
			tr, err := cehttp.New(m...)
			if err != nil {
				t.Fatalf("Unable to create transport, %v", err)
			}
			innermostResponse := "Original"
			origResponse := makeRequestToServer(t, tr, innermostResponse)

			// Verify that the response is all the middlewares run in the correct order (as a stack).
			response := string(origResponse)
			for i := len(tc.middleware) - 1; i >= 0; i-- {
				expected := tc.middleware[i]
				if !strings.HasPrefix(response, expected) {
					t.Fatalf("Incorrect prefix at offset %d. Expected %s. Actual %s", i, tc.middleware[i], string(origResponse))
				}
				response = strings.TrimPrefix(response, expected)
			}
			if response != innermostResponse {
				t.Fatalf("Incorrect prefix at last offset. Expected '%s'. Actual %s", innermostResponse, string(origResponse))
			}
		})
	}
}

// makeRequestToServer starts the transport and makes a request to it, pointing at a custom path that will return
// responseText.
func makeRequestToServer(t *testing.T, tr *cehttp.Transport, responseText string) string {
	// Create a custom path that will be used to respond with responseText.
	tr.Handler = http.NewServeMux()
	path := "/123"
	tr.Handler.HandleFunc(path, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(responseText))
	})

	// Start the server.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go tr.StartReceiver(ctx)

	// Give some time for the receiver to start. One second was chosen arbitrarily.
	time.Sleep(time.Second)

	// Make the request.
	port := tr.GetPort()
	r, err := http.Post(fmt.Sprintf("http://localhost:%d%s", port, path), "text", &bytes.Buffer{})
	if err != nil {
		t.Fatalf("Error posting: %v", err)
	}
	rb, err := ioutil.ReadAll(r.Body)
	if err != nil {
		t.Fatalf("Error reading: %v", err)
	}
	return string(rb)
}

type namedHandler struct {
	name string
	next http.Handler
}

func (h *namedHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte(h.name))
	h.next.ServeHTTP(w, r)
}
