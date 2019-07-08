package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/cloudevents/sdk-go/pkg/client"

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/cloudevents/sdk-go/pkg/transport"
	cehttp "github.com/cloudevents/sdk-go/pkg/transport/http"
	"github.com/google/uuid"
)

// Conversion Test:

//         Data -> POST -> Wire Format -> Receive -> Error -> Convert -> Got
// Given:   ^                                                            ^==Want
// Data is a payload.

type ConversionTest struct {
	now       time.Time
	data      interface{}
	convertFn client.ConvertFn
	asSent    *TapValidation
	asRecv    *TapValidation
	want      *cloudevents.Event
}

type ConversionTestCases map[string]ConversionTest

func UnitTestConvert(ctx context.Context, m transport.Message, err error) (*cloudevents.Event, error) {
	if msg, ok := m.(*cehttp.Message); ok {
		tx := cehttp.TransportContextFrom(ctx)

		// Make a new event and convert the message payload.
		event := cloudevents.NewEvent()
		event.SetSource("github.com/cloudevents/test/http/conversion")
		event.SetType(fmt.Sprintf("io.cloudevents.conversion.http.%s", strings.ToLower(tx.Method)))
		event.SetID("321-CBA")
		event.SetExtension(unitTestIDKey, msg.Header.Get(unitTestIDKey))
		event.Data = msg.Body

		return &event, nil
	}
	return nil, err
}

func ClientConversion(t *testing.T, tc ConversionTest, topts ...cehttp.Option) {
	tap := NewTap()
	server := httptest.NewServer(tap)
	defer server.Close()

	if len(topts) == 0 {
		topts = append(topts, cehttp.WithBinaryEncoding())
	}
	topts = append(topts, cehttp.WithPort(0)) // random port
	trans, err := cehttp.New(
		topts...,
	)
	if err != nil {
		t.Fatal(err)
	}

	tap.handler = trans

	ce, err := client.New(
		trans,
		client.WithEventDefaulter(AlwaysThen(tc.now)),
		client.WithConverterFn(tc.convertFn),
	)
	if err != nil {
		t.Fatal(err)
	}

	testID := uuid.New().String()

	recvCtx, recvCancel := context.WithCancel(context.Background())
	go func() {
		t.Log(ce.StartReceiver(recvCtx, func(got *cloudevents.Event) {
			assertEventEquality(t, "got event", tc.want, got)
		}))
	}()

	b, err := json.Marshal(tc.data)
	if err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequest("POST", server.URL, bytes.NewBuffer(b))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(unitTestIDKey, testID)
	got, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	_ = got

	recvCancel()

	if req, ok := tap.req[testID]; ok {
		assertTappedEquality(t, "http request", tc.asSent, &req)
	}

	if resp, ok := tap.resp[testID]; ok {
		assertTappedEquality(t, "http response", tc.asRecv, &resp)
	}
}
