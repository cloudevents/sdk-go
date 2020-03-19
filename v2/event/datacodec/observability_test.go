package datacodec

import "testing"

func TestMethodName(t *testing.T) {
	if reportEncode.MethodName() != "encode" {
		t.Fail()
	}
	if reportDecode.MethodName() != "decode" {
		t.Fail()
	}
	if observed(1337).MethodName() != "unknown" {
		t.Fail()
	}
}

func TestLatencyMs(t *testing.T) {
	if reportEncode.LatencyMs() == nil {
		t.Fail()
	}
	if reportDecode.LatencyMs() == nil {
		t.Fail()
	}
}
