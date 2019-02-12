package types

import "reflect"

// Allocates allocates a new instance of type t and returns:
// asPtr is of type t if t is a pointer type and of type &t otherwise
// asValue is a Value of type t pointing to the same data as asPtr
func Allocate(t reflect.Type) (asPtr interface{}, asValue reflect.Value) {
	if t == nil {
		return nil, reflect.Value{}
	}
	if t.Kind() == reflect.Ptr {
		reflectPtr := reflect.New(t.Elem())
		asPtr = reflectPtr.Interface()
		asValue = reflectPtr
	} else {
		reflectPtr := reflect.New(t)
		asPtr = reflectPtr.Interface()
		asValue = reflectPtr.Elem()
	}
	return
}
