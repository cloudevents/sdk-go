package spec_test

import (
	"testing"

	"github.com/cloudevents/sdk-go/v1/binding/spec"
	ce "github.com/cloudevents/sdk-go/v1/cloudevents"
	"github.com/stretchr/testify/assert"
)

func TestVersions(t *testing.T) {
	assert := assert.New(t)
	versions := spec.New()
	assert.Equal([]string{"specversion", "cloudEventsVersion"}, versions.SpecVersionNames())

	want := []string{"1.0", "0.3", "0.2", "0.1"}
	all := versions.Versions()
	assert.Equal(len(want), len(all))
	for i, s := range want {
		assert.Equal(s, all[i].String())
		assert.Equal(s, all[i].NewContext().GetSpecVersion())
		converted := all[i].Convert(ce.EventContextV01{}.AsV01())
		assert.Equal(s, converted.GetSpecVersion(), "%v %v %v", i, s, converted)
	}
	assert.Equal(want[0], versions.Latest().NewContext().GetSpecVersion())
}
