/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/extensions"
	cehttp "github.com/cloudevents/sdk-go/v2/protocol/http"
)

func main() {
	ctx := cloudevents.ContextWithTarget(context.Background(), "http://localhost:8080/")

	p, err := cloudevents.NewHTTP()
	if err != nil {
		log.Fatalf("failed to create protocol: %s", err.Error())
	}

	c, err := cloudevents.NewClient(p, cloudevents.WithTimeNow(), cloudevents.WithUUIDs())
	if err != nil {
		log.Fatalf("failed to create client, %v", err)
	}

	// Initial event in a logical flow
	correlationID := "txn-abc-123"
	fmt.Printf("[Correlation ID: %s]\n", correlationID)

	e1 := cloudevents.NewEvent()
	e1.SetID("order-123")
	e1.SetType("com.example.order.placed")
	e1.SetSource("https://example.com/orders")
	_ = e1.SetData(cloudevents.ApplicationJSON, map[string]interface{}{
		"orderId":    "123",
		"customerId": "456",
	})

	// Add correlation extension
	ext1 := extensions.CorrelationExtension{
		CorrelationID: correlationID,
	}
	ext1.AddCorrelationAttributes(&e1)

	send(c, ctx, e1, "└── ", "Order Placed")

	// Event B: Payment Processed (triggered by order A)
	e2 := cloudevents.NewEvent()
	e2.SetID("payment-789")
	e2.SetType("com.example.payment.processed")
	e2.SetSource("https://example.com/payments")
	_ = e2.SetData(cloudevents.ApplicationJSON, map[string]interface{}{
		"amount":   150.0,
		"currency": "USD",
	})

	ext2 := extensions.CorrelationExtension{
		CorrelationID: correlationID,
		CausationID:   e1.ID(),
	}
	ext2.AddCorrelationAttributes(&e2)

	send(c, ctx, e2, "    ├── ", "Payment Processed")

	// Event C: Inventory Checked (triggered by order A)
	e3 := cloudevents.NewEvent()
	e3.SetID("inventory-456")
	e3.SetType("com.example.inventory.checked")
	e3.SetSource("https://example.com/inventory")
	_ = e3.SetData(cloudevents.ApplicationJSON, map[string]interface{}{
		"items":     []string{"sku-001", "sku-002"},
		"available": true,
	})

	ext3 := extensions.CorrelationExtension{
		CorrelationID: correlationID,
		CausationID:   e1.ID(),
	}
	ext3.AddCorrelationAttributes(&e3)

	send(c, ctx, e3, "    └── ", "Inventory Checked")

	// Event D: Shipping Scheduled (triggered by inventory check C)
	e4 := cloudevents.NewEvent()
	e4.SetID("shipping-012")
	e4.SetType("com.example.shipping.scheduled")
	e4.SetSource("https://example.com/shipping")
	_ = e4.SetData(cloudevents.ApplicationJSON, map[string]interface{}{
		"carrier":           "FastShip",
		"estimatedDelivery": "2024-01-15",
	})

	ext4 := extensions.CorrelationExtension{
		CorrelationID: correlationID,
		CausationID:   e3.ID(),
	}
	ext4.AddCorrelationAttributes(&e4)

	send(c, ctx, e4, "        └── ", "Shipping Scheduled")

	// Event E: Notification Sent (triggered by shipping D)
	e5 := cloudevents.NewEvent()
	e5.SetID("notify-email-890")
	e5.SetType("com.example.notification.email")
	e5.SetSource("https://example.com/notifications")
	_ = e5.SetData(cloudevents.ApplicationJSON, map[string]interface{}{
		"recipient": "customer@example.com",
		"template":  "order-fulfilled",
	})

	ext5 := extensions.CorrelationExtension{
		CorrelationID: correlationID,
		CausationID:   e4.ID(),
	}
	ext5.AddCorrelationAttributes(&e5)

	send(c, ctx, e5, "            └── ", "Notification Sent")
}

func send(c cloudevents.Client, ctx context.Context, e cloudevents.Event, prefix string, label string) {
	res := c.Send(ctx, e)
	if cloudevents.IsUndelivered(res) {
		fmt.Printf("%sID: %s (%s) [FAILED: %v]\n", prefix, e.ID(), label, res)
		return
	}
	var httpResult *cehttp.Result
	if cloudevents.ResultAs(res, &httpResult) {
		status := fmt.Sprintf("%d", httpResult.StatusCode)
		if httpResult.StatusCode != http.StatusOK && httpResult.StatusCode != http.StatusAccepted {
			status = fmt.Sprintf("FAILED %d: %s", httpResult.StatusCode, fmt.Sprintf(httpResult.Format, httpResult.Args...))
		}
		fmt.Printf("%sID: %s (%s) [%s]\n", prefix, e.ID(), label, status)
		return
	}
	fmt.Printf("%sID: %s (%s) [%s]\n", prefix, e.ID(), label, res.Error())
}
