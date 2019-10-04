package http

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

var (
	// Headers that are added to the response, but we don't want to check in our assertions.
	unimportantHeaders = []string{
		"accept-encoding",
		"content-length",
		"user-agent",
		"connection",
		"test-ce-id",
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
	if diff := cmp.Diff(expected, actual, cmpopts.IgnoreFields(cloudevents.Event{}, "Data", "DataEncoded", "DataBinary")); diff != "" {
		t.Errorf("Unexpected difference in %s (-want, +got): %v", ctx, diff)
	}
	if expected == nil || actual == nil {
		return
	}
	if diff := cmp.Diff(expected.Data, actual.Data); diff != "" {
		t.Errorf("Unexpected data difference in %s (-want, +got): %v", ctx, diff)
	}
}

func assertEventEquality(t *testing.T, ctx string, expected, actual *cloudevents.Event) {
	if diff := cmp.Diff(expected, actual, cmpopts.IgnoreFields(cloudevents.Event{}, "Data", "DataEncoded", "DataBinary")); diff != "" {
		t.Errorf("Unexpected difference in %s (-want, +got): %v", ctx, diff)
	}
	if expected == nil || actual == nil {
		return
	}
	data := make(map[string]string)
	err := actual.DataAs(&data)
	if err != nil {
		t.Error(err)
	}
	if diff := cmp.Diff(expected.Data, data); diff != "" {
		t.Errorf("Unexpected data difference in %s (-want, +got): %v", ctx, diff)
	}
}

func assertTappedEquality(t *testing.T, ctx string, expected, actual *TapValidation) {
	canonicalizeHeaders(expected, actual)
	if diff := cmp.Diff(expected, actual, cmpopts.IgnoreFields(TapValidation{}, "ContentLength")); diff != "" {
		t.Errorf("Unexpected difference in %s (-want, +got): %v", ctx, diff)
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
