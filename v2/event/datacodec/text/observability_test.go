package text

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
