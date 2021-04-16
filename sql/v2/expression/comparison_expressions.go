package expression

import (
	cesql "github.com/cloudevents/sdk-go/sql/v2"
	"github.com/cloudevents/sdk-go/sql/v2/runtime"
	cloudevents "github.com/cloudevents/sdk-go/v2"
)

type equalExpression struct {
	baseBinaryExpression
	equal bool
}

func (s equalExpression) Evaluate(event cloudevents.Event) (interface{}, error) {
	leftVal, err := s.left.Evaluate(event)
	if err != nil {
		return nil, err
	}

	rightVal, err := s.right.Evaluate(event)
	if err != nil {
		return nil, err
	}

	leftVal, err = runtime.Cast(leftVal, runtime.TypeFromVal(rightVal))
	if err != nil {
		return nil, err
	}

	return (leftVal == rightVal) == s.equal, nil
}

func NewEqualExpression(left cesql.Expression, right cesql.Expression) cesql.Expression {
	return equalExpression{
		baseBinaryExpression: baseBinaryExpression{
			left:  left,
			right: right,
		},
		equal: true,
	}
}

func NewNotEqualExpression(left cesql.Expression, right cesql.Expression) cesql.Expression {
	return equalExpression{
		baseBinaryExpression: baseBinaryExpression{
			left:  left,
			right: right,
		},
		equal: false,
	}
}
