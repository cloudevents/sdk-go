/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package amqp

import (
	"net/url"
	"os"
	"testing"

	"github.com/Azure/go-amqp"

	"github.com/stretchr/testify/require"

	protocolamqp "github.com/cloudevents/sdk-go/protocol/amqp/v2"
	clienttest "github.com/cloudevents/sdk-go/v2/client/test"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/cloudevents/sdk-go/v2/test"
)

func TestSendEvent(t *testing.T) {
	test.EachEvent(t, test.Events(), func(t *testing.T, eventIn event.Event) {
		eventIn = test.ConvertEventExtensionsToString(t, eventIn)
		clienttest.SendReceive(t, func() interface{} {
			return protocolFactory(t)
		}, eventIn, func(e event.Event) {
			test.AssertEventEquals(t, eventIn, test.ConvertEventExtensionsToString(t, e))
		})
	})
}

func TestSenderReceiverEvent(t *testing.T) {
	test.EachEvent(t, test.Events(), func(t *testing.T, eventIn event.Event) {
		eventIn = test.ConvertEventExtensionsToString(t, eventIn)
		clienttest.SendReceive(t, func() interface{} {
			s := senderProtocolFactory(t)
			r := receiverProtocolFactory(t)
			s.Receiver = r.Receiver
			return s
		}, eventIn, func(e event.Event) {
			test.AssertEventEquals(t, eventIn, test.ConvertEventExtensionsToString(t, e))
		})
	})
}

func senderProtocolFactory(t *testing.T) *protocolamqp.Protocol {
	c, ss, a := testClient(t)

	p, err := protocolamqp.NewSenderProtocolFromClient(c, ss, a)
	require.NoError(t, err)

	return p
}

func receiverProtocolFactory(t *testing.T) *protocolamqp.Protocol {
	c, ss, a := testClient(t)

	p, err := protocolamqp.NewReceiverProtocolFromClient(c, ss, a)
	require.NoError(t, err)

	return p
}

// Some test require an AMQP broker or router. If the connection fails
// the tests are skipped. The env variable TEST_AMQP_URL can be set to the
// test URL, otherwise the default is "/test"
//
// On option is http://qpid.apache.org/components/dispatch-router/indexthtml.
// It can be installed from source or from RPMs, see https://qpid.apache.org/packages.html
// Run `qdrouterd` and the tests will work with no further config.
func testClient(t *testing.T) (client *amqp.Client, session *amqp.Session, addr string) {
	t.Helper()
	addr = "test"
	s := os.Getenv("TEST_AMQP_URL")
	if u, err := url.Parse(s); err == nil && u.Path != "" {
		addr = u.Path
	}
	client, err := amqp.Dial(s)
	if err != nil {
		t.Skipf("ampq.Dial(%#v): %v", s, err)
	}
	session, err = client.NewSession()
	require.NoError(t, err)
	return client, session, addr
}

func protocolFactory(t *testing.T) *protocolamqp.Protocol {
	c, ss, a := testClient(t)

	p, err := protocolamqp.NewProtocolFromClient(c, ss, a)
	require.NoError(t, err)

	return p
}
