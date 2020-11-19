package client_test

import (
	"context"
	"errors"
	cehttp "github.com/cloudevents/sdk-go/v2/protocol/http"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/client"
	"github.com/stretchr/testify/require"
)

func TestEventReceiverServeHTTP_WithContext(t *testing.T) {
	type ctxKey string
	const ctxKeyTest ctxKey = "testKey"

	middleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			ctx = context.WithValue(ctx, ctxKeyTest, "testValue")
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}

	eventReceiver := func(ctx context.Context) error {
		v, ok := ctx.Value(ctxKeyTest).(string)
		if !ok {
			t.Errorf("invalid context value type: %v", v)
			return errors.New("invalid context")
		}
		if v != "testValue" {
			t.Errorf("invalid context value: %s", v)
			return errors.New("invalid context")
		}
		return nil
	}

	p, err := cloudevents.NewHTTP()
	if err != nil {
		t.Fatal(err)
	}
	httpHandler, err := client.NewHTTPReceiveHandler(context.Background(), p, eventReceiver)
	if err != nil {
		t.Fatal(err)
	}
	c, err := cloudevents.NewDefaultClient()
	if err != nil {
		t.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.Handle("/test", middleware(httpHandler))
	ts := httptest.NewServer(mux)
	defer ts.Close()

	event := cloudevents.NewEvent()
	event.SetSource("testSource")
	event.SetType("testType")
	ctx := context.Background()
	ctx = cloudevents.ContextWithTarget(ctx, ts.URL+"/test")

	result := c.Send(ctx, event)
	require.True(t, cloudevents.IsACK(result))
}

func TestEventReceiverServeHTTP_Options(t *testing.T) {
	p, err := cloudevents.NewHTTP()
	if err != nil {
		t.Fatal(err)
	}
	httpHandler, err := client.NewHTTPReceiveHandler(context.Background(), p, func() {})
	if err != nil {
		t.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.Handle("/test", httpHandler)
	ts := httptest.NewServer(mux)
	defer ts.Close()

	req, err := http.NewRequest(http.MethodOptions, ts.URL+"/test", nil)
	require.NoError(t, err)
	res, err := ts.Client().Do(req)
	t.Logf("foo")
	require.NoError(t, err)
	require.NotNil(t, res)
}

func TestEventReceiverServeHTTP_Webhook(t *testing.T) {
	p, err := cloudevents.NewHTTP(cloudevents.WithDefaultOptionsHandlerFunc([]string{http.MethodPost}, cehttp.DefaultAllowedRate, []string{"*"}, false))
	if err != nil {
		t.Fatal(err)
	}
	p.OptionsHandlerFn = p.OptionsHandler
	httpHandler, err := client.NewHTTPReceiveHandler(context.Background(), p, func() {})
	if err != nil {
		t.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.Handle("/test", httpHandler)
	ts := httptest.NewServer(mux)
	defer ts.Close()

	req, err := http.NewRequest(http.MethodOptions, ts.URL+"/test", nil)
	require.NoError(t, err)
	res, err := ts.Client().Do(req)
	t.Logf("foo")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.StatusCode)
	require.Equal(t, strconv.Itoa(cehttp.DefaultAllowedRate), res.Header.Get("WebHook-Allowed-Rate"))
}
