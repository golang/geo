// Copyright 2016 Google Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package s2

import (
	"fmt"
	"math"
	"math/big"
	"testing"

	"github.com/golang/geo/r3"
	"github.com/golang/geo/s1"
)

func TestPredicatesEpsilonForDigits(t *testing.T) {
	tests := []struct {
		have int
		want float64
	}{
		{
			have: 0,
			want: 1.0,
		},
		{
			have: 24,
			want: math.Ldexp(1.0, -24),
		},
		{
			have: 53,
			want: math.Ldexp(1.0, -53),
		},
		{
			have: 64,
			want: math.Ldexp(1.0, -64),
		},
		{
			have: 106,
			want: math.Ldexp(1.0, -106),
		},
		{
			have: 113,
			want: math.Ldexp(1.0, -113),
		},
	}

	for _, test := range tests {
		got := epsilonForDigits(test.have)
		if !float64Eq(got, test.want) {
			t.Errorf("epsilonForDigits(%d) = %g, want %g", test.have, got, test.want)
		}
	}
}

func TestRoundingEpsilon(t *testing.T) {
	var f32 float32
	var f64 float64

	const f32Epsilon = 1.1920928955078125e-7

	if got, want := roundingEpsilon(f32), f32Epsilon*0.5; got != want {
		t.Errorf("roundingEpsilon(float32) = %g, want %g", got, want)
	}

	if got, want := roundingEpsilon(f64), dblEpsilon*0.5; got != want {
		t.Errorf("roundingEpsilon(float64) = %g, want %g", got, want)
	}
}

func TestPredicatesSign(t *testing.T) {
	tests := []struct {
		p1x, p1y, p1z, p2x, p2y, p2z, p3x, p3y, p3z float64
		want                                        bool
	}{
		{1, 0, 0, 0, 1, 0, 0, 0, 1, true},
		{0, 1, 0, 0, 0, 1, 1, 0, 0, true},
		{0, 0, 1, 1, 0, 0, 0, 1, 0, true},
		{1, 1, 0, 0, 1, 1, 1, 0, 1, true},
		{-3, -1, 4, 2, -1, -3, 1, -2, 0, true},

		// All degenerate cases of Sign(). Let M_1, M_2, ... be the sequence of
		// submatrices whose determinant sign is tested by that function. Then the
		// i-th test below is a 3x3 matrix M (with rows A, B, C) such that:
		//
		//    det(M) = 0
		//    det(M_j) = 0 for j < i
		//    det(M_i) != 0
		//    A < B < C in lexicographic order.
		// det(M_1) = b0*c1 - b1*c0
		{-3, -1, 0, -2, 1, 0, 1, -2, 0, false},
		// det(M_2) = b2*c0 - b0*c2
		{-6, 3, 3, -4, 2, -1, -2, 1, 4, false},
		// det(M_3) = b1*c2 - b2*c1
		{0, -1, -1, 0, 1, -2, 0, 2, 1, false},
		// From this point onward, B or C must be zero, or B is proportional to C.
		// det(M_4) = c0*a1 - c1*a0
		{-1, 2, 7, 2, 1, -4, 4, 2, -8, false},
		// det(M_5) = c0
		{-4, -2, 7, 2, 1, -4, 4, 2, -8, false},
		// det(M_6) = -c1
		{0, -5, 7, 0, -4, 8, 0, -2, 4, false},
		// det(M_7) = c2*a0 - c0*a2
		{-5, -2, 7, 0, 0, -2, 0, 0, -1, false},
		// det(M_8) = c2
		{0, -2, 7, 0, 0, 1, 0, 0, 2, false},
	}

	for _, test := range tests {
		p1 := Point{r3.Vector{X: test.p1x, Y: test.p1y, Z: test.p1z}}
		p2 := Point{r3.Vector{X: test.p2x, Y: test.p2y, Z: test.p2z}}
		p3 := Point{r3.Vector{X: test.p3x, Y: test.p3y, Z: test.p3z}}
		result := Sign(p1, p2, p3)
		if result != test.want {
			t.Errorf("Sign(%v, %v, %v) = %v, want %v", p1, p2, p3, result, test.want)
		}
		if test.want {
			// For these cases we can test the reversibility condition
			result = Sign(p3, p2, p1)
			if result == test.want {
				t.Errorf("Sign(%v, %v, %v) = %v, want %v", p3, p2, p1, result, !test.want)
			}
		}
	}
}

// Points used in the various RobustSign tests.
var (
	// The following points happen to be *exactly collinear* along a line that it
	// approximate tangent to the surface of the unit sphere. In fact, C is the
	// exact midpoint of the line segment AB. All of these points are close
	// enough to unit length to satisfy r3.Vector.IsUnit().
	poA = Point{r3.Vector{X: 0.72571927877036835, Y: 0.46058825605889098, Z: 0.51106749730504852}}
	poB = Point{r3.Vector{X: 0.7257192746638208, Y: 0.46058826573818168, Z: 0.51106749441312738}}
	poC = Point{r3.Vector{X: 0.72571927671709457, Y: 0.46058826089853633, Z: 0.51106749585908795}}

	// The points "x1" and "x2" are exactly proportional, i.e. they both lie
	// on a common line through the origin. Both points are considered to be
	// normalized, and in fact they both satisfy (x == x.Normalize()).
	// Therefore the triangle (x1, x2, -x1) consists of three distinct points
	// that all lie on a common line through the origin.
	x1 = Point{r3.Vector{X: 0.99999999999999989, Y: 1.4901161193847655e-08, Z: 0}}
	x2 = Point{r3.Vector{X: 1, Y: 1.4901161193847656e-08, Z: 0}}

	// Here are two more points that are distinct, exactly proportional, and
	// that satisfy (x == x.Normalize()).
	x3 = Point{r3.Vector{X: 1, Y: 1, Z: 1}.Normalize()}
	x4 = Point{x3.Mul(0.99999999999999989)}

	// The following three points demonstrate that Normalize() is not idempotent, i.e.
	// y0.Normalize() != y0.Normalize().Normalize(). Both points are exactly proportional.
	y0 = Point{r3.Vector{X: 1, Y: 1, Z: 0}}
	y1 = Point{y0.Normalize()}
	y2 = Point{y1.Normalize()}
)

func TestPredicatesRobustSignEqualities(t *testing.T) {
	tests := []struct {
		p1, p2 Point
		want   bool
	}{
		{Point{poC.Sub(poA.Vector)}, Point{poB.Sub(poC.Vector)}, true},
		{x1, Point{x1.Normalize()}, true},
		{x2, Point{x2.Normalize()}, true},
		{x3, Point{x3.Normalize()}, true},
		{x4, Point{x4.Normalize()}, true},
		{x3, x4, false},
		{y1, y2, false},
		{y2, Point{y2.Normalize()}, true},
	}

	for _, test := range tests {
		if got := test.p1.Vector == test.p2.Vector; got != test.want {
			t.Errorf("Testing equality for RobustSign. %v = %v, got %v want %v", test.p1, test.p2, got, test.want)
		}
	}
}

func TestPredicatesRobustSign(t *testing.T) {
	x := Point{r3.Vector{X: 1, Y: 0, Z: 0}}
	y := Point{r3.Vector{X: 0, Y: 1, Z: 0}}
	z := Point{r3.Vector{X: 0, Y: 0, Z: 1}}

	tests := []struct {
		p1, p2, p3 Point
		want       Direction
	}{
		// Simple collinear points test cases.
		// a == b != c
		{x, x, z, Indeterminate},
		// a != b == c
		{x, y, y, Indeterminate},
		// c == a != b
		{z, x, z, Indeterminate},
		// CCW
		{x, y, z, CounterClockwise},
		// CW
		{z, y, x, Clockwise},

		// Edge cases:
		// The following points happen to be *exactly collinear* along a line that it
		// approximate tangent to the surface of the unit sphere. In fact, C is the
		// exact midpoint of the line segment AB. All of these points are close
		// enough to unit length to satisfy IsUnitLength().
		{
			poA, poB, poC, Clockwise,
		},

		// The points "x1" and "x2" are exactly proportional, i.e. they both lie
		// on a common line through the origin. Both points are considered to be
		// normalized, and in fact they both satisfy (x == x.Normalize()).
		// Therefore the triangle (x1, x2, -x1) consists of three distinct points
		// that all lie on a common line through the origin.
		{
			x1, x2, Point{x1.Mul(-1.0)}, CounterClockwise,
		},

		// Here are two more points that are distinct, exactly proportional, and
		// that satisfy (x == x.Normalize()).
		{
			x3, x4, Point{x3.Mul(-1.0)}, Clockwise,
		},

		// The following points demonstrate that Normalize() is not idempotent,
		// i.e. y0.Normalize() != y0.Normalize().Normalize(). Both points satisfy
		// IsNormalized(), though, and the two points are exactly proportional.
		{
			y1, y2, Point{y1.Mul(-1.0)}, CounterClockwise,
		},
	}

	for _, test := range tests {
		result := RobustSign(test.p1, test.p2, test.p3)
		if result != test.want {
			t.Errorf("RobustSign(%v, %v, %v) got %v, want %v",
				test.p1, test.p2, test.p3, result, test.want)
		}
		// Test RobustSign(b,c,a) == RobustSign(a,b,c) for all a,b,c
		rotated := RobustSign(test.p2, test.p3, test.p1)
		if rotated != result {
			t.Errorf("RobustSign(%v, %v, %v) vs Rotated RobustSign(%v, %v, %v) got %v, want %v",
				test.p1, test.p2, test.p3, test.p2, test.p3, test.p1, rotated, result)
		}
		// Test RobustSign(c,b,a) == -RobustSign(a,b,c) for all a,b,c
		var want Direction
		switch result {
		case CounterClockwise:
			want = Clockwise
		case Clockwise:
			want = CounterClockwise
		case Indeterminate:
			want = Indeterminate
		}
		reversed := RobustSign(test.p3, test.p2, test.p1)
		if reversed != want {
			t.Errorf("RobustSign(%v, %v, %v) vs Reversed RobustSign(%v, %v, %v) got %v, want %v",
				test.p1, test.p2, test.p3, test.p3, test.p2, test.p1, reversed, -1*result)
		}
	}

	// Test cases that should not be indeterminate.
	if got := RobustSign(poA, poB, poC); got == Indeterminate {
		t.Errorf("RobustSign(%v,%v,%v) = %v, want not Indeterminate", poA, poA, poA, got)
	}
	if got := RobustSign(x1, x2, Point{x1.Mul(-1)}); got == Indeterminate {
		t.Errorf("RobustSign(%v,%v,%v) = %v, want not Indeterminate", x1, x2, x1.Mul(-1), got)
	}
	if got := RobustSign(x3, x4, Point{x3.Mul(-1)}); got == Indeterminate {
		t.Errorf("RobustSign(%v,%v,%v) = %v, want not Indeterminate", x3, x4, x3.Mul(-1), got)
	}
	if got := RobustSign(y1, y2, Point{y1.Mul(-1)}); got == Indeterminate {
		t.Errorf("RobustSign(%v,%v,%v) = %v, want not Indeterminate", x1, x2, y1.Mul(-1), got)
	}
}

func TestPredicatesStableSignFailureRate(t *testing.T) {
	const iters = 1000

	// Verify that stableSign is able to handle most cases where the three
	// points are as collinear as possible. (For reference, triageSign fails
	// almost 100% of the time on this test.)
	//
	// Note that the failure rate *decreases* as the points get closer together,
	// and the decrease is approximately linear. For example, the failure rate
	// is 0.4% for collinear points spaced 1km apart, but only 0.0004% for
	// collinear points spaced 1 meter apart.
	//
	//  1km spacing: <  1% (actual is closer to 0.4%)
	// 10km spacing: < 10% (actual is closer to 4%)
	want := 0.01
	spacing := 1.0

	// Estimate the probability that stableSign will not be able to compute
	// the determinant sign of a triangle A, B, C consisting of three points
	// that are as collinear as possible and spaced the given distance apart
	// by counting up the times it returns Indeterminate.
	failureCount := 0
	m := math.Tan(spacing / earthRadiusKm)
	for iter := 0; iter < iters; iter++ {
		f := randomFrame()
		a := f.col(0)
		x := f.col(1)

		b := Point{a.Sub(x.Mul(m)).Normalize()}
		c := Point{a.Add(x.Mul(m)).Normalize()}
		sign := stableSign(a, b, c)
		if sign != Indeterminate {
			if got := exactSign(a, b, c, true); got != sign {
				t.Errorf("exactSign(%v, %v, %v, true) = %v, want %v", a, b, c, got, sign)
			}
		} else {
			failureCount++
		}
	}

	rate := float64(failureCount) / float64(iters)
	if rate >= want {
		t.Errorf("stableSign failure rate for spacing %v km = %v, want %v", spacing, rate, want)
	}
}

func TestPredicatesSymbolicallyPerturbedSign(t *testing.T) {
	// The purpose of this test is simply to get code coverage of
	// SymbolicallyPerturbedSign().  Let M_1, M_2, ... be the sequence of
	// submatrices whose determinant sign is tested by that function. Then the
	// i-th test below is a 3x3 matrix M (with rows A, B, C) such that:
	//
	//    det(M) = 0
	//    det(M_j) = 0 for j < i
	//    det(M_i) != 0
	//    A < B < C in lexicographic order.
	//
	// Checked that reversing the sign of any of the "return" statements in
	// SymbolicallyPerturbedSign will cause this test to fail.
	tests := []struct {
		a, b, c Point
		want    Direction
	}{
		{
			// det(M_1) = b0*c1 - b1*c0
			a:    Point{r3.Vector{X: -3, Y: -1, Z: 0}},
			b:    Point{r3.Vector{X: -2, Y: 1, Z: 0}},
			c:    Point{r3.Vector{X: 1, Y: -2, Z: 0}},
			want: CounterClockwise,
		},
		{
			// det(M_2) = b2*c0 - b0*c2
			want: CounterClockwise,
			a:    Point{r3.Vector{X: -6, Y: 3, Z: 3}},
			b:    Point{r3.Vector{X: -4, Y: 2, Z: -1}},
			c:    Point{r3.Vector{X: -2, Y: 1, Z: 4}},
		},
		{
			// det(M_3) = b1*c2 - b2*c1
			want: CounterClockwise,
			a:    Point{r3.Vector{X: 0, Y: -1, Z: -1}},
			b:    Point{r3.Vector{X: 0, Y: 1, Z: -2}},
			c:    Point{r3.Vector{X: 0, Y: 2, Z: 1}},
		},
		// From this point onward, B or C must be zero, or B is proportional to C.
		{
			// det(M_4) = c0*a1 - c1*a0
			want: CounterClockwise,
			a:    Point{r3.Vector{X: -1, Y: 2, Z: 7}},
			b:    Point{r3.Vector{X: 2, Y: 1, Z: -4}},
			c:    Point{r3.Vector{X: 4, Y: 2, Z: -8}},
		},
		{
			// det(M_5) = c0
			want: CounterClockwise,
			a:    Point{r3.Vector{X: -4, Y: -2, Z: 7}},
			b:    Point{r3.Vector{X: 2, Y: 1, Z: -4}},
			c:    Point{r3.Vector{X: 4, Y: 2, Z: -8}},
		},
		{
			// det(M_6) = -c1
			want: CounterClockwise,
			a:    Point{r3.Vector{X: 0, Y: -5, Z: 7}},
			b:    Point{r3.Vector{X: 0, Y: -4, Z: 8}},
			c:    Point{r3.Vector{X: 0, Y: -2, Z: 4}},
		},
		{
			// det(M_7) = c2*a0 - c0*a2
			want: CounterClockwise,
			a:    Point{r3.Vector{X: -5, Y: -2, Z: 7}},
			b:    Point{r3.Vector{X: 0, Y: 0, Z: -2}},
			c:    Point{r3.Vector{X: 0, Y: 0, Z: -1}},
		},
		{
			// det(M_8) = c2
			want: CounterClockwise,
			a:    Point{r3.Vector{X: 0, Y: -2, Z: 7}},
			b:    Point{r3.Vector{X: 0, Y: 0, Z: 1}},
			c:    Point{r3.Vector{X: 0, Y: 0, Z: 2}},
		},
		// From this point onward, C must be zero.
		{
			// det(M_9) = a0*b1 - a1*b0
			want: CounterClockwise,
			a:    Point{r3.Vector{X: -3, Y: 1, Z: 7}},
			b:    Point{r3.Vector{X: -1, Y: -4, Z: 1}},
			c:    Point{r3.Vector{X: 0, Y: 0, Z: 0}},
		},
		{
			// det(M_10) = -b0
			want: CounterClockwise,
			a:    Point{r3.Vector{X: -6, Y: -4, Z: 7}},
			b:    Point{r3.Vector{X: -3, Y: -2, Z: 1}},
			c:    Point{r3.Vector{X: 0, Y: 0, Z: 0}},
		},
		{
			// det(M_11) = b1
			want: Clockwise,
			a:    Point{r3.Vector{X: 0, Y: -4, Z: 7}},
			b:    Point{r3.Vector{X: 0, Y: -2, Z: 1}},
			c:    Point{r3.Vector{X: 0, Y: 0, Z: 0}},
		},
		{
			// det(M_12) = a0
			want: Clockwise,
			a:    Point{r3.Vector{X: -1, Y: -4, Z: 5}},
			b:    Point{r3.Vector{X: 0, Y: 0, Z: -3}},
			c:    Point{r3.Vector{X: 0, Y: 0, Z: 0}},
		},
		{
			// det(M_13) = 1
			want: CounterClockwise,
			a:    Point{r3.Vector{X: 0, Y: -4, Z: 5}},
			b:    Point{r3.Vector{X: 0, Y: 0, Z: -5}},
			c:    Point{r3.Vector{X: 0, Y: 0, Z: 0}},
		},
	}
	// Given 3 points A, B, C that are exactly coplanar with the origin and where
	// A < B < C in lexicographic order, verify that ABC is counterclockwise (if
	// expected == 1) or clockwise (if expected == -1) using expensiveSign().

	for _, test := range tests {
		if test.a.Cmp(test.b.Vector) != -1 {
			t.Errorf("%v >= %v, want <", test.a, test.b)
		}
		if test.b.Cmp(test.c.Vector) != -1 {
			t.Errorf("%v >= %v, want <", test.b, test.c)
		}
		if got := test.a.Dot(test.b.Cross(test.c.Vector)); !float64Eq(got, 0) {
			t.Errorf("%v.Dot(%v.Cross(%v)) = %v, want 0", test.a, test.b, test.c, got)
		}

		if got := expensiveSign(test.a, test.b, test.c); got != test.want {
			t.Errorf("expensiveSign(%v, %v, %v) = %v, want %v", test.a, test.b, test.c, got, test.want)
		}
		if got := expensiveSign(test.b, test.c, test.a); got != test.want {
			t.Errorf("expensiveSign(%v, %v, %v) = %v, want %v", test.b, test.c, test.a, got, test.want)
		}
		if got := expensiveSign(test.c, test.a, test.b); got != test.want {
			t.Errorf("expensiveSign(%v, %v, %v) = %v, want %v", test.c, test.a, test.b, got, test.want)
		}

		if got := expensiveSign(test.c, test.b, test.a); got != -test.want {
			t.Errorf("expensiveSign(%v, %v, %v) = %v, want %v", test.c, test.b, test.a, got, -test.want)
		}
		if got := expensiveSign(test.b, test.a, test.c); got != -test.want {
			t.Errorf("expensiveSign(%v, %v, %v) = %v, want %v", test.b, test.a, test.c, got, -test.want)
		}
		if got := expensiveSign(test.a, test.c, test.b); got != -test.want {
			t.Errorf("expensiveSign(%v, %v, %v) = %v, want %v", test.a, test.c, test.b, got, -test.want)
		}
	}
}

// compareDistancesFunc defines a type of function that can be used in the compare distances tests.
type compareDistancesFunc func(x, a, b Point) int

// triageCompareMinusSin2Distance wrapper to invert X for use when angles > 90.
func triageCompareMinusSin2Distance(x, a, b Point) int {
	return -triageCompareSin2Distances(Point{x.Mul(-1)}, a, b)
}

type precision int

const (
	doublePrecision precision = iota
	exactPrecision
	symbolicPrecision
)

func (p precision) String() string {
	switch p {
	case doublePrecision:
		return "double"
	case exactPrecision:
		return "exact"
	case symbolicPrecision:
		return "symbolic"
	default:
		panic(fmt.Sprintf("invalid precision value %d", p))
	}
}

func TestPredicatesCompareDistancesCoverage(t *testing.T) {
	// This test attempts to exercise all the code paths in all precisions.
	tests := []struct {
		x, a, b  Point
		distFunc compareDistancesFunc
		wantSign int
		wantPrec precision
	}{
		// Test triageCompareSin2Distances.
		{
			x:        PointFromCoords(1, 1, 1),
			a:        PointFromCoords(1, 1-1e-15, 1),
			b:        PointFromCoords(1, 1, 1+2e-15),
			distFunc: triageCompareSin2Distances,
			wantSign: -1,
			wantPrec: doublePrecision,
		},
		{
			x:        PointFromCoords(1, 1, 0),
			a:        PointFromCoords(1, 1-1e-15, 1e-21),
			b:        PointFromCoords(1, 1-1e-15, 0),
			distFunc: triageCompareSin2Distances,
			wantSign: 1,
			wantPrec: doublePrecision,
		},
		{
			x:        Point{r3.Vector{X: 2, Y: 0, Z: 0}},
			a:        Point{r3.Vector{X: 2, Y: -1, Z: 0}},
			b:        Point{r3.Vector{X: 2, Y: 1, Z: 1e-100}},
			distFunc: triageCompareSin2Distances,
			wantSign: -1,
			wantPrec: exactPrecision,
		},
		{
			x:        PointFromCoords(1, 0, 0),
			a:        PointFromCoords(1, -1, 0),
			b:        PointFromCoords(1, 1, 0),
			distFunc: triageCompareSin2Distances,
			wantSign: 1,
			wantPrec: symbolicPrecision,
		},
		{
			x:        PointFromCoords(1, 0, 0),
			a:        PointFromCoords(1, 0, 0),
			b:        PointFromCoords(1, 0, 0),
			distFunc: triageCompareSin2Distances,
			wantSign: 0,
			wantPrec: symbolicPrecision,
		},

		// triageCompareCosDistances
		{
			x:        PointFromCoords(1, 1, 1),
			a:        PointFromCoords(1, -1, 0),
			b:        PointFromCoords(-1, 1, 3e-15),
			distFunc: triageCompareCosDistances,
			wantSign: 1,
			wantPrec: doublePrecision,
		},
		{
			x:        PointFromCoords(1, 0, 0),
			a:        PointFromCoords(1, 1e-30, 0),
			b:        PointFromCoords(-1, 1e-40, 0),
			wantSign: -1,
			wantPrec: doublePrecision,
			distFunc: triageCompareCosDistances,
		},
		{
			x:        PointFromCoords(1, 1, 1),
			a:        PointFromCoords(1, -1, 0),
			b:        PointFromCoords(-1, 1, 1e-100),
			wantSign: 1,
			wantPrec: exactPrecision,
			distFunc: triageCompareCosDistances,
		},
		{
			x:        PointFromCoords(1, 1, 1),
			a:        PointFromCoords(1, -1, 0),
			b:        PointFromCoords(-1, 1, 0),
			wantSign: -1,
			wantPrec: symbolicPrecision,
			distFunc: triageCompareCosDistances,
		},
		{
			x:        PointFromCoords(1, 1, 1),
			a:        PointFromCoords(1, -1, 0),
			b:        PointFromCoords(1, -1, 0),
			distFunc: triageCompareCosDistances,
			wantSign: 0,
			wantPrec: symbolicPrecision,
		},
		// Test triageCompareSin2Distances using distances greater than 90 degrees.
		{
			x:        PointFromCoords(1, 1, 0),
			a:        PointFromCoords(-1, -1+1e-15, 0),
			b:        PointFromCoords(-1, -1, 0),
			distFunc: triageCompareMinusSin2Distance,
			wantSign: -1,
			wantPrec: doublePrecision,
		},
		{
			x:        PointFromCoords(-1, -1, 0),
			a:        PointFromCoords(1, 1-1e-15, 0),
			b:        PointFromCoords(1, 1-1e-15, 1e-21),
			distFunc: triageCompareMinusSin2Distance,
			wantSign: 1,
			wantPrec: doublePrecision,
		},
		{
			x:        PointFromCoords(-1, -1, 0),
			a:        PointFromCoords(2, 1, 0),
			b:        PointFromCoords(2, 1, 1e-30),
			distFunc: triageCompareMinusSin2Distance,
			wantSign: 1,
			wantPrec: exactPrecision,
		},
		{
			x:        PointFromCoords(-1, -1, 0),
			a:        PointFromCoords(2, 1, 0),
			b:        PointFromCoords(1, 2, 0),
			distFunc: triageCompareMinusSin2Distance,
			wantSign: -1,
			wantPrec: symbolicPrecision,
		},
	}

	for _, test := range tests {
		x := test.x
		a := test.a
		b := test.b

		// Verifies that CompareDistances(x, a, b) == wantSign, and furthermore
		// checks that the minimum required precision is wantPrec when the
		// distance calculation method for distFunc is used.

		// Don't normalize the arguments unless necessary (to allow testing points
		// that differ only in magnitude).
		if !x.IsUnit() {
			x = Point{x.Normalize()}
		}
		if !a.IsUnit() {
			a = Point{a.Normalize()}
		}
		if !b.IsUnit() {
			b = Point{b.Normalize()}
		}

		sign := test.distFunc(x, a, b)
		exactSign := exactCompareDistances(r3.PreciseVectorFromVector(x.Vector), r3.PreciseVectorFromVector(a.Vector), r3.PreciseVectorFromVector(b.Vector))

		actualSign := exactSign
		if exactSign == 0 {
			actualSign = symbolicCompareDistances(x, a, b)
		}

		// Check that the signs are correct (if non-zero), and also that if sign
		// is non-zero then so are the rest, etc.
		if test.wantSign != actualSign {
			t.Errorf("actual sign = %v, want %d", actualSign, test.wantSign)
		}
		if exactSign != 0 && exactSign != actualSign {
			t.Errorf("symbolic comparison was used, got sign %v, want %v", actualSign, exactSign)
		}

		var actualPrec precision
		if sign != 0 {
			actualPrec = doublePrecision
		} else if exactSign != 0 {
			actualPrec = exactPrecision
		} else {
			actualPrec = symbolicPrecision
		}

		if test.wantPrec != actualPrec {
			t.Errorf("got precision %s, want %s", test.wantPrec, actualPrec)
		}

		// Make sure that the top-level function returns the expected result.
		if got := CompareDistances(x, a, b); got != test.wantSign {
			t.Errorf("CompareDistances(%v, %v, %v) = %v, want %v", x, a, b, got, test.wantSign)
		}

		// Check that reversing the arguments negates the result.
		if got := CompareDistances(x, b, a); got != -test.wantSign {
			t.Errorf("CompareDistances(%v, %v, %v) = %v, want %v", x, b, a, got, -test.wantSign)
		}
	}
}

// compareDistanceFunc defines a type of function that can be used in the compare distance tests.
type compareDistanceFunc func(x, a Point, r float64) int

func TestPredicatesCompareDistanceCoverage(t *testing.T) {
	// This test attempts to exercise all the code paths in all precisions.
	tests := []struct {
		x, y     Point
		r        s1.ChordAngle
		distFunc compareDistanceFunc
		wantSign int
		wantPrec precision
	}{
		// Test triageCompareSin2Distance.
		{
			x:        PointFromCoords(1, 1, 1),
			y:        PointFromCoords(1, 1-1e-15, 1),
			r:        s1.ChordAngleFromAngle(1e-15),
			distFunc: triageCompareSin2Distance,
			wantSign: -1,
			wantPrec: doublePrecision,
		},
		{
			x:        PointFromCoords(1, 0, 0),
			y:        PointFromCoords(1, 1, 0),
			r:        s1.ChordAngleFromAngle(math.Pi / 4),
			distFunc: triageCompareSin2Distance,
			wantSign: -1,
			wantPrec: exactPrecision,
		},
		{
			x:        PointFromCoords(1, 1e-40, 0),
			y:        PointFromCoords(1+dblEpsilon, 1e-40, 0),
			r:        s1.ChordAngleFromAngle(0.9 * dblEpsilon * 1e-40),
			distFunc: triageCompareSin2Distance,
			wantSign: 1,
			wantPrec: exactPrecision,
		},
		{
			x:        PointFromCoords(1, 1e-40, 0),
			y:        PointFromCoords(1+dblEpsilon, 1e-40, 0),
			r:        s1.ChordAngleFromAngle(1.1 * dblEpsilon * 1e-40),
			distFunc: triageCompareSin2Distance,
			wantSign: -1,
			wantPrec: exactPrecision,
		},
		{
			x:        PointFromCoords(1, 0, 0),
			y:        PointFromCoords(1+dblEpsilon, 0, 0),
			r:        s1.ChordAngle(0),
			distFunc: triageCompareSin2Distance,
			wantSign: 0,
			wantPrec: exactPrecision,
		},
		// Test TriageCompareCosDistance.
		{
			x:        PointFromCoords(1, 0, 0),
			y:        PointFromCoords(1, 1e-8, 0),
			r:        s1.ChordAngle(1e-7),
			distFunc: triageCompareCosDistance,
			wantSign: -1,
			wantPrec: doublePrecision,
		},
		{
			x:        PointFromCoords(1, 0, 0),
			y:        PointFromCoords(-1, 1e-8, 0),
			r:        s1.ChordAngle(math.Pi - 1e-7),
			distFunc: triageCompareCosDistance,
			wantSign: 1,
			wantPrec: doublePrecision,
		},
		{
			x:        PointFromCoords(1, 1, 0),
			y:        PointFromCoords(1, -1-2*dblEpsilon, 0),
			r:        s1.RightChordAngle,
			distFunc: triageCompareCosDistance,
			wantSign: 1,
			wantPrec: doublePrecision,
		},
		{
			x:        PointFromCoords(1, 1, 0),
			y:        PointFromCoords(1, -1-dblEpsilon, 0),
			r:        s1.RightChordAngle,
			distFunc: triageCompareCosDistance,
			wantSign: 1,
			wantPrec: exactPrecision,
		},
		{
			x:        PointFromCoords(1, 1, 0),
			y:        PointFromCoords(1, -1, 1e-30),
			r:        s1.RightChordAngle,
			distFunc: triageCompareCosDistance,
			wantSign: 0,
			wantPrec: exactPrecision,
		},
		{
			// The angle between these two points is exactly 60 degrees.
			x:        PointFromCoords(1, 1, 0),
			y:        PointFromCoords(0, 1, 1),
			r:        s1.ChordAngleFromSquaredLength(1),
			distFunc: triageCompareCosDistance,
			wantSign: 0,
			wantPrec: exactPrecision,
		},
	}

	for d, test := range tests {
		x := test.x
		y := test.y
		r := test.r

		// Verifies that CompareDistance(x, y, r) == wantSign, and furthermore
		// checks that the minimum required precision is wantPrec when the
		// distance calculation method defined by distFunc is used.

		// Don't normalize the arguments unless necessary (to allow testing points
		// that differ only in magnitude).
		if !x.IsUnit() {
			x = Point{x.Normalize()}
		}
		if !y.IsUnit() {
			y = Point{y.Normalize()}
		}

		sign := test.distFunc(x, y, float64(r))
		exactSign := exactCompareDistance(r3.PreciseVectorFromVector(x.Vector), r3.PreciseVectorFromVector(y.Vector), big.NewFloat(float64(r)).SetPrec(big.MaxPrec))
		actualSign := exactSign

		// Check that the signs are correct (if non-zero), and also that if sign
		// is non-zero then so are the rest, etc.
		if test.wantSign != actualSign {
			t.Errorf("%d. actual sign = %v, want %d", d, actualSign, test.wantSign)
		}

		var actualPrec precision
		if sign != 0 {
			actualPrec = doublePrecision
		} else {
			actualPrec = exactPrecision
		}

		if test.wantPrec != actualPrec {
			t.Errorf("%d. got precision %s, want %s", d, actualPrec, test.wantPrec)
		}

		// Make sure that the top-level function returns the expected result.
		if got := CompareDistance(x, y, r); got != test.wantSign {
			t.Errorf("%d. CompareDistance(%v, %v, %v) = %v, want %v", d, x, y, r, got, test.wantSign)
		}

		// Mathematically, if d(X, Y) < r then d(-X, Y) > (Pi - r).  Unfortunately
		// there can be rounding errors when computing the supplementary distance,
		// so to ensure the two distances are exactly supplementary we need to do
		// the following.
		rSupp := s1.StraightChordAngle - r
		r = s1.StraightChordAngle - rSupp
		if got, want := -CompareDistance(x, y, r), CompareDistance(Point{x.Mul(-1)}, y, rSupp); got != want {
			t.Errorf("%d. CompareDistance(%v, %v, %v) = %v, CompareDistance(%v, %v, %v) = %v, should be the same", d, x, y, r, got, x.Mul(-1), y, rSupp, want)
		}
	}
}

// testCompareDistancesConsistency checks that the result at one level of precision
// is consistent with the result at the next higher level of precision. It returns
// the minimum precision that yielded a non-zero result.
func testCompareDistancesConsistency(t *testing.T, x, a, b Point, distFunc compareDistancesFunc) precision {
	dblSign := distFunc(x, a, b)
	exactSign := exactCompareDistances(r3.PreciseVectorFromVector(x.Vector), r3.PreciseVectorFromVector(a.Vector), r3.PreciseVectorFromVector(b.Vector))
	if dblSign != 0 {
		if exactSign != dblSign {
			t.Errorf("triageCompareDistances(%v, %v, %v) should be consistent with exactCompareDistances. got %v, want %v", x, a, b, dblSign, exactSign)
		}
		return symbolicPrecision
	}

	if exactSign != 0 {
		if got := CompareDistances(x, a, b); exactSign != got {
			t.Errorf("exactCompareDistances(%v, %v, %v) should have symbolicCompareDistance result agree. got %v, want %v", x, a, b, got, exactSign)
		}
		return exactPrecision
	}

	// Unlike the other methods, SymbolicCompareDistances has the
	// precondition that the exact sign must be zero.
	symbolicSign := symbolicCompareDistances(x, a, b)
	if got, want := CompareDistances(x, a, b), symbolicSign; got != want {
		t.Errorf("symbolicCompareDistances(%v, %v, %v) should be consistent with CompareDistances. got %v, want %v", x, a, b, got, want)
	}
	return symbolicPrecision
}

// choosePointNearPlaneOrAxes returns a random Point that is often near the
// intersection of one of the coodinates planes or coordinate axes with the unit
// sphere. (It is possible to represent very small perturbations near such points.)
func choosePointNearPlaneOrAxes() Point {
	p := randomPoint()
	if oneIn(3) {
		p.X *= math.Pow(1e-50, randomFloat64())
	}
	if oneIn(3) {
		p.Y *= math.Pow(1e-50, randomFloat64())
	}
	if oneIn(3) {
		p.Z *= math.Pow(1e-50, randomFloat64())
	}
	return Point{p.Normalize()}
}

func TestPredicatesCompareDistancesConsistency(t *testing.T) {
	const iters = 1000

	// This test chooses random point pairs that are nearly equidistant from a
	// target point, and then checks that the answer given by a method at one
	// level of precision is consistent with the answer given at the next higher
	// level of precision.
	//
	// The code below checks that the Cos, Sin2, and MinusSin2 methods
	// are consistent across their entire valid range of inputs, and also
	// simulates the logic in CompareDistance that chooses which method to use.

	// Test a specific case for equidistant points.
	if got, want := testCompareDistancesConsistency(t, PointFromCoords(1, 0, 0), PointFromCoords(0, -1, 0), PointFromCoords(0, 1, 0), triageCompareCosDistances), symbolicPrecision; got != want {
		t.Errorf("CompareDistances with 2 equidistant points didn't use symbolic compare, got %q want %q", got, want)
	}

	for iter := 0; iter < iters; iter++ {
		x := choosePointNearPlaneOrAxes()
		dir := choosePointNearPlaneOrAxes()
		r := s1.Angle(math.Pi / 2 * math.Pow(1e-30, randomFloat64()))
		if oneIn(2) {
			r = s1.Angle(math.Pi/2) - r
		}
		if oneIn(2) {
			r = s1.Angle(math.Pi/2) + r
		}

		a := InterpolateAtDistance(r, x, dir)
		b := InterpolateAtDistance(r, x, Point{dir.Mul(-1)})
		testCompareDistancesConsistency(t, x, a, b, triageCompareCosDistances)

		// The Sin2 method is only valid if both distances are less than 90
		// degrees, and similarly for the MinusSin2 method. (In the actual
		// implementation these methods are only used if both distances are less
		// than 45 degrees or greater than 135 degrees respectively.)
		if r.Radians() < math.Pi/2-1e-14 {
			prec := testCompareDistancesConsistency(t, x, a, b, triageCompareSin2Distances)
			if r.Degrees() < 45 {
				if a == b && symbolicPrecision != prec {
					t.Errorf("CompareDistances(%v, %v, %v) method for degenerate points = %s, want %s", x, a, b, prec, symbolicPrecision)
				}
			}
		} else if r.Radians() > math.Pi/2+1e-14 {
			// Use minus sin for > 45.
			testCompareDistancesConsistency(t, x, a, b, triageCompareMinusSin2Distance)
		}
	}
}

func TestPredicatesCompareDistanceConsistency(t *testing.T) {
	// This test chooses random inputs such that the distance between points X
	// and Y is very close to the threshold distance "r".  It then checks that
	// the answer given by a method at one level of precision is consistent with
	// the answer given at the next higher level of precision.  See also the
	// comments in the CompareDistances consistency test.
	const iters = 1000

	for iter := 0; iter < iters; iter++ {
		x := choosePointNearPlaneOrAxes()
		dir := choosePointNearPlaneOrAxes()
		r := s1.Angle(math.Pi / 2 * math.Pow(1e-30, randomFloat64()))
		if oneIn(2) {
			r = s1.Angle(math.Pi/2) - r
		}
		if oneIn(5) {
			r = s1.Angle(math.Pi/2) + r
		}
		y := InterpolateAtDistance(r, x, dir)

		// Checks that the result at one level of precision is consistent with the
		// result at the next higher level of precision.  Returns the minimum
		// precision that yielded a non-zero result.
		dblSign := triageCompareCosDistance(x, y, float64(s1.ChordAngleFromAngle(r)))
		exactSign := exactCompareDistance(r3.PreciseVectorFromVector(x.Vector), r3.PreciseVectorFromVector(y.Vector), big.NewFloat(float64(s1.ChordAngleFromAngle(r))).SetPrec(big.MaxPrec))
		if dblSign != 0 && dblSign != exactSign {
			t.Errorf("triageCompareCosDistance(%v, %v, %v) = %v, want %v", x, y, r, dblSign, exactSign)
		}
		if got := CompareDistance(x, y, s1.ChordAngleFromAngle(r)); exactSign != got {
			t.Errorf("CompareDistance(%v, %v, %v) = %v, want %v", x, y, r, got, exactSign)
		}

		if r.Radians() < math.Pi/2-1e-14 {
			dblSign = triageCompareSin2Distance(x, y, float64(s1.ChordAngleFromAngle(r)))
			exactSign = exactCompareDistance(r3.PreciseVectorFromVector(x.Vector), r3.PreciseVectorFromVector(y.Vector), big.NewFloat(float64(s1.ChordAngleFromAngle(r))).SetPrec(big.MaxPrec))
			if dblSign != 0 && dblSign != exactSign {
				t.Errorf("triageCompareSin2Distance(%v, %v, %v) = %v, want %v", x, y, r, dblSign, exactSign)
			}
		}
	}
}

// Verifies that SignDotProd(a, b) == expected, and that the minimum
// required precision is "expected_prec".
func TestPredicatesSignDotProd(t *testing.T) {
	tests := []struct {
		a, b      Point
		want      int
		precision string
	}{
		{
			// Orthogonal
			a:         PointFromCoords(1, 0, 0),
			b:         PointFromCoords(0, 1, 0),
			want:      0,
			precision: "EXACT",
		},
		{
			//  NearlyOrthogonalPositive
			a:         PointFromCoords(1, 0, 0),
			b:         PointFromCoords(dblEpsilon, 1, 0),
			want:      1,
			precision: "EXACT",
		},
		{
			//  NearlyOrthogonalPositive
			a:         PointFromCoords(1, 0, 0),
			b:         PointFromCoords(1e-45, 1, 0),
			want:      1,
			precision: "EXACT",
		},
		{
			// NearlyOrthogonalNegative
			a:         PointFromCoords(1, 0, 0),
			b:         PointFromCoords(-dblEpsilon, 1, 0),
			want:      -1,
			precision: "EXACT",
		},
		{
			// NearlyOrthogonalNegative
			a:         PointFromCoords(1, 0, 0),
			b:         PointFromCoords(-1e-45, 1, 0),
			want:      -1,
			precision: "EXACT",
		},
	}

	for _, test := range tests {
		got := SignDotProd(test.a, test.b)
		if got != test.want {
			t.Errorf("SignDotProd(%+v, %+v) = %d, wnat %d", test.a, test.b, got, test.want)
		}

		gotPrec := "EXACT"
		if triageSignDotProd(test.a, test.b) != 0 {
			gotPrec = "DOUBLE"
		}
		if test.precision != gotPrec {
			t.Errorf("triageSignDotProd precision = %q, wanted %q", gotPrec, test.precision)
		}
	}
}
func TestPredicatesCircleEdgeIntersectionOrdering(t *testing.T) {
	// Verifies that CircleEdgeIntersectionOrdering(a, b, c, d, n, m) == expected,
	// and that the minimum required precision is "expected_prec".

	// Two cells who's left and right edges are on the prime meridian,
	// cell0 := CellFromCellID(CellIDFromToken("054"))
	cell1 := CellFromCellID(CellIDFromToken("1ac"))

	// And then the three neighbors above them.
	// cella := CellFromCellID(CellIDFromToken("0fc"))
	cellb := CellFromCellID(CellIDFromToken("104"))
	//      cellc := CellFromCellID(CellIDFromToken("10c"))

	// Top, left and right edges of the cell as unnormalized vectors.
	e3 := cell1.EdgeRaw(3)
	e2 := cell1.EdgeRaw(2)
	e1 := cell1.EdgeRaw(1) // EdgeRaw
	c1 := cell1.Center()
	cb := cellb.Center()
	yeps := r3.Vector{X: 0, Y: epsilon, Z: 0}

	tests := []struct {
		a, b, c, d, m, n Point
		want             int
		prec             string
	}{
		{
			// The same edge should cross at the same spot exactly.
			a:    c1,
			b:    cb,
			c:    c1,
			d:    cb,
			m:    e2,
			n:    e1,
			want: 0,
			prec: "DOUBLE",
		},
		{
			// Simple case where the crossings aren't too close, AB should cross after CD.
			a:    c1,
			b:    cellb.Vertex(3),
			c:    c1,
			d:    cellb.Vertex(2),
			m:    e2,
			n:    e1,
			want: 1,
			prec: "DOUBLE",
		},
		// Swapping the boundary we're comparing against should negate the sign.
		{
			a:    c1,
			b:    cellb.Vertex(3),
			c:    c1,
			d:    cellb.Vertex(2),
			m:    e2,
			n:    e3,
			want: -1,
			prec: "DOUBLE",
		},

		// As should swapping the edge ordering.
		{
			a:    c1,
			b:    cellb.Vertex(2),
			c:    c1,
			d:    cellb.Vertex(3),
			m:    e2,
			n:    e1,
			want: -1,
			prec: "DOUBLE",
		},
		{
			a:    c1,
			b:    cellb.Vertex(2),
			c:    c1,
			d:    cellb.Vertex(3),
			m:    e2,
			n:    e3,
			want: 1,
			prec: "DOUBLE",
		},

		// Nearly the same edge but with one endpoint perturbed enough to require
		// long double precision.
		{
			a:    c1,
			b:    Point{cb.Add(yeps)},
			c:    c1,
			d:    cb,
			m:    e2,
			n:    e1,
			want: -1,
			prec: "EXACT",
		},
		{
			a:    c1,
			b:    Point{cb.Sub(yeps)},
			c:    c1,
			d:    cb,
			m:    e2,
			n:    e1,
			want: 1,
			prec: "EXACT",
		},
		{
			a:    c1,
			b:    cb,
			c:    c1,
			d:    Point{cb.Add(yeps)},
			m:    e2,
			n:    e1,
			want: 1,
			prec: "EXACT",
		},
		{
			a:    c1,
			b:    cb,
			c:    c1,
			d:    Point{cb.Sub(yeps)},
			m:    e2,
			n:    e1,
			want: -1,
			prec: "EXACT",
		},
	}

	for i, test := range tests {
		got := CircleEdgeIntersectionOrdering(test.a, test.b, test.c, test.d, test.m, test.n)

		if got != test.want {
			t.Errorf("%d: CircleEdgeIntersectionOrdering(%v, %v, %v, %v, %v, %v) = %d, want %d ",
				i, test.a, test.b, test.c, test.d, test.m, test.n, got, test.want)
		}

		// We triage in double precision and then fall back to exact for 0.
		actualPrec := "EXACT"
		if triageIntersectionOrdering(test.a, test.b, test.c, test.d, test.m, test.n) != 0 {
			actualPrec = "DOUBLE"
		} else {
			// We got zero, check for duplicate/reverse duplicate edges before falling
			// back to more precision.
			if (test.a == test.c && test.b == test.d) || (test.a == test.d && test.b == test.c) {
				actualPrec = "DOUBLE"
			}
		}
		if actualPrec != test.prec {
			t.Errorf("%d: CircleEdgeIntersectionOrdering(%v, %v, %v, %v, %v, %v( used precision %q, wanted %q",
				i, test.a, test.b, test.c, test.d, test.m, test.n, actualPrec, test.prec)
		}
	}
}

// ---------------------------------- Benchmarks ---------------------------

func BenchmarkSign(b *testing.B) {
	p1 := Point{r3.Vector{X: -3, Y: -1, Z: 4}}
	p2 := Point{r3.Vector{X: 2, Y: -1, Z: -3}}
	p3 := Point{r3.Vector{X: 1, Y: -2, Z: 0}}
	for i := 0; i < b.N; i++ {
		Sign(p1, p2, p3)
	}
}

// BenchmarkRobustSignSimple runs the benchmark for points that satisfy the first
// checks in RobustSign to compare the performance to that of Sign().
func BenchmarkRobustSignSimple(b *testing.B) {
	p1 := Point{r3.Vector{X: -3, Y: -1, Z: 4}}
	p2 := Point{r3.Vector{X: 2, Y: -1, Z: -3}}
	p3 := Point{r3.Vector{X: 1, Y: -2, Z: 0}}
	for i := 0; i < b.N; i++ {
		RobustSign(p1, p2, p3)
	}
}

// BenchmarkRobustSignNearCollinear runs the benchmark for points that are almost but not
// quite collinear, so the tests have to use most of the calculations of RobustSign
// before getting to an answer.
func BenchmarkRobustSignNearCollinear(b *testing.B) {
	for i := 0; i < b.N; i++ {
		RobustSign(poA, poB, poC)
	}
}

// TODO(rsned): Differences from C++
//
// TEST(epsilon_for_digits, recursion) {
// TEST(rounding_epsilon, vs_numeric_limits) {
// TEST(Sign, CollinearPoints) {
// TEST(Sign, StableSignUnderflow) {
// TEST_F(SignTest, StressTest) {
// TEST_F(StableSignTest, FailureRate) {
// TEST(Sign, SymbolicPerturbationCodeCoverage) {
// TEST(CompareDistances, Coverage) {
// TEST(CompareDistance, Coverage) {
// TEST(CompareEdgeDistance, Coverage) {
// TEST(CompareEdgeDistance, Consistency) {
// TEST(CompareEdgePairDistance, Coverage) {
// TEST(CompareEdgeDirections, Coverage) {
// TEST(CircleEdgeIntersectionSign, Works) {
// TEST(CompareEdgeDirections, Consistency) {
// TEST(EdgeCircumcenterSign, Coverage) {
// TEST(EdgeCircumcenterSign, Consistency) {
// TEST(VoronoiSiteExclusion, Coverage) {
// TEST(VoronoiSiteExclusion, Consistency) {
