/*
Copyright 2016 Google Inc. All rights reserved.

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

// This file contains various predicates that are guaranteed to produce
// correct, consistent results. They are also relatively efficient. This is
// achieved by computing conservative error bounds and falling back to high
// precision or even exact arithmetic when the result is uncertain. Such
// predicates are useful in implementing robust algorithms.
//
// See also EdgeCrosser, which implements various exact
// edge-crossing predicates more efficiently than can be done here.

import (
	"math"

	"github.com/golang/geo/r3"
)

const (
	// epsilon is a small number that represents a reasonable level of noise between two
	// values that can be considered to be equal.
	epsilon = 1e-15
	// dblEpsilon is a smaller number for values that require more precision.
	dblEpsilon = 2.220446049250313e-16

	// maxDeterminantError is the maximum error in computing (AxB).C where all vectors
	// are unit length. Using standard inequalities, it can be shown that
	//
	//  fl(AxB) = AxB + D where |D| <= (|AxB| + (2/sqrt(3))*|A|*|B|) * e
	//
	// where "fl()" denotes a calculation done in floating-point arithmetic,
	// |x| denotes either absolute value or the L2-norm as appropriate, and
	// e is a reasonably small value near the noise level of floating point
	// number accuracy. Similarly,
	//
	//  fl(B.C) = B.C + d where |d| <= (|B.C| + 2*|B|*|C|) * e .
	//
	// Applying these bounds to the unit-length vectors A,B,C and neglecting
	// relative error (which does not affect the sign of the result), we get
	//
	//  fl((AxB).C) = (AxB).C + d where |d| <= (3 + 2/sqrt(3)) * e
	maxDeterminantError = 1.8274 * dblEpsilon

	// detErrorMultiplier is the factor to scale the magnitudes by when checking
	// for the sign of set of points with certainty. Using a similar technique to
	// the one used for maxDeterminantError, the error is at most:
	//
	//   |d| <= (3 + 6/sqrt(3)) * |A-C| * |B-C| * e
	//
	// If the determinant magnitude is larger than this value then we know
	// its sign with certainty.
	detErrorMultiplier = 3.2321 * dblEpsilon
)

// Direction is an indication of the ordering of a set of points.
type Direction int

// These are the three options for the direction of a set of points.
const (
	Clockwise        Direction = -1
	Indeterminate              = 0
	CounterClockwise           = 1
)

// Sign returns true if the points A, B, C are strictly counterclockwise,
// and returns false if the points are clockwise or collinear (i.e. if they are all
// contained on some great circle).
//
// Due to numerical errors, situations may arise that are mathematically
// impossible, e.g. ABC may be considered strictly CCW while BCA is not.
// However, the implementation guarantees the following:
//
// If Sign(a,b,c), then !Sign(c,b,a) for all a,b,c.
func Sign(a, b, c Point) bool {
	// NOTE(dnadasi): In the C++ API the equivalent method here was known as "SimpleSign".

	// We compute the signed volume of the parallelepiped ABC. The usual
	// formula for this is (A ⨯ B) · C, but we compute it here using (C ⨯ A) · B
	// in order to ensure that ABC and CBA are not both CCW. This follows
	// from the following identities (which are true numerically, not just
	// mathematically):
	//
	//     (1) x ⨯ y == -(y ⨯ x)
	//     (2) -x · y == -(x · y)
	return c.Cross(a.Vector).Dot(b.Vector) > 0
}

// RobustSign returns a Direction representing the ordering of the points.
// CounterClockwise is returned if the points are in counter-clockwise order,
// Clockwise for clockwise, and Indeterminate if any two points are the same (collinear),
// or the sign could not completely be determined.
//
// This function has additional logic to make sure that the above properties hold even
// when the three points are coplanar, and to deal with the limitations of
// floating-point arithmetic.
//
// RobustSign satisfies the following conditions:
//
//  (1) RobustSign(a,b,c) == Indeterminate if and only if a == b, b == c, or c == a
//  (2) RobustSign(b,c,a) == RobustSign(a,b,c) for all a,b,c
//  (3) RobustSign(c,b,a) == -RobustSign(a,b,c) for all a,b,c
//
// In other words:
//
//  (1) The result is Indeterminate if and only if two points are the same.
//  (2) Rotating the order of the arguments does not affect the result.
//  (3) Exchanging any two arguments inverts the result.
//
// On the other hand, note that it is not true in general that
// RobustSign(-a,b,c) == -RobustSign(a,b,c), or any similar identities
// involving antipodal points.
func RobustSign(a, b, c Point) Direction {
	sign := triageSign(a, b, c)
	if sign == Indeterminate {
		sign = expensiveSign(a, b, c)
	}
	return sign
}

// stableSign reports the direction sign of the points in a numerically stable way.
// Unlike triageSign, this method can usually compute the correct determinant sign
// even when all three points are as collinear as possible. For example if three
// points are spaced 1km apart along a random line on the Earth's surface using
// the nearest representable points, there is only a 0.4% chance that this method
// will not be able to find the determinant sign. The probability of failure
// decreases as the points get closer together; if the collinear points are 1 meter
// apart, the failure rate drops to 0.0004%.
//
// This method could be extended to also handle nearly-antipodal points, but antipodal
// points are rare in practice so it seems better to simply fall back to
// exact arithmetic in that case.
func stableSign(a, b, c Point) Direction {
	ab := a.Sub(b.Vector)
	ab2 := ab.Norm2()
	bc := b.Sub(c.Vector)
	bc2 := bc.Norm2()
	ca := c.Sub(a.Vector)
	ca2 := ca.Norm2()

	// Now compute the determinant ((A-C)x(B-C)).C, where the vertices have been
	// cyclically permuted if necessary so that AB is the longest edge. (This
	// minimizes the magnitude of cross product.)  At the same time we also
	// compute the maximum error in the determinant.

	// The two shortest edges, pointing away from their common point.
	var e1, e2, op r3.Vector
	if ab2 >= bc2 && ab2 >= ca2 {
		// AB is the longest edge.
		e1, e2, op = ca, bc, c.Vector
	} else if bc2 >= ca2 {
		// BC is the longest edge.
		e1, e2, op = ab, ca, a.Vector
	} else {
		// CA is the longest edge.
		e1, e2, op = bc, ab, b.Vector
	}

	det := e1.Cross(e2).Dot(op)
	maxErr := detErrorMultiplier * math.Sqrt(e1.Norm2()*e2.Norm2())

	// If the determinant isn't zero, within maxErr, we know definitively the point ordering.
	if det > maxErr {
		return CounterClockwise
	}
	if det < -maxErr {
		return Clockwise
	}
	return Indeterminate
}

// triageSign returns the direction sign of the points. It returns Indeterminate if two
// points are identical or the result is uncertain. Uncertain cases can be resolved, if
// desired, by calling expensiveSign.
//
// The purpose of this method is to allow additional cheap tests to be done without
// calling expensiveSign.
func triageSign(a, b, c Point) Direction {
	det := a.Cross(b.Vector).Dot(c.Vector)
	if det > maxDeterminantError {
		return CounterClockwise
	}
	if det < -maxDeterminantError {
		return Clockwise
	}
	return Indeterminate
}

// expensiveSign reports the direction sign of the points. It returns Indeterminate
// if two of the input points are the same. It uses multiple-precision arithmetic
// to ensure that its results are always self-consistent.
func expensiveSign(a, b, c Point) Direction {
	// Return Indeterminate if and only if two points are the same.
	// This ensures RobustSign(a,b,c) == Indeterminate if and only if a == b, b == c, or c == a.
	// ie. Property 1 of RobustSign.
	if a == b || b == c || c == a {
		return Indeterminate
	}

	// Next we try recomputing the determinant still using floating-point
	// arithmetic but in a more precise way. This is more expensive than the
	// simple calculation done by triageSign, but it is still *much* cheaper
	// than using arbitrary-precision arithmetic. This optimization is able to
	// compute the correct determinant sign in virtually all cases except when
	// the three points are truly collinear (e.g., three points on the equator).
	detSign := stableSign(a, b, c)
	if detSign != Indeterminate {
		return detSign
	}

	// Otherwise fall back to exact arithmetic and symbolic permutations.
	return exactSign(a, b, c, false)
}

// exactSign reports the direction sign of the points using exact precision arithmetic.
func exactSign(a, b, c Point, perturb bool) Direction {
	// In the C++ version, the final computation is performed using OpenSSL's
	// Bignum exact precision math library. The existence of an equivalent
	// library in Go is indeterminate. In C++, using the exact precision library
	// to solve this stage is ~300x slower than the above checks.
	// TODO(roberts): Select and incorporate an appropriate Go exact precision
	// floating point library for the remaining calculations.
	return Indeterminate
}
