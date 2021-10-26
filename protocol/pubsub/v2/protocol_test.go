package pubsub

import (
	"context"
	"fmt"
	"testing"

	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/pubsub/pstest"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/option"
	pb "google.golang.org/genproto/googleapis/pubsub/v1"
	"google.golang.org/grpc"

	"github.com/cloudevents/sdk-go/v2/test"
)

type testPubsubClient struct {
	srv  *pstest.Server
	conn *grpc.ClientConn
}

func (pc *testPubsubClient) NewWithOrderInterceptor(ctx context.Context, projectID, orderingKey string) (*pubsub.Client, error) {
	pc.srv = pstest.NewServer()
	conn, err := grpc.Dial(pc.srv.Addr, grpc.WithInsecure(), grpc.WithUnaryInterceptor(orderingKeyInterceptor(orderingKey)))
	if err != nil {
		return nil, err
	}
	pc.conn = conn
	return pubsub.NewClient(ctx, projectID, option.WithGRPCConn(conn))
}

func (pc *testPubsubClient) Close() {
	pc.srv.Close()
	pc.conn.Close()
}

func orderingKeyInterceptor(orderingKey string) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if method == "/google.pubsub.v1.Publisher/Publish" {
			pr, _ := req.(*pb.PublishRequest)
			for _, m := range pr.Messages {
				if m.OrderingKey != orderingKey {
					return fmt.Errorf("expecting ordering key %q, got %q", orderingKey, m.OrderingKey)
				}
			}
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func TestPublishMessageHasOrderingKey(t *testing.T) {
	require := require.New(t)
	ctx := context.Background()
	pc := &testPubsubClient{}
	defer pc.Close()

	projectID, topicID, orderingKey := "test-project", "test-topic", "foobar"

	client, err := pc.NewWithOrderInterceptor(ctx, projectID, orderingKey)
	require.NoError(err, "create pubsub client")
	defer client.Close()

	prot, err := New(ctx,
		WithClient(client),
		WithProjectID(projectID),
		WithTopicID(topicID),
		WithMessageOrdering(),
		AllowCreateTopic(true),
	)
	require.NoError(err, "create protocol")

	err = prot.Send(WithOrderingKey(ctx, orderingKey), test.FullMessage())
	require.NoError(err)
}
