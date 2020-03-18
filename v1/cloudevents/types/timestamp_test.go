package types_test

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"testing"
	"time"

	"github.com/cloudevents/sdk-go/v1/cloudevents/types"
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
