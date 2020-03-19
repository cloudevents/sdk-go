package types_test

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
	"time"

	"github.com/cloudevents/sdk-go/v2/types"
	"github.com/stretchr/testify/assert"
)

func TestTimestampParseString(t *testing.T) {
	ok := func(s string, want time.Time) {
		t.Helper()
		got, err := types.ParseTimestamp(s)
		assert.NoError(t, err)
		assert.Equal(t, want, got.Time)
		assert.Equal(t, s, got.String())
	}
	ok("1984-02-28T15:04:05Z", time.Date(1984, 02, 28, 15, 04, 05, 0, time.UTC))
	ok("1984-02-28T15:04:05.999999999Z", time.Date(1984, 02, 28, 15, 04, 05, 999999999, time.UTC))

	// empty string
	{
		got, err := types.ParseTimestamp("")
		assert.NoError(t, err)
		require.Nil(t, got)
	}

	bad := func(s, wanterr string) {
		t.Helper()
		_, err := types.ParseTime(s)
		assert.EqualError(t, err, wanterr)
	}
	bad("", "cannot convert \"\" to time.Time: not in RFC3339 format")
	bad("2019-02-28", "cannot convert \"2019-02-28\" to time.Time: not in RFC3339 format")
}

func TestJsonMarshalUnmarshalTimestamp(t *testing.T) {
	ok := func(ts string) {
		t.Helper()
		tt, err := types.ParseTime(ts)
		assert.NoError(t, err)
		got, err := json.Marshal(types.Timestamp{Time: tt})
		assert.NoError(t, err)
		assert.Equal(t, fmt.Sprintf(`"%s"`, ts), string(got))
	}
	ok("1984-02-28T15:04:05Z")
	ok("1984-02-28T15:04:05.999999999Z")

	bad := func(s, wanterr string) {
		t.Helper()
		var ts types.Timestamp
		err := json.Unmarshal([]byte(s), &ts)
		assert.EqualError(t, err, wanterr)
	}
	bad("", "unexpected end of JSON input")
	bad("2019-02-28", "invalid character '-' after top-level value")
}

func TestJsonMarshalUnmarshalTimestamp_direct(t *testing.T) {
	ok := func(s string) {
		t.Helper()
		tt, err := types.ParseTime(s)
		assert.NoError(t, err)
		ts := &types.Timestamp{Time: tt}
		got, err := ts.MarshalJSON()
		assert.NoError(t, err)
		assert.Equal(t, fmt.Sprintf(`"%s"`, s), string(got))
	}
	ok("1984-02-28T15:04:05Z")
	ok("1984-02-28T15:04:05.999999999Z")

	bad := func(s, wanterr string) {
		t.Helper()
		var ts types.Timestamp
		err := ts.UnmarshalJSON([]byte(s))
		assert.EqualError(t, err, wanterr)
	}
	bad("", "unexpected end of JSON input")
	bad("2019-02-28", "invalid character '-' after top-level value")

	// ok, empty time
	{
		ts := &types.Timestamp{}
		got, err := ts.MarshalJSON()
		assert.NoError(t, err)
		assert.Equal(t, fmt.Sprintf(`"%s"`, ""), string(got))
	}

	// bad bytes
	{
		ts := &types.Timestamp{}
		err := ts.UnmarshalJSON([]byte(`"not a time"`))
		assert.EqualError(t, err, "cannot convert \"not a time\" to time.Time: not in RFC3339 format")
	}

	// incorrect iso
	{
		ts := &types.Timestamp{}
		err := ts.UnmarshalJSON([]byte(`"Mon Jan _2 15:04:05 2006"`))
		assert.EqualError(t, err, "cannot convert \"Mon Jan _2 15:04:05 2006\" to time.Time: not in RFC3339 format")
	}
}

func TestXMLMarshalUnmarshalTimestamp(t *testing.T) {
	ok := func(tstr string) {
		t.Helper()
		tt, err := types.ParseTime(tstr)
		assert.NoError(t, err)
		got, err := xml.Marshal(types.Timestamp{Time: tt})
		assert.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("<Timestamp>%s</Timestamp>", tstr), string(got))
		var ts types.Timestamp
		err = xml.Unmarshal(got, &ts)
		assert.NoError(t, err)
		assert.Equal(t, tt, ts.Time)
	}
	ok("1984-02-28T15:04:05Z")
	ok("1984-02-28T15:04:05.999999999Z")

	bad := func(s, wanterr string) {
		t.Helper()
		var ts types.Timestamp
		err := xml.Unmarshal([]byte(s), &ts)
		assert.EqualError(t, err, wanterr)
	}
	bad("", "EOF")
	bad("2019-02-28", "EOF")
	bad("<Timestamp>2019-02-28</Timestamp>", "cannot convert \"2019-02-28\" to time.Time: not in RFC3339 format")
	bad("<Timestamp></Timestamp>", "cannot convert \"\" to time.Time: not in RFC3339 format")
}
