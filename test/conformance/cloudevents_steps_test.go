/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package conformance

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/cucumber/godog"
	"github.com/google/go-cmp/cmp"
)

var currentEvent *event.Event

func CloudEventsFeatureContext(ctx *godog.ScenarioContext) {
	ctx.BeforeScenario(func(*godog.Scenario) {
		currentEvent = nil
	})

	ctx.Step(`^the attributes are:$`, func(attributes *godog.Table) error {
		for _, row := range attributes.Rows {
			key := row.Cells[0].Value
			value := row.Cells[1].Value

			var actual string
			switch key {
			case "key":
				// ignore the header
				continue
			case "specversion":
				actual = currentEvent.SpecVersion()
			case "id":
				actual = currentEvent.ID()
			case "type":
				actual = currentEvent.Type()
			case "source":
				actual = currentEvent.Source()
			case "time":
				actual = currentEvent.Time().Format(time.RFC3339)
			case "datacontenttype":
				actual = currentEvent.DataContentType()
			default:
				return fmt.Errorf("Unknown key '%s'", key)
			}

			if diff := cmp.Diff(value, actual); diff != "" {
				return fmt.Errorf("unexpected '%s' (-want, +got) = %v", key, diff)
			}
		}

		return nil
	})

	ctx.Step(`^the data is equal to the following JSON:$`, func(jsonData *godog.DocString) error {
		actualBytes := currentEvent.Data()

		var expectedJSONAsInterface, actualJSONAsInterface interface{}

		if err := json.Unmarshal([]byte(jsonData.Content), &expectedJSONAsInterface); err != nil {
			return fmt.Errorf("Expected value ('%s') is not valid json.\nJSON parsing error: '%s'", jsonData.Content, err.Error())
		}

		if err := json.Unmarshal(actualBytes, &actualJSONAsInterface); err != nil {
			return fmt.Errorf("Input ('%s') needs to be valid json.\nJSON parsing error: '%s'", string(actualBytes), err.Error())
		}

		if diff := cmp.Diff(expectedJSONAsInterface, actualJSONAsInterface); diff != "" {
			return fmt.Errorf("unexpected  (-want, +got) = %v", diff)
		}

		return nil
	})
}
