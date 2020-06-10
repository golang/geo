// Copyright 2018 Google Inc. All rights reserved.
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

	"github.com/golang/geo/s1"
)

func TestConvexHullQueryNoPoints(t *testing.T) {
	query := NewConvexHullQuery()
	result := query.ConvexHull()
	if !result.IsEmpty() {
		t.Errorf("ConvexHullQuery with no geometry should return an empty hull")
	}
}

func TestConvexHullQueryOnePoint(t *testing.T) {
	query := NewConvexHullQuery()
	p := PointFromCoords(0, 0, 1)
	query.AddPoint(p)
	result := query.ConvexHull()
	if got, want := len(result.vertices), 3; got != want {
		t.Errorf("len(query.ConvexHull()) = %d, want %d", got, want)
	}
	if !result.IsNormalized() {
		t.Errorf("ConvexHull should be normalized but wasn't")
	}

	if !loopHasVertex(result, p) {
		t.Errorf("ConvexHull doesn't have vertex %v, but should", p)
	}

	// Add some duplicate points and check that the result is the same.
	query.AddPoint(p)
	query.AddPoint(p)
	result2 := query.ConvexHull()
	if !result2.Equal(result) {
		t.Errorf("adding duplicate points to the ConvexHull should not change the result.")
	}
}

func TestConvexHullQueryTwoPoints(t *testing.T) {
	query := NewConvexHullQuery()
	p := PointFromCoords(0, 0, 1)
	q := PointFromCoords(0, 1, 0)
	query.AddPoint(p)
	query.AddPoint(q)
	result := query.ConvexHull()
	if got, want := len(result.vertices), 3; got != want {
		t.Errorf("len(query.ConvexHull()) = %d, want %d", got, want)
	}
	if !result.IsNormalized() {
		t.Errorf("ConvexHull should be normalized but wasn't")
	}

	if !loopHasVertex(result, p) {
		t.Errorf("ConvexHull doesn't have vertex %v, but should", p)
	}
	if !loopHasVertex(result, q) {
		t.Errorf("ConvexHull doesn't have vertex %v, but should", q)
	}
	// Add some duplicate points and check that the result is the same.
	query.AddPoint(q)
	query.AddPoint(p)
	query.AddPoint(p)
	result2 := query.ConvexHull()
	if !result2.Equal(result) {
		t.Errorf("adding duplicate points to the ConvexHull should not change the result.")
	}
}

func TestConvexHullAntipodalPoints(t *testing.T) {
	query := NewConvexHullQuery()
	query.AddPoint(PointFromCoords(0, 0, 1))
	query.AddPoint(PointFromCoords(0, 0, -1))
	result := query.ConvexHull()
	if !result.IsFull() {
		t.Errorf("antipodal points should return a Full Polygon, got: %v", result)
	}
}

func loopHasVertex(l *Loop, p Point) bool {
	for _, v := range l.vertices {
		if v == p {
			return true
		}
	}
	return false
}

func TestConvexHullQueryEmptyLoop(t *testing.T) {
	query := NewConvexHullQuery()
	query.AddLoop(EmptyLoop())
	result := query.ConvexHull()
	if !result.IsEmpty() {
		t.Errorf("ConvexHull of Empty Loop should be the Empty Loop")
	}
}

func TestConvexHullQueryFullLoop(t *testing.T) {
	query := NewConvexHullQuery()
	query.AddLoop(FullLoop())
	result := query.ConvexHull()
	if !result.IsFull() {
		t.Errorf("ConvexHull of Full Loop should be the Full Loop")
	}
}

func TestConvexHullQueryEmptyPolygon(t *testing.T) {
	query := NewConvexHullQuery()
	query.AddPolygon(PolygonFromLoops([]*Loop{}))
	result := query.ConvexHull()
	if !result.IsEmpty() {
		t.Errorf("ConvexHull of an empty Polygon should be the Empty Loop")
	}
}

func TestConvexHullQueryNonConvexPoints(t *testing.T) {
	// Generate a point set such that the only convex region containing them is
	// the entire sphere. In other words, you can generate any point on the
	// sphere by repeatedly linearly interpolating between the points. (The
	// four points of a tetrahedron would also work, but this is easier.)
	query := NewConvexHullQuery()
	for face := 0; face < 6; face++ {
		query.AddPoint(CellIDFromFace(face).Point())
	}
	result := query.ConvexHull()
	if !result.IsFull() {
		t.Errorf("ConvexHull of all faces should be the Full Loop, got %v", result)
	}
}

func TestConvexHullQuerySimplePolyline(t *testing.T) {
	// A polyline is handled identically to a point set, so there is no need
	// for special testing other than code coverage.
	polyline := makePolyline("0:1, 0:9, 1:6, 2:6, 3:10, 4:10, 5:5, 4:0, 3:0, 2:5, 1:5")
	query := NewConvexHullQuery()
	query.AddPolyline(polyline)
	result := query.ConvexHull()
	want := makeLoop("0:1, 0:9, 3:10, 4:10, 5:5, 4:0, 3:0")
	if !result.BoundaryEqual(want) {
		t.Errorf("ConvexHull from %v = %v, want %v", polyline, result, want)
	}
}

func TestConvexHullQueryLoopsAroundNorthPole(t *testing.T) {
	tests := []struct {
		radius   s1.Angle
		numVerts int
	}{
		// Test loops of various sizes around the north pole.
		{radius: 1 * s1.Degree, numVerts: 3},
		{radius: 89 * s1.Degree, numVerts: 3},
		// The following two loops should yield the full loop.
		{radius: 91 * s1.Degree, numVerts: 3},
		{radius: 179 * s1.Degree, numVerts: 3},

		{radius: 10 * s1.Degree, numVerts: 100},
		{radius: 89 * s1.Degree, numVerts: 1000},
	}

	for _, test := range tests {
		query := NewConvexHullQuery()
		loop := RegularLoop(PointFromCoords(0, 0, 1), test.radius, test.numVerts)
		query.AddLoop(loop)
		result := query.ConvexHull()

		if test.radius > s1.Angle(math.Pi/2) {
			if !result.IsFull() {
				t.Errorf("ConvexHull of a Loop with radius > 90 should be the Full Loop")
			}
		} else {
			if !result.BoundaryEqual(loop) {
				t.Errorf("ConvexHull of a north pole loop = %v, want %v", result, loop)
			}
		}
	}
}

func TestConvexHullQueryPointsInsideHull(t *testing.T) {
	// Repeatedly build the convex hull of a set of points, then add more points
	// inside that loop and build the convex hull again. The result should
	// always be the same.
	const iters = 1000
	for iter := 0; iter < iters; iter++ {
		// Choose points from within a cap of random size, up to but not including
		// an entire hemisphere.
		c := randomCap(1e-15, 1.999*math.Pi)
		numPoints1 := randomUniformInt(100) + 3

		query := NewConvexHullQuery()
		for i := 0; i < numPoints1; i++ {
			query.AddPoint(samplePointFromCap(c))
		}
		hull := query.ConvexHull()

		// When the convex hull is nearly a hemisphere, the algorithm sometimes
		// returns a full cap instead. This is because it first computes a
		// bounding rectangle for all the input points/edges and then converts it
		// to a bounding cap, which sometimes yields a non-convex cap (radius
		// larger than 90 degrees). This should not be a problem in practice
		// (since most convex hulls are not hemispheres), but in order make this
		// test pass reliably it means that we need to reject convex hulls whose
		// bounding cap (when computed from a bounding rectangle) is not convex.
		//
		// TODO(roberts): This test can still fail (about 1 iteration in 500,000)
		// because the Rect.CapBound implementation does not guarantee
		// that A.Contains(B) implies A.CapBound().Contains(B.CapBound()).
		if hull.CapBound().Height() >= 1 {
			continue
		}

		// Otherwise, add more points inside the convex hull.
		const numPoints2 = 1000
		for i := 0; i < numPoints2; i++ {
			p := samplePointFromCap(c)
			if hull.ContainsPoint(p) {
				query.AddPoint(p)
			}
		}
		// Finally, build a new convex hull and check that it hasn't changed.
		hull2 := query.ConvexHull()
		if !hull2.BoundaryEqual(hull) {
			t.Errorf("%v.BoundaryEqual(%v) = false, but should be true", hull2, hull)
		}
	}
}
