package json

import (
	"encoding/json"
	"fmt"
	"strconv"
)

func Decode(in, out interface{}) error {
	if in == nil {
		return nil
	}

	b, ok := in.([]byte)
	if !ok {
		var err error
		b, err = json.Marshal(in)
		if err != nil {
			return fmt.Errorf("[json] failed to marshal in: %s", err.Error())
		}
	}

	if len(b) > 1 && (b[0] == byte('"') || (b[0] == byte('\\') && b[1] == byte('"'))) {
		s, err := strconv.Unquote(string(b))
		if err != nil {
			return err
		}
		if len(s) > 0 && s[0] == '{' {
			// looks like json, use it
			b = []byte(s)
		}
	}

	if err := json.Unmarshal(b, out); err != nil {
		return fmt.Errorf("[json] found bytes %q, but failed to unmarshal: %s", string(b), err.Error())
	}
	return nil
}

func Encode(in interface{}) ([]byte, error) {
	if b, ok := in.([]byte); ok {
		// check to see if it is a pre-encoded byte string.
		if len(b) > 0 && b[0] == byte('"') {
			return b, nil
		}
	}

	return json.Marshal(in)
}
