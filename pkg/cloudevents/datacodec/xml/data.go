package xml

import (
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"log"
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
			return fmt.Errorf("[xml] failed to marshal in: %s", err.Error())
		}
	}

	// If the message is encoded as a base64 block as a string, we need to
	// decode that first before trying to unmarshal the bytes
	if len(b) > 0 && b[0] == byte('"') {
		bs, err := base64.StdEncoding.DecodeString(string(b[1 : len(b)-1]))
		if err != nil {
			return err
		}
		b = bs

		log.Printf("popping quotes made %s", string(b))
	}

	if err := xml.Unmarshal(b, out); err != nil {
		return fmt.Errorf("[xml] found bytes, but failed to unmarshal: %s %s", err.Error(), string(b))
	}
	return nil
}

func Encode(in interface{}) ([]byte, error) {
	if b, ok := in.([]byte); ok {
		log.Printf("asked to encode bytes... wrong? %s", string(b))
	}

	return xml.Marshal(in)
}
