/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package sql

import (
	"testing"

	cesql "github.com/cloudevents/sdk-go/sql/v2/parser"
	cloudevents "github.com/cloudevents/sdk-go/v2"
)

func TestShortCircuitAND(t *testing.T) {

	sql := "(EXISTS revisiontype AND revisiontype=Branch) OR (branch='master')"

	expression, err := cesql.Parse(string(sql))
	evt := cloudevents.NewEvent("1.0")
	evt.SetID("evt-1")
	evt.SetType("what-ever")
	evt.SetSource("/event")
	evt.SetExtension("branch", "master")
	val, err := expression.Evaluate(evt)

	if err != nil {
		t.Errorf("err should be nil: %s", err.Error())
	} else {
		if !val.(bool) {
			t.Errorf("should be true ,but :%s", val)
		}
	}
}

func TestShortCircuitOR(t *testing.T) {

	sql := "(branch='master' OR revisiontype=Branch)"

	expression, err := cesql.Parse(string(sql))
	evt := cloudevents.NewEvent("1.0")
	evt.SetID("evt-1")
	evt.SetType("what-ever")
	evt.SetSource("/event")
	evt.SetExtension("branch", "master")
	val, err := expression.Evaluate(evt)

	if err != nil {
		t.Errorf("err should be nil: %s", err.Error())
	} else {
		if !val.(bool) {
			t.Errorf("should be true ,but :%s", val)
		}
	}
}
