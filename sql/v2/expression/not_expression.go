package expression

import (
	cesql "github.com/cloudevents/sdk-go/sql/v2"
	"github.com/cloudevents/sdk-go/sql/v2/runtime"
	cloudevents "github.com/cloudevents/sdk-go/v2"
)

type notExpression baseUnaryExpression

func (l notExpression) Evaluate(event cloudevents.Event) (interface{}, error) {
	val, err := l.child.Evaluate(event)
	if err != nil {
		return nil, err
	}

	val, err = runtime.Cast(val, runtime.BooleanType)
	if err != nil {
		return nil, err
	}

	return !(val.(bool)), nil
}

func NewNotExpression(child cesql.Expression) cesql.Expression {
	return notExpression{child: child}
}
