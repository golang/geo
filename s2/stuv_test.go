package s2

import (
	"testing"
)

func TestSTUV(t *testing.T) {
	if x := stToUV(uvToST(.125)); x != .125 {
		t.Error("stToUV(uvToST(.125) == ", x)
	}
	if x := uvToST(stToUV(.125)); x != .125 {
		t.Error("uvToST(stToUV(.125) == ", x)
	}
}
