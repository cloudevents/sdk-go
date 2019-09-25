package types

import (
	"encoding/base64"
	"fmt"
	"math"
	"net/url"
	"reflect"
	"strconv"
	"time"
)

// Value holds the value of a CloudEvents attribute.
//
// Value can hold the native type or the string representation.
// The To...() methods extract the value as a native type.
type Value struct{ v interface{} }

// Interface returns the value as an interface{}.
//
// The return value may be nil if the Value is empty, otherwise it will always
// be one of the native types: bool, int32, string, []byte, url.URL, time.Time.
func (v Value) Interface() interface{} { return v.v }

// ValueOf validates the type and range of x and wraps it in a Value.
// x can be any type that is convertible to a CloudEvents attribute.
func ValueOf(x interface{}) (Value, error) {
	switch x := x.(type) {

	case Value: // Already a Value, no wrapping
		return x, nil

	case bool, int32, string, []byte, url.URL, time.Time:
		return Value{x}, nil // Already a preferred type, no conversion needed.

	case uint, uintptr, uint8, uint16, uint32, uint64:
		u := reflect.ValueOf(x).Uint()
		if u > math.MaxInt32 {
			return Value{}, rangeErr(x)
		}
		return Value{int32(u)}, nil
	case int, int8, int16, int64:
		i := reflect.ValueOf(x).Int()
		if i > math.MaxInt32 || i < math.MinInt32 {
			return Value{}, rangeErr(x)
		}
		return Value{int32(i)}, nil
	case float32, float64:
		f := reflect.ValueOf(x).Float()
		if f > math.MaxInt32 || f < math.MinInt32 {
			return Value{}, rangeErr(x)
		}
		return Value{int32(f)}, nil

	case URIRef:
		return Value{x.URL}, nil
	case URI:
		return Value{x.URL}, nil
	case URLRef:
		return Value{x.URL}, nil

	case Timestamp:
		return Value{x.Time}, nil

	case *url.URL, *URIRef, *URLRef, *URI, *Timestamp, *time.Time:
		rx := reflect.ValueOf(x)
		if rx.IsNil() {
			return Value{}, valueErr("", x)
		}
		return ValueOf(rx.Elem().Interface())

	default:
		return Value{}, fmt.Errorf("%T is not a CloudEvents type", x)
	}
}

func ValueOfBool(x bool) Value           { return Value{x} }
func ValueOfInteger(x int32) Value       { return Value{x} }
func ValueOfString(x string) Value       { return Value{x} }
func ValueOfBinary(x []byte) Value       { return Value{x} }
func ValueOfURIRef(x URIRef) Value       { return Value{x} }
func ValueOfTimestamp(x time.Time) Value { return Value{x} }

// String returns the canonical string encoded representation of the value.
// Use ToString() to extract the value only if it is of type string.
func (v Value) String() string {
	switch x := v.Interface().(type) {
	case bool:
		return strconv.FormatBool(x)
	case int32:
		return strconv.Itoa(int(x))
	case string:
		return x
	case []byte:
		return base64.StdEncoding.EncodeToString(x)
	case url.URL:
		return x.String()
	case time.Time:
		return x.UTC().Format(time.RFC3339Nano)
	case nil:
		return "<nil>"
	default:
		return fmt.Sprintf("<unknown:%T>", v.v)
	}
}

// ToBool returns a bool value, decoding from string if necessary.
func (v Value) ToBool() (bool, error) {
	switch x := v.v.(type) {
	case bool:
		return x, nil
	case string:
		return strconv.ParseBool(x)
	default:
		return false, typeErr("Bool", x)
	}
}

// ToInteger returns an int32 value, decoding from string if necessary.
func (v Value) ToInteger() (int32, error) {
	switch x := v.v.(type) {
	case int32:
		return x, nil
	case string:
		// Accept floating-point but truncate to int32 as per CE spec.
		f, err := strconv.ParseFloat(x, 64)
		if f > math.MaxInt32 || f < math.MinInt32 {
			return 0, rangeErr(x)
		}
		return int32(f), err
	default:
		return 0, typeErr("Integer", x)
	}
}

// ToString returns a string value. It does NOT encode non-string, use String() for that.
func (v Value) ToString() (string, error) {
	switch x := v.v.(type) {
	case string:
		return x, nil
	default:
		return "", typeErr("String", x)
	}
}

// ToBinary returns a []byte value, decoding from string if necessary.
func (v Value) ToBinary() ([]byte, error) {
	switch x := v.v.(type) {
	case []byte:
		return x, nil
	case string:
		return base64.StdEncoding.DecodeString(x)
	default:
		return nil, typeErr("Binary", x)
	}
}

// ToURIRef returns a url.URL value, decoding from string if necessary.
func (v Value) ToURIRef() (url.URL, error) {
	switch x := v.v.(type) {
	case url.URL:
		return x, nil
	case string:
		u, err := url.Parse(x)
		if err != nil {
			return url.URL{}, err
		}
		return *u, nil
	default:
		return url.URL{}, typeErr("URI-Reference", x)
	}
}

// ToTimestamp returns a Timestamp value, decoding from string if necessary.
func (v Value) ToTimestamp() (time.Time, error) {
	switch x := v.v.(type) {
	case time.Time:
		return x, nil
	case string:
		ts, err := time.Parse(time.RFC3339Nano, x)
		if err != nil {
			return time.Time{}, err
		}
		return ts, nil
	default:
		return time.Time{}, typeErr("Timestamp", x)
	}
}

func rangeErr(v interface{}) error {
	return fmt.Errorf("%v is out of range for Integer", v)
}

func typeErr(name string, v interface{}) error {
	return fmt.Errorf("cannot convert %T to %s", v, name)
}

func valueErr(name string, v interface{}) error {
	if name == "" {
		return fmt.Errorf("invalid CloudEvents value: %#v", v)
	}
	return fmt.Errorf("invalid CloudEvents %s: %#v", name, v)
}
