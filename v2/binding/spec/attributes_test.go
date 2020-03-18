package spec_test

import (
	"testing"
	"time"

	"github.com/cloudevents/sdk-go/v2/binding/spec"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
)

func TestAttributes03(t *testing.T) {
	assert := require.New(t)
	v := spec.WithPrefix("x:").Version("0.3")
	subject := v.Attribute("x:subject")
	assert.Equal(spec.Subject, subject.Kind())
	assert.Equal("x:subject", subject.PrefixedName())

	c := v.NewContext()
	assert.Equal("0.3", c.GetSpecVersion())
	assert.NoError(subject.Set(c, "foobar"))
	got := subject.Get(c)
	assert.Equal("foobar", got)
	assert.Equal("foobar", c.GetSubject())

	now := time.Now()
	atime := v.Attribute("x:time")
	assert.Equal(spec.Time, atime.Kind())
	assert.NoError(atime.Set(c, now))
	tm := atime.Get(c)
	assert.Empty(cmp.Diff(now, tm))
	assert.Empty(cmp.Diff(now, c.GetTime()))

	nosuch := v.Attribute("nosuch")
	assert.Nil(nosuch)

	err := subject.Set(c, 1)
	assert.EqualError(err, "invalid value for subject: 1")
	err = atime.Set(c, "foo")
	assert.EqualError(err, `invalid value for time: "foo"`)
}

func TestAttributes1(t *testing.T) {
	assert := require.New(t)
	v := spec.WithPrefix("x:").Version("1.0")
	c := v.NewContext()
	id := v.Attribute("x:id")
	assert.NoError(id.Set(c, "foobar"))
	s := id.Get(c)
	assert.Equal("foobar", s)
	assert.Equal("foobar", c.GetID())

	now := time.Now()
	atime := v.Attribute("x:time")
	assert.Equal(spec.Time, atime.Kind())
	assert.NoError(atime.Set(c, now))
	tm := atime.Get(c)
	assert.Empty(cmp.Diff(now, tm))
	assert.Empty(cmp.Diff(now, c.GetTime()))

	nosuch := v.Attribute("nosuch")
	assert.Nil(nosuch)

	err := id.Set(c, 1)
	assert.EqualError(err, "invalid value for id: 1")
	err = atime.Set(c, "foo")
	assert.EqualError(err, "invalid value for time: \"foo\"")
}

func TestAttributesBadVersions(t *testing.T) {
	assert := require.New(t)
	v := spec.WithPrefix("x:").Version("0.x")
	assert.Nil(v)
}
