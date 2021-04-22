/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package amqp

import "github.com/cloudevents/sdk-go/v2/types"

func safeAMQPPropertiesUnwrap(val interface{}) (interface{}, error) {
	v, err := types.Validate(val)
	if err != nil {
		return nil, err
	}
	switch t := v.(type) {
	case types.URI: // Use string form of URLs.
		v = t.String()
	case types.URIRef: // Use string form of URLs.
		v = t.String()
	case types.Timestamp: // Use string form of URLs.
		v = t.Time
	case int32: // Use AMQP long for Integer as per CE spec.
		v = int64(t)
	}

	return v, nil
}
