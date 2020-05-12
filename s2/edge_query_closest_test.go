// Copyright 2019 Google Inc. All rights reserved.
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
	"testing"

	"github.com/golang/geo/r3"
	"github.com/golang/geo/s1"
)

func TestClosestEdgeQueryNoEdges(t *testing.T) {
	index := &ShapeIndex{}
	query := NewClosestEdgeQuery(index, nil)
	target := NewMinDistanceToPointTarget(PointFromCoords(1, 0, 0))
	edge := query.findEdge(target, query.opts)

	if edge.shapeID != -1 {
		t.Errorf("shapeID for empty index should be -1, got %v", edge.shapeID)
	}
	if edge.edgeID != -1 {
		t.Errorf("edgeID for empty index should be -1, got %v", edge.edgeID)
	}
	if got, want := edge.Distance(), s1.InfChordAngle(); got != want {
		t.Errorf("edge.Distance = %+v, want %+v", got, want)
	}

	if got, want := query.Distance(target), s1.InfChordAngle(); got != want {
		t.Errorf("query.Distance(%v) = %+v, want %+v", target, got, want)
	}
}

func TestClosestEdgeQueryBasicTest(t *testing.T) {
	index := makeShapeIndex("1:1 | 1:2 | 1:3 # #")
	opts := NewClosestEdgeQueryOptions().
		MaxResults(1).
		DistanceLimit(s1.ChordAngleFromAngle(s1.Angle(3) * s1.Degree)).
		MaxError(s1.ChordAngleFromAngle(s1.Angle(0.001) * s1.Degree))

	query := NewClosestEdgeQuery(index, opts)
	target := NewMinDistanceToPointTarget(parsePoint("2:2"))
	result := query.findEdge(target, query.opts)

	if got, want := result.edgeID, int32(1); got != want {
		t.Errorf("query.findEdge(%v).edgeID = %v, want %v", target, got, want)
	}
	if got, want := query.Distance(target).Angle().Degrees(), 1.0; !float64Near(got, want, epsilon) {
		t.Errorf("query.Distance(%v) = %v, want %v", target, got, want)
	}
	if !query.IsDistanceLess(target, s1.ChordAngleFromAngle(s1.Angle(1.5)*s1.Degree)) {
		t.Errorf("query.IsDistanceLess(%v, 1.5) should be true but wasn't", target)
	}
}

func TestClosestEdgeQueryDistanceEqualToLimit(t *testing.T) {
	// Tests the behavior of IsDistanceLess, IsDistanceLessOrEqual, and
	// IsConservativeDistanceLessOrEqual (and the corresponding Options) when
	// the distance to the target exactly equals the chosen limit.
	p0 := parsePoint("23:12")
	p1 := parsePoint("47:11")
	index := NewShapeIndex()
	pv := PointVector([]Point{p0})
	index.Add(Shape(&pv))
	query := NewClosestEdgeQuery(index, nil)

	// Start with two identical points and a zero distance.
	target0 := NewMinDistanceToPointTarget(p0)
	dist0 := s1.ChordAngle(0)
	if query.IsDistanceLess(target0, dist0) {
		t.Errorf("query.IsDistanceLess(%v, %v) = true, want false", target0, dist0)
	}
	if !query.IsDistanceLess(target0, dist0.Successor()) {
		t.Errorf("query.IsDistanceLess(%v, %v) = false, want true", target0, dist0)
	}
	if !query.IsConservativeDistanceLessOrEqual(target0, dist0) {
		t.Errorf("query.IsConservativeDistanceLessOrEqual(%v, %v) = false, want true", target0, dist0)
	}

	// Now try two points separated by a non-zero distance.
	target1 := NewMinDistanceToPointTarget(p1)
	dist1 := ChordAngleBetweenPoints(p0, p1)
	if query.IsDistanceLess(target1, dist1) {
		t.Errorf("query.IsDistanceLess(%v, %v) = true, want false", target1, dist1)
	}
	if !query.IsDistanceLess(target1, dist1.Successor()) {
		t.Errorf("query.IsDistanceLess(%v, %v) = false, want true", target1, dist1)
	}
	if !query.IsConservativeDistanceLessOrEqual(target1, dist1) {
		t.Errorf("query.IsConservativeDistanceLessOrEqual(%v, %v) = false, want true", target1, dist1)
	}
}

func TestClosestEdgeQueryTrueDistanceLessThanChordAngleDistance(t *testing.T) {
	// Tests that IsConservativeDistanceLessOrEqual returns points where the
	// true distance is slightly less than the one computed by ChordAngle.
	//
	// The points below had the worst error from among 100,000 random pairs.
	p0 := Point{r3.Vector{0.78516762584829192, -0.50200400690845970, -0.36263449417782678}}
	p1 := Point{r3.Vector{0.78563011732429433, -0.50187655940493503, -0.36180828883938054}}
	pv := &PointVector{p0}

	index := NewShapeIndex()
	index.Add(pv)
	query := NewClosestEdgeQuery(index, nil)

	// The ChordAngle distance is ~4 ulps greater than the true distance.
	dist := ChordAngleBetweenPoints(p0, p1)
	limit := dist.Predecessor().Predecessor().Predecessor().Predecessor()
	if got, want := CompareDistance(p0, p1, limit), 0; got >= want {
		t.Errorf("CompareDistance(%v, %v, %v) = %v, want >= %v", p0, p1, limit, got, want)
	}

	// Verify that IsConservativeDistanceLessOrEqual() still returns "p1".
	target := NewMinDistanceToPointTarget(p1)
	if query.IsDistanceLess(target, limit) {
		t.Errorf("query.IsDistanceLess(%v, %v) = true, want false", target, dist)
	}
	if query.IsDistanceLess(target, limit.Successor()) {
		t.Errorf("query.IsDistanceLessOrEqual(%v, %v) = true, want false", target, dist)
	}
	if !query.IsConservativeDistanceLessOrEqual(target, limit) {
		t.Errorf("query.IsConservativeDistanceLessOrEqual(%v, %v) = false, want true", target, dist)
	}
}

func TestClosestEdgeQueryTargetPointInsideIndexedPolygon(t *testing.T) {
	// Tests a target point in the interior of an indexed polygon.
	// (The index also includes a polyline loop with no interior.)
	index := makeShapeIndex("# 0:0, 0:5, 5:5, 5:0 # 0:10, 0:15, 5:15, 5:10")
	opts := NewClosestEdgeQueryOptions().
		IncludeInteriors(true).
		DistanceLimit(s1.ChordAngleFromAngle(s1.Angle(1) * s1.Degree))
	query := NewClosestEdgeQuery(index, opts)

	target := NewMinDistanceToPointTarget(parsePoint("2:12"))

	results := query.FindEdges(target)
	if len(results) != 1 {
		t.Fatalf("len(results) = %v, want 1", len(results))
	}

	r0 := results[0]
	if r0.Distance() != 0 {
		t.Errorf("result[0].Distance = %v, want 0", r0.Distance())
	}
	if r0.shapeID != 1 {
		t.Errorf("result[0].shapeID = %v, want 1", r0.shapeID)
	}
	if r0.edgeID != -1 {
		t.Errorf("result[0].edgeID = %v, want -1", r0.edgeID)
	}
	if !r0.IsInterior() {
		t.Errorf("first edge should have been interior to the polygon")
	}
	if r0.IsEmpty() {
		t.Errorf("result should not have been empty")
	}
}

func TestClosestEdgeQueryTargetPolygonContainingIndexedPoints(t *testing.T) {
	// Two points are contained within a polyline loop (no interior) and two
	// points are contained within a polygon.
	index := makeShapeIndex("2:2 | 3:3 | 1:11 | 3:13 # #")
	opts := NewClosestEdgeQueryOptions().
		UseBruteForce(false).
		DistanceLimit(s1.ChordAngleFromAngle(s1.Angle(1) * s1.Degree))
	query := NewClosestEdgeQuery(index, opts)

	targetIndex := makeShapeIndex("# 0:0, 0:5, 5:5, 5:0 # 0:10, 0:15, 5:15, 5:10")
	target := NewMinDistanceToShapeIndexTarget(targetIndex)

	target.setIncludeInteriors(true)
	results := query.FindEdges(target)

	// All points should be returned since we did not specify MaxResults.
	if len(results) != 2 {
		t.Errorf("2 shapes should have matched. Got %d shapes", len(results))
	}

	r0 := results[0]
	if r0.Distance() != 0 {
		t.Errorf("result[0].Distance != 0, want 0")
	}
	if r0.shapeID != 0 {
		t.Errorf("result[0].shapeID = %v, want 0", r0.shapeID)
	}
	if r0.edgeID != 2 {
		t.Errorf("result[0].edgeID = %v, want 0", r0.edgeID) // 1:11
	}

	r1 := results[1]
	if r1.Distance() != 0 {
		t.Errorf("result[1].Distance != 0, want 0")
	}
	if r1.shapeID != 0 {
		t.Errorf("result[1].shapeID = %v, want 0", r1.shapeID)
	}
	if r1.edgeID != 3 {
		t.Errorf("result[1].edgeID = %v, want 3", r1.edgeID) // 3:13
	}
}

func BenchmarkEdgeQueryFindEdgesClosestFractal(b *testing.B) {
	// Test searching within the general vicinity of the indexed shapes.
	opts := &edgeQueryBenchmarkOptions{
		fact:                     fractalLoopShapeIndexGenerator,
		includeInteriors:         false,
		targetType:               queryTypePoint,
		numTargetEdges:           0,
		chooseTargetFromIndex:    false,
		radiusKm:                 1000,
		maxDistanceFraction:      -1,
		maxErrorFraction:         -1,
		targetRadiusFraction:     0.0,
		centerSeparationFraction: -2.0,
	}

	benchmarkEdgeQueryFindClosest(b, opts)
}

func BenchmarkEdgeQueryFindEdgesClosestInterior(b *testing.B) {
	// Test searching within the general vicinity of the indexed shapes including interiors.
	opts := &edgeQueryBenchmarkOptions{
		fact:                     fractalLoopShapeIndexGenerator,
		includeInteriors:         true,
		targetType:               queryTypePoint,
		numTargetEdges:           0,
		chooseTargetFromIndex:    false,
		radiusKm:                 1000,
		maxDistanceFraction:      -1,
		maxErrorFraction:         -1,
		targetRadiusFraction:     0.0,
		centerSeparationFraction: -2.0,
	}

	benchmarkEdgeQueryFindClosest(b, opts)
}

func BenchmarkEdgeQueryFindEdgesClosestErrorPoint01Percent(b *testing.B) {
	// Test searching with an error tolerance.  Allowing 1% error makes searches
	// 6x faster in the case of regular loops with a large number of vertices.
	opts := &edgeQueryBenchmarkOptions{
		fact:                     fractalLoopShapeIndexGenerator,
		includeInteriors:         false,
		targetType:               queryTypePoint,
		numTargetEdges:           0,
		chooseTargetFromIndex:    false,
		radiusKm:                 1000,
		maxDistanceFraction:      -1,
		maxErrorFraction:         0.01,
		targetRadiusFraction:     0.0,
		centerSeparationFraction: -2.0,
	}

	benchmarkEdgeQueryFindClosest(b, opts)
}

func BenchmarkEdgeQueryFindEdgesClosestErrorPoint1Percent(b *testing.B) {
	// Test searching with an error tolerance.  Allowing 1% error makes searches
	// 6x faster in the case of regular loops with a large number of vertices.
	opts := &edgeQueryBenchmarkOptions{
		fact:                     fractalLoopShapeIndexGenerator,
		includeInteriors:         false,
		targetType:               queryTypePoint,
		numTargetEdges:           0,
		chooseTargetFromIndex:    false,
		radiusKm:                 1000,
		maxDistanceFraction:      -1,
		maxErrorFraction:         0.1,
		targetRadiusFraction:     0.0,
		centerSeparationFraction: -2.0,
	}

	benchmarkEdgeQueryFindClosest(b, opts)
}

// TODO(roberts): Remaining tests to implement.
//
// TestClosestEdgeQueryTestReuseOfQuery) {
// TestClosestEdgeQueryTargetPointInsideIndexedPolygon) {
// TestClosestEdgeQueryTargetPointOutsideIndexedPolygon) {
// TestClosestEdgeQueryTargetPolygonContainingIndexedPoints) {
// TestClosestEdgeQueryEmptyTargetOptimized) {
// TestClosestEdgeQueryEmptyPolygonTarget) {
// TestClosestEdgeQueryFullLaxPolygonTarget) {
// TestClosestEdgeQueryFullS2PolygonTarget) {
// TestClosestEdgeQueryIsConservativeDistanceLessOrEqual) {
// TestClosestEdgeQueryCircleEdges) {
// TestClosestEdgeQueryFractalEdges) {
// TestClosestEdgeQueryPointCloudEdges) {
// TestClosestEdgeQueryConservativeCellDistanceIsUsed) {
//
// More of the Benchmarking code.
