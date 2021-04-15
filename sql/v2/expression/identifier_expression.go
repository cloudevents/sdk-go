package expression

import (
	"fmt"

	cesql "github.com/cloudevents/sdk-go/sql/v2"
	"github.com/cloudevents/sdk-go/sql/v2/runtime"
	cloudevents "github.com/cloudevents/sdk-go/v2"
)

type identifierExpression struct {
	identifier string
}

func (l identifierExpression) Evaluate(event cloudevents.Event) (interface{}, error) {
	value := runtime.GetAttribute(event, l.identifier)
	if value == nil {
		return nil, fmt.Errorf("missing attribute '%s'", l.identifier)
	}

	return value, nil
}

func NewIdentifierExpression(identifier string) cesql.Expression {
	return identifierExpression{identifier: identifier}
}
