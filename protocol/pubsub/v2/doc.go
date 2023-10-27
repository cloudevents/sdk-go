/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

/*
Package pubsub implements a Pub/Sub binding using google.cloud.com/go/pubsub module

PubSub Messages can be modified beyond what CloudEvents cover by using `WithOrderingKey`
or `WithCustomAttributes`. See function docs for more details.
*/
package pubsub
