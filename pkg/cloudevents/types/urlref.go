package types

import (
	"encoding/json"
	"fmt"
	"net/url"
)

type URLRef struct {
	url.URL
}

func ParseURLRef(u string) *URLRef {
	if u == "" {
		return nil
	}
	pu, err := url.Parse(u)
	if err != nil {
		return nil
	}
	return &URLRef{URL: *pu}
}

// This allows json marshaling to always be in RFC3339Nano format.
func (u URLRef) MarshalJSON() ([]byte, error) {
	rfc3339 := fmt.Sprintf("%q", u.String())
	return []byte(rfc3339), nil
}

func (u *URLRef) UnmarshalJSON(b []byte) error {
	var ref string
	if err := json.Unmarshal(b, &ref); err != nil {
		return err
	}
	*u = *ParseURLRef(ref)
	return nil
}
