/*
Package convert provides conversions specified by the CloudEvents type system.

Each CloudEvents type has a preferred Go type, a set of convertible Go types,
and a canonical string-encoding. This package provides conversions between these
representations.

In the table below "types" refers to package "github.com/cloudevents/sdk-go/pkg/cloudevents/types".

+----------------+---------+-----------------------------------+
|CloudEvents Type|Go type  |Compatible types                   |
+----------------+---------+-----------------------------------+
|Bool            |bool     |bool                               |
+----------------+---------+-----------------------------------+
|Integer         |int32    |Any numeric type with value in     |
|                |         |range of int32                     |
+----------------+---------+-----------------------------------+
|String          |string   |string                             |
+----------------+---------+-----------------------------------+
|Binary          |[]byte   |[]byte                             |
+----------------+---------+-----------------------------------+
|URI             |url.URL  |url.URL, types.URIRef.             |
|                |         |Must be an absolute URI.           |
+----------------+---------+-----------------------------------+
|URI-Reference   |url.URL  |url.URL, types.URIRef              |
+----------------+---------+-----------------------------------+
|Timestamp       |time.Time|time.Time, types.Timestamp         |
+----------------+---------+-----------------------------------+

*/
package convert

import (
	"encoding/base64"
	"errors"
	"fmt"
	"math"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/cloudevents/sdk-go/pkg/cloudevents/types"
)

// Bool converts from bool or a "true"/"false" string.
func ToBool(v interface{}) (ret bool, err error) {
	name := "Bool"
	switch v := v.(type) {
	case bool:
		ret = v
	case string:
		switch v {
		case "true":
			ret = true
		case "false":
			ret = false
		default:
			err = valueErr(name, v)
		}
	default:
		err = typeErr(name, v)
	}
	return ret, err
}

// Integer converts from any numeric value that fits in range of int32 or
// a string in strconv.Atoi format.
func ToInteger(v interface{}) (ret int32, err error) {
	name := "Integer"
	switch v := v.(type) {
	case uint, uintptr, uint8, uint16, uint32, uint64:
		u := reflect.ValueOf(v).Uint()
		if u > math.MaxInt32 {
			err = valueErr(name, v)
		}
		ret = int32(u)
	case int, int8, int16, int32, int64:
		i := reflect.ValueOf(v).Int()
		if i > math.MaxInt32 || i < math.MinInt32 {
			err = valueErr(name, v)
		}
		ret = int32(i)
	case float32, float64:
		f := reflect.ValueOf(v).Float()
		if f > math.MaxInt32 || f < math.MinInt32 {
			err = valueErr(name, v)
		}
		ret = int32(f)
	case string:
		// Accept floating-point but truncate to int32 as per CE spec.
		var f float64
		if f, err = strconv.ParseFloat(v, 64); err == nil {
			ret, err = ToInteger(f)
		}
		if err != nil {
			err = valueErr(name, v) // Use original v in error message.
		}
	default:
		err = typeErr(name, v)
	}
	return ret, err
}

// Binary converts from []byte or base64-encoded string.
//
// NOTE: if v is []byte, the same []byte is returned, bytes are not copied.
func ToBinary(v interface{}) (ret []byte, err error) {
	name := "Binary"
	switch v := v.(type) {
	case []byte:
		ret = v
	case string:
		ret, err = base64.StdEncoding.DecodeString(v)
		if err != nil {
			err = valueErr(name, v)
		}
	default:
		err = typeErr(name, v)
	}
	return ret, err
}

// URIReference converts from url.URL, types.URIref or string.
func ToURIReference(v interface{}) (ret url.URL, err error) {
	name := "URI-Reference"
	switch v := v.(type) {
	case url.URL:
		ret = v
	case *url.URL:
		if v == nil {
			err = valueErr(name, v)
		} else {
			ret = *v
		}
	case types.URIRef:
		ret = v.URL
	case *types.URIRef:
		if v == nil {
			err = valueErr(name, v)
		} else {
			ret = v.URL
		}
	case types.URLRef:
		ret = v.URL
	case *types.URLRef:
		if v == nil {
			err = valueErr(name, v)
		} else {
			ret = v.URL
		}
	case string:
		u, err2 := url.Parse(v)
		if err2 != nil {
			err = valueErr(name, v)
		} else {
			ret = *u
		}
	default:
		err = typeErr(name, v)
	}
	return ret, err
}

// URI converts from url.URL, types.URIref or string. The URI must be absolute.
func ToURI(v interface{}) (ret url.URL, err error) {
	name := "URI"
	ret, err = ToURIReference(v)
	if err != nil {
		err = errors.New(strings.Replace(err.Error(), "URI-Reference", name, -1))
	} else if !ret.IsAbs() {
		err = fmt.Errorf("%s: %s", valueErr(name, v), "not an absolute URI")
	}
	return ret, err
}

// Timestamp converts from time.Time or string in time.RFC3339Nano format.
func ToTimestamp(v interface{}) (ret time.Time, err error) {
	name := "Timestamp"
	switch v := v.(type) {
	case time.Time:
		ret = v
	case *time.Time:
		if v == nil {
			err = valueErr(name, v)
		} else {
			ret = *v
		}
	case types.Timestamp:
		ret = v.Time
	case *types.Timestamp:
		if v == nil {
			err = valueErr(name, v)
		} else {
			ret = v.Time
		}
	case string:
		ret, err = time.Parse(time.RFC3339Nano, v)
		if err != nil {
			err = valueErr(name, v)
		}
	default:
		err = typeErr(name, v)
	}
	return ret, err
}

// Normalize converts v to the preferred types that it is convertible to:
// bool, int32, string, []byte, url.URL or time.Time.
func Normalize(v interface{}) (ret interface{}, err error) {
	switch v.(type) {
	case bool:
		return ToBool(v)
	case uint, uintptr, uint8, uint16, uint32, uint64, int, int8, int16, int32, int64, float32, float64:
		return ToInteger(v)
	case string:
		return v, nil
	case []byte:
		return ToBinary(v)
	case url.URL, *url.URL, types.URLRef, *types.URLRef, types.URIRef, *types.URIRef:
		return ToURIReference(v)

	case time.Time, *time.Time, types.Timestamp, *types.Timestamp:
		return ToTimestamp(v)
	default:
		return nil, fmt.Errorf("%T is not a CloudEvents-compatible type", v)
	}
}

// StringOf returns the canonical string-encoding for v.
func StringOf(v interface{}) (ret string, err error) {
	v, err = Normalize(v)
	if err != nil {
		return "", err
	}
	switch v := v.(type) {
	case bool:
		if v {
			ret = "true"
		} else {
			ret = "false"
		}
	case int32:
		ret = strconv.Itoa(int(v))
	case string:
		ret = v
	case []byte:
		ret = base64.StdEncoding.EncodeToString(v)
	case url.URL:
		ret = v.String()
	case time.Time:
		ret = v.UTC().Format(time.RFC3339Nano)
	default:
		err = fmt.Errorf("%T is not a CloudEvents-compatible type", v)
	}
	return ret, err
}

func typeErr(name string, v interface{}) error {
	return fmt.Errorf("%T is not compatible with CloudEvents %s", v, name)
}

func valueErr(name string, v interface{}) error {
	return fmt.Errorf("invalid value for CloudEvents %s: %#v", name, v)
}
