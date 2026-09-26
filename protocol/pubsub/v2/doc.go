/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

/*
Package pubsub implements a Pub/Sub binding using google.cloud.com/go/pubsub module

PubSub Messages can be modified beyond what CloudEvents cover by using `WithOrderingKey`
or `WithCustomAttributes`. See function docs for more details.

In binary mode, content-type and ce-datacontenttype both describe the event data's
media type. If both are present, their values must match exactly; ReadBinary
rejects conflicting values. When reading, legacy Content-Type is used only when
neither is present. Binary messages are written with ce-datacontenttype and, for
backward compatibility with existing consumers, Content-Type. Use
ce-datacontenttype when filtering binary-mode events by data content type.

For structured JSON, the Google Cloud Pub/Sub binding defines the content-type
attribute as the CloudEvents envelope's media type (for example,
application/cloudevents+json; charset=utf-8). The datacontenttype property inside
the JSON describes the event data's media type. The binding permits this property
to also be included as a ce-datacontenttype Pub/Sub attribute, but that duplication
is optional in structured mode and cannot be assumed when defining filters.

When reading messages, content-type determines the encoding. If it is absent,
legacy Content-Type is used instead. A known CloudEvents event format selects
structured mode even when ce-specversion is also present.
*/
package pubsub
