/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package http

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/require"

	cloudevents "github.com/cloudevents/sdk-go/v2"
)

var (
	// Headers that are added to the response, but we don't want to check in our assertions.
	unimportantHeaders = []string{
		"accept-encoding",
		"content-length",
		"user-agent",
		"connection",
		"test-ce-id",
		"traceparent",
		"tracestate",
	}
)

func toBytes(body map[string]interface{}) []byte {
	b, err := json.Marshal(body)
	if err != nil {
		return []byte(fmt.Sprintf(`{"error":%q}`, err.Error()))
	}
	return b
}

func assertEventEqualityExact(t *testing.T, ctx string, expected, actual *cloudevents.Event) {
	if diff := cmp.Diff(expected, actual, cmpopts.IgnoreFields(cloudevents.Event{}, "DataEncoded", "DataBase64")); diff != "" {
		t.Errorf("Unexpected difference in %s (-want, +got): %v", ctx, diff)
	}
	if expected == nil || actual == nil {
		return
	}
	if diff := cmp.Diff(expected.Data(), actual.Data()); diff != "" {
		t.Errorf("Unexpected data difference in %s (-want, +got): %v", ctx, diff)
	}
}

func assertEventEquality(t *testing.T, ctx string, expected, actual *cloudevents.Event) {
	if diff := cmp.Diff(expected, actual, cmpopts.IgnoreFields(cloudevents.Event{}, "DataEncoded", "DataBase64")); diff != "" {
		t.Errorf("Unexpected difference in %s (-want, +got): %v", ctx, diff)
	}
	if expected == nil || actual == nil {
		return
	}
	if diff := cmp.Diff(expected.Data(), actual.Data()); diff != "" {
		t.Errorf("Unexpected data difference in %s (-want, +got): %v", ctx, diff)
	}
}

func assertTappedEquality(t *testing.T, ctx string, expected, actual *TapValidation) {
	canonicalizeHeaders(expected, actual)
	if diff := cmp.Diff(expected, actual, cmpopts.IgnoreFields(TapValidation{}, "ContentLength", "Body", "BodyContains")); diff != "" {
		t.Errorf("Unexpected difference in %s (-want, +got): %v", ctx, diff)
	}

	if expected.Body != "" {
		require.Equal(t, expected.Body, actual.Body)
	}
	if expected.BodyContains != nil {
		for _, bc := range expected.BodyContains {
			require.Contains(t, actual.Body, bc)
		}
	}

}

func canonicalizeHeaders(rvs ...*TapValidation) {
	// HTTP header names are case-insensitive, so normalize them to lower case for comparison.
	for _, rv := range rvs {
		if rv == nil || rv.Header == nil {
			continue
		}
		header := rv.Header
		for n, v := range header {
			delete(header, n)
			ln := strings.ToLower(n)

			if isImportantHeader(ln) {
				header[ln] = v
			}
		}
	}
}

func isImportantHeader(h string) bool {
	for _, v := range unimportantHeaders {
		if v == h {
			return false
		}
	}
	return true
}

func strptr(s string) *string {
	return &s
}
