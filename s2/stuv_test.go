package s2

import (
	"testing"

	"github.com/golang/geo/r3"
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

func TestFaceXYZtoUVW(t *testing.T) {
	var (
		origin = Point{r3.Vector{0, 0, 0}}
		posX   = Point{r3.Vector{1, 0, 0}}
		negX   = Point{r3.Vector{-1, 0, 0}}
		posY   = Point{r3.Vector{0, 1, 0}}
		negY   = Point{r3.Vector{0, -1, 0}}
		posZ   = Point{r3.Vector{0, 0, 1}}
		negZ   = Point{r3.Vector{0, 0, -1}}
	)

	for face := 0; face < 6; face++ {
		if got := faceXYZtoUVW(face, origin); got != origin {
			t.Errorf("faceXYZtoUVW(%d, %v) = %v, want %v", face, origin, got, origin)
		}

		if got := faceXYZtoUVW(face, uAxis(face)); got != posX {
			t.Errorf("faceXYZtoUVW(%d, %v) = %v, want %v", face, uAxis(face), got, posX)
		}

		if got := faceXYZtoUVW(face, Point{uAxis(face).Mul(-1)}); got != negX {
			t.Errorf("faceXYZtoUVW(%d, %v) = %v, want %v", face, uAxis(face).Mul(-1), got, negX)
		}

		if got := faceXYZtoUVW(face, vAxis(face)); got != posY {
			t.Errorf("faceXYZtoUVW(%d, %v) = %v, want %v", face, vAxis(face), got, posY)
		}

		if got := faceXYZtoUVW(face, Point{vAxis(face).Mul(-1)}); got != negY {
			t.Errorf("faceXYZtoUVW(%d, %v) = %v, want %v", face, vAxis(face).Mul(-1), got, negY)
		}

		if got := faceXYZtoUVW(face, unitNorm(face)); got != posZ {
			t.Errorf("faceXYZtoUVW(%d, %v) = %v, want %v", face, unitNorm(face), got, posZ)
		}

		if got := faceXYZtoUVW(face, Point{unitNorm(face).Mul(-1)}); got != negZ {
			t.Errorf("faceXYZtoUVW(%d, %v) = %v, want %v", face, unitNorm(face).Mul(-1), got, negZ)
		}
	}
}

func TestUVWAxis(t *testing.T) {
	for face := 0; face < 6; face++ {
		// Check that the axes are consistent with faceUVtoXYZ.
		if faceUVToXYZ(face, 1, 0).Sub(faceUVToXYZ(face, 0, 0)) != uAxis(face).Vector {
			t.Errorf("face 1,0 - face 0,0 should equal uAxis")
		}
		if faceUVToXYZ(face, 0, 1).Sub(faceUVToXYZ(face, 0, 0)) != vAxis(face).Vector {
			t.Errorf("faceUVToXYZ(%d, 0, 1).Sub(faceUVToXYZ(%d, 0, 0)) != vAxis(%d), should be equal.", face, face, face)
		}
		if faceUVToXYZ(face, 0, 0) != unitNorm(face).Vector {
			t.Errorf("faceUVToXYZ(%d, 0, 0) != unitNorm(%d), should be equal", face, face)
		}

		// Check that every face coordinate frame is right-handed.
		if got := uAxis(face).Vector.Cross(vAxis(face).Vector).Dot(unitNorm(face).Vector); got != 1 {
			t.Errorf("right-handed check failed. got %d, want 1", got)
		}

		// Check that GetUVWAxis is consistent with GetUAxis, GetVAxis, GetNorm.
		if uAxis(face) != uvwAxis(face, 0) {
			t.Errorf("uAxis(%d) != uvwAxis(%d, 0), should be equal", face, face)
		}
		if vAxis(face) != uvwAxis(face, 1) {
			t.Errorf("vAxis(%d) != uvwAxis(%d, 1), should be equal", face, face)
		}
		if unitNorm(face) != uvwAxis(face, 2) {
			t.Errorf("unitNorm(%d) != uvwAxis(%d, 2), should be equal", face, face)
		}
	}
}
