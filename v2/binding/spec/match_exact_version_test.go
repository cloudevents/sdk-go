package spec_test

import (
	"net/textproto"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cloudevents/sdk-go/v2/binding/spec"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/cloudevents/sdk-go/v2/test"
)

func TestMatchExactVersion(t *testing.T) {
	test.EachEvent(t, test.AllVersions([]event.Event{test.FullEvent()}), func(t *testing.T, e event.Event) {
		e = e.Clone()
		s := spec.WithPrefixMatchExact(
			func(s string) string {
				if s == "datacontenttype" {
					return "Content-Type"
				} else {
					return textproto.CanonicalMIMEHeaderKey("Ce-" + s)
				}
			},
			"Ce-",
		)
		sv := s.Version(e.SpecVersion())
		require.NotNil(t, sv)

		require.Equal(t, e.ID(), sv.Attribute("Ce-Id").Get(e.Context))
		require.Equal(t, "id", sv.Attribute("Ce-Id").Name())

		require.Equal(t, e.DataContentType(), sv.Attribute("Content-Type").Get(e.Context))
		require.Equal(t, "datacontenttype", sv.Attribute("Content-Type").Name())
	})
}
