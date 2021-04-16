package expression

import cesql "github.com/cloudevents/sdk-go/sql/v2"

type baseUnaryExpression struct {
	child cesql.Expression
}

type baseBinaryExpression struct {
	left  cesql.Expression
	right cesql.Expression
}
