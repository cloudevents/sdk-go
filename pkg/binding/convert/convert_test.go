package convert_test

import (
	"math"
	"net/url"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/cloudevents/sdk-go/pkg/binding/convert"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/types"
	"github.com/stretchr/testify/assert"
)

var (
	absURI   = url.URL{Scheme: "http", Host: "example.com", Path: "/foo"}
	relURI   = url.URL{Path: "/foo"}
	timeStr  = "2020-03-21T12:34:56.78Z"
	someTime = func() time.Time {
		tm, err := time.Parse(time.RFC3339Nano, timeStr)
		if err != nil {
			panic(err)
		}
		return tm
	}()
)

type tester struct {
	testing.TB
	convertFn interface{}
}

// Call a convert.To... function, use reflection since return types differ.
func (t *tester) convert(v interface{}) (interface{}, error) {
	args := make([]reflect.Value, 1)
	if v == nil {
		args[0] = reflect.Zero(reflect.TypeOf(t.convertFn).In(0)) // Beware the Zero Value!
	} else {
		args[0] = reflect.ValueOf(v)
	}
	result := reflect.ValueOf(t.convertFn).Call(args)
	err, _ := result[1].Interface().(error)
	return result[0].Interface(), err
}

// Verify round trip: convertible -> preferred -> string -> preferred
func (t *tester) ok(in, want interface{}, wantStr string) {
	t.Helper()
	got, err := t.convert(in) // Direct conversion
	assert.NoError(t, err)
	assert.Equal(t, want, got)

	got, err = convert.Normalize(in) // ValueOf conversion
	assert.NoError(t, err)
	assert.Equal(t, want, got)

	gotStr, err := convert.StringOf(in) // Canonical string
	assert.NoError(t, err)
	assert.Equal(t, wantStr, gotStr)

	got, err = t.convert(gotStr) // String back to value
	assert.NoError(t, err)
	assert.Equal(t, want, got)
}

// Verify expected error.
func (t *tester) err(in interface{}, wantErr string) {
	t.Helper()
	_, err := t.convert(in)
	assert.EqualError(t, err, wantErr)
}

// Verify string->value conversion.
func (t *tester) str(str string, want interface{}) {
	t.Helper()
	got, err := t.convert(str)
	assert.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestBool(t *testing.T) {
	x := tester{t, convert.ToBool}
	x.ok(true, true, "true")
	x.ok(false, false, "false")

	x.err("TRUE", "invalid value for CloudEvents Bool: \"TRUE\"")
	x.err(0, "int is not compatible with CloudEvents Bool")
	x.err(nil, "<nil> is not compatible with CloudEvents Bool")
}

func TestInteger(t *testing.T) {
	x := tester{t, convert.ToInteger}
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

	x.err(math.MaxInt32+1, "invalid value for CloudEvents Integer: 2147483648")
	x.err(uint32(math.MaxInt32+1), "invalid value for CloudEvents Integer: 0x80000000")
	x.err(int64(math.MaxInt32+1), "invalid value for CloudEvents Integer: 2147483648")
	x.err(int64(math.MinInt32-1), "invalid value for CloudEvents Integer: -2147483649")
	x.err(float64(math.MinInt32-1), "invalid value for CloudEvents Integer: -2.147483649e+09")
	x.err(float64(math.MaxInt32+1), "invalid value for CloudEvents Integer: 2.147483648e+09")
	// Float32 doesn't keep all the bits of an int32 so we need to exaggerate for range error.
	x.err(float64(2*math.MinInt32), "invalid value for CloudEvents Integer: -4.294967296e+09")
	x.err(float64(-2*math.MaxInt32), "invalid value for CloudEvents Integer: -4.294967294e+09")

	x.err("X", "invalid value for CloudEvents Integer: \"X\"")
	x.err(true, "bool is not compatible with CloudEvents Integer")
	x.err(nil, "<nil> is not compatible with CloudEvents Integer")
}

func TestString(t *testing.T) {
	// There is no convert.ToString to avoid confusion with convert.StringOf.
	// Not much to test except that ValueOf/StringOf round-trip a string.
	x := tester{t, func(v interface{}) (interface{}, error) { return v.(string), nil }}
	x.ok("hello", "hello", "hello")
}

func TestBinary(t *testing.T) {
	x := tester{t, convert.ToBinary}
	x.ok([]byte("hello"), []byte("hello"), "aGVsbG8=")
	x.ok([]byte{}, []byte{}, "")
	// Asymmetic case: ToBinary([]byte(nil)) returns []byte(nil),
	// but ToBinary("") returns []byte{}
	// Logically equivalent but not assert.Equal().
	x.str("", []byte{})

	x.err("XXX", "invalid value for CloudEvents Binary: \"XXX\"")
	x.err(nil, "<nil> is not compatible with CloudEvents Binary")
}

func TestURIReference(t *testing.T) {
	x := tester{t, convert.ToURIReference}
	x.ok(absURI, absURI, absURI.String())
	x.ok(&relURI, relURI, relURI.String())
	x.ok(types.URIRef{URL: relURI}, relURI, relURI.String())
	x.ok(&types.URIRef{URL: absURI}, absURI, absURI.String())

	x.str("http://hello/world", url.URL{Scheme: "http", Host: "hello", Path: "/world"})
	x.str("/world", url.URL{Path: "/world"})
	x.str("world", url.URL{Path: "world"})

	x.err("%x", "invalid value for CloudEvents URI-Reference: \"%x\"")
	x.err(nil, "<nil> is not compatible with CloudEvents URI-Reference")
	x.err((*url.URL)(nil), "invalid value for CloudEvents URI-Reference: (*url.URL)(nil)")
	x.err((*types.URIRef)(nil), "invalid value for CloudEvents URI-Reference: (*types.URIRef)(nil)")
}

func TestURI(t *testing.T) {
	// ToURI just does ToURIReference and adds an IsAbs() test.

	x := tester{t, convert.ToURI}
	x.ok(absURI, absURI, absURI.String())
	x.ok(&types.URIRef{URL: absURI}, absURI, absURI.String())

	x.str("http://hello/world", url.URL{Scheme: "http", Host: "hello", Path: "/world"})

	x.err("%x", "invalid value for CloudEvents URI: \"%x\"")
	x.err("/world", "invalid value for CloudEvents URI: \"/world\": not an absolute URI")
}

func TestTimestamp(t *testing.T) {
	x := tester{t, convert.ToTimestamp}
	x.ok(someTime, someTime, timeStr)
	x.ok(&someTime, someTime, timeStr)
	x.ok(types.Timestamp{Time: someTime}, someTime, timeStr)
	x.ok(&types.Timestamp{Time: someTime}, someTime, timeStr)

	x.str(timeStr, someTime)

	x.err(nil, "<nil> is not compatible with CloudEvents Timestamp")
	x.err(5, "int is not compatible with CloudEvents Timestamp")
	x.err((*time.Time)(nil), "invalid value for CloudEvents Timestamp: (*time.Time)(nil)")
	x.err((*types.Timestamp)(nil), "invalid value for CloudEvents Timestamp: (*types.Timestamp)(nil)")
	x.err("not a time", "invalid value for CloudEvents Timestamp: \"not a time\"")
}

func TestIncompatible(t *testing.T) {
	// Values that won't convert at all.
	x := tester{t, convert.Normalize}
	x.err(nil, "<nil> is not a CloudEvents-compatible type")
	x.err(complex(0, 0), "complex128 is not a CloudEvents-compatible type")
	x.err(map[string]interface{}{}, "map[string]interface {} is not a CloudEvents-compatible type")
	x.err(struct{ i int }{i: 9}, "struct { i int } is not a CloudEvents-compatible type")
	x.err((*int32)(nil), "*int32 is not a CloudEvents-compatible type")
}
