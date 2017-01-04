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

import (
	"math"
	"testing"

	"github.com/golang/geo/r3"
)

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
		p1 := PointFromCoords(test.p1x, test.p1y, test.p1z)
		p2 := PointFromCoords(test.p2x, test.p2y, test.p2z)
		p3 := PointFromCoords(test.p3x, test.p3y, test.p3z)
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
		// enough to unit length to satisfy S2::IsUnitLength().
		{
			// Until we get ExactSign, this will only return Indeterminate.
			// It should be Clockwise.
			poA, poB, poC, Indeterminate,
		},

		// The points "x1" and "x2" are exactly proportional, i.e. they both lie
		// on a common line through the origin. Both points are considered to be
		// normalized, and in fact they both satisfy (x == x.Normalize()).
		// Therefore the triangle (x1, x2, -x1) consists of three distinct points
		// that all lie on a common line through the origin.
		{
			// Until we get ExactSign, this will only return Indeterminate.
			// It should be CounterClockwise.
			x1, x2, Point{x1.Mul(-1.0)}, Indeterminate,
		},

		// Here are two more points that are distinct, exactly proportional, and
		// that satisfy (x == x.Normalize()).
		{
			// Until we get ExactSign, this will only return Indeterminate.
			// It should be Clockwise.
			x3, x4, Point{x3.Mul(-1.0)}, Indeterminate,
		},

		// The following points demonstrate that Normalize() is not idempotent,
		// i.e. y0.Normalize() != y0.Normalize().Normalize(). Both points satisfy
		// S2::IsNormalized(), though, and the two points are exactly proportional.
		{
			// Until we get ExactSign, this will only return Indeterminate.
			// It should be CounterClockwise.
			y1, y2, Point{y1.Mul(-1.0)}, Indeterminate,
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
		want := Clockwise
		if result == Clockwise {
			want = CounterClockwise
		} else if result == Indeterminate {
			want = Indeterminate
		}
		reversed := RobustSign(test.p3, test.p2, test.p1)
		if reversed != want {
			t.Errorf("RobustSign(%v, %v, %v) vs Reversed RobustSign(%v, %v, %v) got %v, want %v",
				test.p1, test.p2, test.p3, test.p3, test.p2, test.p1, reversed, -1*result)
		}
	}

	// Test cases that should not be indeterminate.
	/*
		Uncomment these tests once RobustSign is completed.
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
	*/
}

func TestPredicatesStableSignFailureRate(t *testing.T) {
	const earthRadiusKm = 6371.01
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
			// TODO(roberts): Once exactSign is implemented, uncomment this case.
			//if got := exactSign(a, b, c, true); got != sign {
			//	t.Errorf("exactSign(%v, %v, %v, true) = %v, want %v", a, b, c, got, sign)
			//}
		} else {
			failureCount++
		}
	}

	rate := float64(failureCount) / float64(iters)
	if rate >= want {
		t.Errorf("stableSign failure rate for spacing %v km = %v, want %v", spacing, rate, want)
	}
}

func BenchmarkSign(b *testing.B) {
	p1 := PointFromCoords(-3, -1, 4)
	p2 := PointFromCoords(2, -1, -3)
	p3 := PointFromCoords(1, -2, 0)
	for i := 0; i < b.N; i++ {
		Sign(p1, p2, p3)
	}
}

// BenchmarkRobustSignSimple runs the benchmark for points that satisfy the first
// checks in RobustSign to compare the performance to that of Sign().
func BenchmarkRobustSignSimple(b *testing.B) {
	p1 := PointFromCoords(-3, -1, 4)
	p2 := PointFromCoords(2, -1, -3)
	p3 := PointFromCoords(1, -2, 0)
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
