/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package runtime

import (
	"os"
	"strings"
	"testing"

	cesql "github.com/cloudevents/sdk-go/sql/v2"
	"github.com/cloudevents/sdk-go/sql/v2/function"
	cloudevents "github.com/cloudevents/sdk-go/v2"
)

var HasPrefixCustomFunction cesql.Function

func TestMain(m *testing.M) {
	HasPrefixCustomFunction = function.NewFunction(
		"HASPREFIX",
		[]cesql.Type{cesql.StringType, cesql.StringType},
		nil,
		func(event cloudevents.Event, i []interface{}) (interface{}, error) {
			str := i[0].(string)
			prefix := i[1].(string)

			return strings.HasPrefix(str, prefix), nil
		},
	)
	os.Exit(m.Run())
}

func Test_functionTable_AddFunction(t *testing.T) {

	type args struct {
		function cesql.Function
	}
	tests := []struct {
		name    string
		table   functionTable
		args    args
		wantErr bool
	}{
		{
			name:  "Add custom fixedArgs func",
			table: globalFunctionTable,
			args: args{
				function: HasPrefixCustomFunction,
			},
			wantErr: false,
		},
		{
			name:  "Fail add custom fixedArgs func if it exists",
			table: globalFunctionTable,
			args: args{
				function: HasPrefixCustomFunction,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.table.AddFunction(tt.args.function); (err != nil) != tt.wantErr {
				t.Errorf("functionTable.AddFunction() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
