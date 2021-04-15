package test

import (
	"os"
	"path"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"

	cesql "github.com/cloudevents/sdk-go/sql/v2"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/binding/spec"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/cloudevents/sdk-go/v2/test"
)

var TCKFileNames = []string{
	"binary_math_operators",
	"binary_logical_operators",
	"binary_comparison_operators",
	"case_sensitivity",
	"casting_functions",
	"context_attributes_access",
	"exists_expression",
	"in_expression",
	"integer_builtin_functions",
	"like_expression",
	"literals",
	"negate_operator",
	"not_operator",
	"parse_errors",
	"spec_examples",
	"string_builtin_functions",
	"sub_expression",
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
	Name  string        `yaml:"name"`
	Tests []TckTestCase `yaml:"tests"`
}

type TckTestCase struct {
	Name       string `yaml:"name"`
	Expression string `yaml:"expression"`

	Result interface{} `yaml:"result"`
	Error  ErrorType   `yaml:"error"`

	Event          *cloudevents.Event     `yaml:"event"`
	EventOverrides map[string]interface{} `yaml:"eventOverrides"`
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

func TestTCK(t *testing.T) {
	tckFiles := make([]TckFile, 0, len(TCKFileNames))

	_, basePath, _, _ := runtime.Caller(0)
	basePath, _ = path.Split(basePath)

	for _, testFile := range TCKFileNames {
		testFilePath := path.Join(basePath, "tck", testFile+".yaml")

		t.Logf("Loading file %s", testFilePath)

		file, err := os.Open(testFilePath)
		require.NoError(t, err)

		tckFileModel := TckFile{}
		require.NoError(t, yaml.NewDecoder(file).Decode(&tckFileModel))

		tckFiles = append(tckFiles, tckFileModel)
	}

	for i, file := range tckFiles {
		t.Run(file.Name, func(t *testing.T) {
			for j, testCase := range tckFiles[i].Tests {
				t.Run(testCase.Name, func(t *testing.T) {
					t.Parallel()
					testCase := tckFiles[i].Tests[j]

					if testCase.Error == ParseError {
						_, err := cesql.Parse(testCase.Expression)
						require.NotNil(t, err)
					}

					expr, err := cesql.Parse(testCase.Expression)
					require.NoError(t, err)
					require.NotNil(t, expr)

					result, err := expr.Evaluate(testCase.InputEvent(t))

					if testCase.Error != "" {
						require.NotNil(t, err)
					} else {
						require.NoError(t, err)
						require.Equal(t, testCase.Result, result)
					}
				})
			}
		})
	}
}
