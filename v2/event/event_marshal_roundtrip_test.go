/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package event_test

import (
	"encoding/json"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/cloudevents/sdk-go/v2/test"
	"github.com/cloudevents/sdk-go/v2/types"
)

func TestMarshalRoundtrip(t *testing.T) {
	now := types.Timestamp{Time: time.Now().UTC()}
	sourceUrl, _ := url.Parse("http://example.com/source")
	source := &types.URIRef{URL: *sourceUrl}

	schemaUrl, _ := url.Parse("http://example.com/schema")
	schema := &types.URI{URL: *schemaUrl}

	testCases := map[string]event.Event{
		"struct data v1.0 without data content type": func() event.Event {
			e := event.Event{
				Context: event.EventContextV1{
					Type:       "com.example.test",
					Source:     *source,
					DataSchema: schema,
					ID:         "ABC-123",
					Time:       &now,
				}.AsV1(),
			}
			_ = e.SetData(event.ApplicationJSON, DataExample{
				AnInt:   42,
				AString: "testing",
			})
			// Remove the data content type
			e.SetDataContentType("")
			return e
		}(),
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			expected := tc.Clone()

			marshalled, err := json.Marshal(expected)
			require.NoError(t, err)
			have := event.Event{}
			require.NoError(t, json.Unmarshal(marshalled, &have))
			test.AssertEventEquals(t, expected, have)
		})
	}
}
