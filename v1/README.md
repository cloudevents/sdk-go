# NOTE: The v1 directory will be removed for v2.0.0.

We will make a final migration release that will match the v2.0.0 release with
the addition of this legacy directory for migration support.

# Migration Guide

To enable an easier migration, this directory holds a copy of the code in the
branch `release-1.y.z`.

Switch your imports from `github.com/cloudevents/sdk-go` to
`github.com/cloudevents/sdk-go/legacy` and your code should compile again,
letting you get on with the task of migrating to v2.

## Background

In the migration from v1 to v2 of the SDK, there are a lot of API breaking
changes. It is shorter to define what is not a breaking change:

- The `Event` object marshaling results in the same json.
- cloudevents.NewDefault should get an http server working out of the box.
- Most of alias.go file remains with some exceptions.
- Most of the original demos remain in the cmd dir to see how the new
  integrations should be.

Large breaking changes have to do with a strategy change inside the SDK to shift
the control to the integrator, allowing more direct access to the knobs required
to make choices over plumbing those knobs through the SDK down to the original
transports that implement the features integrators are really trying to control.

If you implemented a custom transport, the migration to how protocol bindings
work is covered in the document [TBD](TODO).

## Architectural Changes

New Architectural Layout:

```
Client      <-- Operates on event.Event
   |
   v
Protocol    <-- Operates on binding.Message
   |
   v
3rd Party   <-- Operates out of our control
```

Some Architectural changes that are worth noting:

- Client still exists but it has a new API.
  - client.Request allows for responses from outbound events.
  - client.StartReceiver has a mode that will test for underlying support if the
    receiver function is allowed to produce responses to inbound events.
- Client interface is event.Event focused.
- Protocol layer operates on `binding.Message`
  - This is a change from v1, `transport.Transport` mixed up `event.Event`
    objects into the interface. With the thinking that codecs were specific to a
    transport. But as we implemented bindings, it became clear that there are
    many cases where the cost to convert a 3rd Party message into a
    `event.Event` is too high and it is better to stay in the intermediate state
    of a `binding.Message` (similar to a `transport.Message` but
    `transport.Message` was never exposed in the v1 architecture).
- Setting a transport to emit a specific version of cloudevents is an
  anti-pattern. If a version is required, the burden should be on the integrator
  to implement what they need. The edge cases the SDK had to handle made that
  code unrulely. It is simpler if the SDK does simple things. So outbound event
  encoding is based on the `event.Event` that is passed in.

## Moves and renames

Note these are based on internal packages unless noted as from alias.

- `cloudevents.Event` --> `event.Event`
- `transport.Codec` --> Deleted, the binding concept replaced it.
- `transport.Transport` --> Deleted, the
  protocol.Sender/Receiver/Requester/Responder interfaces replaced it.
