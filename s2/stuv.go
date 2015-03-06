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
	"math"

	"github.com/golang/geo/r3"
)

const (
	// maxSiTi is the maximum value of an si- or ti-coordinate.
	// It is one shift more than maxSize.
	maxSiTi = maxSize << 1
)

// siTiToST converts an si- or ti-value to the corresponding s- or t-value.
// Value is capped at 1.0 because there is no DCHECK in Go.
func siTiToST(si uint64) float64 {
	if si > maxSiTi {
		return 1.0
	}
	return float64(si / maxSiTi)
}

// stToUV converts an s or t value to the corresponding u or v value.
// This is a non-linear transformation from [-1,1] to [-1,1] that
// attempts to make the cell sizes more uniform.
// This uses what the C++ version calls 'the quadratic transform'.
func stToUV(s float64) float64 {
	if s >= 0.5 {
		return (1 / 3.) * (4*s*s - 1)
	}
	return (1 / 3.) * (1 - 4*(1-s)*(1-s))
}

// uvToST is the inverse of the stToUV transformation. Note that it
// is not always true that uvToST(stToUV(x)) == x due to numerical
// errors.
func uvToST(u float64) float64 {
	if u >= 0 {
		return 0.5 * math.Sqrt(1+3*u)
	}
	return 1 - 0.5*math.Sqrt(1-3*u)
}

// face returns face ID from 0 to 5 containing the r. For points on the
// boundary between faces, the result is arbitrary but deterministic.
func face(r r3.Vector) int {
	abs := r.Abs()
	id := 0
	value := r.X
	if abs.Y > abs.X {
		id = 1
		value = r.Y
	}
	if abs.Z > math.Abs(value) {
		id = 2
		value = r.Z
	}
	if value < 0 {
		id += 3
	}
	return id
}

// validFaceXYZToUV given a valid face for the given point r (meaning that
// dot product of r with the face normal is positive), returns
// the corresponding u and v values, which may lie outside the range [-1,1].
func validFaceXYZToUV(face int, r r3.Vector) (float64, float64) {
	switch face {
	case 0:
		return r.Y / r.X, r.Z / r.X
	case 1:
		return -r.X / r.Y, r.Z / r.Y
	case 2:
		return -r.X / r.Z, -r.Y / r.Z
	case 3:
		return r.Z / r.X, r.Y / r.X
	case 4:
		return r.Z / r.Y, -r.X / r.Y
	}
	return -r.Y / r.Z, -r.X / r.Z
}

// xyzToFaceUV converts a direction vector (not necessarily unit length) to
// (face, u, v) coordinates.
func xyzToFaceUV(r r3.Vector) (f int, u, v float64) {
	f = face(r)
	u, v = validFaceXYZToUV(f, r)
	return f, u, v
}

// faceUVToXYZ turns face and UV coordinates into an unnormalized 3 vector.
func faceUVToXYZ(face int, u, v float64) r3.Vector {
	switch face {
	case 0:
		return r3.Vector{1, u, v}
	case 1:
		return r3.Vector{-u, 1, v}
	case 2:
		return r3.Vector{-u, -v, 1}
	case 3:
		return r3.Vector{-1, -v, -u}
	case 4:
		return r3.Vector{v, -1, -u}
	default:
		return r3.Vector{v, u, -1}
	}
}

// faceXYZToUV returns the u and v values (which may lie outside the range
// [-1, 1]) if the dot product of the point p with the given face normal is positive.
func faceXYZToUV(face int, p Point) (u, v float64, ok bool) {
	switch face {
	case 0:
		if p.X <= 0 {
			return 0, 0, false
		}
	case 1:
		if p.Y <= 0 {
			return 0, 0, false
		}
	case 2:
		if p.Z <= 0 {
			return 0, 0, false
		}
	case 3:
		if p.X >= 0 {
			return 0, 0, false
		}
	case 4:
		if p.Y >= 0 {
			return 0, 0, false
		}
	default:
		if p.Z >= 0 {
			return 0, 0, false
		}
	}

	u, v = validFaceXYZToUV(face, p.Vector)
	return u, v, true
}

// faceXYZtoUVW transforms the given point P to the (u,v,w) coordinate frame of the given
// face where the w-axis represents the face normal.
func faceXYZtoUVW(face int, p Point) Point {
	// The result coordinates are simply the dot products of P with the (u,v,w)
	// axes for the given face (see faceUVWAxes).
	switch face {
	case 0:
		return Point{r3.Vector{p.Y, p.Z, p.X}}
	case 1:
		return Point{r3.Vector{-p.X, p.Z, p.Y}}
	case 2:
		return Point{r3.Vector{-p.X, -p.Y, p.Z}}
	case 3:
		return Point{r3.Vector{-p.Z, -p.Y, -p.X}}
	case 4:
		return Point{r3.Vector{-p.Z, p.X, -p.Y}}
	default:
		return Point{r3.Vector{p.Y, p.X, -p.Z}}
	}
}

// uNorm returns the right-handed normal (not necessarily unit length) for an
// edge in the direction of the positive v-axis at the given u-value on
// the given face.  (This vector is perpendicular to the plane through
// the sphere origin that contains the given edge.)
func uNorm(face int, u float64) r3.Vector {
	switch face {
	case 0:
		return r3.Vector{u, -1, 0}
	case 1:
		return r3.Vector{1, u, 0}
	case 2:
		return r3.Vector{1, 0, u}
	case 3:
		return r3.Vector{-u, 0, 1}
	case 4:
		return r3.Vector{0, -u, 1}
	default:
		return r3.Vector{0, -1, -u}
	}
}

// vNorm returns the right-handed normal (not necessarily unit length) for an
// edge in the direction of the positive u-axis at the given v-value on
// the given face.
func vNorm(face int, v float64) r3.Vector {
	switch face {
	case 0:
		return r3.Vector{-v, 0, 1}
	case 1:
		return r3.Vector{0, -v, 1}
	case 2:
		return r3.Vector{0, -1, -v}
	case 3:
		return r3.Vector{v, -1, 0}
	case 4:
		return r3.Vector{1, v, 0}
	default:
		return r3.Vector{1, 0, v}
	}
}

// faceUVWAxes are the U, V, and W axes for each face.
var faceUVWAxes = [6][3]Point{
	{Point{r3.Vector{0, 1, 0}}, Point{r3.Vector{0, 0, 1}}, Point{r3.Vector{1, 0, 0}}},
	{Point{r3.Vector{-1, 0, 0}}, Point{r3.Vector{0, 0, 1}}, Point{r3.Vector{0, 1, 0}}},
	{Point{r3.Vector{-1, 0, 0}}, Point{r3.Vector{0, -1, 0}}, Point{r3.Vector{0, 0, 1}}},
	{Point{r3.Vector{0, 0, -1}}, Point{r3.Vector{0, -1, 0}}, Point{r3.Vector{-1, 0, 0}}},
	{Point{r3.Vector{0, 0, -1}}, Point{r3.Vector{1, 0, 0}}, Point{r3.Vector{0, -1, 0}}},
	{Point{r3.Vector{0, 1, 0}}, Point{r3.Vector{1, 0, 0}}, Point{r3.Vector{0, 0, -1}}},
}

// uvwAxis returns the given axis of the given face.
func uvwAxis(face, axis int) Point {
	return faceUVWAxes[face][axis]
}

// uAxis returns the u-axis for the given face.
func uAxis(face int) Point {
	return uvwAxis(face, 0)
}

// vAxis returns the v-axis for the given face.
func vAxis(face int) Point {
	return uvwAxis(face, 1)
}

// Return the unit-length normal for the given face.
func unitNorm(face int) Point {
	return uvwAxis(face, 2)
}
