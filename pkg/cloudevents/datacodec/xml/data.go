package xml

import (
	"encoding/xml"
	"fmt"
)

func Decode(in, out interface{}) error {
	if in == nil {
		return nil
	}

	b, ok := in.([]byte)
	if !ok {
		var err error
		b, err = xml.Marshal(in)
		if err != nil {
			return fmt.Errorf("failed to marshal in: %s", err.Error())
		}
	}
	if err := xml.Unmarshal(b, out); err != nil {
		return fmt.Errorf("found bytes, but failed to unmarshal: %s", err.Error())
	}
	return nil
}
