package expression

import (
	cesql "github.com/cloudevents/sdk-go/sql/v2"
	"github.com/cloudevents/sdk-go/sql/v2/runtime"
	cloudevents "github.com/cloudevents/sdk-go/v2"
)

type logicExpression struct {
	baseBinaryExpression
	fn func(x, y bool) bool
}

func (s logicExpression) Evaluate(event cloudevents.Event) (interface{}, error) {
	leftVal, err := s.left.Evaluate(event)
	if err != nil {
		return nil, err
	}

	rightVal, err := s.right.Evaluate(event)
	if err != nil {
		return nil, err
	}

	leftVal, err = runtime.Cast(leftVal, runtime.BooleanType)
	if err != nil {
		return nil, err
	}

	rightVal, err = runtime.Cast(rightVal, runtime.BooleanType)
	if err != nil {
		return nil, err
	}

	return s.fn(leftVal.(bool), rightVal.(bool)), nil
}

func NewAndExpression(left cesql.Expression, right cesql.Expression) cesql.Expression {
	return logicExpression{
		baseBinaryExpression: baseBinaryExpression{
			left:  left,
			right: right,
		},
		fn: func(x, y bool) bool {
			return x && y
		},
	}
}

func NewOrExpression(left cesql.Expression, right cesql.Expression) cesql.Expression {
	return logicExpression{
		baseBinaryExpression: baseBinaryExpression{
			left:  left,
			right: right,
		},
		fn: func(x, y bool) bool {
			return x || y
		},
	}
}

func NewXorExpression(left cesql.Expression, right cesql.Expression) cesql.Expression {
	return logicExpression{
		baseBinaryExpression: baseBinaryExpression{
			left:  left,
			right: right,
		},
		fn: func(x, y bool) bool {
			return x != y
		},
	}
}
