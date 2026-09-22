package eventbridge

import (
	"context"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/binding/format"
	"github.com/cloudevents/sdk-go/v2/event"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
)

func WriteMsgInput(
	ctx context.Context,
	m binding.Message,
	msgInput *types.PutEventsRequestEntry,
	transformers ...binding.Transformer,
) error {
	var e *event.Event
	e, err := binding.ToEvent(ctx, m, transformers...)
	if err != nil {
		return err
	}

	data, err := format.JSON.Marshal(e)
	if err != nil {
		return err
	}
	details := string(data)

	msgInput.DetailType = aws.String(e.Type())
	msgInput.Source = aws.String(e.Source())
	msgInput.Time = aws.Time(e.Time())
	msgInput.Detail = aws.String(details)
	if traceparent, ok := e.Extensions()["traceparent"]; ok && traceparent != nil {
		msgInput.TraceHeader = aws.String(traceparent.(string))
	}

	return nil
}
