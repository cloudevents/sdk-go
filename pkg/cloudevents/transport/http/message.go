package http

import (
	"fmt"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport"
	"net/http"
	"net/url"
	"time"
)

// type check that this transport message impl matches the contract
var _ transport.Message = (*Message)(nil)

type Message struct {
	Header http.Header
	Body   []byte
}

func (m Message) CloudEventVersion() string {
	return ""
}

//func (m *Message) ContextAttributes() []string {
//	return nil
//}

func (m Message) Get(key string) (interface{}, bool) {
	return nil, false
}

func (m *Message) Set(key string, value interface{}) error {
	return fmt.Errorf("not implemented")
}

func (m Message) GetInt(key string) (int32, bool) {
	return 0, false
}

func (m *Message) SetInt(key string, value int32) error {
	return fmt.Errorf("not implemented")
}

func (m Message) GetString(key string) (string, bool) {
	return "", false
}

func (m *Message) SetString(key string, value string) error {
	return fmt.Errorf("not implemented")
}

func (m Message) GetBinary(key string) ([]byte, bool) {
	return nil, false
}

func (m *Message) SetBinary(key string, value string) error {
	return fmt.Errorf("not implemented")
}

func (m Message) GetMap(key string) (map[string]interface{}, bool) {
	return nil, false
}

func (m *Message) SetMap(key string, value map[string]interface{}) error {
	return fmt.Errorf("not implemented")
}

func (m Message) GetTime(key string) (time.Time, bool) {
	return time.Time{}, false
}

func (m *Message) SetTime(key string, value time.Time) error {
	return fmt.Errorf("not implemented")
}

func (m Message) GetURL(key string) (url.URL, bool) {
	return url.URL{}, false
}

func (m *Message) SetURL(key string, value url.URL) error {
	return fmt.Errorf("not implemented")
}
