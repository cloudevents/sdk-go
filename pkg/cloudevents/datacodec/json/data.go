package json

import (
	"encoding/json"
	"fmt"
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
	if err := json.Unmarshal(b, out); err != nil {
		return fmt.Errorf("[json] found bytes, but failed to unmarshal: %s", err.Error())
	}
	return nil
}

func Encode(in interface{}) ([]byte, error) {
	return json.Marshal(in)
}
