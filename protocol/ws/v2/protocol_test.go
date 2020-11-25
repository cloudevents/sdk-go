package v2

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
	"nhooyr.io/websocket"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/client"
	. "github.com/cloudevents/sdk-go/v2/test"
)

func pingEvent() cloudevents.Event {
	event := cloudevents.NewEvent()
	event.SetID("1")
	event.SetType("ping")
	event.SetSource("localhost")
	return event
}

func TestClientProtocolPingPong(t *testing.T) {
	server := httptest.NewServer(pingPongHandler(t))
	defer server.Close()

	p, err := Dial(context.TODO(), server.URL, nil)
	require.NoError(t, err)

	ping := pingEvent()
	require.NoError(t, p.Send(context.TODO(), binding.ToMessage(&ping)))

	receivedMessage, err := p.Receive(context.TODO())
	require.NoError(t, err)

	pong, err := binding.ToEvent(context.TODO(), receivedMessage)
	require.NoError(t, err)

	AssertEvent(t, *pong, HasId("2"), HasType("pong"))

	require.NoError(t, p.Close(context.TODO()))
}

func TestClientProtocolPingPongWithClient(t *testing.T) {
	server := httptest.NewServer(pingPongHandler(t))
	defer server.Close()

	p, err := Dial(context.TODO(), server.URL, nil)
	require.NoError(t, err)

	c, err := cloudevents.NewClient(p, client.WithPollGoroutines(1))
	require.NoError(t, err)

	ping := pingEvent()
	require.NoError(t, c.Send(context.TODO(), ping))

	ctx, cancelFn := context.WithCancel(context.TODO())
	var received atomic.Value
	// Start receiver closes the connection when stopped!
	err = c.StartReceiver(ctx, func(event cloudevents.Event) {
		received.Store(event)
		cancelFn()
	})
	require.NoError(t, err)

	pong, ok := received.Load().(cloudevents.Event)
	require.True(t, ok)

	AssertEvent(t, pong, HasId("2"), HasType("pong"))
}

func TestClientServerProtocolPingPong(t *testing.T) {
	server := httptest.NewServer(pingPongProtocolHandler(t))
	defer server.Close()

	p, err := Dial(context.TODO(), server.URL, nil)
	require.NoError(t, err)

	ping := pingEvent()
	require.NoError(t, p.Send(context.TODO(), binding.ToMessage(&ping)))

	receivedMessage, err := p.Receive(context.TODO())
	require.NoError(t, err)

	pong, err := binding.ToEvent(context.TODO(), receivedMessage)
	require.NoError(t, err)

	AssertEvent(t, *pong, HasId("2"), HasType("pong"))

	require.NoError(t, p.Close(context.TODO()))
}

func pingPongProtocolHandler(t *testing.T) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		ctx := request.Context()
		p, err := Accept(ctx, writer, request, &websocket.AcceptOptions{Subprotocols: SupportedSubprotocols})

		require.NoError(t, err)
		require.Equal(t, JsonSubprotocol, p.conn.Subprotocol())

		m, err := p.Receive(ctx)
		require.NoError(t, err)

		ping, err := binding.ToEvent(ctx, m)
		require.NoError(t, err)
		AssertEvent(t, *ping, HasId("1"), HasType("ping"))

		pong := ping.Clone()
		pong.SetID("2")
		pong.SetType("pong")

		require.NoError(t, p.Send(ctx, binding.ToMessage(&pong)))
		require.NoError(t, p.Close(ctx))
	}
}

func pingPongHandler(t *testing.T) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		c, err := websocket.Accept(writer, request, &websocket.AcceptOptions{Subprotocols: SupportedSubprotocols})

		require.NoError(t, err)
		require.Equal(t, JsonSubprotocol, c.Subprotocol())

		messageType, b, err := c.Read(context.TODO())
		require.NoError(t, err)
		require.Equal(t, websocket.MessageText, messageType)

		var ping cloudevents.Event
		require.NoError(t, json.Unmarshal(b, &ping))
		AssertEvent(t, ping, HasId("1"), HasType("ping"))

		pong := ping.Clone()
		pong.SetID("2")
		pong.SetType("pong")

		require.NoError(t, c.Write(context.TODO(), websocket.MessageText, MustJSON(t, pong)))

		ctx := c.CloseRead(context.TODO())
		<-ctx.Done()
	}
}
