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

	"github.com/golang/geo/s1"
)

func TestSimpleCrossing(t *testing.T) {
	tests := []struct {
		a, b, c, d Point
		want       bool
	}{
		{
			// Two regular edges that cross.
			PointFromCoords(1, 2, 1),
			PointFromCoords(1, -3, 0.5),
			PointFromCoords(1, -0.5, -3),
			PointFromCoords(0.1, 0.5, 3),
			true,
		},
		{
			// Two regular edges that cross antipodal points.
			PointFromCoords(1, 2, 1),
			PointFromCoords(1, -3, 0.5),
			PointFromCoords(-1, 0.5, 3),
			PointFromCoords(-0.1, -0.5, -3),
			false,
		},
		{
			// Two edges on the same great circle.
			PointFromCoords(0, 0, -1),
			PointFromCoords(0, 1, 0),
			PointFromCoords(0, 1, 1),
			PointFromCoords(0, 0, 1),
			false,
		},
		{
			// Two edges that cross where one vertex is the OriginPoint.
			PointFromCoords(1, 0, 0),
			OriginPoint(),
			PointFromCoords(1, -0.1, 1),
			PointFromCoords(1, 1, -0.1),
			true,
		},
		{
			// Two edges that cross antipodal points.
			PointFromCoords(1, 0, 0),
			PointFromCoords(0, 1, 0),
			PointFromCoords(0, 0, -1),
			PointFromCoords(-1, -1, 1),
			false,
		},
		{
			// Two edges that share an endpoint.  The Ortho() direction is (-4,0,2),
			// and edge CD is further CCW around (2,3,4) than AB.
			PointFromCoords(2, 3, 4),
			PointFromCoords(-1, 2, 5),
			PointFromCoords(7, -2, 3),
			PointFromCoords(2, 3, 4),
			false,
		},
	}

	for _, test := range tests {
		if got := SimpleCrossing(test.a, test.b, test.c, test.d); got != test.want {
			t.Errorf("SimpleCrossing(%v,%v,%v,%v) = %t, want %t",
				test.a, test.b, test.c, test.d, got, test.want)
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
