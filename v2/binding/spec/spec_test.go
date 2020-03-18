package spec_test

import (
	"testing"

	"github.com/cloudevents/sdk-go/v2/binding/spec"
	"github.com/cloudevents/sdk-go/v2/event"
	tassert "github.com/stretchr/testify/assert"
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
