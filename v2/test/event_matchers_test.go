package test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cloudevents/sdk-go/v2/event"
)

func TestAssertAnyOf(t *testing.T) {
	tests := []struct {
		name      string
		have      event.Event
		anyOf     []EventMatcher
		shouldErr bool
	}{{
		name:      "any of the two ids matching",
		have:      FullEvent(),
		anyOf:     []EventMatcher{HasId("min-event"), HasId("full-event")},
		shouldErr: false,
	}, {
		name:      "any of the two ids matching - reverse",
		have:      FullEvent(),
		anyOf:     []EventMatcher{HasId("full-event"), HasId("min-event")},
		shouldErr: false,
	}, {
		name:      "none matching",
		have:      FullEvent(),
		anyOf:     []EventMatcher{HasId("other-event"), HasId("min-event")},
		shouldErr: true,
	}, {
		name:      "both matching",
		have:      FullEvent(),
		anyOf:     []EventMatcher{HasId("full-event"), HasId("full-event")},
		shouldErr: false,
	},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldErr {
				require.Error(t, AnyOf(tt.anyOf...)(tt.have))
			} else {
				require.NoError(t, AnyOf(tt.anyOf...)(tt.have))
			}
		})
	}
}

func TestAssertContainsExactlyExtensions(t *testing.T) {
	tests := []struct {
		name      string
		have      event.Event
		exts      []string
		shouldErr bool
	}{{
		name:      "match all of them",
		have:      FullEvent(),
		exts:      []string{"exbool", "exint", "exstring", "exbinary", "exurl", "extime"},
		shouldErr: false,
	}, {
		name:      "match none",
		have:      MinEvent(),
		exts:      []string{},
		shouldErr: false,
	}, {
		name:      "no match because one more",
		have:      FullEvent(),
		exts:      []string{"exbool", "exint", "exstring", "exbinary", "exurl"},
		shouldErr: true,
	}, {
		name:      "no match because one less",
		have:      FullEvent(),
		exts:      []string{"exbool", "exint", "exstring", "exbinary", "exurl", "extime", "exother"},
		shouldErr: true,
	},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldErr {
				require.Error(t, ContainsExactlyExtensions(tt.exts...)(tt.have))
			} else {
				require.NoError(t, ContainsExactlyExtensions(tt.exts...)(tt.have))
			}
		})
	}
}
