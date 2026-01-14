# Migration Guide: go-amqp v0.17.0 to v1.x

This document describes the breaking changes when migrating from go-amqp v0.17.0 to v1.x in the CloudEvents AMQP protocol binding.

## Overview

The CloudEvents AMQP protocol binding now uses go-amqp v1.x (stable release) instead of v0.17.0 (pre-release). This brings stability and proper semantic versioning, but requires updates to existing code.

## Breaking Changes

### 1. Connection Type Changed

**Before (v0.17.0):**
```go
var client *amqp.Client
```

**After (v1.x):**
```go
var conn *amqp.Conn
```

### 2. Function Signatures Require Context

All AMQP operations now require a `context.Context` parameter.

**Before:**
```go
conn, err := amqp.Dial(server, options...)
session, err := client.NewSession(options...)
sender, err := session.NewSender(options...)
receiver, err := session.NewReceiver(options...)
```

**After:**
```go
ctx := context.Background()
conn, err := amqp.Dial(ctx, server, options)
session, err := conn.NewSession(ctx, options)
sender, err := session.NewSender(ctx, target, options)
receiver, err := session.NewReceiver(ctx, source, options)
```

### 3. Options Pattern Changed

Options changed from functional (variadic) to struct-based (single pointer).

**Before (functional options):**
```go
protocol, err := amqp.NewProtocol(
    server,
    queue,
    []amqp.ConnOption{amqp.ConnSASLPlain(user, pass)},
    []amqp.SessionOption{},
)
```

**After (struct options):**
```go
protocol, err := amqp.NewProtocol(
    server,
    queue,
    &amqp.ConnOptions{
        SASLType: amqp.SASLTypePlain(user, pass),
    },
    nil, // SessionOptions
)
```

### 4. Protocol Options Updated

Connection and session options are now passed as direct parameters.
Link options (sender/receiver) remain as variadic options.

**Before:**
```go
protocol, err := amqp.NewProtocol(server, queue,
    []amqp.ConnOption{amqp.ConnSASLPlain(user, pass)},
    []amqp.SessionOption{amqp.SessionMaxLinks(100)},
    amqp.WithSenderLinkOption(amqp.LinkSenderSettle(amqp.ModeSettled)),
)
```

**After:**
```go
protocol, err := amqp.NewProtocol(server, queue,
    &amqp.ConnOptions{
        SASLType: amqp.SASLTypePlain(user, pass),
    },
    &amqp.SessionOptions{
        MaxLinks: 100,
    },
    amqp.WithSenderOptions(&amqp.SenderOptions{
        SettlementMode: &amqp.SenderSettleModeSettled,
    }),
)
```

For NewProtocolFromConn (connection already created):
```go
protocol, err := amqp.NewProtocolFromConn(conn, session, queue,
    amqp.WithSenderOptions(&amqp.SenderOptions{
        SettlementMode: &amqp.SenderSettleModeSettled,
    }),
)
```

### 5. Error Handling

**Before:**
```go
condition := amqp.ErrorCondition("my-error")
```

**After:**
```go
condition := amqp.ErrCond("my-error")
```

## Complete Example Migration

### Before (v0.17.0)

```go
package main

import (
    "log"

    "github.com/Azure/go-amqp"
    ceamqp "github.com/cloudevents/sdk-go/protocol/amqp/v2"
)

func main() {
    // This code worked with go-amqp v0.17.0 but breaks with v1.x
    protocol, err := ceamqp.NewProtocol(
        "amqp://localhost:5672",
        "myqueue",
        []amqp.ConnOption{amqp.ConnSASLPlain("user", "pass")},
        []amqp.SessionOption{},
    )
    if err != nil {
        log.Fatal(err)
    }
    defer protocol.Close(context.Background())

    // Use protocol...
}
```

### After (v1.x)

```go
package main

import (
    "context"
    "log"

    "github.com/Azure/go-amqp"
    ceamqp "github.com/cloudevents/sdk-go/protocol/amqp/v2"
)

func main() {
    ctx := context.Background()

    // Updated for go-amqp v1.x
    protocol, err := ceamqp.NewProtocol(
        "amqp://localhost:5672",
        "myqueue",
        &amqp.ConnOptions{
            SASLType: amqp.SASLTypePlain("user", "pass"),
        },
        nil, // SessionOptions - use nil for defaults
        ceamqp.WithSenderOptions(&amqp.SenderOptions{
            SettlementMode: &amqp.SenderSettleModeSettled,
        }),
    )
    if err != nil {
        log.Fatal(err)
    }
    defer protocol.Close(ctx)

    // Use protocol...
}
```

## Azure Service Bus Example

### Before

```go
connOptions := []amqp.ConnOption{
    amqp.ConnSASLPlain(keyName, key),
    amqp.ConnProperty("product", "my-app"),
}

protocol, err := ceamqp.NewProtocol(
    "amqps://myns.servicebus.windows.net",
    "myqueue",
    connOptions,
    []amqp.SessionOption{},
)
```

### After

```go
ctx := context.Background()

connOptions := &amqp.ConnOptions{
    SASLType: amqp.SASLTypePlain(keyName, key),
    Properties: map[string]any{
        "product": "my-app",
    },
}

protocol, err := ceamqp.NewProtocol(
    "amqps://myns.servicebus.windows.net",
    "myqueue",
    connOptions,
    nil, // SessionOptions
)
```

## Testing

If you're using the protocol binding in tests, update your test helpers:

### Before

```go
func setupProtocol(t *testing.T) *ceamqp.Protocol {
    client, _ := amqp.Dial("amqp://localhost:5672")
    session, _ := client.NewSession()
    protocol, err := ceamqp.NewProtocolFromConn(client, session, "test")
    require.NoError(t, err)
    return protocol
}
```

### After

```go
func setupProtocol(t *testing.T) *ceamqp.Protocol {
    ctx := context.Background()
    conn, _ := amqp.Dial(ctx, "amqp://localhost:5672", nil)
    session, _ := conn.NewSession(ctx, nil)
    protocol, err := ceamqp.NewProtocolFromConn(conn, session, "test")
    require.NoError(t, err)
    return protocol
}
```

## Common Issues

### Issue: "undefined: amqp.Client"

**Solution:** Change `*amqp.Client` to `*amqp.Conn`

### Issue: "not enough arguments in call to amqp.Dial"

**Solution:** Add `context.Context` as first parameter and use struct options:
```go
// Before
conn, err := amqp.Dial(addr, opts...)

// After
conn, err := amqp.Dial(ctx, addr, &amqp.ConnOptions{...})
```

### Issue: "undefined: amqp.ConnOption"

**Solution:** Replace functional options with struct fields:
```go
// Before
opt := amqp.ConnSASLPlain(user, pass)

// After
opts := &amqp.ConnOptions{
    SASLType: amqp.SASLTypePlain(user, pass),
}
```

### Issue: "cannot use nil as type []amqp.ConnOption"

**Solution:** Use `nil` instead of `[]amqp.ConnOption(nil)`:
```go
// Before
protocol, err := ceamqp.NewProtocol(addr, queue, nil, nil)

// After (no change needed, but be explicit)
protocol, err := ceamqp.NewProtocol(addr, queue, nil, nil)
```

## Benefits of v1.x

1. **Stable API**: Semantic versioning guarantees
2. **Context support**: Proper cancellation and timeout handling
3. **Better errors**: More detailed error types
4. **Type safety**: Struct options catch errors at compile time
5. **No replace directive**: Works correctly as a dependency

## Additional Resources

- [go-amqp v1.x documentation](https://pkg.go.dev/github.com/Azure/go-amqp@v1.5.1)
- [CloudEvents AMQP Protocol Binding Spec](https://github.com/cloudevents/spec/blob/v1.0.2/cloudevents/bindings/amqp-protocol-binding.md)
- [GitHub Issue #1039](https://github.com/cloudevents/sdk-go/issues/1039)

## Questions?

If you encounter issues during migration, please:
1. Check this guide for common solutions
2. Review the go-amqp v1.x documentation
3. Open an issue on the CloudEvents SDK repository
