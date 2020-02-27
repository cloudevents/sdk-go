package nats

import "fmt"

var BinaryEncodingNotSupported = fmt.Errorf("binary encoding for NATS protocol is not supported")
