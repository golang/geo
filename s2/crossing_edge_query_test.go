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
	"reflect"
	"testing"

	"github.com/golang/geo/s1"
)

func perturbAtDistance(distance s1.Angle, a0, b0 Point) Point {
	x := InterpolateAtDistance(distance, a0, b0)
	if oneIn(2) {
		if oneIn(2) {
			x.X = math.Nextafter(x.X, 1)
		} else {
			x.X = math.Nextafter(x.X, -1)
		}
		if oneIn(2) {
			x.Y = math.Nextafter(x.Y, 1)
		} else {
			x.Y = math.Nextafter(x.Y, -1)
		}
		if oneIn(2) {
			x.Z = math.Nextafter(x.Z, 1)
		} else {
			x.Z = math.Nextafter(x.Z, -1)

		}
		x = Point{x.Normalize()}
	}
	return x
}

// generatePerturbedSubEdges generate sub-edges of some given edge (a,b).
// The length of the sub-edges is distributed exponentially over a large range,
// and the endpoints may be slightly perturbed to one side of (a,b) or the other.
func generatePerturbedSubEdges(a, b Point, count int) []Edge {
	var edges []Edge

	a = Point{a.Normalize()}
	b = Point{b.Normalize()}

	length0 := a.Distance(b)
	for i := 0; i < count; i++ {
		length := length0 * s1.Angle(math.Pow(1e-15, randomFloat64()))
		offset := (length0 - length) * s1.Angle(randomFloat64())
		edges = append(edges, Edge{
			perturbAtDistance(offset, a, b),
			perturbAtDistance(offset+length, a, b),
		})
	}
	return edges
}

// generateCapEdges creates edges whose center is randomly chosen from the given cap, and
// whose length is randomly chosen up to maxLength.
func generateCapEdges(centerCap Cap, maxLength s1.Angle, count int) []Edge {
	var edges []Edge
	for i := 0; i < count; i++ {
		center := samplePointFromCap(centerCap)
		edgeCap := CapFromCenterAngle(center, 0.5*maxLength)
		p1 := samplePointFromCap(edgeCap)
		// Compute p1 reflected through center, and normalize for good measure.
		p2 := Point{(center.Mul(2 * p1.Dot(center.Vector)).Sub(p1.Vector)).Normalize()}
		edges = append(edges, Edge{p1, p2})
	}
	return edges
}

func testCrossingEdgeQueryAllCrossings(t *testing.T, edges []Edge) {
	s := &edgeVectorShape{}
	for _, edge := range edges {
		s.Add(edge.V0, edge.V1)
	}

	// Force more subdivision than usual to make the test more challenging.
	index := NewShapeIndex()
	index.maxEdgesPerCell = 1
	index.Add(s)

	// To check that candidates are being filtered reasonably, we count the
	// total number of candidates that the total number of edge pairs that
	// either intersect or are very close to intersecting.
	var numCandidates, numNearbyPairs int

	for _, edge := range edges {
		a := edge.V0
		b := edge.V1

		query := NewCrossingEdgeQuery(index)
		candidates := query.candidates(a, b, s)

		// Verify that the EdgeMap version of candidates returns the same result.
		edgeMap := query.candidatesEdgeMap(a, b)
		if len(edgeMap) != 1 {
			t.Errorf("there should be only one shape in this map, got %d", len(edgeMap))
			// Skip the next part of the check since we expect only 1
			// value, and you can't get just the first entry in a map.
			continue
		}
		for k, v := range edgeMap {
			if Shape(s) != k {
				t.Errorf("element(%v) of edgeMap should be this shape", k)
			}
			if !reflect.DeepEqual(candidates, v) {
				t.Errorf("edgeMap candidates for this shape = %v, want %v", v, candidates)
			}
		}
		if len(candidates) == 0 {
			t.Errorf("candidates should not be empty")
		}

		// Now check the actual candidates.
		// Check the results are sorted
		for k, v := range candidates {
			if k < len(candidates)-1 {
				if v > candidates[k+1] {
					t.Errorf("candidates was not sorted. candidates[%d] = %v > candidates[%d] = %v", k, v, k+1, candidates[k+1])
				}
			}
		}
		if got := candidates[0]; got < 0 {
			t.Errorf("the first element should not have a negative edge id")
		}
		if got, want := candidates[len(candidates)-1], s.NumEdges(); got >= want {
			t.Errorf("candidates[%d] = %v, want < %v", len(candidates)-1, got, want)
		}

		numCandidates += len(candidates)

		var expectedCrossings, expectedInteriorCrossings []int
		var missingCandidates []int
		for i := 0; i < s.NumEdges(); i++ {
			edge := s.Edge(i)
			c := edge.V0
			d := edge.V1
			sign := CrossingSign(a, b, c, d)
			if sign != DoNotCross {
				expectedCrossings = append(expectedCrossings, i)
				if sign == Cross {
					expectedInteriorCrossings = append(expectedInteriorCrossings, i)
				}
				numNearbyPairs++

				// try to find this edge number in the set of candidates.
				found := false
				for _, v := range candidates {
					if v == i {
						found = true
					}
				}
				// If we couldn't find it, add it to the missing set.
				if !found {
					missingCandidates = append(missingCandidates, i)
				}
			} else {
				maxDist := MaxDiagMetric.Value(maxLevel)
				if DistanceFromSegment(a, c, d).Radians() < maxDist ||
					DistanceFromSegment(b, c, d).Radians() < maxDist ||
					DistanceFromSegment(c, a, b).Radians() < maxDist ||
					DistanceFromSegment(d, a, b).Radians() < maxDist {
					numNearbyPairs++
				}
			}
		}

		if len(missingCandidates) != 0 {
			t.Errorf("missingCandidates should have been empty")
		}

		// test that Crossings returns only the actual crossing edges.
		actualCrossings := query.Crossings(a, b, s, CrossingTypeAll)
		if !reflect.DeepEqual(expectedCrossings, actualCrossings) {
			t.Errorf("query.Crossings(%v, %v, ...) = %v, want %v", a, b, actualCrossings, expectedCrossings)
		}

		// Verify that the edge map version of Crossings returns the same result.
		if edgeMap := query.CrossingsEdgeMap(a, b, CrossingTypeAll); len(edgeMap) != 0 {
			if len(edgeMap) != 1 {
				t.Errorf("query.CrossingsEdgeMap(%v, %v) has %d entried, want 1", a, b, len(edgeMap))
			} else {
				for k, v := range edgeMap {
					if s != k {
						t.Errorf("query.CrossingsEdgeMap(%v, %v, ...) shape = %v, want %v", a, b, k, s)
					}
					if !reflect.DeepEqual(expectedCrossings, v) {
						t.Errorf("query.CrossingsEdgeMap(%v, %v) = %v, want %v", a, b, v, expectedCrossings)
					}
				}
			}
		}

		// Verify that CrossingTypeInterior returns only the interior crossings.
		actualCrossings = query.Crossings(a, b, s, CrossingTypeInterior)
		// TODO(roberts): Move to package "cmp" when s2 can meet the minimum go version.
		// This will handle the case where one slice is nil and the other is non-nil
		// but empty as being equivalent more cleanly.
		if !reflect.DeepEqual(expectedInteriorCrossings, actualCrossings) && (len(actualCrossings) != 0 && len(expectedInteriorCrossings) != 0) {
			t.Errorf("query.Crossings(%v, %v, CrossingTypeInterior) = %v, want %v", a, b, actualCrossings, expectedInteriorCrossings)
		}
	}

	// There is nothing magical about this particular ratio; this check exists
	// to catch changes that dramatically increase the number of candidates.
	if numCandidates > 3*numNearbyPairs {
		t.Errorf("number of candidates should not be dramatically larger than number of nearby pairs. got %v, want < %v", numCandidates, 3*numNearbyPairs)
	}

}

func TestCrossingEdgeQueryCrossingCandidatesPerturbedCubeEdges(t *testing.T) {
	// Test edges that lie in the plane of one of the S2 cube edges. Such edges
	// may lie on the boundary between two cube faces, or pass through a cube
	// vertex, or follow a 45 diagonal across a cube face toward its center.
	//
	// This test is sufficient to demonstrate that padding the cell boundaries
	// is necessary for correctness. (It will fails if ShapeIndexes CellPadding is
	// set to zero.)
	for iter := 0; iter < 10; iter++ {
		face := randomUniformInt(6)
		scale := math.Pow(1e-15, randomFloat64())
		u := scale*2*float64(randomUniformInt(2)) - 1
		v := scale*2*float64(randomUniformInt(2)) - 1

		a := Point{faceUVToXYZ(face, u, v)}
		b := Point{a.Sub(unitNorm(face).Mul(2))}
		// TODO(roberts): This test is currently slow because *every* crossing test
		// needs to invoke ExpensiveSign.
		edges := generatePerturbedSubEdges(a, b, 30)
		testCrossingEdgeQueryAllCrossings(t, edges)
	}
}

// Test edges that lie in the plane of one of the S2 cube face axes. These
// edges are special because one coordinate is zero, and they lie on the
// boundaries between the immediate child cells of the cube face.
func TestCrossingEdgeQueryCandidatesPerturbedCubeFaceAxes(t *testing.T) {
	for iter := 0; iter < 5; iter++ {
		face := randomUniformInt(6)
		scale := math.Pow(1e-15, randomFloat64())
		axis := uvwAxis(face, randomUniformInt(2))
		a := Point{axis.Mul(scale).Add(unitNorm(face).Vector)}
		b := Point{axis.Mul(scale).Sub(unitNorm(face).Vector)}
		edges := generatePerturbedSubEdges(a, b, 30)
		testCrossingEdgeQueryAllCrossings(t, edges)
	}
}

func TestCrossingEdgeQueryCandidatesCapEdgesNearCubeVertex(t *testing.T) {
	// Test a random collection of edges near the S2 cube vertex where the
	// Hilbert curve starts and ends.
	edges := generateCapEdges(CapFromCenterAngle(PointFromCoords(-1, -1, 1), s1.Angle(1e-3)), s1.Angle(1e-4), 1000)
	testCrossingEdgeQueryAllCrossings(t, edges)
}

func TestCrossingEdgeQueryCandidatesDegenerateEdgeOnCellVertexIsItsOwnCandidate(t *testing.T) {
	for iter := 0; iter < 100; iter++ {
		cell := CellFromCellID(randomCellID())
		edges := []Edge{Edge{cell.Vertex(0), cell.Vertex(0)}}
		testCrossingEdgeQueryAllCrossings(t, edges)
	}
}

func TestCrossingEdgeQueryCandidatesCollinearEdgesOnCellBoundaries(t *testing.T) {
	const numEdgeIntervals = 8 // 9*8/2 = 36 edges
	for level := 0; level <= maxLevel; level++ {
		var edges []Edge
		cell := CellFromCellID(randomCellIDForLevel(level))
		i := randomUniformInt(4)
		p1 := cell.Vertex(i % 4)
		p2 := cell.Vertex((i + 1) % 4)
		delta := p2.Sub(p1.Vector).Mul(1 / numEdgeIntervals)
		for i := 0; i <= numEdgeIntervals; i++ {
			for j := 0; j < i; j++ {
				edges = append(edges, Edge{
					Point{p1.Add(delta.Mul(float64(i))).Normalize()},
					Point{p1.Add(delta.Mul(float64(j))).Normalize()},
				})
			}
		}
		testCrossingEdgeQueryAllCrossings(t, edges)
	}
}

func TestCrossingEdgeQueryCrossingsPolylineCrossings(t *testing.T) {
	index := NewShapeIndex()
	// Three zig-zag lines near the equator.
	index.Add(makePolyline("0:0, 2:1, 0:2, 2:3, 0:4, 2:5, 0:6"))
	index.Add(makePolyline("1:0, 3:1, 1:2, 3:3, 1:4, 3:5, 1:6"))
	index.Add(makePolyline("2:0, 4:1, 2:2, 4:3, 2:4, 4:5, 2:6"))
	index.Begin()

	var tests = []struct {
		a0, a1 Point
	}{
		{parsePoint("1:0"), parsePoint("1:4")},
		{parsePoint("5:5"), parsePoint("6:6")},
	}

	for _, test := range tests {
		query := NewCrossingEdgeQuery(index)

		edgeMap := query.CrossingsEdgeMap(test.a0, test.a1, CrossingTypeAll)
		if len(edgeMap) == 0 {
			continue
		}

		for shape, edges := range edgeMap {
			polyline := shape.(*Polyline)
			if len(edges) == 0 {
				t.Errorf("shapes with no crossings should have been filtered out")
			}

			for _, edge := range edges {
				b0 := (*polyline)[edge]
				b1 := (*polyline)[edge+1]
				if got := CrossingSign(test.a0, test.a1, b0, b1); got == DoNotCross {
					t.Errorf("CrossingSign(%v, %v, %v, %v) = %v, want MaybeCross or Cross", test.a0, test.a1, b0, b1, got)
				}
			}
		}

		// Also test that no edges are missing.
		for _, shape := range index.shapes {
			polyline := shape.(*Polyline)
			edges := edgeMap[shape]
			for e := 0; e < len(*polyline)-1; e++ {
				if got := CrossingSign(test.a0, test.a1, (*polyline)[e], (*polyline)[e+1]); got != DoNotCross {
					// Need to count occurrences of the current edge to see
					// if and how many are in the edge map.
					count := 0
					for _, edge := range edges {
						if edge == e {
							count++
						}
					}

					if count != 1 {
						t.Errorf("edge %v should appear once in the set of edges, got %v", e, count)
					}
				}
			}
		}
	}
}

func TestUniqueInts(t *testing.T) {
	tests := []struct {
		have []int
		want []int
	}{
		{
			have: nil,
			want: nil,
		},
		{
			have: []int{},
			want: nil,
		},
		{
			have: []int{1},
			want: []int{1},
		},
		{
			have: []int{3, 2, 1},
			want: []int{1, 2, 3},
		},
		{
			have: []int{4, 4, 4},
			want: []int{4},
		},
		{
			have: []int{1, 2, 3, 4, 2, 3, 5, 4, 6, 1, 2, 3, 4, 5, 7, 8, 1, 3, 1, 2, 3, 9, 3, 2, 1},
			want: []int{1, 2, 3, 4, 5, 6, 7, 8, 9},
		},
	}

	for _, test := range tests {
		if got := uniqueInts(test.have); !reflect.DeepEqual(got, test.want) {
			t.Errorf("uniqueInts(%v) = %v, want %v", test.have, got, test.want)
		}
	}
}
