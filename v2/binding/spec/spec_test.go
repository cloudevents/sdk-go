package spec_test

import (
	"testing"

	"github.com/cloudevents/sdk-go/v2/binding/spec"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/cloudevents/sdk-go/v2/test"
	tassert "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVersions(t *testing.T) {
	assert := tassert.New(t)
	versions := spec.New()
	assert.Equal("specversion", versions.PrefixedSpecVersionName())

	want := []string{"1.0", "0.3"}
	all := versions.Versions()
	assert.Equal(len(want), len(all))
	for i, s := range want {
		assert.Equal(s, all[i].String())
		assert.Equal(s, all[i].NewContext().GetSpecVersion())
		converted := all[i].Convert(event.EventContextV1{}.AsV1())
		assert.Equal(s, converted.GetSpecVersion(), "%v %v %v", i, s, converted)
	}
	assert.Equal(want[0], versions.Latest().NewContext().GetSpecVersion())
}

func TestSetAttribute(t *testing.T) {
	test.EachEvent(t, test.AllVersions([]event.Event{test.MinEvent()}), func(t *testing.T, e event.Event) {
		e = e.Clone()
		s := spec.WithPrefix("ce_")
		sv := s.Version(e.SpecVersion())

		id := "another-id"
		require.NoError(t, sv.SetAttribute(e.Context, "ce_id", id))
		require.Equal(t, id, e.ID())

		extName := "ce_someExt"
		extValue := "extValue"
		require.NoError(t, sv.SetAttribute(e.Context, extName, extValue))
		require.Equal(t, extValue, e.Extensions()["someext"])
	})
}
