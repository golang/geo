/*
Copyright 2015 Google Inc. All rights reserved.

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

	"github.com/golang/geo/r1"
	"github.com/golang/geo/s1"
)

func TestCrossings(t *testing.T) {
	na1 := math.Nextafter(1, 0)
	na2 := math.Nextafter(1, 2)

	tests := []struct {
		msg        string
		a, b, c, d Point
		simple     bool
		vertex     bool
	}{
		{
			"two regular edges that cross",
			PointFromCoords(1, 2, 1),
			PointFromCoords(1, -3, 0.5),
			PointFromCoords(1, -0.5, -3),
			PointFromCoords(0.1, 0.5, 3),
			true,
			true,
		},
		{
			"two regular edges that cross antipodal points",
			PointFromCoords(1, 2, 1),
			PointFromCoords(1, -3, 0.5),
			PointFromCoords(-1, 0.5, 3),
			PointFromCoords(-0.1, -0.5, -3),
			false,
			true,
		},
		{
			"two edges on the same great circle",
			PointFromCoords(0, 0, -1),
			PointFromCoords(0, 1, 0),
			PointFromCoords(0, 1, 1),
			PointFromCoords(0, 0, 1),
			false,
			false,
		},
		{
			"two edges that cross where one vertex is the OriginPoint",
			PointFromCoords(1, 0, 0),
			OriginPoint(),
			PointFromCoords(1, -0.1, 1),
			PointFromCoords(1, 1, -0.1),
			true,
			true,
		},
		{
			"two edges that cross antipodal points",
			PointFromCoords(1, 0, 0),
			PointFromCoords(0, 1, 0),
			PointFromCoords(0, 0, -1),
			PointFromCoords(-1, -1, 1),
			false,
			true,
		},
		{
			"two edges that share an endpoint",
			// The Ortho() direction is (-4,0,2) and edge CD
			// is further CCW around (2,3,4) than AB.
			PointFromCoords(2, 3, 4),
			PointFromCoords(-1, 2, 5),
			PointFromCoords(7, -2, 3),
			PointFromCoords(2, 3, 4),
			false,
			true,
		},
		{
			"two edges that barely cross near the middle of one edge",
			// The edge AB is approximately in the x=y plane, while CD is approximately
			// perpendicular to it and ends exactly at the x=y plane.
			PointFromCoords(1, 1, 1),
			PointFromCoords(1, na1, -1),
			PointFromCoords(11, -12, -1),
			PointFromCoords(10, 10, 1),
			true,
			true,
		},
		{
			"two edges that barely cross near the middle separated by a distance of about 1e-15",
			PointFromCoords(1, 1, 1),
			PointFromCoords(1, na2, -1),
			PointFromCoords(1, -1, 0),
			PointFromCoords(1, 1, 0),
			false,
			false,
		},
		{
			"two edges that barely cross each other near the end of both edges",
			// This example cannot be handled using regular double-precision
			// arithmetic due to floating-point underflow.
			// TODO(roberts): Determine if this case should be dropped from
			// the simplecrossing tests.
			PointFromCoords(0, 0, 1),
			PointFromCoords(2, -1e-323, 1),
			PointFromCoords(1, -1, 1),
			PointFromCoords(1e-323, 0, 1),
			false,
			false,
		},
		{
			"two edges that barely cross each other near the end separated by a distance of about 1e-640",
			PointFromCoords(0, 0, 1),
			PointFromCoords(2, 1e-323, 1),
			PointFromCoords(1, -1, 1),
			PointFromCoords(1e-323, 0, 1),
			false,
			false,
		},
		{
			"two edges that barely cross each other near the middle of one edge",
			// Computing the exact determinant of some of the triangles in this test
			// requires more than 2000 bits of precision.
			// TODO(roberts): Determine if this case should be dropped from
			// the simplecrossing tests.
			PointFromCoords(1, -1e-323, -1e-323),
			PointFromCoords(1e-323, 1, 1e-323),
			PointFromCoords(1, -1, 1e-323),
			PointFromCoords(1, 1, 0),
			false,
			true,
		},
		{
			"two edges that barely cross each other near the middle separated by a distance of about 1e-640",
			PointFromCoords(1, 1e-323, -1e-323),
			PointFromCoords(-1e-323, 1, 1e-323),
			PointFromCoords(1, -1, 1e-323),
			PointFromCoords(1, 1, 0),
			false,
			true,
		},
	}

	for _, test := range tests {
		if got := SimpleCrossing(test.a, test.b, test.c, test.d); got != test.simple {
			t.Errorf("%s: using vertex order (a,b,c,d)\nSimpleCrossing(%v,%v,%v,%v) = %t, want %t",
				test.msg, test.a, test.b, test.c, test.d, got, test.simple)
		}
		if got := SimpleCrossing(test.b, test.a, test.c, test.d); got != test.simple {
			t.Errorf("%s: using vertex order (b,a,c,d)\nSimpleCrossing(%v,%v,%v,%v) = %t, want %t",
				test.msg, test.b, test.a, test.c, test.d, got, test.simple)
		}
		if got := SimpleCrossing(test.a, test.b, test.d, test.c); got != test.simple {
			t.Errorf("%s: using vertex order (a,b,d,c)\nSimpleCrossing(%v,%v,%v,%v) = %t, want %t",
				test.msg, test.a, test.b, test.d, test.c, got, test.simple)
		}
		if got := SimpleCrossing(test.b, test.a, test.d, test.c); got != test.simple {
			t.Errorf("%s: using vertex order (b,a,d,c)\nSimpleCrossing(%v,%v,%v,%v) = %t, want %t",
				test.msg, test.b, test.a, test.d, test.c, got, test.simple)
		}

		if got := SimpleCrossing(test.c, test.d, test.a, test.b); got != test.simple {
			t.Errorf("%s: using vertex order (c,d,a,b)\nSimpleCrossing(%v,%v,%v,%v) = %t, want %t",
				test.msg, test.c, test.d, test.a, test.b, got, test.simple)
		}

		if got := VertexCrossing(test.a, test.b, test.c, test.b); got != test.vertex {
			t.Errorf("%s: VertexCrossing(%v,%v,%v,%v) = %t, want %t",
				test.msg, test.a, test.b, test.c, test.d, got, test.vertex)
		}
	}
}

func TestInterpolate(t *testing.T) {
	// Choose test points designed to expose floating-point errors.
	p1 := Point{PointFromCoords(0.1, 1e-30, 0.3).Normalize()}
	p2 := Point{PointFromCoords(-0.7, -0.55, -1e30).Normalize()}

	tests := []struct {
		a, b Point
		dist float64
		want Point
	}{
		// A zero-length edge.
		{p1, p1, 0, p1},
		{p1, p1, 1, p1},
		// Start, end, and middle of a medium-length edge.
		{p1, p2, 0, p1},
		{p1, p2, 1, p2},
		{p1, p2, 0.5, Point{(p1.Add(p2.Vector)).Mul(0.5).Normalize()}},

		// Test that interpolation is done using distances on the sphere
		// rather than linear distances.
		{
			Point{PointFromCoords(1, 0, 0).Normalize()},
			Point{PointFromCoords(0, 1, 0).Normalize()},
			1.0 / 3.0,
			Point{PointFromCoords(math.Sqrt(3), 1, 0).Normalize()},
		},
		{
			Point{PointFromCoords(1, 0, 0).Normalize()},
			Point{PointFromCoords(0, 1, 0).Normalize()},
			2.0 / 3.0,
			Point{PointFromCoords(1, math.Sqrt(3), 0).Normalize()},
		},
	}

	for _, test := range tests {
		// We allow a bit more than the usual 1e-15 error tolerance because
		// Interpolate() uses trig functions.
		if got := Interpolate(test.dist, test.a, test.b); !pointsApproxEquals(got, test.want, 3e-15) {
			t.Errorf("Interpolate(%v, %v, %v) = %v, want %v", test.dist, test.a, test.b, got, test.want)
		}
	}
}

func TestInterpolateOverLongEdge(t *testing.T) {
	lng := math.Pi - 1e-2
	a := Point{PointFromLatLng(LatLng{0, 0}).Normalize()}
	b := Point{PointFromLatLng(LatLng{0, s1.Angle(lng)}).Normalize()}

	for f := 0.4; f > 1e-15; f *= 0.1 {
		// Test that interpolation is accurate on a long edge (but not so long that
		// the definition of the edge itself becomes too unstable).
		want := Point{PointFromLatLng(LatLng{0, s1.Angle(f * lng)}).Normalize()}
		if got := Interpolate(f, a, b); !pointsApproxEquals(got, want, 3e-15) {
			t.Errorf("long edge Interpolate(%v, %v, %v) = %v, want %v", f, a, b, got, want)
		}

		// Test the remainder of the dist also matches.
		wantRem := Point{PointFromLatLng(LatLng{0, s1.Angle((1 - f) * lng)}).Normalize()}
		if got := Interpolate(1-f, a, b); !pointsApproxEquals(got, wantRem, 3e-15) {
			t.Errorf("long edge Interpolate(%v, %v, %v) = %v, want %v", 1-f, a, b, got, wantRem)
		}
	}
}

func TestInterpolateAntipodal(t *testing.T) {
	p1 := Point{PointFromCoords(0.1, 1e-30, 0.3).Normalize()}

	// Test that interpolation on a 180 degree edge (antipodal endpoints) yields
	// a result with the correct distance from each endpoint.
	for dist := 0.0; dist <= 1.0; dist += 0.125 {
		actual := Interpolate(dist, p1, Point{p1.Mul(-1)})
		if !float64Near(actual.Distance(p1).Radians(), dist*math.Pi, 3e-15) {
			t.Errorf("antipodal points Interpolate(%v, %v, %v) = %v, want %v", dist, p1, Point{p1.Mul(-1)}, actual, dist*math.Pi)
		}
	}
}

func TestRepeatedInterpolation(t *testing.T) {
	// Check that points do not drift away from unit length when repeated
	// interpolations are done.
	for i := 0; i < 100; i++ {
		a := randomPoint()
		b := randomPoint()
		for j := 0; j < 1000; j++ {
			a = Interpolate(0.01, a, b)
		}
		if !a.Vector.IsUnit() {
			t.Errorf("repeated Interpolate(%v, %v, %v) calls did not stay unit length for", 0.01, a, b)
		}
	}
}

var (
	rectErrorLat = 10 * dblEpsilon
	rectErrorLng = dblEpsilon
)

func rectBoundForPoints(a, b Point) Rect {
	bounder := NewRectBounder()
	bounder.AddPoint(a)
	bounder.AddPoint(b)
	return bounder.RectBound()
}

func TestRectBounderMaxLatitudeSimple(t *testing.T) {
	cubeLat := math.Asin(1 / math.Sqrt(3)) // 35.26 degrees
	cubeLatRect := Rect{r1.IntervalFromPoint(-cubeLat).AddPoint(cubeLat),
		s1.IntervalFromEndpoints(-math.Pi/4, math.Pi/4)}

	tests := []struct {
		a, b Point
		want Rect
	}{
		// Check cases where the min/max latitude is attained at a vertex.
		{
			a:    PointFromCoords(1, 1, 1),
			b:    PointFromCoords(1, -1, -1),
			want: cubeLatRect,
		},
		{
			a:    PointFromCoords(1, -1, 1),
			b:    PointFromCoords(1, 1, -1),
			want: cubeLatRect,
		},
	}

	for _, test := range tests {
		if got := rectBoundForPoints(test.a, test.b); !rectsApproxEqual(got, test.want, rectErrorLat, rectErrorLng) {
			t.Errorf("RectBounder for points (%v, %v) near max lat failed: got %v, want %v", test.a, test.b, got, test.want)
		}
	}
}

func TestRectBounderMaxLatitudeEdgeInterior(t *testing.T) {
	// Check cases where the min/max latitude occurs in the edge interior.
	// These tests expect the result to be pretty close to the middle of the
	// allowable error range (i.e., by adding 0.5 * kRectError).

	tests := []struct {
		got  float64
		want float64
	}{
		// Max latitude, CW edge
		{
			math.Pi/4 + 0.5*rectErrorLat,
			rectBoundForPoints(PointFromCoords(1, 1, 1), PointFromCoords(1, -1, 1)).Lat.Hi,
		},
		// Min latitude, CW edge
		{
			-math.Pi/4 - 0.5*rectErrorLat,
			rectBoundForPoints(PointFromCoords(1, -1, -1), PointFromCoords(-1, -1, -1)).Lat.Lo,
		},
		// Max latitude, CCW edge
		{
			math.Pi/4 + 0.5*rectErrorLat,
			rectBoundForPoints(PointFromCoords(1, -1, 1), PointFromCoords(1, 1, 1)).Lat.Hi,
		},
		// Min latitude, CCW edge
		{
			-math.Pi/4 - 0.5*rectErrorLat,
			rectBoundForPoints(PointFromCoords(-1, 1, -1), PointFromCoords(-1, -1, -1)).Lat.Lo,
		},

		// Check cases where the edge passes through one of the poles.
		{
			math.Pi / 2,
			rectBoundForPoints(PointFromCoords(.3, .4, 1), PointFromCoords(-.3, -.4, 1)).Lat.Hi,
		},
		{
			-math.Pi / 2,
			rectBoundForPoints(PointFromCoords(.3, .4, -1), PointFromCoords(-.3, -.4, -1)).Lat.Lo,
		},
	}

	for _, test := range tests {
		if !float64Eq(test.got, test.want) {
			t.Errorf("RectBound for max lat on interior of edge failed; got %v want %v", test.got, test.want)
		}
	}
}

func TestRectBounderMaxLatitudeRandom(t *testing.T) {
	// Check that the maximum latitude of edges is computed accurately to within
	// 3 * dblEpsilon (the expected maximum error). We concentrate on maximum
	// latitudes near the equator and north pole since these are the extremes.

	for iter := 0; iter < 100; iter++ {
		// Construct a right-handed coordinate frame (U,V,W) such that U points
		// slightly above the equator, V points at the equator, and W is slightly
		// offset from the north pole.
		u := randomPoint()
		u.Z = dblEpsilon * 1e-6 * math.Pow(1e12, randomFloat64())

		u = Point{u.Normalize()}
		v := Point{PointFromCoords(0, 0, 1).PointCross(u).Normalize()}
		w := Point{u.PointCross(v).Normalize()}

		// Construct a line segment AB that passes through U, and check that the
		// maximum latitude of this segment matches the latitude of U.
		a := Point{u.Sub(v.Mul(randomFloat64())).Normalize()}
		b := Point{u.Add(v.Mul(randomFloat64())).Normalize()}
		abBound := rectBoundForPoints(a, b)
		if !float64Near(latitude(u).Radians(), abBound.Lat.Hi, rectErrorLat) {
			t.Errorf("bound for line AB not near enough to the latitude of point %v. got %v, want %v",
				u, latitude(u).Radians(), abBound.Lat.Hi)
		}

		// Construct a line segment CD that passes through W, and check that the
		// maximum latitude of this segment matches the latitude of W.
		c := Point{w.Sub(v.Mul(randomFloat64())).Normalize()}
		d := Point{w.Add(v.Mul(randomFloat64())).Normalize()}
		cdBound := rectBoundForPoints(c, d)
		if !float64Near(latitude(w).Radians(), cdBound.Lat.Hi, rectErrorLat) {
			t.Errorf("bound for line CD not near enough to the lat of point %v. got %v, want %v",
				v, latitude(w).Radians(), cdBound.Lat.Hi)
		}
	}
}
