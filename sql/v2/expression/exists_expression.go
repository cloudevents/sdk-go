package expression

import (
	cesql "github.com/cloudevents/sdk-go/sql/v2"
	"github.com/cloudevents/sdk-go/sql/v2/runtime"
	cloudevents "github.com/cloudevents/sdk-go/v2"
)

type existsExpression struct {
	identifier string
}

func (l existsExpression) Evaluate(event cloudevents.Event) (interface{}, error) {
	return runtime.ContainsAttribute(event, l.identifier), nil
}

func NewExistsExpression(identifier string) cesql.Expression {
	return existsExpression{identifier: identifier}
}
