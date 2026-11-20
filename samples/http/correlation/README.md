# Correlation Sample

This sample demonstrates how to use the `CorrelationExtension` to track event relationships and causality in distributed systems.

## Prerequisites

- Go 1.25.0 or later
- Access to a terminal

## Running the Sample

1. Start the receiver in one terminal:
   ```bash
   go run receiver/main.go
   ```

2. Start the sender in another terminal:
   ```bash
   go run sender/main.go
   ```

## Expected Output

### Receiver
The receiver will print the incoming events along with their correlation and causation identifiers:
```
Received Event:
Context Attributes,
  specversion: 1.0
  type: com.example.order.placed
  source: https://example.com/orders
  id: order-123
Extensions,
  correlationid: txn-abc-123
Data,
  {
    "customerId": "456",
    "orderId": "123"
  }
Correlation ID: txn-abc-123
-------------------------------------------------
Received Event:
Context Attributes,
  specversion: 1.0
  type: com.example.payment.processed
  source: https://example.com/payments
  id: payment-789
Extensions,
  causationid: order-123
  correlationid: txn-abc-123
Data,
  {
    "amount": 150,
    "currency": "USD"
  }
Correlation ID: txn-abc-123
Causation ID: order-123
-------------------------------------------------
Received Event:
Context Attributes,
  specversion: 1.0
  type: com.example.inventory.checked
  source: https://example.com/inventory
  id: inventory-456
Extensions,
  causationid: order-123
  correlationid: txn-abc-123
Data,
  {
    "available": true,
    "items": ["sku-001", "sku-002"]
  }
Correlation ID: txn-abc-123
Causation ID: order-123
-------------------------------------------------
Received Event:
Context Attributes,
  specversion: 1.0
  type: com.example.shipping.scheduled
  source: https://example.com/shipping
  id: shipping-012
Extensions,
  causationid: inventory-456
  correlationid: txn-abc-123
Data,
  {
    "carrier": "FastShip",
    "estimatedDelivery": "2024-01-15"
  }
Correlation ID: txn-abc-123
Causation ID: inventory-456
-------------------------------------------------
Received Event:
Context Attributes,
  specversion: 1.0
  type: com.example.notification.email
  source: https://example.com/notifications
  id: notify-email-890
Extensions,
  causationid: shipping-012
  correlationid: txn-abc-123
Data,
  {
    "recipient": "customer@example.com",
    "template": "order-fulfilled"
  }
Correlation ID: txn-abc-123
Causation ID: shipping-012
-------------------------------------------------
```

### Sender
The sender will log its activity, showing the causation relationship in a tree format:
```
[Correlation ID: txn-abc-123]
└── ID: order-123 (Order Placed) [202]
    ├── ID: payment-789 (Payment Processed) [202]
    └── ID: inventory-456 (Inventory Checked) [202]
        └── ID: shipping-012 (Shipping Scheduled) [202]
            └── ID: notify-email-890 (Notification Sent) [202]
```
