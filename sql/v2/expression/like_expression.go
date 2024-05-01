/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package expression

import (
	cesql "github.com/cloudevents/sdk-go/sql/v2"
	"github.com/cloudevents/sdk-go/sql/v2/utils"
	cloudevents "github.com/cloudevents/sdk-go/v2"
)

type likeExpression struct {
	baseUnaryExpression
	pattern string
}

func (l likeExpression) Evaluate(event cloudevents.Event) (interface{}, error) {
	val, err := l.child.Evaluate(event)
	if err != nil {
		return nil, err
	}

	val, err = utils.Cast(val, cesql.StringType)
	if err != nil {
		return nil, err
	}

	return matchString(val.(string), l.pattern), nil

}

func NewLikeExpression(child cesql.Expression, pattern string) (cesql.Expression, error) {
	return likeExpression{
		baseUnaryExpression: baseUnaryExpression{
			child: child,
		},
		pattern: pattern,
	}, nil
}

func matchString(text, pattern string) bool {
	textLen := len(text)
	patternLen := len(pattern)
	textIdx := 0
	patternIdx := 0
	lastWildcardIdx := -1
	lastMatchIdx := 0

	for textIdx < textLen {
		// handle escaped characters -> pattern needs to increment two places here
		if patternIdx < patternLen-1 && pattern[patternIdx] == '\\' &&
			((pattern[patternIdx+1] == '_' || pattern[patternIdx+1] == '%') &&
				pattern[patternIdx+1] == text[textIdx]) {
			patternIdx += 2
			textIdx += 1
			// handle non escaped characters
		} else if patternIdx < patternLen && (pattern[patternIdx] == '_' || pattern[patternIdx] == text[textIdx]) {
			textIdx += 1
			patternIdx += 1
			// handle wildcard characters
		} else if patternIdx < patternLen && pattern[patternIdx] == '%' {
			lastWildcardIdx = patternIdx
			lastMatchIdx = textIdx
			patternIdx += 1
			// greedy match didn't work, try again from the last known match
		} else if lastWildcardIdx != -1 {
			patternIdx = lastWildcardIdx + 1
			lastMatchIdx += 1
			textIdx = lastMatchIdx
		} else {
			return false
		}
	}

	// consume remaining pattern characters as long as they are wildcards
	for patternIdx < patternLen {
		if pattern[patternIdx] != '%' {
			return false
		}

		patternIdx += 1
	}

	return true
}
