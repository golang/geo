// Copyright 2017 Google Inc. All rights reserved.
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
	"math"
	"testing"

	"github.com/golang/geo/r3"
	"github.com/golang/geo/s1"
)

func TestEdgeDistancesCheckDistance(t *testing.T) {
	tests := []struct {
		x, a, b r3.Vector
		distRad float64
		want    r3.Vector
	}{
		{
			x:       r3.Vector{X: 1, Y: 0, Z: 0},
			a:       r3.Vector{X: 1, Y: 0, Z: 0},
			b:       r3.Vector{X: 0, Y: 1, Z: 0},
			distRad: 0,
			want:    r3.Vector{X: 1, Y: 0, Z: 0},
		},
		{
			x:       r3.Vector{X: 0, Y: 1, Z: 0},
			a:       r3.Vector{X: 1, Y: 0, Z: 0},
			b:       r3.Vector{X: 0, Y: 1, Z: 0},
			distRad: 0,
			want:    r3.Vector{X: 0, Y: 1, Z: 0},
		},
		{
			x:       r3.Vector{X: 1, Y: 3, Z: 0},
			a:       r3.Vector{X: 1, Y: 0, Z: 0},
			b:       r3.Vector{X: 0, Y: 1, Z: 0},
			distRad: 0,
			want:    r3.Vector{X: 1, Y: 3, Z: 0},
		},
		{
			x:       r3.Vector{X: 0, Y: 0, Z: 1},
			a:       r3.Vector{X: 1, Y: 0, Z: 0},
			b:       r3.Vector{X: 0, Y: 1, Z: 0},
			distRad: math.Pi / 2,
			want:    r3.Vector{X: 1, Y: 0, Z: 0},
		},
		{
			x:       r3.Vector{X: 0, Y: 0, Z: -1},
			a:       r3.Vector{X: 1, Y: 0, Z: 0},
			b:       r3.Vector{X: 0, Y: 1, Z: 0},
			distRad: math.Pi / 2,
			want:    r3.Vector{X: 1, Y: 0, Z: 0},
		},
		{
			x:       r3.Vector{X: -1, Y: -1, Z: 0},
			a:       r3.Vector{X: 1, Y: 0, Z: 0},
			b:       r3.Vector{X: 0, Y: 1, Z: 0},
			distRad: 0.75 * math.Pi,
			want:    r3.Vector{X: 1, Y: 0, Z: 0},
		},
		{
			x:       r3.Vector{X: 0, Y: 1, Z: 0},
			a:       r3.Vector{X: 1, Y: 0, Z: 0},
			b:       r3.Vector{X: 1, Y: 1, Z: 0},
			distRad: math.Pi / 4,
			want:    r3.Vector{X: 1, Y: 1, Z: 0},
		},
		{
			x:       r3.Vector{X: 0, Y: -1, Z: 0},
			a:       r3.Vector{X: 1, Y: 0, Z: 0},
			b:       r3.Vector{X: 1, Y: 1, Z: 0},
			distRad: math.Pi / 2,
			want:    r3.Vector{X: 1, Y: 0, Z: 0},
		},
		{
			x:       r3.Vector{X: 0, Y: -1, Z: 0},
			a:       r3.Vector{X: 1, Y: 0, Z: 0},
			b:       r3.Vector{X: -1, Y: 1, Z: 0},
			distRad: math.Pi / 2,
			want:    r3.Vector{X: 1, Y: 0, Z: 0},
		},
		{
			x:       r3.Vector{X: -1, Y: -1, Z: 0},
			a:       r3.Vector{X: 1, Y: 0, Z: 0},
			b:       r3.Vector{X: -1, Y: 1, Z: 0},
			distRad: math.Pi / 2,
			want:    r3.Vector{X: -1, Y: 1, Z: 0},
		},
		{
			x:       r3.Vector{X: 1, Y: 1, Z: 1},
			a:       r3.Vector{X: 1, Y: 0, Z: 0},
			b:       r3.Vector{X: 0, Y: 1, Z: 0},
			distRad: math.Asin(math.Sqrt(1.0 / 3.0)),
			want:    r3.Vector{X: 1, Y: 1, Z: 0},
		},
		{
			x:       r3.Vector{X: 1, Y: 1, Z: -1},
			a:       r3.Vector{X: 1, Y: 0, Z: 0},
			b:       r3.Vector{X: 0, Y: 1, Z: 0},
			distRad: math.Asin(math.Sqrt(1.0 / 3.0)),
			want:    r3.Vector{X: 1, Y: 1, Z: 0},
		},
		{
			x:       r3.Vector{X: -1, Y: 0, Z: 0},
			a:       r3.Vector{X: 1, Y: 1, Z: 0},
			b:       r3.Vector{X: 1, Y: 1, Z: 0},
			distRad: 0.75 * math.Pi,
			want:    r3.Vector{X: 1, Y: 1, Z: 0},
		},
		{
			x:       r3.Vector{X: 0, Y: 0, Z: -1},
			a:       r3.Vector{X: 1, Y: 1, Z: 0},
			b:       r3.Vector{X: 1, Y: 1, Z: 0},
			distRad: math.Pi / 2,
			want:    r3.Vector{X: 1, Y: 1, Z: 0},
		},
		{
			x:       r3.Vector{X: -1, Y: 0, Z: 0},
			a:       r3.Vector{X: 1, Y: 0, Z: 0},
			b:       r3.Vector{X: 1, Y: 0, Z: 0},
			distRad: math.Pi,
			want:    r3.Vector{X: 1, Y: 0, Z: 0},
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

		closest := Project(x, a, b)
		if !closest.ApproxEqual(want) {
			t.Errorf("ClosestPoint(%v, %v, %v) = %v, want %v", x, a, b, closest, want)
		}

		if minDistance, ok := UpdateMinDistance(x, a, b, 0); ok {
			t.Errorf("UpdateMinDistance(%v, %v, %v, %v) = %v, want %v", x, a, b, 0, minDistance, 0)
		}

		minDistance, ok := UpdateMinDistance(x, a, b, s1.InfChordAngle())
		if !ok {
			t.Errorf("UpdateMinDistance(%v, %v, %v, %v) = %v, want %v", x, a, b, s1.InfChordAngle(), minDistance, s1.InfChordAngle())
		}

		if !float64Near(test.distRad, minDistance.Angle().Radians(), 1e-15) {
			t.Errorf("MinDistance between %v and %v,%v = %v, want %v within %v", x, a, b, minDistance.Angle().Radians(), test.distRad, 1e-15)
		}
	}
}

func TestEdgeDistancesUpdateMinInteriorDistanceLowerBoundOptimizationIsConservative(t *testing.T) {
	// Verifies that alwaysUpdateMinInteriorDistance computes the lower bound
	// on the true distance conservatively.  (This test used to fail.)
	x := PointFromCoords(-0.017952729194524016, -0.30232422079175203, 0.95303607751077712)
	a := PointFromCoords(-0.017894725505830295, -0.30229974986194175, 0.95304493075220664)
	b := PointFromCoords(-0.017986591360900289, -0.30233851195954353, 0.95303090543659963)

	minDistance, ok := UpdateMinDistance(x, a, b, s1.InfChordAngle())
	if !ok {
		t.Errorf("UpdateMinDistance(%v, %v, %v, %v) = %v, want %v", x, a, b, s1.InfChordAngle(), minDistance, s1.InfChordAngle())
	}
	minDistance = minDistance.Successor()
	minDistance, ok = UpdateMinDistance(x, a, b, minDistance)
	if !ok {
		t.Errorf("UpdateMinDistance(%v, %v, %v, %v) = %v, want %v", x, a, b, s1.InfChordAngle(), minDistance, minDistance)
	}
}

func TestEdgeDistancesUpdateMinInteriorDistanceRejectionTestIsConservative(t *testing.T) {
	// This test checks several representative cases where previously
	// UpdateMinInteriorDistance was failing to update the distance because a
	// rejection test was not being done conservatively.
	//
	// Note that all of the edges AB in this test are nearly antipodal.
	minDist := s1.ChordAngleFromSquaredLength(6.3897233584120815e-26)

	tests := []struct {
		x, a, b Point
		minDist s1.ChordAngle
		want    bool
	}{
		{

			x:       Point{r3.Vector{X: 1, Y: -4.6547732744037044e-11, Z: -5.6374428459823598e-89}},
			a:       Point{r3.Vector{X: 1, Y: -8.9031850507928352e-11, Z: 0}},
			b:       Point{r3.Vector{X: -0.99999999999996347, Y: 2.7030110029169596e-07, Z: 1.555092348806121e-99}},
			minDist: minDist,
			want:    true,
		},
		{
			x:       Point{r3.Vector{X: 1, Y: -4.7617930898495072e-13, Z: 0}},
			a:       Point{r3.Vector{X: -1, Y: -1.6065916409055676e-10, Z: 0}},
			b:       Point{r3.Vector{X: 1, Y: 0, Z: 9.9964883247706732e-35}},
			minDist: minDist,
			want:    true,
		},
		{
			x:       Point{r3.Vector{X: 1, Y: 0, Z: 0}},
			a:       Point{r3.Vector{X: 1, Y: -8.4965026896454536e-11, Z: 0}},
			b:       Point{r3.Vector{X: -0.99999999999966138, Y: 8.2297529603339328e-07, Z: 9.6070344113320997e-21}},
			minDist: minDist,
			want:    true,
		},
	}

	for _, test := range tests {
		if _, ok := UpdateMinDistance(test.x, test.a, test.b, test.minDist); ok != test.want {
			t.Errorf("UpdateMinDistance(%v, %v, %v, %v) = %v, want %v", test.x, test.a, test.b, test.minDist, ok, test.want)
		}
	}
}

func TestEdgeDistancesCheckMaxDistance(t *testing.T) {
	tests := []struct {
		x, a, b r3.Vector
		distRad float64
	}{
		{
			x:       r3.Vector{X: 1, Y: 0, Z: 1},
			a:       r3.Vector{X: 1, Y: 0, Z: 0},
			b:       r3.Vector{X: 0, Y: 1, Z: 0},
			distRad: math.Pi / 2,
		},
		{
			x:       r3.Vector{X: 1, Y: 0, Z: -1},
			a:       r3.Vector{X: 1, Y: 0, Z: 0},
			b:       r3.Vector{X: 0, Y: 1, Z: 0},
			distRad: math.Pi / 2,
		},
		{
			x:       r3.Vector{X: 0, Y: 1, Z: 1},
			a:       r3.Vector{X: 1, Y: 0, Z: 0},
			b:       r3.Vector{X: 0, Y: 1, Z: 0},
			distRad: math.Pi / 2,
		},
		{
			x:       r3.Vector{X: 0, Y: 1, Z: -1},
			a:       r3.Vector{X: 1, Y: 0, Z: 0},
			b:       r3.Vector{X: 0, Y: 1, Z: 0},
			distRad: math.Pi / 2,
		},
		{
			x:       r3.Vector{X: 1, Y: 1, Z: 1},
			a:       r3.Vector{X: 1, Y: 0, Z: 0},
			b:       r3.Vector{X: 0, Y: 1, Z: 0},
			distRad: math.Asin(math.Sqrt(2.0 / 3)),
		},
		{
			x:       r3.Vector{X: 1, Y: 1, Z: -1},
			a:       r3.Vector{X: 1, Y: 0, Z: 0},
			b:       r3.Vector{X: 0, Y: 1, Z: 0},
			distRad: math.Asin(math.Sqrt(2.0 / 3)),
		},
		{
			x:       r3.Vector{X: 1, Y: 0, Z: 0},
			a:       r3.Vector{X: 1, Y: 1, Z: 0},
			b:       r3.Vector{X: 1, Y: -1, Z: 0},
			distRad: math.Pi / 4,
		},
		{
			x:       r3.Vector{X: 0, Y: 1, Z: 0},
			a:       r3.Vector{X: 1, Y: 1, Z: 0},
			b:       r3.Vector{X: 1, Y: 1, Z: 0},
			distRad: math.Pi / 4,
		},
		{
			x:       r3.Vector{X: 0, Y: 0, Z: 1},
			a:       r3.Vector{X: 0, Y: 1, Z: 1},
			b:       r3.Vector{X: 0, Y: -1, Z: 1},
			distRad: math.Pi / 4,
		},
		{
			x:       r3.Vector{X: 0, Y: 0, Z: 1},
			a:       r3.Vector{X: 1, Y: 0, Z: 0},
			b:       r3.Vector{X: 1, Y: 0, Z: -1},
			distRad: 3 * math.Pi / 4,
		},
		{
			x:       r3.Vector{X: 0, Y: 0, Z: 1},
			a:       r3.Vector{X: 1, Y: 0, Z: 0},
			b:       r3.Vector{X: 1, Y: 1, Z: -math.Sqrt2},
			distRad: 3 * math.Pi / 4,
		},
		{
			x:       r3.Vector{X: 0, Y: 0, Z: 1},
			a:       r3.Vector{X: 0, Y: 0, Z: -1},
			b:       r3.Vector{X: 0, Y: 0, Z: -1},
			distRad: math.Pi,
		},
	}

	for _, test := range tests {
		x := Point{test.x.Normalize()}
		a := Point{test.a.Normalize()}
		b := Point{test.b.Normalize()}

		var ok bool
		maxDistance := s1.StraightChordAngle
		if maxDistance, ok = UpdateMaxDistance(x, a, b, maxDistance); ok {
			t.Errorf("UpdateMaxDistance(%v, %v, %v, %v) = %v, want %v", x, a, b, s1.StraightChordAngle, maxDistance, s1.StraightChordAngle)
		}

		maxDistance = s1.NegativeChordAngle
		if maxDistance, ok = UpdateMaxDistance(x, a, b, maxDistance); !ok {
			t.Errorf("UpdateMaxDistance(%v, %v, %v, %v) = %v, want > %v", x, a, b, s1.NegativeChordAngle, maxDistance, s1.NegativeChordAngle)
		}

		if !float64Near(test.distRad, maxDistance.Angle().Radians(), 1e-15) {
			t.Errorf("MaxDistance between %v and %v, %v = %v, want %v within %v", x, a, b, maxDistance.Angle().Radians(), test.distRad, 1e-15)
		}
	}
}

func TestEdgeDistancesInterpolate(t *testing.T) {
	// Choose test points designed to expose floating-point errors.
	p1 := PointFromCoords(0.1, 1e-30, 0.3)
	p2 := PointFromCoords(-0.7, -0.55, -1e30)
	i := PointFromCoords(1, 0, 0)
	j := PointFromCoords(0, 1, 0)

	// Take a small fraction along the curve, 1/1000 of the way.
	p := Interpolate(0.001, i, j)

	tests := []struct {
		a, b Point
		dist float64
		want Point
	}{
		// A zero-length edge.
		{p1, p1, 0, p1},
		{p1, p1, 1, p1},

		// Zero-length edges, actually interpolated.
		{PointFromCoords(1, 0, 0), PointFromCoords(1, 0, 0), 0.5, PointFromCoords(1, 0, 0)},
		{PointFromCoords(1, 0, 0), PointFromCoords(1, 0, 0), math.SmallestNonzeroFloat64, PointFromCoords(1, 0, 0)},
		{p1, p1, 0.5, p1},
		{p1, p1, math.SmallestNonzeroFloat64, p1},

		// Start, end, and middle of a medium-length edge.
		{p1, p2, 0, p1},
		{p1, p2, 1, p2},
		{p1, p2, 0.5, Point{(p1.Add(p2.Vector)).Mul(0.5)}},

		// Test that interpolation is done using distances on the sphere
		// rather than linear distances.
		{
			Point{r3.Vector{X: 1, Y: 0, Z: 0}},
			Point{r3.Vector{X: 0, Y: 1, Z: 0}},
			1.0 / 3.0,
			Point{r3.Vector{X: sqrt3, Y: 1, Z: 0}},
		},
		{
			Point{r3.Vector{X: 1, Y: 0, Z: 0}},
			Point{r3.Vector{X: 0, Y: 1, Z: 0}},
			2.0 / 3.0,
			Point{r3.Vector{X: 1, Y: sqrt3, Z: 0}},
		},

		// InterpolateCanExtrapolate checks

		// Initial vectors at 90 degrees.
		{i, j, 0, Point{r3.Vector{X: 1, Y: 0, Z: 0}}},
		{i, j, 1, Point{r3.Vector{X: 0, Y: 1, Z: 0}}},
		{i, j, 1.5, Point{r3.Vector{X: -1, Y: 1, Z: 0}}},
		{i, j, 2, Point{r3.Vector{X: -1, Y: 0, Z: 0}}},
		{i, j, 3, Point{r3.Vector{X: 0, Y: -1, Z: 0}}},
		{i, j, 4, Point{r3.Vector{X: 1, Y: 0, Z: 0}}},

		// Negative values of t.
		{i, j, -1, Point{r3.Vector{X: 0, Y: -1, Z: 0}}},
		{i, j, -2, Point{r3.Vector{X: -1, Y: 0, Z: 0}}},
		{i, j, -3, Point{r3.Vector{X: 0, Y: 1, Z: 0}}},
		{i, j, -4, Point{r3.Vector{X: 1, Y: 0, Z: 0}}},

		// Initial vectors at 45 degrees.
		{i, Point{r3.Vector{X: 1, Y: 1, Z: 0}}, 2, Point{r3.Vector{X: 0, Y: 1, Z: 0}}},
		{i, Point{r3.Vector{X: 1, Y: 1, Z: 0}}, 3, Point{r3.Vector{X: -1, Y: 1, Z: 0}}},
		{i, Point{r3.Vector{X: 1, Y: 1, Z: 0}}, 4, Point{r3.Vector{X: -1, Y: 0, Z: 0}}},

		// Initial vectors at 135 degrees.
		{i, Point{r3.Vector{X: -1, Y: 1, Z: 0}}, 2, Point{r3.Vector{X: 0, Y: -1, Z: 0}}},

		// Test that we should get back where we started by interpolating
		// the 1/1000th by 1000.
		{i, p, 1000, j},
	}

	for _, test := range tests {
		test.a = Point{test.a.Normalize()}
		test.b = Point{test.b.Normalize()}
		test.want = Point{test.want.Normalize()}
		maxError := s1.Angle(3e-15)

		// We allow a bit more than the usual 1e-15 error tolerance because
		// Interpolate() uses trig functions.
		if got := Interpolate(test.dist, test.a, test.b); !pointsApproxEqual(got, test.want, float64(maxError)) {
			t.Errorf("Interpolate(%v, %v, %v) = %v, want %v", test.dist, test.a, test.b, got, test.want)
		}

		// Now test the other interpolation functions.
		r := s1.Angle(test.dist) * test.a.Distance(test.b)
		if got := PointOnLine(test.a, test.b, r); got.Distance(test.want) > maxError {
			t.Errorf("PointOnLijne((%v, %v, %v) = %v, want %v",
				test.a, test.b, r, got, test.want)
		}

		if test.a.Dot(test.b.Vector) == 0 { // Common in the test cases below.
			if got := PointOnRay(test.a, test.b, r); got.Distance(test.want) > maxError {
				t.Errorf("PointOnRay(%v, %v, %v) = %v, want %v",
					test.a, test.b, r, got, test.want)
			}
		}
		if r.Radians() >= 0 && r.Radians() < 0.99*math.Pi {
			// We don't have the parallel ChordAngle variants of
			// PointOnLine/Ray/Left/Right
			// so the second tests are not added here.
		}
	}
}

func TestEdgeDistancesInterpolateOverLongEdge(t *testing.T) {
	lng := math.Pi - 1e-2
	a := Point{PointFromLatLng(LatLng{0, 0}).Normalize()}
	b := Point{PointFromLatLng(LatLng{0, s1.Angle(lng)}).Normalize()}

	for f := 0.4; f > 1e-15; f *= 0.1 {
		// Test that interpolation is accurate on a long edge (but not so long that
		// the definition of the edge itself becomes too unstable).
		want := Point{PointFromLatLng(LatLng{0, s1.Angle(f * lng)}).Normalize()}
		if got := Interpolate(f, a, b); !pointsApproxEqual(got, want, 3e-15) {
			t.Errorf("long edge Interpolate(%v, %v, %v) = %v, want %v", f, a, b, got, want)
		}

		// Test the remainder of the dist also matches.
		wantRem := Point{PointFromLatLng(LatLng{0, s1.Angle((1 - f) * lng)}).Normalize()}
		if got := Interpolate(1-f, a, b); !pointsApproxEqual(got, wantRem, 3e-15) {
			t.Errorf("long edge Interpolate(%v, %v, %v) = %v, want %v", 1-f, a, b, got, wantRem)
		}
	}
}

func TestEdgeDistancesInterpolateAntipodal(t *testing.T) {
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

func TestEdgeDistancesRepeatedInterpolation(t *testing.T) {
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

func TestEdgeDistanceMinUpdateDistanceMaxError(t *testing.T) {
	tests := []struct {
		actual s1.Angle
		maxErr s1.Angle
	}{
		{0, 1.5e-15},
		{1e-8, 1e-15},
		{1e-5, 1e-15},
		{0.05, 1e-15},
		{math.Pi/2 - 1e-8, 2e-15},
		{math.Pi / 2, 2e-15},
		{math.Pi/2 + 1e-8, 2e-15},
		{math.Pi - 1e-5, 2e-10},
		{math.Pi, 0},
	}

	// This checks that the error returned by minUpdateDistanceMaxError for
	// the distance actual (measured in radians) corresponds to a distance error
	// of less than maxErr (measured in radians).
	//
	// The reason for the awkward phraseology above is that the value returned by
	// minUpdateDistanceMaxError is not a distance; it represents an error in
	// the *squared* distance.
	for _, test := range tests {
		ca := s1.ChordAngleFromAngle(test.actual)
		bound := ca.Expanded(minUpdateDistanceMaxError(ca)).Angle()

		if got := s1.Angle(bound.Radians()) - test.actual; got > test.maxErr {
			t.Errorf("minUpdateDistanceMaxError(%v)-%v = %v> %v, want <=", ca, got, test.actual, test.maxErr)
		}
	}
}

func TestEdgeDistanceUpdateMinInteriorDistanceMaxError(t *testing.T) {
	// Check that the error bound returned by
	// updateMinInteriorDistanceMaxError() is large enough.

	for iter := 0; iter < 10000; iter++ {
		a0 := randomPoint()
		length := s1.Angle(math.Pi*math.Pow(1e-20, randomFloat64())) * s1.Radian
		if oneIn(4) {
			length = math.Pi - length
		}
		a1 := PointOnLine(a0, randomPoint(), length)

		// TODO(rsned): The error bound holds for antipodal points, but the S2
		// predicates used to test the error do not support antipodal points yet.
		if a1.Vector == a0.Mul(-1) {
			continue
		}
		n := Point{a0.Cross(a1.Vector).Normalize()}
		f := math.Pow(1e-20, randomFloat64())
		a := Point{a0.Mul(1 - f).Add(a1.Mul(f)).Normalize()}
		r := s1.Angle(math.Pi / 2.0 * math.Pow(1e-20, randomFloat64()))
		if oneIn(2) {
			r = s1.Angle(math.Pi/2.0) - r
		}
		x := PointOnLine(a, n, r)
		minDist := s1.InfChordAngle()

		var ok bool
		// TODO(rsned): The first param here should be minDist which is needed
		// for the two commented out tests below.  When they are implemented
		// return this to "if minDist, ok == Update...."
		if _, ok = UpdateMinInteriorDistance(x, a0, a1, minDist); !ok {
			iter--
			continue
		}
		// TODO(rsned): Uncomment once predicates has CompareEdgeDistance
		/*
			maxErr := minUpdateDistanceMaxError(minDist)
			if got := CompareEdgeDistance(x, a0, a1, minDist.Expanded(maxErr)); got > 0 {
				t.Errorf("CompareEdgeDistance(%v, %v, %v, %v) = got, want <= 0",
					x, a0, a1, minDist.Expanded(maxErr), got)

			}
			if got := CompareEdgeDistance(x, a0, a1, minDist.Expanded(-maxErr)); got < 0 {
				t.Errorf("CompareEdgeDistance(%v, %v, %v, %v) = got, want >= 0",
					x, a0, a1, minDist.Expanded(-maxErr), got)
			}
		*/
	}
}

func TestEdgeDistancesEdgePairMinDistance(t *testing.T) {
	var zero Point
	tests := []struct {
		a0, a1   Point
		b0, b1   Point
		distRads float64
		wantA    Point
		wantB    Point
	}{
		{
			// One edge is degenerate.
			a0:       PointFromCoords(1, 0, 1),
			a1:       PointFromCoords(1, 0, 1),
			b0:       PointFromCoords(1, -1, 0),
			b1:       PointFromCoords(1, 1, 0),
			distRads: math.Pi / 4,
			wantA:    PointFromCoords(1, 0, 1),
			wantB:    PointFromCoords(1, 0, 0),
		},
		{
			// One edge is degenerate.
			a0:       PointFromCoords(1, -1, 0),
			a1:       PointFromCoords(1, 1, 0),
			b0:       PointFromCoords(1, 0, 1),
			b1:       PointFromCoords(1, 0, 1),
			distRads: math.Pi / 4,
			wantA:    PointFromCoords(1, 0, 0),
			wantB:    PointFromCoords(1, 0, 1),
		},
		{
			// Both edges are degenerate.
			a0:       PointFromCoords(1, 0, 0),
			a1:       PointFromCoords(1, 0, 0),
			b0:       PointFromCoords(0, 1, 0),
			b1:       PointFromCoords(0, 1, 0),
			distRads: math.Pi / 2,
			wantA:    PointFromCoords(1, 0, 0),
			wantB:    PointFromCoords(0, 1, 0),
		},
		{
			// Both edges are degenerate and antipodal.
			a0:       PointFromCoords(1, 0, 0),
			a1:       PointFromCoords(1, 0, 0),
			b0:       PointFromCoords(-1, 0, 0),
			b1:       PointFromCoords(-1, 0, 0),
			distRads: math.Pi,
			wantA:    PointFromCoords(1, 0, 0),
			wantB:    PointFromCoords(-1, 0, 0),
		},
		{
			// Two identical edges.
			a0:       PointFromCoords(1, 0, 0),
			a1:       PointFromCoords(0, 1, 0),
			b0:       PointFromCoords(1, 0, 0),
			b1:       PointFromCoords(0, 1, 0),
			distRads: 0,
			wantA:    zero,
			wantB:    zero,
		},
		{
			// Both edges are degenerate and identical.
			a0:       PointFromCoords(1, 0, 0),
			a1:       PointFromCoords(1, 0, 0),
			b0:       PointFromCoords(1, 0, 0),
			b1:       PointFromCoords(1, 0, 0),
			distRads: 0,
			wantA:    PointFromCoords(1, 0, 0),
			wantB:    PointFromCoords(1, 0, 0),
		},
		// Edges that share exactly one vertex (all 4 possibilities).
		{
			a0:       PointFromCoords(1, 0, 0),
			a1:       PointFromCoords(0, 1, 0),
			b0:       PointFromCoords(0, 1, 0),
			b1:       PointFromCoords(0, 1, 1),
			distRads: 0,
			wantA:    PointFromCoords(0, 1, 0),
			wantB:    PointFromCoords(0, 1, 0),
		},
		{
			a0:       PointFromCoords(0, 1, 0),
			a1:       PointFromCoords(1, 0, 0),
			b0:       PointFromCoords(0, 1, 0),
			b1:       PointFromCoords(0, 1, 1),
			distRads: 0,
			wantA:    PointFromCoords(0, 1, 0),
			wantB:    PointFromCoords(0, 1, 0),
		},
		{
			a0:       PointFromCoords(1, 0, 0),
			a1:       PointFromCoords(0, 1, 0),
			b0:       PointFromCoords(0, 1, 1),
			b1:       PointFromCoords(0, 1, 0),
			distRads: 0,
			wantA:    PointFromCoords(0, 1, 0),
			wantB:    PointFromCoords(0, 1, 0),
		},
		{
			a0:       PointFromCoords(0, 1, 0),
			a1:       PointFromCoords(1, 0, 0),
			b0:       PointFromCoords(0, 1, 1),
			b1:       PointFromCoords(0, 1, 0),
			distRads: 0,
			wantA:    PointFromCoords(0, 1, 0),
			wantB:    PointFromCoords(0, 1, 0),
		},
		{
			// Two edges whose interiors cross.
			a0:       PointFromCoords(1, -1, 0),
			a1:       PointFromCoords(1, 1, 0),
			b0:       PointFromCoords(1, 0, -1),
			b1:       PointFromCoords(1, 0, 1),
			distRads: 0,
			wantA:    PointFromCoords(1, 0, 0),
			wantB:    PointFromCoords(1, 0, 0),
		},
		// The closest distance occurs between two edge endpoints, but more than one
		// endpoint pair is equally distant.
		{
			a0:       PointFromCoords(1, -1, 0),
			a1:       PointFromCoords(1, 1, 0),
			b0:       PointFromCoords(-1, 0, 0),
			b1:       PointFromCoords(-1, 0, 1),
			distRads: math.Acos(-0.5),
			wantA:    zero,
			wantB:    PointFromCoords(-1, 0, 1),
		},
		{
			a0:       PointFromCoords(-1, 0, 0),
			a1:       PointFromCoords(-1, 0, 1),
			b0:       PointFromCoords(1, -1, 0),
			b1:       PointFromCoords(1, 1, 0),
			distRads: math.Acos(-0.5),
			wantA:    PointFromCoords(-1, 0, 1),
			wantB:    zero,
		},
		{
			a0:       PointFromCoords(1, -1, 0),
			a1:       PointFromCoords(1, 1, 0),
			b0:       PointFromCoords(-1, 0, -1),
			b1:       PointFromCoords(-1, 0, 1),
			distRads: math.Acos(-0.5),
			wantA:    zero,
			wantB:    zero,
		},
	}

	// Given two edges a0a1 and b0b1, check that the minimum distance
	// between them is distRads, and that EdgePairClosestPoints returns
	// wantA and wantB as the points that achieve this distance.
	// Point{0, 0, 0} may be passed for wantA or wantB to indicate
	// that both endpoints of the corresponding edge are equally distant,
	// and therefore either one might be returned.
	for _, test := range tests {
		actualA, actualB := EdgePairClosestPoints(test.a0, test.a1, test.b0, test.b1)
		if test.wantA == zero {
			// either end point works.
			if !(actualA == test.a0 || actualA == test.a1) {
				t.Errorf("EdgePairClosestPoints(%v, %v, %v, %v) = %v, want %v or %v", test.a0, test.a1, test.b0, test.b1, actualA, test.a0, test.a1)
			}
		} else {
			if !actualA.ApproxEqual(test.wantA) {
				t.Errorf("EdgePairClosestPoints(%v, %v, %v, %v) = %v, want %v", test.a0, test.a1, test.b0, test.b1, actualA, test.wantA)
			}
		}

		if test.wantB == zero {
			// either end point works.
			if !(actualB == test.b0 || actualB == test.b1) {
				t.Errorf("EdgePairClosestPoints(%v, %v, %v, %v) = %v, want %v or %v", test.a0, test.a1, test.b0, test.b1, actualB, test.b0, test.b1)
			}
		} else {
			if !actualB.ApproxEqual(test.wantB) {
				t.Errorf("EdgePairClosestPoints(%v, %v, %v, %v) = %v, want %v", test.a0, test.a1, test.b0, test.b1, actualB, test.wantB)
			}
		}

		var minDist s1.ChordAngle
		var ok bool
		minDist, ok = updateEdgePairMinDistance(test.a0, test.a1, test.b0, test.b1, minDist)
		if ok {
			t.Errorf("updateEdgePairMinDistance(%v, %v, %v, %v, %v) = %v, want updated to be false", test.a0, test.a1, test.b0, test.b1, 0, minDist)
		}

		minDist = s1.InfChordAngle()
		minDist, ok = updateEdgePairMinDistance(test.a0, test.a1, test.b0, test.b1, minDist)
		if !ok {
			t.Errorf("updateEdgePairMinDistance(%v, %v, %v, %v, %v) = %v, want updated to be true", test.a0, test.a1, test.b0, test.b1, s1.InfChordAngle(), minDist)
		}

		if !float64Near(test.distRads, minDist.Angle().Radians(), epsilon) {
			t.Errorf("minDist %v - %v = %v, want < %v", test.distRads, minDist.Angle().Radians(), (test.distRads - minDist.Angle().Radians()), epsilon)
		}
	}
}

func TestEdgeDistancesEdgePairMaxDistance(t *testing.T) {
	tests := []struct {
		a0, a1   Point
		b0, b1   Point
		distRads float64
	}{
		{
			// Standard situation. Same hemisphere, not degenerate.
			a0:       PointFromCoords(1, 0, 0),
			a1:       PointFromCoords(0, 1, 0),
			b0:       PointFromCoords(1, 1, 0),
			b1:       PointFromCoords(1, 1, 1),
			distRads: math.Acos(1 / sqrt3),
		},
		{

			// One edge is degenerate.
			a0:       PointFromCoords(1, 0, 1),
			a1:       PointFromCoords(1, 0, 1),
			b0:       PointFromCoords(1, -1, 0),
			b1:       PointFromCoords(1, 1, 0),
			distRads: math.Acos(0.5),
		},
		{
			a0:       PointFromCoords(1, -1, 0),
			a1:       PointFromCoords(1, 1, 0),
			b0:       PointFromCoords(1, 0, 1),
			b1:       PointFromCoords(1, 0, 1),
			distRads: math.Acos(0.5),
		},
		{
			// Both edges are degenerate.
			a0:       PointFromCoords(1, 0, 0),
			a1:       PointFromCoords(1, 0, 0),
			b0:       PointFromCoords(0, 1, 0),
			b1:       PointFromCoords(0, 1, 0),
			distRads: math.Pi / 2,
		},
		{
			// Both edges are degenerate and antipodal.
			a0:       PointFromCoords(1, 0, 0),
			a1:       PointFromCoords(1, 0, 0),
			b0:       PointFromCoords(-1, 0, 0),
			b1:       PointFromCoords(-1, 0, 0),
			distRads: math.Pi,
		},
		{
			// Two identical edges.
			a0:       PointFromCoords(1, 0, 0),
			a1:       PointFromCoords(0, 1, 0),
			b0:       PointFromCoords(1, 0, 0),
			b1:       PointFromCoords(0, 1, 0),
			distRads: math.Pi / 2,
		},
		{
			// Both edges are degenerate and identical.
			a0:       PointFromCoords(1, 0, 0),
			a1:       PointFromCoords(1, 0, 0),
			b0:       PointFromCoords(1, 0, 0),
			b1:       PointFromCoords(1, 0, 0),
			distRads: 0,
		},
		{
			// Antipodal reflection of one edge crosses the other edge.
			a0:       PointFromCoords(1, 0, 1),
			a1:       PointFromCoords(1, 0, -1),
			b0:       PointFromCoords(-1, -1, 0),
			b1:       PointFromCoords(-1, 1, 0),
			distRads: math.Pi,
		},
		{
			// One vertex of one edge touches the interior of the antipodal reflection
			// of the other edge.
			a0:       PointFromCoords(1, 0, 1),
			a1:       PointFromCoords(1, 0, 0),
			b0:       PointFromCoords(-1, -1, 0),
			b1:       PointFromCoords(-1, 1, 0),
			distRads: math.Pi,
		},
	}

	for _, test := range tests {
		// Given two edges a0a1 and b0b1, check that the maximum distance between them
		// is distancerads.
		if maxDist, ok := updateEdgePairMaxDistance(test.a0, test.a1, test.b0, test.b1, s1.StraightChordAngle); ok {
			t.Errorf("updateEdgePairMaxDistance(%v, %v, %v, %v, %v) = %v, want updated to be false", test.a0, test.a1, test.b0, test.b1, s1.StraightChordAngle, maxDist)
		}

		maxDist, ok := updateEdgePairMaxDistance(test.a0, test.a1, test.b0, test.b1, s1.NegativeChordAngle)
		if !ok {
			t.Errorf("updateEdgePairMaxDistance(%v, %v, %v, %v, %v) = %v, want updated to be false", test.a0, test.a1, test.b0, test.b1, s1.NegativeChordAngle, maxDist)
		}
		if !float64Near(test.distRads, maxDist.Angle().Radians(), epsilon) {
			t.Errorf("maxDist %v - %v = %v, want < %v", test.distRads, maxDist.Angle().Radians(), (test.distRads - maxDist.Angle().Radians()), epsilon)

		}
	}
}

func TestEdgeDistancesPointToLeft(t *testing.T) {
	a := PointFromLatLng(LatLngFromDegrees(0, 0))
	b := PointFromLatLng(LatLngFromDegrees(0, 5)) // east
	dist := kmToAngle(10 / 1000.0)

	c := PointToLeft(a, b, dist)
	if got := a.Distance(c).Radians(); !float64Near(got, dist.Radians(), epsilon) {
		t.Errorf("PointToLeft(%v, %v, %v) = %v, want %v", a, b, dist, got, dist.Radians())
	}
	// CAB must be a right angle with C to the left of AB.
	if got := TurnAngle(c, a, b).Radians(); !float64Near(got, math.Pi/2.0, epsilon) {
		t.Errorf("TurnAngle(%v, %v, %v) = %v, want %v", c, a, b, got, math.Pi/2.0)
	}
}

func TestEdgeDistancesPointToRight(t *testing.T) {
	a := PointFromLatLng(LatLngFromDegrees(0, 0))
	b := PointFromLatLng(LatLngFromDegrees(0, 5)) // east
	dist := kmToAngle(10 / 1000.0)

	c := PointToRight(a, b, dist)
	if got := a.Distance(c).Radians(); !float64Near(got, dist.Radians(), epsilon) {
		t.Errorf("PointToRight(%v, %v, %v) = %v, want %v", a, b, dist, got, dist.Radians())
	}
	// CAB must be a right angle with C to the right of AB.
	if got := TurnAngle(c, a, b).Radians(); !float64Near(got, -math.Pi/2.0, epsilon) {
		t.Errorf("TurnAngle(%v, %v, %v) = %v, want %v", c, a, b, got, -math.Pi/2.0)
	}
}

// TODO(rsned): Differences from C++
//
// TestEdgeDistancesEdgeBNearEdgeA
// TestProjectError
