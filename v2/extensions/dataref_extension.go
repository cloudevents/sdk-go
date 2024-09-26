/*
 Copyright 2024 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package extensions

import (
	"github.com/cloudevents/sdk-go/v2/event"
	"net/url"
)

const DataRefExtensionKey = "dataref"

type DataRefExtension struct {
	DataRef string `json:"dataref"`
}

func AddDataRefExtension(e *event.Event, dataRef string) error {
	if _, err := url.Parse(dataRef); err != nil {
		return err
	}
	e.SetExtension(DataRefExtensionKey, dataRef)
	return nil
}

func GetDataRefExtension(e event.Event) (DataRefExtension, bool) {
	if dataRefValue, ok := e.Extensions()[DataRefExtensionKey]; ok {
		dataRefStr, ok := dataRefValue.(string)
		if !ok {
			return DataRefExtension{}, false
		}
		return DataRefExtension{DataRef: dataRefStr}, true
	}
	return DataRefExtension{}, false
}
