package types

import "reflect"

// Allocates allocates a new instance of type t and returns:
// asPtr is of type t if t is a pointer type and of type &t otherwise
// asValue is a Value of type t pointing to the same data as asPtr
func Allocate(obj interface{}) (asPtr interface{}, asValue reflect.Value) {
	if obj == nil {
		return nil, reflect.Value{}
	}

	t := reflect.TypeOf(obj)

	//t, ok := obj.(reflect.Type)
	//if ok {
	switch t.Kind() {
	case reflect.Ptr:
		reflectPtr := reflect.New(t.Elem())
		asPtr = reflectPtr.Interface()
		asValue = reflectPtr
	case reflect.Map:
		reflectPtr := reflect.MakeMap(t)
		asPtr = reflectPtr.Interface()
		asValue = reflectPtr
	case reflect.String:

		reflectPtr := reflect.New(t)
		asPtr = ""
		asValue = reflectPtr.Elem()
	default:
		reflectPtr := reflect.New(t)
		asPtr = reflectPtr.Interface()
		asValue = reflectPtr.Elem()
	}
	//}
	return
}
