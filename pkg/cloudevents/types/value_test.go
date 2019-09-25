package types_test

import (
	"fmt"
	"math"
	"net/url"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/cloudevents/sdk-go/pkg/cloudevents/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func ExampleValue() {
	// Convert canonical time string to time
	ts, err := types.ValueOfString("2020-03-21T12:34:56.78Z").ToTimestamp()
	fmt.Printf("%v (%v)\n", ts, err)

	// Then back to string
	fmt.Printf("%q\n", types.ValueOfTimestamp(ts).String())

	// Canonical string of binary value is base64
	v := types.ValueOfBinary([]byte{1, 2, 3, 4})
	fmt.Printf("%T(%v) %q\n", v.Interface(), v.Interface(), v.String())

	// Decode base64 to []byte
	b, err := types.ValueOfString("AQIDBA==").ToBinary()
	fmt.Printf("%T(%v) (%v)\n", b, b, err)

	// Convert float to Integer
	v, err = types.ValueOf(123.456)
	fmt.Printf("%T(%v) %q (%v)\n", v.Interface(), v.Interface(), v.String(), err)

	// Integer conversions are range-checked
	_, err = types.ValueOf(math.MaxUint32)
	fmt.Println(err)

	// Consistent treatment for native or string-encoded values.
	asInt := func(x interface{}) {
		v, _ := types.ValueOf(x)
		i, err := v.ToInteger()
		fmt.Printf("asInt: %v (%v)\n", i, err)
	}
	asInt(42)
	asInt("42")
	asInt("notanint")

	// OUTPUT:
	// 2020-03-21 12:34:56.78 +0000 UTC (<nil>)
	// "2020-03-21T12:34:56.78Z"
	// []uint8([1 2 3 4]) "AQIDBA=="
	// []uint8([1 2 3 4]) (<nil>)
	// int32(123) "123" (<nil>)
	// 4294967295 is out of range for Integer
	// asInt: 42 (<nil>)
	// asInt: 42 (<nil>)
	// asInt: 0 (strconv.ParseFloat: parsing "notanint": invalid syntax)
}

var (
	uriRef    = url.URL{Scheme: "http", Host: "example.com", Path: "/foo"}
	uriRefStr = "http://example.com/foo"
	timeStr   = "2020-03-21T12:34:56.78Z"
	someTime  = func() time.Time {
		tm := types.ParseTimestamp(timeStr)
		return tm.Time
	}()
)

type valueTester struct {
	testing.TB
	convertFn interface{}
}

// Call types.To... function, use reflection since return types differ.
func (t valueTester) convert(v interface{}) (interface{}, error) {
	args := []reflect.Value{reflect.ValueOf(v)}
	result := reflect.ValueOf(t.convertFn).Call(args)
	err, _ := result[1].Interface().(error)
	return result[0].Interface(), err
}

// Verify round trip: convertible -> preferred -> string -> preferred
func (t *valueTester) ok(in, want interface{}, wantStr string) {
	t.Helper()
	v, err := types.ValueOf(in)
	require.NoError(t, err)
	assert.Equal(t, want, v.Interface())
	assert.Equal(t, wantStr, v.String())
	x, err := t.convert(v)
	assert.NoError(t, err)
	assert.Equal(t, want, x)
	sv, err := types.ValueOf(wantStr)
	assert.NoError(t, err)
	got, err := t.convert(sv) // String back to value
	assert.NoError(t, err)
	assert.Equal(t, want, got)
}

// Verify expected error.
func (t *valueTester) err(in interface{}, wantErr string) {
	t.Helper()
	v, err := types.ValueOf(in)
	if err != nil {
		assert.EqualError(t, err, wantErr)
		assert.Nil(t, v.Interface())
	} else {
		_, err = t.convert(v)
		assert.EqualError(t, err, wantErr)
	}
}

// Verify string->value conversion.
func (t *valueTester) str(str string, want interface{}) {
	t.Helper()
	v, _ := types.ValueOf(str)
	got, err := t.convert(v)
	assert.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestBool(t *testing.T) {
	x := valueTester{t, types.Value.ToBool}
	x.ok(true, true, "true")
	x.ok(false, false, "false")

	x.err("notabool", "strconv.ParseBool: parsing \"notabool\": invalid syntax")
	x.err(0, "cannot convert int32 to Bool")
	x.err(nil, "<nil> is not a CloudEvents type")
}

func TestInteger(t *testing.T) {
	x := valueTester{t, types.Value.ToInteger}
	x.ok(42, int32(42), "42")
	x.ok(int8(-8), int32(-8), "-8")
	x.ok(int16(-16), int32(-16), "-16")
	x.ok(int32(-32), int32(-32), "-32")
	x.ok(int64(-64), int32(-64), "-64")
	x.ok(uint(1), int32(1), "1")
	x.ok(uint8(8), int32(8), "8")
	x.ok(uint16(16), int32(16), "16")
	x.ok(uint32(32), int32(32), "32")
	x.ok(uint64(64), int32(64), "64")
	x.ok(float32(123.4), int32(123), "123")
	x.ok(float64(-567.8), int32(-567), "-567")

	x.ok(math.MaxInt32, int32(math.MaxInt32), strconv.Itoa(math.MaxInt32))
	x.ok(math.MinInt32, int32(math.MinInt32), strconv.Itoa(math.MinInt32))
	x.ok(int64(math.MinInt32), int32(math.MinInt32), strconv.Itoa(math.MinInt32))
	x.ok(uint32(math.MaxInt32), int32(math.MaxInt32), strconv.Itoa(math.MaxInt32))
	x.ok(uint64(math.MaxInt32), int32(math.MaxInt32), strconv.Itoa(math.MaxInt32))
	x.ok(float64(math.MaxInt32), int32(math.MaxInt32), strconv.Itoa(math.MaxInt32))
	x.ok(float64(math.MinInt32), int32(math.MinInt32), strconv.Itoa(math.MinInt32))

	x.str("123.456", int32(123))
	x.str("-123.456", int32(-123))
	x.str(".9", int32(0))
	x.str("-.9", int32(0))

	x.err(math.MaxInt32+1, "2147483648 is out of range for Integer")
	x.err(uint32(math.MaxInt32+1), "2147483648 is out of range for Integer")
	x.err(int64(math.MaxInt32+1), "2147483648 is out of range for Integer")
	x.err(int64(math.MinInt32-1), "-2147483649 is out of range for Integer")
	x.err(float64(math.MinInt32-1), "-2.147483649e+09 is out of range for Integer")
	x.err(float64(math.MaxInt32+1), "2.147483648e+09 is out of range for Integer")
	// Float32 doesn't keep all the bits of an int32 so we need to exaggerate fof range error.
	x.err(float64(2*math.MinInt32), "-4.294967296e+09 is out of range for Integer")
	x.err(float64(-2*math.MaxInt32), "-4.294967294e+09 is out of range for Integer")

	x.err("X", "strconv.ParseFloat: parsing \"X\": invalid syntax")
	x.err(true, "cannot convert bool to Integer")
	x.err(nil, "<nil> is not a CloudEvents type")
}

func TestString(t *testing.T) {
	x := valueTester{t, types.Value.ToString}
	x.ok("hello", "hello", "hello")
}

func TestBinary(t *testing.T) {
	x := valueTester{t, types.Value.ToBinary}
	x.ok([]byte("hello"), []byte("hello"), "aGVsbG8=")
	x.ok([]byte{}, []byte{}, "")
	// Asymmetic case: ToBinary([]byte(nil)) returns []byte(nil),
	// but ToBinary("") returns []byte{}
	// Logically equivalent but not assert.Equal().
	x.str("", []byte{})

	x.err("XXX", "illegal base64 data at input byte 0")
	x.err(nil, "<nil> is not a CloudEvents type")
}

func TestURIRef(t *testing.T) {
	x := valueTester{t, types.Value.ToURIRef}
	x.ok(uriRef, uriRef, uriRefStr)
	x.ok(&uriRef, uriRef, uriRefStr)
	x.ok(types.URLRef{URL: uriRef}, uriRef, uriRefStr)
	x.ok(&types.URLRef{URL: uriRef}, uriRef, uriRefStr)
	x.ok(types.URI{uriRef}, uriRef, uriRefStr)
	x.ok(&types.URI{uriRef}, uriRef, uriRefStr)

	x.str("http://hello/world", url.URL{Scheme: "http", Host: "hello", Path: "/world"})
	x.str("/world", url.URL{Path: "/world"})
	x.str("world", url.URL{Path: "world"})

	x.err("%bad %url", "parse %bad %url: invalid URL escape \"%ur\"")
	x.err(nil, "<nil> is not a CloudEvents type")
	x.err((*url.URL)(nil), "invalid CloudEvents value: (*url.URL)(nil)")
	x.err((*types.URIRef)(nil), "invalid CloudEvents value: (*types.URIRef)(nil)")
}

func TestTimestamp(t *testing.T) {
	x := valueTester{t, types.Value.ToTimestamp}
	x.ok(someTime, someTime, timeStr)
	x.ok(&someTime, someTime, timeStr)
	x.ok(someTime, someTime, timeStr)
	x.ok(&someTime, someTime, timeStr)

	x.str(timeStr, someTime)

	x.err(nil, "<nil> is not a CloudEvents type")
	x.err(5, "cannot convert int32 to Timestamp")
	x.err((*time.Time)(nil), "invalid CloudEvents value: (*time.Time)(nil)")
	x.err((*types.Timestamp)(nil), "invalid CloudEvents value: (*types.Timestamp)(nil)")
	x.err("not a time", "parsing time \"not a time\" as \"2006-01-02T15:04:05.999999999Z07:00\": cannot parse \"not a time\" as \"2006\"")
}

func TestIncompatible(t *testing.T) {
	// Values that won't convert at all.
	x := valueTester{t, nil}
	x.err(nil, "<nil> is not a CloudEvents type")
	x.err(complex(0, 0), "complex128 is not a CloudEvents type")
	x.err(map[string]interface{}{}, "map[string]interface {} is not a CloudEvents type")
	x.err(struct{ i int }{i: 9}, "struct { i int } is not a CloudEvents type")
	x.err((*int32)(nil), "*int32 is not a CloudEvents type")
}
