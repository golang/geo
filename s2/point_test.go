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
	"testing"

	"github.com/golang/geo/r3"
	"github.com/golang/geo/s1"
)

func TestOriginPoint(t *testing.T) {
	if math.Abs(OriginPoint().Norm()-1) > 1e-15 {
		t.Errorf("Origin point norm = %v, want 1", OriginPoint().Norm())
	}

	// The point chosen below is about 66km from the north pole towards the East
	// Siberian Sea. The purpose of the stToUV(2/3) calculation is to keep the
	// origin as far away as possible from the longitudinal edges of large
	// Cells. (The line of longitude through the chosen point is always 1/3
	// or 2/3 of the way across any Cell with longitudinal edges that it
	// passes through.)
	p := PointFromCoords(-0.01, 0.01*stToUV(2.0/3), 1)
	if !p.ApproxEqual(OriginPoint()) {
		t.Errorf("Origin point should fall in the Siberian Sea, but does not.")
	}

	// Check that the origin is not too close to either pole.
	// The Earth's mean radius in kilometers (according to NASA).
	const earthRadiusKm = 6371.01
	if dist := math.Acos(OriginPoint().Z) * earthRadiusKm; dist <= 50 {
		t.Errorf("Origin point is to close to the North Pole. Got %v, want >= 50km", dist)
	}

}

func TestPointCross(t *testing.T) {
	tests := []struct {
		p1x, p1y, p1z, p2x, p2y, p2z float64
	}{
		{1, 0, 0, 1, 0, 0},
		{1, 0, 0, 0, 1, 0},
		{0, 1, 0, 1, 0, 0},
		{1, 2, 3, -4, 5, -6},
	}
	for _, test := range tests {
		p1 := PointFromCoords(test.p1x, test.p1y, test.p1z)
		p2 := PointFromCoords(test.p2x, test.p2y, test.p2z)
		result := p1.PointCross(p2)
		if !float64Eq(result.Norm(), 1) {
			t.Errorf("|%v ⨯ %v| = %v, want 1", p1, p2, result.Norm())
		}
		if x := result.Dot(p1.Vector); !float64Eq(x, 0) {
			t.Errorf("|(%v ⨯ %v) · %v| = %v, want 0", p1, p2, p1, x)
		}
		if x := result.Dot(p2.Vector); !float64Eq(x, 0) {
			t.Errorf("|(%v ⨯ %v) · %v| = %v, want 0", p1, p2, p2, x)
		}
	}
}

func TestSign(t *testing.T) {
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
	x = Point{r3.Vector{1, 0, 0}}
	y = Point{r3.Vector{0, 1, 0}}
	z = Point{r3.Vector{0, 0, 1}}

	// The following points happen to be *exactly collinear* along a line that it
	// approximate tangent to the surface of the unit sphere. In fact, C is the
	// exact midpoint of the line segment AB. All of these points are close
	// enough to unit length to satisfy r3.Vector.IsUnit().
	poA = Point{r3.Vector{0.72571927877036835, 0.46058825605889098, 0.51106749730504852}}
	poB = Point{r3.Vector{0.7257192746638208, 0.46058826573818168, 0.51106749441312738}}
	poC = Point{r3.Vector{0.72571927671709457, 0.46058826089853633, 0.51106749585908795}}

	// The points "x1" and "x2" are exactly proportional, i.e. they both lie
	// on a common line through the origin. Both points are considered to be
	// normalized, and in fact they both satisfy (x == x.Normalize()).
	// Therefore the triangle (x1, x2, -x1) consists of three distinct points
	// that all lie on a common line through the origin.
	x1 = Point{r3.Vector{0.99999999999999989, 1.4901161193847655e-08, 0}}
	x2 = Point{r3.Vector{1, 1.4901161193847656e-08, 0}}

	// Here are two more points that are distinct, exactly proportional, and
	// that satisfy (x == x.Normalize()).
	x3 = Point{r3.Vector{1, 1, 1}.Normalize()}
	x4 = Point{x3.Mul(0.99999999999999989)}

	// The following three points demonstrate that Normalize() is not idempotent, i.e.
	// y0.Normalize() != y0.Normalize().Normalize(). Both points are exactly proportional.
	y0 = Point{r3.Vector{1, 1, 0}}
	y1 = Point{y0.Normalize()}
	y2 = Point{y1.Normalize()}
)

// TODO(roberts): This test is missing the actual RobustSign() parts of the checks from C++
// test method RobustSign::CollinearPoints.
func TestRobustSignEqualities(t *testing.T) {
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

func TestRobustSign(t *testing.T) {
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
}

func TestPointDistance(t *testing.T) {
	tests := []struct {
		x1, y1, z1 float64
		x2, y2, z2 float64
		want       float64 // radians
	}{
		{1, 0, 0, 1, 0, 0, 0},
		{1, 0, 0, 0, 1, 0, math.Pi / 2},
		{1, 0, 0, 0, 1, 1, math.Pi / 2},
		{1, 0, 0, -1, 0, 0, math.Pi},
		{1, 2, 3, 2, 3, -1, 1.2055891055045298},
	}
	for _, test := range tests {
		p1 := PointFromCoords(test.x1, test.y1, test.z1)
		p2 := PointFromCoords(test.x2, test.y2, test.z2)
		if a := p1.Distance(p2).Radians(); !float64Eq(a, test.want) {
			t.Errorf("%v.Distance(%v) = %v, want %v", p1, p2, a, test.want)
		}
		if a := p2.Distance(p1).Radians(); !float64Eq(a, test.want) {
			t.Errorf("%v.Distance(%v) = %v, want %v", p2, p1, a, test.want)
		}
	}
}

func TestPointApproxEqual(t *testing.T) {
	tests := []struct {
		x1, y1, z1 float64
		x2, y2, z2 float64
		want       bool
	}{
		{1, 0, 0, 1, 0, 0, true},
		{1, 0, 0, 0, 1, 0, false},
		{1, 0, 0, 0, 1, 1, false},
		{1, 0, 0, -1, 0, 0, false},
		{1, 2, 3, 2, 3, -1, false},
		{1, 0, 0, 1 * (1 + epsilon), 0, 0, true},
		{1, 0, 0, 1 * (1 - epsilon), 0, 0, true},
		{1, 0, 0, 1 + epsilon, 0, 0, true},
		{1, 0, 0, 1 - epsilon, 0, 0, true},
		{1, 0, 0, 1, epsilon, 0, true},
		{1, 0, 0, 1, epsilon, epsilon, false},
		{1, epsilon, 0, 1, -epsilon, epsilon, false},
	}
	for _, test := range tests {
		p1 := PointFromCoords(test.x1, test.y1, test.z1)
		p2 := PointFromCoords(test.x2, test.y2, test.z2)
		if got := p1.ApproxEqual(p2); got != test.want {
			t.Errorf("%v.ApproxEqual(%v), got %v want %v", p1, p2, got, test.want)
		}
	}
}

var (
	pz   = PointFromCoords(0, 0, 1)
	p000 = PointFromCoords(1, 0, 0)
	p045 = PointFromCoords(1, 1, 0)
	p090 = PointFromCoords(0, 1, 0)
	p180 = PointFromCoords(-1, 0, 0)
	// Degenerate triangles.
	pr = PointFromCoords(0.257, -0.5723, 0.112)
	pq = PointFromCoords(-0.747, 0.401, 0.2235)

	// For testing the Girard area fall through case.
	g1 = PointFromCoords(1, 1, 1)
	g2 = Point{g1.Add(pr.Mul(1e-15)).Normalize()}
	g3 = Point{g1.Add(pq.Mul(1e-15)).Normalize()}
)

func TestPointArea(t *testing.T) {
	epsilon := 1e-10
	tests := []struct {
		a, b, c  Point
		want     float64
		nearness float64
	}{
		{p000, p090, pz, math.Pi / 2.0, 0},
		// This test case should give 0 as the epsilon, but either Go or C++'s value for Pi,
		// or the accuracy of the multiplications along the way, cause a difference ~15 decimal
		// places into the result, so it is not quite a difference of 0.
		{p045, pz, p180, 3.0 * math.Pi / 4.0, 1e-14},
		// Make sure that Area has good *relative* accuracy even for very small areas.
		{PointFromCoords(epsilon, 0, 1), PointFromCoords(0, epsilon, 1), pz, 0.5 * epsilon * epsilon, 1e-14},
		// Make sure that it can handle degenerate triangles.
		{pr, pr, pr, 0.0, 0},
		{pr, pq, pr, 0.0, 1e-15},
		{p000, p045, p090, 0.0, 0},
		// Try a very long and skinny triangle.
		{p000, PointFromCoords(1, 1, epsilon), p090, 5.8578643762690495119753e-11, 1e-9},
		// TODO(roberts):
		// C++ includes a 10,000 loop of perterbations to test out the Girard area
		// computation is less than some noise threshold.
		// Do we need that many? Will one or two suffice?
		{g1, g2, g3, 0.0, 1e-15},
	}
	for _, test := range tests {
		if got := PointArea(test.a, test.b, test.c); !float64Near(got, test.want, test.nearness) {
			t.Errorf("PointArea(%v, %v, %v), got %v want %v", test.a, test.b, test.c, got, test.want)
		}
	}
}

func TestPointAreaQuarterHemisphere(t *testing.T) {
	tests := []struct {
		a, b, c, d, e Point
		want          float64
	}{
		// Triangles with near-180 degree edges that sum to a quarter-sphere.
		{PointFromCoords(1, 0.1*epsilon, epsilon), p000, p045, p180, pz, math.Pi},
		// Four other triangles that sum to a quarter-sphere.
		{PointFromCoords(1, 1, epsilon), p000, p045, p180, pz, math.Pi},
		// TODO(roberts):
		// C++ Includes a loop of 100 perturbations on a hemisphere for more tests.
	}
	for _, test := range tests {
		area := PointArea(test.a, test.b, test.c) +
			PointArea(test.a, test.c, test.d) +
			PointArea(test.a, test.d, test.e) +
			PointArea(test.a, test.e, test.b)

		if !float64Eq(area, test.want) {
			t.Errorf("Adding up 4 quarter hemispheres with PointArea(), got %v want %v", area, test.want)
		}
	}
}

func TestPlanarCentroid(t *testing.T) {
	tests := []struct {
		name             string
		p0, p1, p2, want Point
	}{
		{
			name: "xyz axis",
			p0:   PointFromCoords(0, 0, 1),
			p1:   PointFromCoords(0, 1, 0),
			p2:   PointFromCoords(1, 0, 0),
			want: PointFromCoords(1./3, 1./3, 1./3),
		},
		{
			name: "Same point",
			p0:   PointFromCoords(1, 0, 0),
			p1:   PointFromCoords(1, 0, 0),
			p2:   PointFromCoords(1, 0, 0),
			want: PointFromCoords(1, 0, 0),
		},
	}

	for _, test := range tests {
		got := PlanarCentroid(test.p0, test.p1, test.p2)
		if !got.ApproxEqual(test.want) {
			t.Errorf("%s: PlanarCentroid(%v, %v, %v) = %v, want %v", test.name, test.p0, test.p1, test.p2, got, test.want)
		}
	}
}

func TestTrueCentroid(t *testing.T) {
	// Test TrueCentroid with very small triangles. This test assumes that
	// the triangle is small enough so that it is nearly planar.
	// The centroid of a planar triangle is at the intersection of its
	// medians, which is two-thirds of the way along each median.
	for i := 0; i < 100; i++ {
		f := randomFrame()
		p := f.col(0)
		x := f.col(1)
		y := f.col(2)
		d := 1e-4 * math.Pow(1e-4, randomFloat64())

		// Make a triangle with two equal sides.
		p0 := Point{p.Sub(x.Mul(d)).Normalize()}
		p1 := Point{p.Add(x.Mul(d)).Normalize()}
		p2 := Point{p.Add(y.Mul(d * 3)).Normalize()}
		want := Point{p.Add(y.Mul(d)).Normalize()}

		got := TrueCentroid(p0, p1, p2).Normalize()
		if got.Distance(want.Vector) >= 2e-8 {
			t.Errorf("TrueCentroid(%v, %v, %v).Normalize() = %v, want %v", p0, p1, p2, got, want)
		}

		// Make a triangle with a right angle.
		p0 = p
		p1 = Point{p.Add(x.Mul(d * 3)).Normalize()}
		p2 = Point{p.Add(y.Mul(d * 6)).Normalize()}
		want = Point{p.Add(x.Add(y.Mul(2)).Mul(d)).Normalize()}

		got = TrueCentroid(p0, p1, p2).Normalize()
		if got.Distance(want.Vector) >= 2e-8 {
			t.Errorf("TrueCentroid(%v, %v, %v).Normalize() = %v, want %v", p0, p1, p2, got, want)
		}
	}
}

func TestPointRegularPoints(t *testing.T) {
	// Conversion to/from degrees has a little more variability than the default epsilon.
	const epsilon = 1e-13
	center := PointFromLatLng(LatLngFromDegrees(80, 135))
	radius := s1.Degree * 20
	pts := regularPoints(center, radius, 4)

	if len(pts) != 4 {
		t.Errorf("regularPoints with 4 vertices should have 4 vertices, got %d", len(pts))
	}

	lls := []LatLng{
		LatLngFromPoint(pts[0]),
		LatLngFromPoint(pts[1]),
		LatLngFromPoint(pts[2]),
		LatLngFromPoint(pts[3]),
	}
	cll := LatLngFromPoint(center)

	// Make sure that the radius is correct.
	wantDist := 20.0
	for i, ll := range lls {
		if got := cll.Distance(ll).Degrees(); !float64Near(got, wantDist, epsilon) {
			t.Errorf("Vertex %d distance from center = %v, want %v", i, got, wantDist)
		}
	}

	// Make sure the angle between each point is correct.
	wantAngle := math.Pi / 2
	for i := 0; i < len(pts); i++ {
		// Mod the index by 4 to wrap the values at each end.
		v0, v1, v2 := pts[(4+i+1)%4], pts[(4+i)%4], pts[(4+i-1)%4]
		if got := float64(v0.Sub(v1.Vector).Angle(v2.Sub(v1.Vector))); !float64Eq(got, wantAngle) {
			t.Errorf("(%v-%v).Angle(%v-%v) = %v, want %v", v0, v1, v1, v2, got, wantAngle)
		}
	}

	// Make sure that all edges of the polygon have the same length.
	wantLength := 27.990890717782829
	for i := 0; i < len(lls); i++ {
		ll1, ll2 := lls[i], lls[(i+1)%4]
		if got := ll1.Distance(ll2).Degrees(); !float64Near(got, wantLength, epsilon) {
			t.Errorf("%v.Distance(%v) = %v, want %v", ll1, ll2, got, wantLength)
		}
	}

	// Spot check an actual coordinate now that we know the points are spaced
	// evenly apart at the same angles and radii.
	if got, want := lls[0].Lat.Degrees(), 62.162880741097204; !float64Near(got, want, epsilon) {
		t.Errorf("%v.Lat = %v, want %v", lls[0], got, want)
	}
	if got, want := lls[0].Lng.Degrees(), 103.11051028343407; !float64Near(got, want, epsilon) {
		t.Errorf("%v.Lng = %v, want %v", lls[0], got, want)
	}
}

func BenchmarkPointArea(b *testing.B) {
	for i := 0; i < b.N; i++ {
		PointArea(p000, p090, pz)
	}
}

func BenchmarkPointAreaGirardCase(b *testing.B) {
	for i := 0; i < b.N; i++ {
		PointArea(g1, g2, g3)
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
