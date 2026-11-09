/*
 Copyright 2026 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package event_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cloudevents/sdk-go/v2/event"
)

func TestUnmarshalNumericExtension(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		want    interface{}
		wantErr bool
	}{
		{name: "positive fraction", value: "42.42", wantErr: true},
		{name: "negative fraction", value: "-42.42", wantErr: true},
		{name: "fraction below one", value: "0.5", wantErr: true},
		{name: "fraction in exponent notation", value: "5e-1", wantErr: true},
		{name: "zero", value: "0", want: int32(0)},
		{name: "positive integer", value: "42", want: int32(42)},
		{name: "negative integer", value: "-42", want: int32(-42)},
		{name: "maximum integer", value: "2147483647", want: int32(2147483647)},
		{name: "minimum integer", value: "-2147483648", want: int32(-2147483648)},
		{name: "string", value: `"42.42"`, want: "42.42"},
		{name: "boolean", value: "true", want: true},
	}
	for _, version := range []string{event.CloudEventsVersionV03, event.CloudEventsVersionV1} {
		for _, extensionFirst := range []bool{false, true} {
			for _, tt := range tests {
				t.Run(fmt.Sprintf("%s/extensionFirst=%t/%s", version, extensionFirst, tt.name), func(t *testing.T) {
					attributes := fmt.Sprintf(`"specversion":%q,"id":"1","type":"test.type","source":"http://localhost/source"`, version)
					extension := `"number":` + tt.value
					body := "{" + attributes + "," + extension + "}"
					if extensionFirst {
						body = "{" + extension + "," + attributes + "}"
					}

					var e event.Event
					err := json.Unmarshal([]byte(body), &e)
					if tt.wantErr {
						require.Error(t, err)
						var validationErr event.ValidationError
						require.True(t, errors.As(err, &validationErr), "expected ValidationError, got %T: %v", err, err)
						require.Contains(t, validationErr, "number")
						return
					}
					require.NoError(t, err)
					require.Equal(t, tt.want, e.Extensions()["number"])
				})
			}
		}
	}
}

func TestUnmarshalFractionalJSONData(t *testing.T) {
	for _, version := range []string{event.CloudEventsVersionV03, event.CloudEventsVersionV1} {
		t.Run(version, func(t *testing.T) {
			body := fmt.Sprintf(`{"specversion":%q,"id":"1","type":"test.type","source":"http://localhost/source","datacontenttype":"application/json","data":{"number":42.42}}`, version)
			var e event.Event
			require.NoError(t, json.Unmarshal([]byte(body), &e))
			var data map[string]float64
			require.NoError(t, e.DataAs(&data))
			require.Equal(t, 42.42, data["number"])
		})
	}
}
