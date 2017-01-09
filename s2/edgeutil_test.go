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
	"fmt"
	"math"
	"testing"

	"github.com/golang/geo/r1"
	"github.com/golang/geo/r2"
	"github.com/golang/geo/r3"
	"github.com/golang/geo/s1"
)

func TestEdgeutilCrossings(t *testing.T) {
	na1 := math.Nextafter(1, 0)
	na2 := math.Nextafter(1, 2)

	tests := []struct {
		msg          string
		a, b, c, d   Point
		simpleTest   bool
		robust       Crossing
		vertex       bool
		edgeOrVertex bool
	}{
		{
			"two regular edges that cross",
			Point{r3.Vector{1, 2, 1}},
			Point{r3.Vector{1, -3, 0.5}},
			Point{r3.Vector{1, -0.5, -3}},
			Point{r3.Vector{0.1, 0.5, 3}},
			true,
			Cross,
			true,
			true,
		},
		{
			"two regular edges that cross antipodal points",
			Point{r3.Vector{1, 2, 1}},
			Point{r3.Vector{1, -3, 0.5}},
			Point{r3.Vector{-1, 0.5, 3}},
			Point{r3.Vector{-0.1, -0.5, -3}},
			true,
			DoNotCross,
			true,
			false,
		},
		{
			"two edges on the same great circle",
			Point{r3.Vector{0, 0, -1}},
			Point{r3.Vector{0, 1, 0}},
			Point{r3.Vector{0, 1, 1}},
			Point{r3.Vector{0, 0, 1}},
			true,
			DoNotCross,
			false,
			false,
		},
		{
			"two edges that cross where one vertex is the OriginPoint",
			Point{r3.Vector{1, 0, 0}},
			OriginPoint(),
			Point{r3.Vector{1, -0.1, 1}},
			Point{r3.Vector{1, 1, -0.1}},
			true,
			Cross,
			true,
			true,
		},
		{
			"two edges that cross antipodal points",
			Point{r3.Vector{1, 0, 0}},
			Point{r3.Vector{0, 1, 0}},
			Point{r3.Vector{0, 0, -1}},
			Point{r3.Vector{-1, -1, 1}},
			true,
			DoNotCross,
			true,
			false,
		},
		{
			"two edges that share an endpoint",
			// The Ortho() direction is (-4,0,2) and edge CD
			// is further CCW around (2,3,4) than AB.
			Point{r3.Vector{2, 3, 4}},
			Point{r3.Vector{-1, 2, 5}},
			Point{r3.Vector{7, -2, 3}},
			Point{r3.Vector{2, 3, 4}},
			true,
			MaybeCross,
			true,
			false,
		},
		{
			"two edges that barely cross near the middle of one edge",
			// The edge AB is approximately in the x=y plane, while CD is approximately
			// perpendicular to it and ends exactly at the x=y plane.
			Point{r3.Vector{1, 1, 1}},
			Point{r3.Vector{1, na1, -1}},
			Point{r3.Vector{11, -12, -1}},
			Point{r3.Vector{10, 10, 1}},
			false,
			DoNotCross, // TODO(sbeckman): Should be 1, fix once exactSign is implemented.
			true,
			false, // TODO(sbeckman): Should be true, fix once exactSign is implemented.
		},
		{
			"two edges that barely cross near the middle separated by a distance of about 1e-15",
			Point{r3.Vector{1, 1, 1}},
			Point{r3.Vector{1, na2, -1}},
			Point{r3.Vector{1, -1, 0}},
			Point{r3.Vector{1, 1, 0}},
			false,
			DoNotCross,
			false,
			false,
		},
		{
			"two edges that barely cross each other near the end of both edges",
			// This example cannot be handled using regular double-precision
			// arithmetic due to floating-point underflow.
			Point{r3.Vector{0, 0, 1}},
			Point{r3.Vector{2, -1e-323, 1}},
			Point{r3.Vector{1, -1, 1}},
			Point{r3.Vector{1e-323, 0, 1}},
			false,
			DoNotCross, // TODO(sbeckman): Should be 1, fix once exactSign is implemented.
			false,
			false, // TODO(sbeckman): Should be true, fix once exactSign is implemented.
		},
		{
			"two edges that barely cross each other near the end separated by a distance of about 1e-640",
			Point{r3.Vector{0, 0, 1}},
			Point{r3.Vector{2, 1e-323, 1}},
			Point{r3.Vector{1, -1, 1}},
			Point{r3.Vector{1e-323, 0, 1}},
			false,
			DoNotCross,
			false,
			false,
		},
		{
			"two edges that barely cross each other near the middle of one edge",
			// Computing the exact determinant of some of the triangles in this test
			// requires more than 2000 bits of precision.
			Point{r3.Vector{1, -1e-323, -1e-323}},
			Point{r3.Vector{1e-323, 1, 1e-323}},
			Point{r3.Vector{1, -1, 1e-323}},
			Point{r3.Vector{1, 1, 0}},
			false,
			Cross,
			true,
			true,
		},
		{
			"two edges that barely cross each other near the middle separated by a distance of about 1e-640",
			Point{r3.Vector{1, 1e-323, -1e-323}},
			Point{r3.Vector{-1e-323, 1, 1e-323}},
			Point{r3.Vector{1, -1, 1e-323}},
			Point{r3.Vector{1, 1, 0}},
			false,
			Cross, // TODO(sbeckman): Should be -1, fix once exactSign is implemented.
			true,
			true, // TODO(sbeckman): Should be false, fix once exactSign is implemented.
		},
	}

	for _, test := range tests {
		if err := testCrossing(test.a, test.b, test.c, test.d, test.robust, test.vertex, test.edgeOrVertex, test.simpleTest); err != nil {
			t.Errorf("%s: %v", test.msg, err)
		}
		if err := testCrossing(test.b, test.a, test.c, test.d, test.robust, test.vertex, test.edgeOrVertex, test.simpleTest); err != nil {
			t.Errorf("%s: %v", test.msg, err)
		}
		if err := testCrossing(test.a, test.b, test.d, test.c, test.robust, test.vertex, test.edgeOrVertex, test.simpleTest); err != nil {
			t.Errorf("%s: %v", test.msg, err)
		}
		if err := testCrossing(test.b, test.a, test.c, test.d, test.robust, test.vertex, test.edgeOrVertex, test.simpleTest); err != nil {
			t.Errorf("%s: %v", test.msg, err)
		}
		if err := testCrossing(test.a, test.b, test.a, test.b, MaybeCross, true, true, false); err != nil {
			t.Errorf("%s: %v", test.msg, err)
		}
		if err := testCrossing(test.c, test.d, test.a, test.b, test.robust, test.vertex, test.edgeOrVertex != (test.robust == MaybeCross), test.simpleTest); err != nil {
			t.Errorf("%s: %v", test.msg, err)
		}

		if got := VertexCrossing(test.a, test.b, test.c, test.b); got != test.vertex {
			t.Errorf("%s: VertexCrossing(%v,%v,%v,%v) = %t, want %t", test.msg, test.a, test.b, test.c, test.d, got, test.vertex)
		}
	}
}

func testCrossing(a, b, c, d Point, robust Crossing, vertex, edgeOrVertex, simple bool) error {
	input := fmt.Sprintf("a: %v, b: %v, c: %v, d: %v", a, b, c, d)
	if got, want := SimpleCrossing(a, b, c, d), robust == Cross; simple && got != want {
		return fmt.Errorf("%v, SimpleCrossing(a, b, c, d) = %t, want %t", input, got, want)
	}

	crosser := NewChainEdgeCrosser(a, b, c)
	if got, want := crosser.ChainCrossingSign(d), robust; got != want {
		return fmt.Errorf("%v, ChainCrossingSign(d) = %d, want %d", input, got, want)
	}
	if got, want := crosser.ChainCrossingSign(c), robust; got != want {
		return fmt.Errorf("%v, ChainCrossingSign(c) = %d, want %d", input, got, want)
	}
	if got, want := crosser.CrossingSign(d, c), robust; got != want {
		return fmt.Errorf("%v, CrossingSign(d, c) = %d, want %d", input, got, want)
	}
	if got, want := crosser.CrossingSign(c, d), robust; got != want {
		return fmt.Errorf("%v, CrossingSign(c, d) = %d, want %d", input, got, want)
	}

	crosser.RestartAt(c)
	if got, want := crosser.EdgeOrVertexChainCrossing(d), edgeOrVertex; got != want {
		return fmt.Errorf("%v, EdgeOrVertexChainCrossing(d) = %t, want %t", input, got, want)
	}
	if got, want := crosser.EdgeOrVertexChainCrossing(c), edgeOrVertex; got != want {
		return fmt.Errorf("%v, EdgeOrVertexChainCrossing(c) = %t, want %t", input, got, want)
	}
	if got, want := crosser.EdgeOrVertexCrossing(d, c), edgeOrVertex; got != want {
		return fmt.Errorf("%v, EdgeOrVertexCrossing(d, c) = %t, want %t", input, got, want)
	}
	if got, want := crosser.EdgeOrVertexCrossing(c, d), edgeOrVertex; got != want {
		return fmt.Errorf("%v, EdgeOrVertexCrossing(c, d) = %t, want %t", input, got, want)
	}
	return nil
}

func TestEdgeutilInterpolate(t *testing.T) {
	// Choose test points designed to expose floating-point errors.
	p1 := PointFromCoords(0.1, 1e-30, 0.3)
	p2 := PointFromCoords(-0.7, -0.55, -1e30)

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
		{p1, p2, 0.5, Point{(p1.Add(p2.Vector)).Mul(0.5)}},

		// Test that interpolation is done using distances on the sphere
		// rather than linear distances.
		{
			Point{r3.Vector{1, 0, 0}},
			Point{r3.Vector{0, 1, 0}},
			1.0 / 3.0,
			Point{r3.Vector{math.Sqrt(3), 1, 0}},
		},
		{
			Point{r3.Vector{1, 0, 0}},
			Point{r3.Vector{0, 1, 0}},
			2.0 / 3.0,
			Point{r3.Vector{1, math.Sqrt(3), 0}},
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

func TestEdgeutilInterpolateOverLongEdge(t *testing.T) {
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

func TestEdgeutilInterpolateAntipodal(t *testing.T) {
	p1 := PointFromCoords(0.1, 1e-30, 0.3)

	// Test that interpolation on a 180 degree edge (antipodal endpoints) yields
	// a result with the correct distance from each endpoint.
	for dist := 0.0; dist <= 1.0; dist += 0.125 {
		actual := Interpolate(dist, p1, Point{p1.Mul(-1)})
		if !float64Near(actual.Distance(p1).Radians(), dist*math.Pi, 3e-15) {
			t.Errorf("antipodal points Interpolate(%v, %v, %v) = %v, want %v", dist, p1, Point{p1.Mul(-1)}, actual, dist*math.Pi)
		}
	}
}

func TestEdgeutilRepeatedInterpolation(t *testing.T) {
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

func rectBoundForPoints(a, b Point) Rect {
	bounder := NewRectBounder()
	bounder.AddPoint(a)
	bounder.AddPoint(b)
	return bounder.RectBound()
}

func TestEdgeutilRectBounderMaxLatitudeSimple(t *testing.T) {
	cubeLat := math.Asin(1 / math.Sqrt(3)) // 35.26 degrees
	cubeLatRect := Rect{r1.IntervalFromPoint(-cubeLat).AddPoint(cubeLat),
		s1.IntervalFromEndpoints(-math.Pi/4, math.Pi/4)}

	tests := []struct {
		a, b Point
		want Rect
	}{
		// Check cases where the min/max latitude is attained at a vertex.
		{
			a:    Point{r3.Vector{1, 1, 1}},
			b:    Point{r3.Vector{1, -1, -1}},
			want: cubeLatRect,
		},
		{
			a:    Point{r3.Vector{1, -1, 1}},
			b:    Point{r3.Vector{1, 1, -1}},
			want: cubeLatRect,
		},
	}

	for _, test := range tests {
		if got := rectBoundForPoints(test.a, test.b); !rectsApproxEqual(got, test.want, rectErrorLat, rectErrorLng) {
			t.Errorf("RectBounder for points (%v, %v) near max lat failed: got %v, want %v", test.a, test.b, got, test.want)
		}
	}
}

func TestEdgeutilRectBounderMaxLatitudeEdgeInterior(t *testing.T) {
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
			rectBoundForPoints(Point{r3.Vector{1, 1, 1}}, Point{r3.Vector{1, -1, 1}}).Lat.Hi,
		},
		// Min latitude, CW edge
		{
			-math.Pi/4 - 0.5*rectErrorLat,
			rectBoundForPoints(Point{r3.Vector{1, -1, -1}}, Point{r3.Vector{-1, -1, -1}}).Lat.Lo,
		},
		// Max latitude, CCW edge
		{
			math.Pi/4 + 0.5*rectErrorLat,
			rectBoundForPoints(Point{r3.Vector{1, -1, 1}}, Point{r3.Vector{1, 1, 1}}).Lat.Hi,
		},
		// Min latitude, CCW edge
		{
			-math.Pi/4 - 0.5*rectErrorLat,
			rectBoundForPoints(Point{r3.Vector{-1, 1, -1}}, Point{r3.Vector{-1, -1, -1}}).Lat.Lo,
		},

		// Check cases where the edge passes through one of the poles.
		{
			math.Pi / 2,
			rectBoundForPoints(Point{r3.Vector{.3, .4, 1}}, Point{r3.Vector{-.3, -.4, 1}}).Lat.Hi,
		},
		{
			-math.Pi / 2,
			rectBoundForPoints(Point{r3.Vector{.3, .4, -1}}, Point{r3.Vector{-.3, -.4, -1}}).Lat.Lo,
		},
	}

	for _, test := range tests {
		if !float64Eq(test.got, test.want) {
			t.Errorf("RectBound for max lat on interior of edge failed; got %v want %v", test.got, test.want)
		}
	}
}

func TestEdgeutilRectBounderMaxLatitudeRandom(t *testing.T) {
	// Check that the maximum latitude of edges is computed accurately to within
	// 3 * dblEpsilon (the expected maximum error). We concentrate on maximum
	// latitudes near the equator and north pole since these are the extremes.

	for i := 0; i < 100; i++ {
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

func TestEdgeutilExpandForSubregions(t *testing.T) {
	// Test the full and empty bounds.
	if !ExpandForSubregions(FullRect()).IsFull() {
		t.Errorf("Subregion Bound of full rect should be full")
	}
	if !ExpandForSubregions(EmptyRect()).IsEmpty() {
		t.Errorf("Subregion Bound of empty rect should be empty")
	}

	tests := []struct {
		xLat, xLng, yLat, yLng float64
		wantFull               bool
	}{
		// Cases where the bound does not straddle the equator (but almost does),
		// and spans nearly 180 degrees in longitude.
		{3e-16, 0, 1e-14, math.Pi, true},
		{9e-16, 0, 1e-14, math.Pi, false},
		{1e-16, 7e-16, 1e-14, math.Pi, true},
		{3e-16, 14e-16, 1e-14, math.Pi, false},
		{1e-100, 14e-16, 1e-14, math.Pi, true},
		{1e-100, 22e-16, 1e-14, math.Pi, false},
		// Cases where the bound spans at most 90 degrees in longitude, and almost
		// 180 degrees in latitude.  Note that DBL_EPSILON is about 2.22e-16, which
		// implies that the double-precision value just below Pi/2 can be written as
		// (math.Pi/2 - 2e-16).
		{-math.Pi / 2, -1e-15, math.Pi/2 - 7e-16, 0, true},
		{-math.Pi / 2, -1e-15, math.Pi/2 - 30e-16, 0, false},
		{-math.Pi/2 + 4e-16, 0, math.Pi/2 - 2e-16, 1e-7, true},
		{-math.Pi/2 + 30e-16, 0, math.Pi / 2, 1e-7, false},
		{-math.Pi/2 + 4e-16, 0, math.Pi/2 - 4e-16, math.Pi / 2, true},
		{-math.Pi / 2, 0, math.Pi/2 - 30e-16, math.Pi / 2, false},
		// Cases where the bound straddles the equator and spans more than 90
		// degrees in longitude.  These are the cases where the critical distance is
		// between a corner of the bound and the opposite longitudinal edge.  Unlike
		// the cases above, here the bound may contain nearly-antipodal points (to
		// within 3.055 * DBL_EPSILON) even though the latitude and longitude ranges
		// are both significantly less than (math.Pi - 3.055 * DBL_EPSILON).
		{-math.Pi / 2, 0, math.Pi/2 - 1e-8, math.Pi - 1e-7, true},
		{-math.Pi / 2, 0, math.Pi/2 - 1e-7, math.Pi - 1e-7, false},
		{-math.Pi/2 + 1e-12, -math.Pi + 1e-4, math.Pi / 2, 0, true},
		{-math.Pi/2 + 1e-11, -math.Pi + 1e-4, math.Pi / 2, 0, true},
	}

	for _, tc := range tests {
		in := RectFromLatLng(LatLng{s1.Angle(tc.xLat), s1.Angle(tc.xLng)})
		in = in.AddPoint(LatLng{s1.Angle(tc.yLat), s1.Angle(tc.yLng)})
		got := ExpandForSubregions(in)

		// Test that the bound is actually expanded.
		if !got.Contains(in) {
			t.Errorf("Subregion bound of (%f, %f, %f, %f) should contain original rect", tc.xLat, tc.xLng, tc.yLat, tc.yLng)
		}
		if in.Lat == validRectLatRange && in.Lat.ContainsInterval(got.Lat) {
			t.Errorf("Subregion bound of (%f, %f, %f, %f) shouldn't be contained by original rect", tc.xLat, tc.xLng, tc.yLat, tc.yLng)
		}

		// We check the various situations where the bound contains nearly-antipodal points. The tests are organized into pairs
		// where the two bounds are similar except that the first bound meets the nearly-antipodal criteria while the second does not.
		if got.IsFull() != tc.wantFull {
			t.Errorf("Subregion Bound of (%f, %f, %f, %f).IsFull should be %t", tc.xLat, tc.xLng, tc.yLat, tc.yLng, tc.wantFull)
		}
	}

	rectTests := []struct {
		xLat, xLng, yLat, yLng float64
		wantRect               Rect
	}{
		{1.5, -math.Pi / 2, 1.5, math.Pi/2 - 2e-16, Rect{r1.Interval{1.5, 1.5}, s1.FullInterval()}},
		{1.5, -math.Pi / 2, 1.5, math.Pi/2 - 7e-16, Rect{r1.Interval{1.5, 1.5}, s1.Interval{-math.Pi / 2, math.Pi/2 - 7e-16}}},
		// Check for cases where the bound is expanded to include one of the poles
		{-math.Pi/2 + 1e-15, 0, -math.Pi/2 + 1e-15, 0, Rect{r1.Interval{-math.Pi / 2, -math.Pi/2 + 1e-15}, s1.FullInterval()}},
		{math.Pi/2 - 1e-15, 0, math.Pi/2 - 1e-15, 0, Rect{r1.Interval{math.Pi/2 - 1e-15, math.Pi / 2}, s1.FullInterval()}},
	}

	for _, tc := range rectTests {
		// Now we test cases where the bound does not contain nearly-antipodal
		// points, but it does contain points that are approximately 180 degrees
		// apart in latitude.
		in := RectFromLatLng(LatLng{s1.Angle(tc.xLat), s1.Angle(tc.xLng)})
		in = in.AddPoint(LatLng{s1.Angle(tc.yLat), s1.Angle(tc.yLng)})
		got := ExpandForSubregions(in)
		if !rectsApproxEqual(got, tc.wantRect, rectErrorLat, rectErrorLng) {
			t.Errorf("Subregion Bound of (%f, %f, %f, %f) = (%v) should be %v", tc.xLat, tc.xLng, tc.yLat, tc.yLng, got, tc.wantRect)
		}
	}
}

func TestEdgeutilIntersectsFace(t *testing.T) {
	tests := []struct {
		a    pointUVW
		want bool
	}{
		{pointUVW{r3.Vector{2.05335e-06, 3.91604e-22, 2.90553e-06}}, false},
		{pointUVW{r3.Vector{-3.91604e-22, -2.05335e-06, -2.90553e-06}}, false},
		{pointUVW{r3.Vector{0.169258, -0.169258, 0.664013}}, false},
		{pointUVW{r3.Vector{0.169258, -0.169258, -0.664013}}, false},
		{pointUVW{r3.Vector{math.Sqrt(2.0 / 3.0), -math.Sqrt(2.0 / 3.0), 3.88578e-16}}, true},
		{pointUVW{r3.Vector{-3.88578e-16, -math.Sqrt(2.0 / 3.0), math.Sqrt(2.0 / 3.0)}}, true},
	}

	for _, test := range tests {
		if got := test.a.intersectsFace(); got != test.want {
			t.Errorf("%v.intersectsFace() = %v, want %v", test.a, got, test.want)
		}
	}
}

func TestEdgeutilIntersectsOppositeEdges(t *testing.T) {
	tests := []struct {
		a    pointUVW
		want bool
	}{
		{pointUVW{r3.Vector{0.169258, -0.169258, 0.664013}}, false},
		{pointUVW{r3.Vector{0.169258, -0.169258, -0.664013}}, false},

		{pointUVW{r3.Vector{-math.Sqrt(4.0 / 3.0), 0, -math.Sqrt(4.0 / 3.0)}}, true},
		{pointUVW{r3.Vector{math.Sqrt(4.0 / 3.0), 0, math.Sqrt(4.0 / 3.0)}}, true},

		{pointUVW{r3.Vector{-math.Sqrt(2.0 / 3.0), -math.Sqrt(2.0 / 3.0), 1.66533453694e-16}}, false},
		{pointUVW{r3.Vector{math.Sqrt(2.0 / 3.0), math.Sqrt(2.0 / 3.0), -1.66533453694e-16}}, false},
	}
	for _, test := range tests {
		if got := test.a.intersectsOppositeEdges(); got != test.want {
			t.Errorf("%v.intersectsOppositeEdges() = %v, want %v", test.a, got, test.want)
		}
	}
}

func TestEdgeutilExitAxis(t *testing.T) {
	tests := []struct {
		a    pointUVW
		want axis
	}{
		{pointUVW{r3.Vector{0, -math.Sqrt(2.0 / 3.0), math.Sqrt(2.0 / 3.0)}}, axisU},
		{pointUVW{r3.Vector{0, math.Sqrt(4.0 / 3.0), -math.Sqrt(4.0 / 3.0)}}, axisU},
		{pointUVW{r3.Vector{-math.Sqrt(4.0 / 3.0), -math.Sqrt(4.0 / 3.0), 0}}, axisV},
		{pointUVW{r3.Vector{math.Sqrt(4.0 / 3.0), math.Sqrt(4.0 / 3.0), 0}}, axisV},
		{pointUVW{r3.Vector{math.Sqrt(2.0 / 3.0), -math.Sqrt(2.0 / 3.0), 0}}, axisV},
		{pointUVW{r3.Vector{1.67968702783622, 0, 0.870988820096491}}, axisV},
		{pointUVW{r3.Vector{0, math.Sqrt2, math.Sqrt2}}, axisU},
	}

	for _, test := range tests {
		if got := test.a.exitAxis(); got != test.want {
			t.Errorf("%v.exitAxis() = %v, want %v", test.a, got, test.want)
		}
	}
}

func TestEdgeutilExitPoint(t *testing.T) {
	tests := []struct {
		a        pointUVW
		exitAxis axis
		want     r2.Point
	}{
		{pointUVW{r3.Vector{-3.88578058618805e-16, -math.Sqrt(2.0 / 3.0), math.Sqrt(2.0 / 3.0)}}, axisU, r2.Point{-1, 1}},
		{pointUVW{r3.Vector{math.Sqrt(4.0 / 3.0), -math.Sqrt(4.0 / 3.0), 0}}, axisV, r2.Point{-1, -1}},
		{pointUVW{r3.Vector{-math.Sqrt(4.0 / 3.0), -math.Sqrt(4.0 / 3.0), 0}}, axisV, r2.Point{-1, 1}},
		{pointUVW{r3.Vector{-6.66134e-16, math.Sqrt(4.0 / 3.0), -math.Sqrt(4.0 / 3.0)}}, axisU, r2.Point{1, 1}},
	}

	for _, test := range tests {
		if got := test.a.exitPoint(test.exitAxis); !r2PointsApproxEquals(got, test.want, epsilon) {
			t.Errorf("%v.exitPoint() = %v, want %v", test.a, got, test.want)
		}
	}
}

// testClipToPaddedFace performs a comprehensive set of tests across all faces and
// with random padding for the given points.
//
// We do this by defining an (x,y) coordinate system for the plane containing AB,
// and converting points along the great circle AB to angles in the range
// [-Pi, Pi]. We then accumulate the angle intervals spanned by each
// clipped edge; the union over all 6 faces should approximately equal the
// interval covered by the original edge.
func testClipToPaddedFace(t *testing.T, a, b Point) {
	a = Point{a.Normalize()}
	b = Point{b.Normalize()}
	if a.Vector == b.Mul(-1) {
		return
	}

	norm := Point{a.PointCross(b).Normalize()}
	aTan := Point{norm.Cross(a.Vector)}

	padding := 0.0
	if !oneIn(10) {
		padding = 1e-10 * math.Pow(1e-5, randomFloat64())
	}

	xAxis := a
	yAxis := aTan

	// Given the points A and B, we expect all angles generated from the clipping
	// to fall within this range.
	expectedAngles := s1.Interval{0, float64(a.Angle(b.Vector))}
	if expectedAngles.IsInverted() {
		expectedAngles = s1.Interval{expectedAngles.Hi, expectedAngles.Lo}
	}
	maxAngles := expectedAngles.Expanded(faceClipErrorRadians)
	var actualAngles s1.Interval

	for face := 0; face < 6; face++ {
		aUV, bUV, intersects := ClipToPaddedFace(a, b, face, padding)
		if !intersects {
			continue
		}

		aClip := Point{faceUVToXYZ(face, aUV.X, aUV.Y).Normalize()}
		bClip := Point{faceUVToXYZ(face, bUV.X, bUV.Y).Normalize()}

		desc := fmt.Sprintf("on face %d, a=%v, b=%v, aClip=%v, bClip=%v,", face, a, b, aClip, bClip)

		if got := math.Abs(aClip.Dot(norm.Vector)); got > faceClipErrorRadians {
			t.Errorf("%s abs(%v.Dot(%v)) = %v, want <= %v", desc, aClip, norm, got, faceClipErrorRadians)
		}
		if got := math.Abs(bClip.Dot(norm.Vector)); got > faceClipErrorRadians {
			t.Errorf("%s abs(%v.Dot(%v)) = %v, want <= %v", desc, bClip, norm, got, faceClipErrorRadians)
		}

		if float64(aClip.Angle(a.Vector)) > faceClipErrorRadians {
			if got := math.Max(math.Abs(aUV.X), math.Abs(aUV.Y)); !float64Eq(got, 1+padding) {
				t.Errorf("%s the largest component of %v = %v, want %v", desc, aUV, got, 1+padding)
			}
		}
		if float64(bClip.Angle(b.Vector)) > faceClipErrorRadians {
			if got := math.Max(math.Abs(bUV.X), math.Abs(bUV.Y)); !float64Eq(got, 1+padding) {
				t.Errorf("%s the largest component of %v = %v, want %v", desc, bUV, got, 1+padding)
			}
		}

		aAngle := math.Atan2(aClip.Dot(yAxis.Vector), aClip.Dot(xAxis.Vector))
		bAngle := math.Atan2(bClip.Dot(yAxis.Vector), bClip.Dot(xAxis.Vector))

		// Rounding errors may cause bAngle to be slightly less than aAngle.
		// We handle this by constructing the interval with FromPointPair,
		// which is okay since the interval length is much less than math.Pi.
		faceAngles := s1.IntervalFromEndpoints(aAngle, bAngle)
		if faceAngles.IsInverted() {
			faceAngles = s1.Interval{faceAngles.Hi, faceAngles.Lo}
		}
		if !maxAngles.ContainsInterval(faceAngles) {
			t.Errorf("%s %v.ContainsInterval(%v) = false, but should have contained this interval", desc, maxAngles, faceAngles)
		}
		actualAngles = actualAngles.Union(faceAngles)
	}
	if !actualAngles.Expanded(faceClipErrorRadians).ContainsInterval(expectedAngles) {
		t.Errorf("the union of all angle segments should be larger than the expected angle")
	}
}

func TestEdgeutilFaceClipping(t *testing.T) {
	// Start with a few simple cases.
	// An edge that is entirely contained within one cube face:
	testClipToPaddedFace(t, Point{r3.Vector{1, -0.5, -0.5}}, Point{r3.Vector{1, 0.5, 0.5}})
	testClipToPaddedFace(t, Point{r3.Vector{1, 0.5, 0.5}}, Point{r3.Vector{1, -0.5, -0.5}})
	// An edge that crosses one cube edge:
	testClipToPaddedFace(t, Point{r3.Vector{1, 0, 0}}, Point{r3.Vector{0, 1, 0}})
	testClipToPaddedFace(t, Point{r3.Vector{0, 1, 0}}, Point{r3.Vector{1, 0, 0}})
	// An edge that crosses two opposite edges of face 0:
	testClipToPaddedFace(t, Point{r3.Vector{0.75, 0, -1}}, Point{r3.Vector{0.75, 0, 1}})
	testClipToPaddedFace(t, Point{r3.Vector{0.75, 0, 1}}, Point{r3.Vector{0.75, 0, -1}})
	// An edge that crosses two adjacent edges of face 2:
	testClipToPaddedFace(t, Point{r3.Vector{1, 0, 0.75}}, Point{r3.Vector{0, 1, 0.75}})
	testClipToPaddedFace(t, Point{r3.Vector{0, 1, 0.75}}, Point{r3.Vector{1, 0, 0.75}})
	// An edges that crosses three cube edges (four faces):
	testClipToPaddedFace(t, Point{r3.Vector{1, 0.9, 0.95}}, Point{r3.Vector{-1, 0.95, 0.9}})
	testClipToPaddedFace(t, Point{r3.Vector{-1, 0.95, 0.9}}, Point{r3.Vector{1, 0.9, 0.95}})

	// Comprehensively test edges that are difficult to handle, especially those
	// that nearly follow one of the 12 cube edges.
	biunit := r2.Rect{r1.Interval{-1, 1}, r1.Interval{-1, 1}}

	for i := 0; i < 1000; i++ {
		// Choose two adjacent cube corners P and Q.
		face := randomUniformInt(6)
		i := randomUniformInt(4)
		j := (i + 1) & 3
		p := Point{faceUVToXYZ(face, biunit.Vertices()[i].X, biunit.Vertices()[i].Y)}
		q := Point{faceUVToXYZ(face, biunit.Vertices()[j].X, biunit.Vertices()[j].Y)}

		// Now choose two points that are nearly in the plane of PQ, preferring
		// points that are near cube corners, face midpoints, or edge midpoints.
		a := perturbedCornerOrMidpoint(p, q)
		b := perturbedCornerOrMidpoint(p, q)
		testClipToPaddedFace(t, a, b)
	}
}

// getFraction returns the fraction t of the given point X on the line AB such that
// x = (1-t)*a + t*b. Returns 0 if A = B.
func getFraction(t *testing.T, x, a, b r2.Point) float64 {
	// A bound for the error in edge clipping plus the error in the calculation
	// (which is similar to EdgeIntersectsRect).
	errorDist := (edgeClipErrorUVDist + intersectsRectErrorUVDist)
	if a == b {
		return 0.0
	}
	dir := b.Sub(a).Normalize()
	if got := math.Abs(x.Sub(a).Dot(dir.Ortho())); got > errorDist {
		t.Errorf("getFraction(%v, %v, %v) = %v, which exceeds errorDist %v", x, a, b, got, errorDist)
	}
	return x.Sub(a).Dot(dir)
}

// randomPointFromInterval returns a randomly selected point from the given interval
// with one of three possible choices. All cases have reasonable probability for any
// interval. The choices are: randomly choose a value inside the interval, choose a
// value outside the interval, or select one of the two endpoints.
func randomPointFromInterval(clip r1.Interval) float64 {
	if oneIn(5) {
		if oneIn(2) {
			return clip.Lo
		}
		return clip.Hi
	}

	switch randomUniformInt(3) {
	case 0:
		return clip.Lo - randomFloat64()
	case 1:
		return clip.Hi + randomFloat64()
	default:
		return clip.Lo + randomFloat64()*clip.Length()
	}
}

// Given a rectangle "clip", choose a point that may lie in the rectangle interior, along an extended edge, exactly at a vertex, or in one of the eight regions exterior to "clip" that are separated by its extended edges.  Also sometimes return points that are exactly on one of the extended diagonals of "clip". All cases are reasonably likely to occur for any given rectangle "clip".
func chooseRectEndpoint(clip r2.Rect) r2.Point {
	if oneIn(10) {
		// Return a point on one of the two extended diagonals.
		diag := randomUniformInt(2)
		t := randomUniformFloat64(-1, 2)
		return clip.Vertices()[diag].Mul(1 - t).Add(clip.Vertices()[diag+2].Mul(t))
	}
	return r2.Point{randomPointFromInterval(clip.X), randomPointFromInterval(clip.Y)}
}

// Choose a random point in the rectangle defined by points A and B, sometimes
// returning a point on the edge AB or the points A and B themselves.
func choosePointInRect(a, b r2.Point) r2.Point {
	if oneIn(5) {
		if oneIn(2) {
			return a
		}
		return b
	}

	if oneIn(3) {
		return a.Add(b.Sub(a).Mul(randomFloat64()))
	}
	return r2.Point{randomUniformFloat64(a.X, b.X), randomUniformFloat64(a.Y, b.Y)}
}

// Given a point P representing a possibly clipped endpoint A of an edge AB,
// verify that clip contains P, and that if clipping occurred (i.e., P != A)
// then P is on the boundary of clip.
func checkPointOnBoundary(t *testing.T, p, a r2.Point, clip r2.Rect) {
	if got := clip.ContainsPoint(p); !got {
		t.Errorf("%v.ContainsPoint(%v) = %v, want true", clip, p, got)
	}
	if p != a {
		p1 := r2.Point{math.Nextafter(p.X, a.X), math.Nextafter(p.Y, a.Y)}
		if got := clip.ContainsPoint(p1); got {
			t.Errorf("%v.ContainsPoint(%v) = %v, want false", clip, p1, got)
		}
	}
}

func TestEdgeutilEdgeClipping(t *testing.T) {
	// A bound for the error in edge clipping plus the error in the
	// EdgeIntersectsRect calculation below.
	errorDist := (edgeClipErrorUVDist + intersectsRectErrorUVDist)
	testRects := []r2.Rect{
		// Test clipping against random rectangles.
		r2.RectFromPoints(
			r2.Point{randomUniformFloat64(-1, 1), randomUniformFloat64(-1, 1)},
			r2.Point{randomUniformFloat64(-1, 1), randomUniformFloat64(-1, 1)}),
		r2.RectFromPoints(
			r2.Point{randomUniformFloat64(-1, 1), randomUniformFloat64(-1, 1)},
			r2.Point{randomUniformFloat64(-1, 1), randomUniformFloat64(-1, 1)}),
		r2.RectFromPoints(
			r2.Point{randomUniformFloat64(-1, 1), randomUniformFloat64(-1, 1)},
			r2.Point{randomUniformFloat64(-1, 1), randomUniformFloat64(-1, 1)}),
		r2.RectFromPoints(
			r2.Point{randomUniformFloat64(-1, 1), randomUniformFloat64(-1, 1)},
			r2.Point{randomUniformFloat64(-1, 1), randomUniformFloat64(-1, 1)}),
		r2.RectFromPoints(
			r2.Point{randomUniformFloat64(-1, 1), randomUniformFloat64(-1, 1)},
			r2.Point{randomUniformFloat64(-1, 1), randomUniformFloat64(-1, 1)}),

		// Also clip against one-dimensional, singleton, and empty rectangles.
		r2.Rect{r1.Interval{-0.7, -0.7}, r1.Interval{0.3, 0.35}},
		r2.Rect{r1.Interval{0.2, 0.5}, r1.Interval{0.3, 0.3}},
		r2.Rect{r1.Interval{-0.7, 0.3}, r1.Interval{0, 0}},
		r2.RectFromPoints(r2.Point{0.3, 0.8}),
		r2.EmptyRect(),
	}

	for _, r := range testRects {
		for i := 0; i < 1000; i++ {
			a := chooseRectEndpoint(r)
			b := chooseRectEndpoint(r)

			aClip, bClip, intersects := ClipEdge(a, b, r)
			if !intersects {
				if edgeIntersectsRect(a, b, r.ExpandedByMargin(-errorDist)) {
					t.Errorf("edgeIntersectsRect(%v, %v, %v.ExpandedByMargin(%v) = true, want false", a, b, r, -errorDist)
				}
			} else {
				if !edgeIntersectsRect(a, b, r.ExpandedByMargin(errorDist)) {
					t.Errorf("edgeIntersectsRect(%v, %v, %v.ExpandedByMargin(%v) = false, want true", a, b, r, errorDist)
				}

				// Check that the clipped points lie on the edge AB, and
				// that the points have the expected order along the segment AB.
				if gotA, gotB := getFraction(t, aClip, a, b), getFraction(t, bClip, a, b); gotA > gotB {
					t.Errorf("getFraction(%v,%v,%v) = %v, getFraction(%v, %v, %v) = %v; %v < %v = false, want true", aClip, a, b, gotA, bClip, a, b, gotB, gotA, gotB)
				}

				// Check that the clipped portion of AB is as large as possible.
				checkPointOnBoundary(t, aClip, a, r)
				checkPointOnBoundary(t, bClip, b, r)
			}

			// Choose an random initial bound to pass to clipEdgeBound.
			initialClip := r2.RectFromPoints(choosePointInRect(a, b), choosePointInRect(a, b))
			bound := clippedEdgeBound(a, b, initialClip)
			if bound.IsEmpty() {
				// Precondition of clipEdgeBound not met
				continue
			}
			maxBound := bound.Intersection(r)
			if bound, intersects := clipEdgeBound(a, b, r, bound); !intersects {
				if edgeIntersectsRect(a, b, maxBound.ExpandedByMargin(-errorDist)) {
					t.Errorf("edgeIntersectsRect(%v, %v, %v.ExpandedByMargin(%v) = true, want false", a, b, maxBound.ExpandedByMargin(-errorDist), -errorDist)
				}
			} else {
				if !edgeIntersectsRect(a, b, maxBound.ExpandedByMargin(errorDist)) {
					t.Errorf("edgeIntersectsRect(%v, %v, %v.ExpandedByMargin(%v) = false, want true", a, b, maxBound.ExpandedByMargin(errorDist), errorDist)
				}
				// check that the bound is as large as possible.
				ai := 0
				if a.X > b.X {
					ai = 1
				}
				aj := 0
				if a.Y > b.Y {
					aj = 1
				}
				checkPointOnBoundary(t, bound.VertexIJ(ai, aj), a, maxBound)
				checkPointOnBoundary(t, bound.VertexIJ(1-ai, 1-aj), b, maxBound)
			}
		}
	}
}

func TestCheckDistance(t *testing.T) {
	// Uncomment once Distance / UpdateMinDistance are implemented.
	//var zeroChordAngle s1.ChordAngle
	tests := []struct {
		x, a, b r3.Vector
		distRad float64
		want    r3.Vector
	}{
		{
			x:       r3.Vector{1, 0, 0},
			a:       r3.Vector{1, 0, 0},
			b:       r3.Vector{0, 1, 0},
			distRad: 0,
			want:    r3.Vector{1, 0, 0},
		},
		{
			x:       r3.Vector{0, 1, 0},
			a:       r3.Vector{1, 0, 0},
			b:       r3.Vector{0, 1, 0},
			distRad: 0,
			want:    r3.Vector{0, 1, 0},
		},
		{
			x:       r3.Vector{1, 3, 0},
			a:       r3.Vector{1, 0, 0},
			b:       r3.Vector{0, 1, 0},
			distRad: 0,
			want:    r3.Vector{1, 3, 0},
		},
		{
			x:       r3.Vector{0, 0, 1},
			a:       r3.Vector{1, 0, 0},
			b:       r3.Vector{0, 1, 0},
			distRad: math.Pi / 2,
			want:    r3.Vector{1, 0, 0},
		},
		{
			x:       r3.Vector{0, 0, -1},
			a:       r3.Vector{1, 0, 0},
			b:       r3.Vector{0, 1, 0},
			distRad: math.Pi / 2,
			want:    r3.Vector{1, 0, 0},
		},
		{
			x:       r3.Vector{-1, -1, 0},
			a:       r3.Vector{1, 0, 0},
			b:       r3.Vector{0, 1, 0},
			distRad: 0.75 * math.Pi,
			want:    r3.Vector{1, 0, 0},
		},
		{
			x:       r3.Vector{0, 1, 0},
			a:       r3.Vector{1, 0, 0},
			b:       r3.Vector{1, 1, 0},
			distRad: math.Pi / 4,
			want:    r3.Vector{1, 1, 0},
		},
		{
			x:       r3.Vector{0, -1, 0},
			a:       r3.Vector{1, 0, 0},
			b:       r3.Vector{1, 1, 0},
			distRad: math.Pi / 2,
			want:    r3.Vector{1, 0, 0},
		},
		{
			x:       r3.Vector{0, -1, 0},
			a:       r3.Vector{1, 0, 0},
			b:       r3.Vector{-1, 1, 0},
			distRad: math.Pi / 2,
			want:    r3.Vector{1, 0, 0},
		},
		{
			x:       r3.Vector{-1, -1, 0},
			a:       r3.Vector{1, 0, 0},
			b:       r3.Vector{-1, 1, 0},
			distRad: math.Pi / 2,
			want:    r3.Vector{-1, 1, 0},
		},
		{
			x:       r3.Vector{1, 1, 1},
			a:       r3.Vector{1, 0, 0},
			b:       r3.Vector{0, 1, 0},
			distRad: math.Asin(math.Sqrt(1.0 / 3.0)),
			want:    r3.Vector{1, 1, 0},
		},
		{
			x:       r3.Vector{1, 1, -1},
			a:       r3.Vector{1, 0, 0},
			b:       r3.Vector{0, 1, 0},
			distRad: math.Asin(math.Sqrt(1.0 / 3.0)),
			want:    r3.Vector{1, 1, 0}},
		{
			x:       r3.Vector{-1, 0, 0},
			a:       r3.Vector{1, 1, 0},
			b:       r3.Vector{1, 1, 0},
			distRad: 0.75 * math.Pi,
			want:    r3.Vector{1, 1, 0},
		},
		{
			x:       r3.Vector{0, 0, -1},
			a:       r3.Vector{1, 1, 0},
			b:       r3.Vector{1, 1, 0},
			distRad: math.Pi / 2,
			want:    r3.Vector{1, 1, 0},
		},
		{
			x:       r3.Vector{-1, 0, 0},
			a:       r3.Vector{1, 0, 0},
			b:       r3.Vector{1, 0, 0},
			distRad: math.Pi,
			want:    r3.Vector{1, 0, 0},
		},
	}

	for _, test := range tests {
		x := Point{test.x.Normalize()}
		a := Point{test.a.Normalize()}
		b := Point{test.b.Normalize()}
		want := Point{test.want.Normalize()}

		if d := DistanceFromSegment(x, a, b).Radians(); !float64Near(d, test.distRad, 1e-15) {
			t.Errorf("DistanceFromSegment(%v, %v, %v) = %v, want %v", x, a, b, d, test.distRad)
		}

		closest := ClosestPoint(x, a, b)
		if !closest.ApproxEqual(want) {
			t.Errorf("ClosestPoint(%v, %v, %v) = %v, want %v", x, a, b, closest, want)
		}

		// Uncomment these once Distance / UpdateMinDistance are implemented.
		//minDistance := zeroChordAngle
		//if minDistance, ok := UpdateMinDistance(x, a, b, minDistance); ok {
		//	t.Errorf("UpdateMinDistance(%x, %v, %v, %v) = %v, want %v", x, a, b, zeroChordAngle, minDistance, zeroChordAngle)
		//}
		//
		//minDistance = s1.InfChordAngle()
		//if minDistance, ok := UpdateMinDistance(x, a, b, minDistance); !ok {
		//	t.Errorf("UpdateMinDistance(%x, %v, %v, %v) = %v, want %v", x, a, b, s1.InfChordAngle(), minDistance, s1.InfChordAngle())
		//}
		//
		//if !float64Near(test.distRad, minDistance.Angle().Radians(), 1e-15) {
		//	t.Errorf("%v != %v", minDistance.Angle().Radians(), test.distRad)
		//}
	}
}

func TestEdgeUtilWedges(t *testing.T) {
	// For simplicity, all of these tests use an origin of (0, 0, 1).
	// This shouldn't matter as long as the lower-level primitives are
	// implemented correctly.
	ab1 := Point{r3.Vector{0, 0, 1}}

	tests := []struct {
		desc           string
		a0, a1, b0, b1 Point
		contains       bool
		intersects     bool
		relation       WedgeRel
	}{
		{
			desc:       "Intersection in one wedge",
			a0:         Point{r3.Vector{-1, 0, 10}},
			a1:         Point{r3.Vector{1, 2, 10}},
			b0:         Point{r3.Vector{0, 1, 10}},
			b1:         Point{r3.Vector{1, -2, 10}},
			contains:   false,
			intersects: true,
			relation:   WedgeProperlyOverlaps,
		},
		{
			desc:       "Intersection in two wedges",
			a0:         Point{r3.Vector{-1, -1, 10}},
			a1:         Point{r3.Vector{1, -1, 10}},
			b0:         Point{r3.Vector{1, 0, 10}},
			b1:         Point{r3.Vector{-1, 1, 10}},
			contains:   false,
			intersects: true,
			relation:   WedgeProperlyOverlaps,
		},
		{
			desc:       "Normal containment",
			a0:         Point{r3.Vector{-1, -1, 10}},
			a1:         Point{r3.Vector{1, -1, 10}},
			b0:         Point{r3.Vector{-1, 0, 10}},
			b1:         Point{r3.Vector{1, 0, 10}},
			contains:   true,
			intersects: true,
			relation:   WedgeProperlyContains,
		},
		{
			desc:       "Containment with equality on one side",
			a0:         Point{r3.Vector{2, 1, 10}},
			a1:         Point{r3.Vector{-1, -1, 10}},
			b0:         Point{r3.Vector{2, 1, 10}},
			b1:         Point{r3.Vector{1, -5, 10}},
			contains:   true,
			intersects: true,
			relation:   WedgeProperlyContains,
		},
		{
			desc:       "Containment with equality on the other side",
			a0:         Point{r3.Vector{2, 1, 10}},
			a1:         Point{r3.Vector{-1, -1, 10}},
			b0:         Point{r3.Vector{1, -2, 10}},
			b1:         Point{r3.Vector{-1, -1, 10}},
			contains:   true,
			intersects: true,
			relation:   WedgeProperlyContains,
		},
		{
			desc:       "Containment with equality on both sides",
			a0:         Point{r3.Vector{-2, 3, 10}},
			a1:         Point{r3.Vector{4, -5, 10}},
			b0:         Point{r3.Vector{-2, 3, 10}},
			b1:         Point{r3.Vector{4, -5, 10}},
			contains:   true,
			intersects: true,
			relation:   WedgeEquals,
		},
		{
			desc:       "Disjoint with equality on one side",
			a0:         Point{r3.Vector{-2, 3, 10}},
			a1:         Point{r3.Vector{4, -5, 10}},
			b0:         Point{r3.Vector{4, -5, 10}},
			b1:         Point{r3.Vector{-2, -3, 10}},
			contains:   false,
			intersects: false,
			relation:   WedgeIsDisjoint,
		},
		{
			desc:       "Disjoint with equality on the other side",
			a0:         Point{r3.Vector{-2, 3, 10}},
			a1:         Point{r3.Vector{0, 5, 10}},
			b0:         Point{r3.Vector{4, -5, 10}},
			b1:         Point{r3.Vector{-2, 3, 10}},
			contains:   false,
			intersects: false,
			relation:   WedgeIsDisjoint,
		},
		{
			desc:       "Disjoint with equality on both sides",
			a0:         Point{r3.Vector{-2, 3, 10}},
			a1:         Point{r3.Vector{4, -5, 10}},
			b0:         Point{r3.Vector{4, -5, 10}},
			b1:         Point{r3.Vector{-2, 3, 10}},
			contains:   false,
			intersects: false,
			relation:   WedgeIsDisjoint,
		},
		{
			desc:       "B contains A with equality on one side",
			a0:         Point{r3.Vector{2, 1, 10}},
			a1:         Point{r3.Vector{1, -5, 10}},
			b0:         Point{r3.Vector{2, 1, 10}},
			b1:         Point{r3.Vector{-1, -1, 10}},
			contains:   false,
			intersects: true,
			relation:   WedgeIsProperlyContained,
		},

		{
			desc:       "B contains A with equality on the other side",
			a0:         Point{r3.Vector{2, 1, 10}},
			a1:         Point{r3.Vector{1, -5, 10}},
			b0:         Point{r3.Vector{-2, 1, 10}},
			b1:         Point{r3.Vector{1, -5, 10}},
			contains:   false,
			intersects: true,
			relation:   WedgeIsProperlyContained,
		},
	}

	for _, test := range tests {
		if got := WedgeContains(test.a0, ab1, test.a1, test.b0, test.b1); got != test.contains {
			t.Errorf("%s: WedgeContains(%v, %v, %v, %v, %v) = %t, want %t", test.desc, test.a0, ab1, test.a1, test.b0, test.b1, got, test.contains)
		}
		if got := WedgeIntersects(test.a0, ab1, test.a1, test.b0, test.b1); got != test.intersects {
			t.Errorf("%s: WedgeIntersects(%v, %v, %v, %v, %v) = %t, want %t", test.desc, test.a0, ab1, test.a1, test.b0, test.b1, got, test.intersects)
		}
		if got := WedgeRelation(test.a0, ab1, test.a1, test.b0, test.b1); got != test.relation {
			t.Errorf("%s: WedgeRelation(%v, %v, %v, %v, %v) = %v, want %v", test.desc, test.a0, ab1, test.a1, test.b0, test.b1, got, test.relation)
		}
	}
}
