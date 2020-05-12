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

	"github.com/golang/geo/s1"
)

func TestFurthestEdgeQueryNoEdges(t *testing.T) {
	index := &ShapeIndex{}
	query := NewFurthestEdgeQuery(index, nil)
	target := NewMaxDistanceToPointTarget(PointFromCoords(1, 0, 0))
	edge := query.findEdge(target, query.opts)

	if edge.shapeID != -1 {
		t.Errorf("shapeID for empty index should be -1, got %v", edge.shapeID)
	}
	if edge.edgeID != -1 {
		t.Errorf("edgeID for empty index should be -1, got %v", edge.edgeID)
	}
	if got, want := edge.Distance(), s1.NegativeChordAngle; got != want {
		t.Errorf("edge.Distance = %+v, want %+v", got, want)
	}
	if got, want := query.Distance(target), s1.NegativeChordAngle; got != want {
		t.Errorf("query.Distance(%v) = %+v, want %+v", target, got, want)
	}
}

func TestFurthestEdgeQueryBasicTest(t *testing.T) {
	index := makeShapeIndex("0:1 | 0:2 | 0:3 # #")
	query := NewFurthestEdgeQuery(index, NewFurthestEdgeQueryOptions().
		MaxResults(3).
		DistanceLimit(s1.ChordAngleFromAngle(s1.Angle(1)*s1.Degree)).
		MaxError(s1.ChordAngleFromAngle(s1.Angle(0.001)*s1.Degree)))
	target := NewMaxDistanceToPointTarget(parsePoint("0:4"))

	result := query.findEdge(target, query.opts)

	if result.edgeID != 0 {
		t.Errorf("query.findEdge(%v).edgeID = %v, want %v", target, result.edgeID, 0)
	}

	if got, want := query.Distance(target).Angle().Degrees(), 3.0; !float64Near(got, want, epsilon) {
		t.Errorf("query.Distance(%v) = %v, want %v", target, got, want)
	}
	if !query.IsDistanceGreater(target, s1.ChordAngleFromAngle(s1.Angle(1.5)*s1.Degree)) {
		t.Errorf("query.IsDistanceGreater(%v, 1.5) should be true but wasn't", target)
	}
}

func TestFurthestEdgeQueryAntipodalPointInsideIndexedPolygon(t *testing.T) {
	// Tests a target point antipodal to the interior of an indexed polygon.
	// (The index also includes a polyline loop with no interior.)
	index := makeShapeIndex("# 0:0, 0:5, 5:5, 5:0 # 0:10, 0:15, 5:15, 5:10")

	// First check that with include_interiors set to true, the distance is 180.
	query := NewFurthestEdgeQuery(index, NewFurthestEdgeQueryOptions().
		IncludeInteriors(true).
		DistanceLimit(s1.ChordAngleFromAngle(s1.Angle(178)*s1.Degree)))
	target := NewMaxDistanceToPointTarget(Point{parsePoint("2:12").Mul(-1)})

	results := query.FindEdges(target)

	if len(results) < 1 {
		t.Errorf("aa")
	}
	result := results[0]
	if result.Distance() != s1.StraightChordAngle {
		t.Errorf("result.Distance = %v, want %v", result.Distance(), s1.StraightChordAngle)
	}
	// Should find the polygon shape (id = 1).
	if result.shapeID != 1 {
		t.Errorf("result.shapeID of result should be 1, was %v", result.shapeID)
	}
	// Should find the interior, so no specific edge id.
	if result.edgeID != -1 {
		t.Errorf("result.edgeID = %v, want -1", result.edgeID)
	}

	// Next check that with include_interiors set to false, the distance is less
	// than 180 for the same target and index.
	query.opts.IncludeInteriors(false)
	results = query.FindEdges(target)

	if len(results) <= 0 {
		t.Errorf("findEdges returned no results, should have been at least 1")
		return
	}
	result = results[0]
	if result.Distance() > s1.StraightChordAngle {
		t.Errorf("result.Distance = %v, want less than %v", result.Distance(), s1.StraightChordAngle)
	}
	if result.shapeID != 1 {
		t.Errorf("result.shapeID = %v, want 1", result.shapeID)
	}
	// Found a specific edge, so id should be positive.
	if result.edgeID == -1 {
		t.Errorf("result.edgeID = %v, want >= 0", result.edgeID)
	}
}

func TestFurthestEdgeQueryAntipodalPointOutsideIndexedPolygon(t *testing.T) {
	// Tests a target point antipodal to the interior of a polyline loop with no
	// interior.  The index also includes a polygon almost antipodal to the
	// target, but with all edges closer than the min_distance threshold.
	index := makeShapeIndex("# 0:0, 0:5, 5:5, 5:0 # 0:10, 0:15, 5:15, 5:10")
	query := NewFurthestEdgeQuery(index, NewFurthestEdgeQueryOptions().
		IncludeInteriors(true).
		DistanceLimit(s1.ChordAngleFromAngle(s1.Angle(179)*s1.Degree)))
	target := NewMaxDistanceToPointTarget(Point{parsePoint("2:2").Mul(-1)})

	results := query.FindEdges(target)
	if len(results) != 0 {
		t.Errorf("len(query.FindEdges(target)) = %d, want 0", len(results))
	}
}

func TestFurthestEdgeQueryTargetPolygonContainingIndexedPoints(t *testing.T) {
	// Two points are contained within a polyline loop (no interior) and two
	// points are contained within a polygon.
	index := makeShapeIndex("2:2 | 4:4 | 1:11 | 3:12 # #")
	query := NewFurthestEdgeQuery(index, NewFurthestEdgeQueryOptions().UseBruteForce(false))

	targetIndex := makeShapeIndex("# 0:0, 0:5, 5:5, 5:0 # 0:10, 0:15, 5:15, 5:10")
	target := NewMaxDistanceToShapeIndexTarget(targetIndex)
	target.setIncludeInteriors(true)
	target.setUseBruteForce(true)
	results := query.FindEdges(target)

	// All points should be returned since we did not specify max_results.
	if len(results) != 4 {
		t.Errorf("All 4 shapes should have matched. Got %d shapes", len(results))
	}

	r0 := results[0]
	if r0.Distance() == 0 {
		t.Errorf("result[0].Distance = 0, want non-zero")
	}
	if r0.shapeID != 0 {
		t.Errorf("result[0].shapeID = %v, want 0", r0.shapeID)
	}
	if r0.edgeID != 0 {
		t.Errorf("result[0].edgeID = %v, want 0", r0.edgeID) // 2:2 (to 5:15)
	}

	r1 := results[1]
	if r1.Distance() == 0 {
		t.Errorf("result[1].Distance = 0, want non-zero")
	}
	if r1.shapeID != 0 {
		t.Errorf("result[1].shapeID = %v, want 0", r1.shapeID)
	}
	if r1.edgeID != 3 {
		t.Errorf("result[1].edgeID = %v, want 3", r1.edgeID) // 3:12 (to 0:0)
	}
}

func TestFurthestEdgeQueryAntipodalPolygonContainingIndexedPoints(t *testing.T) {
	// Two antipodal points are contained within a polyline loop (no interior)
	// and two antipodal points are contained within a polygon.
	index := NewShapeIndex()
	pts := parsePoints("2:2, 3:3, 1:11, 3:13")
	antipodalPts := reflectPoints(pts)
	pv := PointVector(antipodalPts)
	index.Add(&pv)

	query := NewFurthestEdgeQuery(index, NewFurthestEdgeQueryOptions().
		DistanceLimit(s1.ChordAngleFromAngle(s1.Angle(179)*s1.Degree)),
	)
	targetIndex := makeShapeIndex("# 0:0, 0:5, 5:5, 5:0 # 0:10, 0:15, 5:15, 5:10")
	target := NewMaxDistanceToShapeIndexTarget(targetIndex)
	target.setIncludeInteriors(true)
	results := query.FindEdges(target)

	if len(results) != 2 {
		t.Errorf("2 shapes should have matched. Got %d shapes", len(results))
	}

	r0 := results[0]
	if r0.Distance() != s1.StraightChordAngle {
		t.Errorf("result[1].Distance = %v, want %v", r0.Distance(), s1.StraightChordAngle)
	}
	if r0.shapeID != 0 {
		t.Errorf("result[0].shapeID = %v, want 0", r0.shapeID)
	}
	if r0.edgeID != 2 {
		t.Errorf("result[0].edgeID = %v, want 2", r0.edgeID) // 1:11
	}

	r1 := results[1]
	if r1.Distance() != s1.StraightChordAngle {
		t.Errorf("result[1].Distance = %v, want %v", r1.Distance(), s1.StraightChordAngle)
	}
	if r1.shapeID != 0 {
		t.Errorf("result[1].shapeID = %v, want 0", r1.shapeID)
	}
	if r1.edgeID != 3 {
		t.Errorf("result[1].edgeID = %v, want 3", r1.edgeID) // 3:13
	}
}

func TestFurthestEdgeQueryEmptyPolygonTarget(t *testing.T) {
	// Verifies that distances are measured correctly to empty polygon targets.
	emptyPolygonIndex := makeShapeIndex("# # empty")
	pointIndex := makeShapeIndex("1:1 # #")
	fullPolygonIndex := makeShapeIndex("# # full")

	target := NewMaxDistanceToShapeIndexTarget(emptyPolygonIndex)

	emptyQuery := NewFurthestEdgeQuery(emptyPolygonIndex, NewFurthestEdgeQueryOptions().IncludeInteriors(true))
	if got, want := emptyQuery.Distance(target), s1.NegativeChordAngle; got != want {
		t.Errorf("emptyQuery.Distance(%v) = %v, want %v", target, got, want)
	}

	pointQuery := NewFurthestEdgeQuery(pointIndex, NewFurthestEdgeQueryOptions().IncludeInteriors(true))
	if got, want := pointQuery.Distance(target), s1.NegativeChordAngle; got != want {
		t.Errorf("pointQuery.Distance(%v) = %v, want %v", target, got, want)
	}

	fullQuery := NewFurthestEdgeQuery(fullPolygonIndex, NewFurthestEdgeQueryOptions().IncludeInteriors(true))
	if got, want := fullQuery.Distance(target), s1.NegativeChordAngle; got != want {
		t.Errorf("fullQuery.Distance(%v) = %v, want %v", target, got, want)
	}
}

func TestFurthestEdgeQueryFullPolygonTarget(t *testing.T) {
	// Verifies that distances are measured correctly to full Polygon targets
	// (which use a different representation of "full" than LaxPolygon does).
	emptyPolygonIndex := makeShapeIndex("# # empty")
	pointIndex := makeShapeIndex("1:1 # #")
	fullPolygonIndex := makeShapeIndex("# #")
	fullPolygonIndex.Add(makePolygon("full", true))

	target := NewMaxDistanceToShapeIndexTarget(fullPolygonIndex)

	emptyQuery := NewFurthestEdgeQuery(emptyPolygonIndex, NewFurthestEdgeQueryOptions().IncludeInteriors(true))
	if got, want := emptyQuery.Distance(target), s1.NegativeChordAngle; got != want {
		t.Errorf("emptyQuery.Distance(%v) = %v, want %v", target, got, want)
	}

	pointQuery := NewFurthestEdgeQuery(pointIndex, NewFurthestEdgeQueryOptions().IncludeInteriors(true))
	if got, want := pointQuery.Distance(target), s1.StraightChordAngle; got != want {
		t.Errorf("pointQuery.Distance(%v) = %v, want %v", target, got, want)
	}

	fullQuery := NewFurthestEdgeQuery(fullPolygonIndex, NewFurthestEdgeQueryOptions().IncludeInteriors(true))
	if got, want := fullQuery.Distance(target), s1.StraightChordAngle; got != want {
		t.Errorf("fullQuery.Distance(%v) = %v, want %v", target, got, want)
	}
}

/*
func TestFurthestEdgeQueryCircleEdges(t *testing.T) {
	testEdgeQueryWithGenerator(t,
		NewFurthestEdgeQuery,
		NewFurthestEdgeQueryOptions,
		loopShapeIndexGenerator, edgeQueryTestNumIndexes,
		edgeQueryTestNumEdges, edgeQueryTestNumQueries)
}
func TestFurthestEdgeQueryFractalEdges(t *testing.T) {
	testEdgeQueryWithGenerator(t,
		NewFurthestEdgeQuery,
		NewFurthestEdgeQueryOptions,
		fractalLoopShapeIndexGenerator, edgeQueryTestNumIndexes,
		edgeQueryTestNumEdges, edgeQueryTestNumQueries)
}
func TestFurthestEdgeQueryPointCloudEdges(t *testing.T) {
	testEdgeQueryWithGenerator(t,
		NewFurthestEdgeQuery,
		NewFurthestEdgeQueryOptions,
		pointCloudShapeIndexGenerator, edgeQueryTestNumIndexes,
		edgeQueryTestNumEdges, edgeQueryTestNumQueries)
}
*/

// TODO(roberts): Remaining tests to implement.
//
// func TestFurthestEdgeQueryDistanceEqualToLimit(t *testing.T) {}
// func TestFurthestEdgeQueryTrueDistanceGreaterThanChordAngleDistance(t *testing.T) { }
// func TestFurthestEdgeQueryFullLaxPolygonTarget(t *testing.T) {}
