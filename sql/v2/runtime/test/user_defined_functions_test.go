/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package runtime_test

import (
	"io"
	"os"
	"path"
	"runtime"
	"strings"
	"testing"

	cesql "github.com/cloudevents/sdk-go/sql/v2"
	"github.com/cloudevents/sdk-go/sql/v2/function"
	"github.com/cloudevents/sdk-go/sql/v2/parser"
	ceruntime "github.com/cloudevents/sdk-go/sql/v2/runtime"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/binding/spec"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/cloudevents/sdk-go/v2/test"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

var TCKFileNames = []string{
	"user_defined_functions",
}

var TCKUserDefinedFunctions = []cesql.Function{
	function.NewFunction(
		"HASPREFIX",
		[]cesql.Type{cesql.StringType, cesql.StringType},
		nil,
		func(event cloudevents.Event, i []interface{}) (interface{}, error) {
			str := i[0].(string)
			prefix := i[1].(string)

			return strings.HasPrefix(str, prefix), nil
		},
	),
	function.NewFunction(
		"KONKAT",
		[]cesql.Type{},
		cesql.TypePtr(cesql.StringType),
		func(event cloudevents.Event, i []interface{}) (interface{}, error) {
			var sb strings.Builder
			for _, v := range i {
				sb.WriteString(v.(string))
			}
			return sb.String(), nil
		},
	),
}

type ErrorType string

const (
	ParseError              ErrorType = "parse"
	MathError               ErrorType = "math"
	CastError               ErrorType = "cast"
	MissingAttributeError   ErrorType = "missingAttribute"
	MissingFunctionError    ErrorType = "missingFunction"
	FunctionEvaluationError ErrorType = "functionEvaluation"
)

type TckFile struct {
	Name  string        `json:"name"`
	Tests []TckTestCase `json:"tests"`
}

type TckTestCase struct {
	Name       string `json:"name"`
	Expression string `json:"expression"`

	Result interface{} `json:"result"`
	Error  ErrorType   `json:"error"`

	Event          *cloudevents.Event     `json:"event"`
	EventOverrides map[string]interface{} `json:"eventOverrides"`
}

func (tc TckTestCase) InputEvent(t *testing.T) cloudevents.Event {
	var inputEvent cloudevents.Event
	if tc.Event != nil {
		inputEvent = *tc.Event
	} else {
		inputEvent = test.FullEvent()
	}

	// Make sure the event is v1
	inputEvent.SetSpecVersion(event.CloudEventsVersionV1)

	for k, v := range tc.EventOverrides {
		require.NoError(t, spec.V1.SetAttribute(inputEvent.Context, k, v))
	}

	return inputEvent
}

func (tc TckTestCase) ExpectedResult() interface{} {
	switch tc.Result.(type) {
	case int:
		return int32(tc.Result.(int))
	case float64:
		return int32(tc.Result.(float64))
	case bool:
		return tc.Result.(bool)
	}
	return tc.Result
}

func Test_functionTable_AddFunction(t *testing.T) {

	type args struct {
		functions []cesql.Function
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Add user functions to global table",

			args: args{
				functions: TCKUserDefinedFunctions,
			},
			wantErr: false,
		},
		{
			name: "Fail add user functions to global table",
			args: args{
				functions: TCKUserDefinedFunctions,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, fn := range tt.args.functions {
				if err := ceruntime.AddFunction(fn); (err != nil) != tt.wantErr {
					t.Errorf("functionTable.AddFunction() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
		})
	}
}

func Test_UserFunctions(t *testing.T) {
	tckFiles := make([]TckFile, 0, len(TCKFileNames))

	_, basePath, _, _ := runtime.Caller(0)
	basePath, _ = path.Split(basePath)

	for _, testFile := range TCKFileNames {
		testFilePath := path.Join(basePath, "tck", testFile+".yaml")

		t.Logf("Loading file %s", testFilePath)
		file, err := os.Open(testFilePath)
		require.NoError(t, err)

		fileBytes, err := io.ReadAll(file)
		require.NoError(t, err)

		tckFileModel := TckFile{}
		require.NoError(t, yaml.Unmarshal(fileBytes, &tckFileModel))

		tckFiles = append(tckFiles, tckFileModel)
	}

	for i, file := range tckFiles {
		i := i
		t.Run(file.Name, func(t *testing.T) {
			for j, testCase := range tckFiles[i].Tests {
				j := j
				testCase := testCase
				t.Run(testCase.Name, func(t *testing.T) {
					t.Parallel()
					testCase := tckFiles[i].Tests[j]

					t.Logf("Test expression: '%s'", testCase.Expression)

					if testCase.Error == ParseError {
						_, err := parser.Parse(testCase.Expression)
						require.NotNil(t, err)
						return
					}

					expr, err := parser.Parse(testCase.Expression)
					require.NoError(t, err)
					require.NotNil(t, expr)

					inputEvent := testCase.InputEvent(t)
					result, err := expr.Evaluate(inputEvent)

					if testCase.Error != "" {
						require.NotNil(t, err)
					} else {
						require.NoError(t, err)
						require.Equal(t, testCase.ExpectedResult(), result)
					}
				})
			}
		})
	}
}
