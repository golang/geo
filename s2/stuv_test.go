/*
Copyright 2014 Google Inc. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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

func TestFaceXYZToUV(t *testing.T) {
	var (
		point    = Point{r3.Vector{1.1, 1.2, 1.3}}
		pointNeg = Point{r3.Vector{-1.1, -1.2, -1.3}}
	)

	tests := []struct {
		face  int
		point Point
		u     float64
		v     float64
		ok    bool
	}{
		{0, point, 1.09090909090909, 1.18181818181818, true},
		{0, pointNeg, 0, 0, false},
		{1, point, -0.916666666666666, 1.08333333333333, true},
		{1, pointNeg, 0, 0, false},
		{2, point, -0.846153846153846, -0.923076923076923, true},
		{2, pointNeg, 0, 0, false},
		{3, point, 0, 0, false},
		{3, pointNeg, 1.18181818181818, 1.09090909090909, true},
		{4, point, 0, 0, false},
		{4, pointNeg, 1.08333333333333, -0.91666666666666, true},
		{5, point, 0, 0, false},
		{5, pointNeg, -0.923076923076923, -0.846153846153846, true},
	}

	for _, test := range tests {
		if u, v, ok := faceXYZToUV(test.face, test.point); !float64Eq(u, test.u) || !float64Eq(v, test.v) || ok != test.ok {
			t.Errorf("faceXYZToUV(%d, %v) = %f, %f, %t, want %f, %f, %t", test.face, test.point, u, v, ok, test.u, test.v, test.ok)
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
