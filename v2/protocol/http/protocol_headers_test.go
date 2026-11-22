/*
 Copyright 2026 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package http

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMakeRequestDoesNotMutateCustomHeaders(t *testing.T) {
	custom := http.Header{"X-Custom": {"context"}}
	ctx := WithCustomHeader(context.Background(), custom)
	p, err := New()
	require.NoError(t, err)
	p.RequestTemplate = &http.Request{Header: http.Header{"X-Custom": {"template"}}}

	for range 2 {
		req := p.makeRequest(ctx)
		assert.Equal(t, []string{"context", "template"}, req.Header.Values("X-Custom"))
	}
	assert.Equal(t, http.Header{"X-Custom": {"context"}}, custom)
}
