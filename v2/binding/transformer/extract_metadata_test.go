package transformer

import (
	"context"
	"fmt"
	"testing"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/binding/test"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/stretchr/testify/assert"
)

func ExampleExtractExtensions() {
	e := event.New("1.0")
	e.Context = &event.EventContextV1{}
	e.SetID("id")
	e.SetSource("source")
	e.SetType("type")
	e.SetExtension("someextension", "somevalue")
	e.SetExtension("otherextension", "othervalue") // extension will be ignored
	extractor := ExtractExtensions{"someextension": nil}
	_, err := binding.ToEvent(context.TODO(), binding.ToMessage(&e), extractor)
	if err != nil {
		fmt.Printf("Error running extractor: %v", err)
	}
	fmt.Println("extractor:", extractor)
	//Output:
	//extractor: map[someextension:somevalue]
}

func TestExtractExtensions(t *testing.T) {
	t.Parallel()

	e := test.MinEvent()
	e.Context = e.Context.AsV1()
	e.SetExtension("extension1", "extension1-val")
	e.SetExtension("extension2", false)

	tcs := []struct {
		name      string
		event     *event.Event
		extractor ExtractExtensions
		want      ExtractExtensions
	}{
		{
			name:  "no extensions extracted",
			event: &e,
			extractor: ExtractExtensions{
				"extension3": nil,
			},
			want: ExtractExtensions{
				"extension3": nil,
			},
		},
		{
			name:  "one extension extracted",
			event: &e,
			extractor: ExtractExtensions{
				"extension3": nil,
				"extension1": nil,
			},
			want: ExtractExtensions{
				"extension1": "extension1-val",
				"extension3": nil,
			},
		},
		{
			name:  "two extensions extracted",
			event: &e,
			extractor: ExtractExtensions{
				"extension1": nil,
				"extension2": nil,
				"extension3": nil,
			},
			want: ExtractExtensions{
				"extension1": "extension1-val",
				"extension2": false,
				"extension3": nil,
			},
		},
	}
	for _, tc := range tcs {
		eventExtractor := make(ExtractExtensions)
		binaryExtractor := make(ExtractExtensions)
		structuredExtractor := make(ExtractExtensions)
		for k, v := range tc.extractor {
			eventExtractor[k] = v
			binaryExtractor[k] = v
			structuredExtractor[k] = v
		}
		testArgs := []test.TransformerTestArgs{
			{
				Name:         "event",
				InputEvent:   tc.event.Clone(),
				AssertFunc:   assertExtractExtensions(tc.event, eventExtractor, tc.want),
				Transformers: []binding.TransformerFactory{eventExtractor},
			},
			{
				Name:         "binary",
				InputMessage: test.MustCreateMockBinaryMessage(*tc.event),
				AssertFunc:   assertExtractExtensions(tc.event, binaryExtractor, tc.want),
				Transformers: []binding.TransformerFactory{binaryExtractor},
			},
			{
				Name:         "structured",
				InputMessage: test.MustCreateMockStructuredMessage(*tc.event),
				AssertFunc:   assertExtractExtensions(tc.event, structuredExtractor, tc.want),
				Transformers: []binding.TransformerFactory{structuredExtractor},
			},
		}
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			test.RunTransformerTests(t, context.TODO(), testArgs)
		})
	}
}

func assertExtractExtensions(wantEvent *event.Event, extractor ExtractExtensions, wantExtractor ExtractExtensions) func(*testing.T, event.Event) {
	return func(t *testing.T, haveEvent event.Event) {
		test.AssertEventEquals(t, *wantEvent, haveEvent)
		assert.Equal(t, wantExtractor, extractor)
	}
}
