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
	q := NewConvexHullQuery()
	expectTrue(q.ConvexHull().IsEmpty(), t)
}

func loopHasVertex(loop *Loop, point Point) bool {
	_, ok := loop.findVertex(point)
	return ok
}

func TestConvexHullQueryOnePoint(t *testing.T) {
	query := NewConvexHullQuery()
	p := PointFromCoords(0, 0, 1)
	query.AddPoint(p)

	res := query.ConvexHull()
	expectInt(3, res.NumVertices(), t)
	expectTrue(res.IsNormalized(), t)
	expectTrue(loopHasVertex(res, p), t)

	// Add some duplicate points and check that the result is the same.
	query.AddPoint(p)
	query.AddPoint(p)
	expectTrue(res.BoundaryEqual(query.ConvexHull()), t)
}

func TestConvexHullQueryTwoPoints(t *testing.T) {
	query := NewConvexHullQuery()
	p := PointFromCoords(0, 0, 1)
	q := PointFromCoords(0, 1, 0)
	query.AddPoint(p)
	query.AddPoint(q)

	res := query.ConvexHull()
	expectInt(3, res.NumVertices(), t)
	expectTrue(res.IsNormalized(), t)
	expectTrue(loopHasVertex(res, p), t)
	expectTrue(loopHasVertex(res, q), t)

	query.AddPoint(p)
	query.AddPoint(p)
	query.AddPoint(q)
	expectTrue(res.BoundaryEqual(query.ConvexHull()), t)
}

func TestConvexHullQueryEmptyLoop(t *testing.T) {
	query := NewConvexHullQuery()
	query.AddLoop(EmptyLoop())
	res := query.ConvexHull()
	expectTrue(res.IsEmpty(), t)
}

func TestConvexHullQueryFullLoop(t *testing.T) {
	query := NewConvexHullQuery()
	query.AddLoop(FullLoop())
	res := query.ConvexHull()
	expectTrue(res.IsFull(), t)
}

func TestConvexHullQueryEmptyPolygon(t *testing.T) {
	query := NewConvexHullQuery()
	empty := PolygonFromLoops(make([]*Loop, 0))
	query.AddPolygon(empty)
	res := query.ConvexHull()
	expectTrue(res.IsEmpty(), t)
}

func TestConvexHullQueryNonConvexPoints(t *testing.T) {
	// Generate a point set such that the only convex region containing them is
	// the entire sphere.  In other words, you can generate any point on the
	// sphere by repeatedly linearly interpolating between the points.  (The
	// four points of a tetrahedron would also work, but this is easier.)
	query := NewConvexHullQuery()
	for i := 0; i < 6; i++ {
		query.AddPoint(CellIDFromFace(i).Point())
	}
	res := query.ConvexHull()
	expectTrue(res.IsFull(), t)
}

func TestConvexHullQuerySimplePolyline(t *testing.T) {
	// A polyline is handling identically to a point set, so there is no need
	// for special testing other than code coverage.
	query := NewConvexHullQuery()
	polyline := makePolyline("0:1, 0:9, 1:6, 2:6, 3:10, 4:10, 5:5, 4:0, 3:0, 2:5, 1:5")
	for _, p := range *polyline {
		query.AddPoint(p)
	}
	//query.AddPolyline(*polyline)
	res := query.ConvexHull()
	expected := makeLoop("0:1, 0:9, 3:10, 4:10, 5:5, 4:0, 3:0")

	expectTrue(res.BoundaryEqual(expected), t)
}

func testConvexHullQueryNorthPoleLoop(radius s1.Angle, numVertices int, t *testing.T) {
	// A polyline is handling identically to a point set, so there is no need
	// for special testing other than code coverage.
	query := NewConvexHullQuery()
	loop := RegularLoop(PointFromCoords(0, 0, 1), radius, numVertices)
	query.AddLoop(loop)
	res := query.ConvexHull()
	if radius.Radians() > 2*math.Pi {
		expectTrue(res.IsFull(), t)
	} else {
		expectTrue(res.BoundaryEqual(loop), t)
	}
}

func TestConvexHullQueryLoopsAroundNorthPole(t *testing.T) {
	// Test loops of various sizes around the north pole.
	testConvexHullQueryNorthPoleLoop(s1.Degree, 3, t)
	testConvexHullQueryNorthPoleLoop(89*s1.Degree, 3, t)

	// The following two loops should yield the full loop.
	testConvexHullQueryNorthPoleLoop(91*s1.Degree, 3, t)
	testConvexHullQueryNorthPoleLoop(179*s1.Degree, 3, t)

	testConvexHullQueryNorthPoleLoop(10*s1.Degree, 100, t)
	testConvexHullQueryNorthPoleLoop(89*s1.Degree, 1000, t)
}

func TestConvexHullPointsInsideHull(t *testing.T) {
	// Repeatedly build the convex hull of a set of points, then add more points
	// inside that loop and build the convex hull again.  The result should
	// always be the same.
	for i := 0; i < 1000; i++ {
		// Choose points from within a cap of random size, up to but not including
		// an entire hemisphere.
		cap := randomCap(1e-15, 1.999*math.Pi)
		query := NewConvexHullQuery()

		numPts := randomUniformInt(100) + 3
		for j := 0; j < numPts; j++ {
			query.AddPoint(samplePointFromCap(cap))
		}

		hull := query.ConvexHull()

		// When the convex hull is nearly a hemisphere, the algorithm sometimes
		// returns a full cap instead.  This is because it first computes a
		// bounding rectangle for all the input points/edges and then converts it
		// to a bounding cap, which sometimes yields a non-convex cap (radius
		// larger than 90 degrees).  This should not be a problem in practice
		// (since most convex hulls are not hemispheres), but in order make this
		// test pass reliably it means that we need to reject convex hulls whose
		// bounding cap (when computed from a bounding rectangle) is not convex.
		//
		// TODO(ericv): This test can still fail (about 1 iteration in 500,000)
		// because the S2LatLngRect::GetCapBound implementation does not guarantee
		// that A.Contains(B) implies A.GetCapBound().Contains(B.GetCapBound()).
		if hull.CapBound().Height() >= 1 {
			continue
		}

		// Otherwise, add more points inside the convex hull.
		for j := 0; j < 1000; j++ {
			p := samplePointFromCap(cap)
			if hull.ContainsPoint(p) {
				query.AddPoint(p)
			}
		}

		hull2 := query.ConvexHull()
		expectTrue(hull2.BoundaryEqual(hull), t)
	}
}

func expectTrue(b bool, t *testing.T) {
	if !b {
		t.Fatalf("Expected true, got false")
	}
}

func expectInt(expected, actual int, t *testing.T) {
	if expected != actual {
		t.Fatalf("Expected %d, got %d", expected, actual)
	}
}
