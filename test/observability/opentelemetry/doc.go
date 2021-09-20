/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

/*
Package opentelemetry validates the carrier and observability service instrumentation with the default SDK.

This package is in a separate module from the instrumentation it tests to
isolate the dependency of the default OTel SDK and not impose this as a transitive
dependency for users.
*/
package opentelemetry
