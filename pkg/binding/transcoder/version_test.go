package transcoder

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/binding/spec"
	"github.com/cloudevents/sdk-go/pkg/binding/test"
	ce "github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/types"
)

func TestVersionTranscoder(t *testing.T) {
	var testEventV02 = cloudevents.Event{
		Context: cloudevents.EventContextV02{
			Source:      types.URLRef{URL: url.URL{Path: "source"}},
			ContentType: cloudevents.StringOfApplicationJSON(),
			ID:          "id",
			Type:        "type",
		}.AsV02(),
	}

	var testEventV1 = testEventV02
	testEventV1.Context = testEventV02.Context.AsV1()

	data := []byte("\"data\"")
	err := testEventV02.SetData(data)
	if err != nil {
		panic(err)
	}
	err = testEventV1.SetData(data)
	if err != nil {
		panic(err)
	}

	tests := []struct {
		name    string
		message binding.Message
		want    ce.Event
	}{
		{
			name:    "V02 -> V1 with Structured message",
			message: binding.NewMockStructuredMessage(copyEventContext(testEventV02)),
			want:    copyEventContext(testEventV1),
		},
		{
			name:    "V02 -> V1 with Binary message",
			message: binding.NewMockBinaryMessage(copyEventContext(testEventV02)),
			want:    copyEventContext(testEventV1),
		},
		{
			name:    "V02 -> V1 with Event message",
			message: binding.EventMessage(copyEventContext(testEventV02)),
			want:    copyEventContext(testEventV1),
		},
	}
	for _, tt := range tests {
		tt := tt // Don't use range variable inside scope
		factory := Version(spec.V1)
		t.Run(tt.name, func(t *testing.T) {
			e, _, _, err := binding.ToEvent(tt.message, factory)
			assert.NoError(t, err)
			test.AssertEventEquals(t, tt.want, e)
		})
	}
}

func copyEventContext(e ce.Event) ce.Event {
	newE := ce.Event{}
	newE.Context = e.Context.Clone()
	newE.DataEncoded = e.DataEncoded
	newE.Data = e.Data
	newE.DataBinary = e.DataBinary
	return newE
}
