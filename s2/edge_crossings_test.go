/*
Copyright 2017 Google Inc. All rights reserved.

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

// The various Crossing methods are tested via s2edge_crosser_test

// testIntersectionExact is a helper for the tests to return a positively
// oriented intersection Point of the two line segments (a0,a1) and (b0,b1).
func testIntersectionExact(a0, a1, b0, b1 Point) Point {
	x := intersectionExact(a0, a1, b0, b1)
	if x.Dot((a0.Add(a1.Vector)).Add(b0.Add(b1.Vector))) < 0 {
		x = Point{x.Mul(-1)}
	}
	return x
}

var distanceAbsError = s1.Angle(3 * dblEpsilon)

func TestEdgeutilIntersectionError(t *testing.T) {
	// We repeatedly construct two edges that cross near a random point "p", and
	// measure the distance from the actual intersection point "x" to the
	// exact intersection point and also to the edges.

	var maxPointDist, maxEdgeDist s1.Angle
	for iter := 0; iter < 5000; iter++ {
		// We construct two edges AB and CD that intersect near "p".  The angle
		// between AB and CD (expressed as a slope) is chosen randomly between
		// 1e-15 and 1e15 such that its logarithm is uniformly distributed.
		// Similarly, two edge lengths approximately between 1e-15 and 1 are
		// chosen.  The edge endpoints are chosen such that they are often very
		// close to the other edge (i.e., barely crossing).  Taken together this
		// ensures that we test both long and very short edges that intersect at
		// both large and very small angles.
		//
		// Sometimes the edges we generate will not actually cross, in which case
		// we simply try again.
		f := randomFrame()
		p := f.col(0)
		d1 := f.col(1)
		d2 := f.col(2)

		slope := 1e-15 * math.Pow(1e30, randomFloat64())
		d2 = Point{d1.Add(d2.Mul(slope)).Normalize()}
		var a, b, c, d Point

		// Find a pair of segments that cross.
		for {
			abLen := math.Pow(1e-15, randomFloat64())
			cdLen := math.Pow(1e-15, randomFloat64())
			aFraction := math.Pow(1e-5, randomFloat64())
			if oneIn(2) {
				aFraction = 1 - aFraction
			}
			cFraction := math.Pow(1e-5, randomFloat64())
			if oneIn(2) {
				cFraction = 1 - cFraction
			}
			a = Point{p.Sub(d1.Mul(aFraction * abLen)).Normalize()}
			b = Point{p.Add(d1.Mul((1 - aFraction) * abLen)).Normalize()}
			c = Point{p.Sub(d2.Mul(cFraction * cdLen)).Normalize()}
			d = Point{p.Add(d2.Mul((1 - cFraction) * cdLen)).Normalize()}
			if NewEdgeCrosser(a, b).CrossingSign(c, d) == Cross {
				break
			}
		}

		// Each constructed edge should be at most 1.5 * dblEpsilon away from the
		// original point P.
		if got, want := DistanceFromSegment(p, a, b), s1.Angle(1.5*dblEpsilon)+distanceAbsError; got > want {
			t.Errorf("DistanceFromSegment(%v, %v, %v) = %v, want %v", p, a, b, got, want)
		}
		if got, want := DistanceFromSegment(p, c, d), s1.Angle(1.5*dblEpsilon)+distanceAbsError; got > want {
			t.Errorf("DistanceFromSegment(%v, %v, %v) = %v, want %v", p, c, d, got, want)
		}

		// Verify that the expected intersection point is close to both edges and
		// also close to the original point P. (It might not be very close to P
		// if the angle between the edges is very small.)
		expected := testIntersectionExact(a, b, c, d)
		if got, want := DistanceFromSegment(expected, a, b), s1.Angle(3*dblEpsilon)+distanceAbsError; got > want {
			t.Errorf("DistanceFromSegment(%v, %v, %v) = %v, want %v", expected, a, b, got, want)
		}
		if got, want := DistanceFromSegment(expected, c, d), s1.Angle(3*dblEpsilon)+distanceAbsError; got > want {
			t.Errorf("DistanceFromSegment(%v, %v, %v) = %v, want %v", expected, c, d, got, want)
		}
		if got, want := expected.Distance(p), s1.Angle(3*dblEpsilon/slope)+intersectionError; got > want {
			t.Errorf("%v.Distance(%v) = %v, want %v", expected, p, got, want)
		}

		// Now we actually test the Intersection() method.
		actual := Intersection(a, b, c, d)
		distAB := DistanceFromSegment(actual, a, b)
		distCD := DistanceFromSegment(actual, c, d)
		pointDist := expected.Distance(actual)
		if got, want := distAB, intersectionError+distanceAbsError; got > want {
			t.Errorf("DistanceFromSegment(%v, %v, %v) = %v want <= %v", actual, a, b, got, want)
		}
		if got, want := distCD, intersectionError+distanceAbsError; got > want {
			t.Errorf("DistanceFromSegment(%v, %v, %v) = %v want <= %v", actual, c, d, got, want)
		}
		if got, want := pointDist, intersectionError; got > want {
			t.Errorf("%v.Distance(%v) = %v want <= %v", expected, actual, got, want)
		}
		maxEdgeDist = maxAngle(maxEdgeDist, maxAngle(distAB, distCD))
		maxPointDist = maxAngle(maxPointDist, pointDist)
	}
}

// TODO(roberts): Differences from C++:
// TestEdgeCrossingsGrazingIntersections
// TestEdgeCrossingsGetIntersectionInvariants
