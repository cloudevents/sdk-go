package types_test

import (
	"fmt"
	"math"
	"net/url"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/cloudevents/sdk-go/v1/cloudevents/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Example() {
	// Handle a time value that may be in native or canonical string form.
	printTime := func(v interface{}) {
		t, err := types.ToTime(v)
		fmt.Printf("%v %v\n", t, err)
	}
	printTime(time.Date(1969, 3, 21, 12, 24, 0, 0, time.UTC))
	printTime("2020-03-21T12:34:56.78Z")

	// Convert numeric values to common 32-bit integer form
	printInt := func(v interface{}) {
		i, err := types.ToInteger(v)
		fmt.Printf("%v %v\n", i, err)
	}
	printInt(123.456)
	printInt("456")
	printInt(int64(99999))
	// But not illegal or out-of-range values
	printInt(math.MaxInt32 + 1)
	printInt("not an int")

	// OUTPUT:
	// 1969-03-21 12:24:00 +0000 UTC <nil>
	// 2020-03-21 12:34:56.78 +0000 UTC <nil>
	// 123 <nil>
	// 456 <nil>
	// 99999 <nil>
	// 0 cannot convert 2147483648 to int32: out of range
	// 0 strconv.ParseFloat: parsing "not an int": invalid syntax
}

var (
	testURL    = &url.URL{Scheme: "http", Host: "example.com", Path: "/foo"}
	testURLstr = "http://example.com/foo"
	timeStr    = "2020-03-21T12:34:56.78Z"
	someTime   = time.Date(2020, 3, 21, 12, 34, 56, 780000000, time.UTC)
)

type valueTester struct {
	testing.TB
	convertFn interface{}
}

// Call types.To... function, use reflection since return types differ.
func (t valueTester) convert(v interface{}) (interface{}, error) {
	rf := reflect.ValueOf(t.convertFn)
	args := []reflect.Value{reflect.ValueOf(v)}
	if v == nil {
		args[0] = reflect.Zero(rf.Type().In(0)) // Avoid the zero argument reflection trap.
	}
	result := rf.Call(args)
	err, _ := result[1].Interface().(error)
	return result[0].Interface(), err
}

// Verify round trip: convertible -> wrapped -> string -> wrapped
func (t *valueTester) ok(in, want interface{}, wantStr string) {
	t.Helper()
	got, err := types.Validate(in)
	require.NoError(t, err)
	assert.Equal(t, want, got)

	gotStr, err := types.Format(in)
	require.NoError(t, err)
	assert.Equal(t, wantStr, gotStr)

	x, err := t.convert(gotStr)
	assert.NoError(t, err)
	x2, err := types.Validate(x)
	assert.NoError(t, err)
	assert.Equal(t, want, x2)
}

// Verify round trip with exception: convertible -> wrapped -> string -> different wrapped
func (t *valueTester) okWithDifferentFromString(in, want interface{}, wantStr string, wantAfterStr interface{}) {
	t.Helper()
	got, err := types.Validate(in)
	require.NoError(t, err)
	assert.Equal(t, want, got)

	gotStr, err := types.Format(in)
	require.NoError(t, err)
	assert.Equal(t, wantStr, gotStr)

	x, err := t.convert(gotStr)
	assert.NoError(t, err)
	x2, err := types.Validate(x)
	assert.NoError(t, err)
	assert.Equal(t, wantAfterStr, x2)
}

// Verify expected error.
func (t *valueTester) err(in interface{}, wantErr string) {
	t.Helper()
	_, err := t.convert(in)
	assert.EqualError(t, err, wantErr)
}

// Verify string->value conversion.
func (t *valueTester) str(str string, want interface{}) {
	t.Helper()
	got, err := t.convert(str)
	assert.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestBool(t *testing.T) {
	x := valueTester{t, types.ToBool}
	x.ok(true, true, "true")
	x.ok(false, false, "false")

	x.err("notabool", "strconv.ParseBool: parsing \"notabool\": invalid syntax")
	x.err(0, "cannot convert 0 to bool")
	x.err(nil, "invalid CloudEvents value: <nil>")
}

func TestInteger(t *testing.T) {
	x := valueTester{t, types.ToInteger}
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
	i := new(uint16)
	*i = 24 // non-nil pointers allowed
	x.ok(i, int32(24), "24")

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

	x.err(math.MaxInt32+1, "cannot convert 2147483648 to int32: out of range")
	x.err(uint32(math.MaxInt32+1), "cannot convert 0x80000000 to int32: out of range")
	x.err(int64(math.MaxInt32+1), "cannot convert 2147483648 to int32: out of range")
	x.err(int64(math.MinInt32-1), "cannot convert -2147483649 to int32: out of range")
	x.err(float64(math.MinInt32-1), "cannot convert -2.147483649e+09 to int32: out of range")
	x.err(float64(math.MaxInt32+1), "cannot convert 2.147483648e+09 to int32: out of range")
	// Float32 doesn't keep all the bits of an int32 so we need to exaggerate fof range error.
	x.err(float64(2*math.MinInt32), "cannot convert -4.294967296e+09 to int32: out of range")
	x.err(float64(-2*math.MaxInt32), "cannot convert -4.294967294e+09 to int32: out of range")

	x.err("X", "strconv.ParseFloat: parsing \"X\": invalid syntax")
	x.err(true, "cannot convert true to int32")
	x.err(nil, "invalid CloudEvents value: <nil>")
}

func TestString(t *testing.T) {
	x := valueTester{t, types.ToString}
	x.ok("hello", "hello", "hello")
	s := new(string)
	*s = "foo" // non-nil pointers allowed
	x.ok(s, "foo", "foo")
}

func TestBinary(t *testing.T) {
	x := valueTester{t, types.ToBinary}
	x.ok([]byte("hello"), []byte("hello"), "aGVsbG8=")
	x.ok([]byte{}, []byte{}, "")
	// Asymmetic case: ToBinary([]byte(nil)) returns []byte(nil),
	// but ToBinary("") returns []byte{}
	// Logically equivalent but not assert.Equal().
	x.str("", []byte{})

	x.err("XXX", "illegal base64 data at input byte 0")
	x.err(nil, "invalid CloudEvents value: <nil>")
}

func TestURL(t *testing.T) {
	x := valueTester{t, types.ToURL}
	x.ok(testURL, types.URI{*testURL}, testURLstr)
	x.ok(*testURL, types.URI{*testURL}, testURLstr)
	x.okWithDifferentFromString(types.URLRef{URL: *testURL}, types.URIRef{*testURL}, testURLstr, types.URI{*testURL})
	x.okWithDifferentFromString(&types.URLRef{URL: *testURL}, types.URIRef{*testURL}, testURLstr, types.URI{*testURL})
	x.okWithDifferentFromString(types.URIRef{URL: *testURL}, types.URIRef{*testURL}, testURLstr, types.URI{*testURL})
	x.okWithDifferentFromString(&types.URIRef{URL: *testURL}, types.URIRef{*testURL}, testURLstr, types.URI{*testURL})
	x.ok(types.URI{URL: *testURL}, types.URI{*testURL}, testURLstr)
	x.ok(&types.URI{URL: *testURL}, types.URI{*testURL}, testURLstr)

	x.str("http://hello/world", &url.URL{Scheme: "http", Host: "hello", Path: "/world"})
	x.str("/world", &url.URL{Path: "/world"})
	x.str("world", &url.URL{Path: "world"})

	x.err("%bad %url", "parse %bad %url: invalid URL escape \"%ur\"")
	x.err(nil, "invalid CloudEvents value: <nil>")
	x.err((*url.URL)(nil), "invalid CloudEvents value: (*url.URL)(nil)")
	x.err((*types.URIRef)(nil), "invalid CloudEvents value: (*types.URIRef)(nil)")
}

func TestTime(t *testing.T) {
	x := valueTester{t, types.ToTime}
	x.ok(someTime, types.Timestamp{someTime}, timeStr)
	x.ok(&someTime, types.Timestamp{someTime}, timeStr)
	x.ok(types.Timestamp{someTime}, types.Timestamp{someTime}, timeStr)
	x.ok(&types.Timestamp{someTime}, types.Timestamp{someTime}, timeStr)

	x.str(timeStr, someTime)

	x.err(nil, "invalid CloudEvents value: <nil>")
	x.err(5, "cannot convert 5 to time.Time")
	x.err((*time.Time)(nil), "invalid CloudEvents value: (*time.Time)(nil)")
	x.err((*types.Timestamp)(nil), "invalid CloudEvents value: (*types.Timestamp)(nil)")
	x.err("not a time", "parsing time \"not a time\" as \"2006-01-02T15:04:05.999999999Z07:00\": cannot parse \"not a time\" as \"2006\"")
}

func TestIncompatible(t *testing.T) {
	// Values that won't convert at all.
	x := valueTester{t, types.Validate}
	x.err(nil, "invalid CloudEvents value: <nil>")
	x.err(complex(0, 0), "invalid CloudEvents value: (0+0i)")
	x.err(map[string]interface{}{}, "invalid CloudEvents value: map[string]interface {}{}")
	x.err(struct{ i int }{i: 9}, "invalid CloudEvents value: struct { i int }{i:9}")
	x.err((*int32)(nil), "invalid CloudEvents value: (*int32)(nil)")
}
