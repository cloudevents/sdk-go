package spec_test

import (
	"testing"
	"time"

	"github.com/cloudevents/sdk-go/v1/binding/spec"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
)

func TestAttributes03(t *testing.T) {
	assert := assert.New(t)
	v, err := spec.WithPrefix("x:").Version("0.3")
	assert.NoError(err)
	subject := v.Attribute("x:subject")
	assert.Equal(spec.Subject, subject.Kind())
	assert.Equal("x:subject", subject.Name())

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

	err = subject.Set(c, 1)
	assert.EqualError(err, "invalid value for subject: 1")
	err = atime.Set(c, "foo")
	assert.EqualError(err, `invalid value for time: "foo"`)
}

func TestAttributes02(t *testing.T) {
	assert := assert.New(t)
	v, err := spec.WithPrefix("x:").Version("0.2")
	assert.NoError(err)
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

	err = id.Set(c, 1)
	assert.EqualError(err, "invalid value for id: 1")
	err = atime.Set(c, "foo")
	assert.EqualError(err, "invalid value for time: \"foo\"")
}

func TestAttributes01(t *testing.T) {
	assert := assert.New(t)
	v, err := spec.WithPrefix("x:").Version("0.1")
	assert.NoError(err)
	c := v.NewContext()
	contentType := v.Attribute("x:contentType")
	assert.Equal(spec.DataContentType, contentType.Kind())
	assert.NoError(contentType.Set(c, "foobar"))
	s := contentType.Get(c)
	assert.Equal("foobar", s)
	assert.Equal("foobar", c.GetDataContentType())

	nosuch := v.Attribute("x:subject")
	assert.Nil(nosuch)
}

func TestAttributesBadVersions(t *testing.T) {
	assert := assert.New(t)
	v, err := spec.WithPrefix("x:").Version("0.x")
	assert.Nil(v)
	assert.EqualError(err, `invalid spec version "0.x"`)
}
