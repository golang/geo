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

func TestUVNorms(t *testing.T) {
	step := 1 / 1024.0
	for face := 0; face < 6; face++ {
		for x := -1.0; x <= 1; x += step {
			if !float64Eq(float64(faceUVToXYZ(face, x, -1).Cross(faceUVToXYZ(face, x, 1)).Angle(uNorm(face, x))), 0.0) {
				t.Errorf("UNorm not orthogonal to the face(%d)", face)
			}
			if !float64Eq(float64(faceUVToXYZ(face, -1, x).Cross(faceUVToXYZ(face, 1, x)).Angle(vNorm(face, x))), 0.0) {
				t.Errorf("VNorm not orthogonal to the face(%d)", face)
			}
		}
	}

}
