package s2

import (
	"math"

	"code.google.com/p/gos2/r3"
)

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
	case 5:
		return r3.Vector{v, u, -1}
	}
	return r3.Vector{0, 0, 0}
}
