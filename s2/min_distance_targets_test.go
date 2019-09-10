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
	"reflect"
	"testing"
)

func TestDistanceTargetMinCellTargetUpdateDistanceToCellWhenEqual(t *testing.T) {
	var minDist minDistance

	targetCell := CellFromCellID(cellIDFromPoint(parsePoint("0:1")))
	target := NewMinDistanceToCellTarget(targetCell)
	dist := minDist.infinity()
	cell := CellFromCellID(cellIDFromPoint(parsePoint("0:0")))

	// First call should pass.
	dist0, ok := target.updateDistanceToCell(cell, dist)
	if !ok {
		t.Errorf("target.updateDistanceToCell(%v, %v) should have succeeded", cell, dist)
	}
	// Second call should fail.
	if _, ok := target.updateDistanceToCell(cell, dist0); ok {
		t.Errorf("target.updateDistanceToCell(%v, %v) should have failed", cell, dist0)
	}
}

func TestDistanceTargetMinCellTargetUpdateDistanceToEdgeWhenEqual(t *testing.T) {
	var minDist minDistance

	targetCell := CellFromCellID(cellIDFromPoint(parsePoint("0:1")))
	target := NewMinDistanceToCellTarget(targetCell)
	dist := minDist.infinity()
	pts := parsePoints("0:-1, 0:1")
	edge := Edge{pts[0], pts[1]}

	// First call should pass.
	dist0, ok := target.updateDistanceToEdge(edge, dist)
	if !ok {
		t.Errorf("target.updateDistanceToEdge(%v, %v) should have succeeded", edge, dist)
	}
	// Second call should fail.
	if _, ok := target.updateDistanceToEdge(edge, dist0); ok {
		t.Errorf("target.updateDistanceToEdge(%v, %v) should have failed", edge, dist0)
	}
}

func TestDistanceTargetMinCellTargetVisitContainingShapes(t *testing.T) {
	// Only shapes 2 and 4 should contain a very small cell near 1:1.
	index := makeShapeIndex("1:1 # 1:1, 2:2 # 0:0, 0:3, 3:0 | 6:6, 6:9, 9:6 | -1:-1, -1:5, 5:-1")
	targetCell := CellFromCellID(cellIDFromPoint(parsePoint("1:1")))
	target := NewMinDistanceToCellTarget(targetCell)

	if got, want := containingShapesForTarget(target, index, 1), []int{2}; !reflect.DeepEqual(got, want) {
		t.Errorf("containingShapesForTarget(%v, %q, 1) = %+v, want %+v", target, shapeIndexDebugString(index), got, want)
	}
	if got, want := containingShapesForTarget(target, index, 5), []int{2, 4}; !reflect.DeepEqual(got, want) {
		t.Errorf("containingShapesForTarget(%v, %q, 5) = %+v, want %+v", target, shapeIndexDebugString(index), got, want)
	}

	// For a larger cell that properly contains one or more index cells, all
	// shapes that intersect the first such cell in CellID order are returned.
	// In the test below, this happens to again be the 1st and 3rd polygons
	// (whose shape_ids are 2 and 4).
	target2 := NewMinDistanceToCellTarget(CellFromCellID(targetCell.ID().Parent(5)))
	if got, want := containingShapesForTarget(target2, index, 5), []int{2, 4}; !reflect.DeepEqual(got, want) {
		t.Errorf("containingShapesForTarget(%v, %q, 5) = %+v, want %+v", target2, shapeIndexDebugString(index), got, want)
	}
}

func TestDistanceTargetMinCellUnionTargetUpdateDistanceToCellWhenEqual(t *testing.T) {
	// TODO(roberts): Uncomment when implented.
	/*
		var minDist minDistance

		targetCellUnion := CellUnion([]CellID{cellIDFromPoint(parsePoint("0:1"))})
		target := NewMinDistanceToCellUnionTarget(targetCellUnion)
		dist := minDist.infinity()
		cell := CellFromCellID(cellIDFromPoint(parsePoint("0:0")))

		// First call should pass.
		dist0, ok := target.updateDistanceToCell(cell, dist)
		if !ok {
			t.Errorf("target.updateDistanceToCell(%v, %v) should have succeeded", cell, dist)
		}
		// Second call should fail.
		if dist1, ok := target.updateDistanceToCell(cell, dist0); ok {
			t.Errorf("target.updateDistanceToCell(%v, %v) should have failed", cell, dist0)
		}
	*/
}

func TestDistanceTargetMinCellUnionTargetUpdateDistanceToEdgeWhenEqual(t *testing.T) {
	// TODO(roberts): Uncomment when implented.
	/*
		var minDist minDistance

		targetCellUnion := CellUnion([]CellID{cellIDFromPoint(parsePoint("0:1"))})
		target := NewMinDistanceToCellUnionTarget(targetCellUnion)
		dist := minDist.infinity()
		pts := parsePoints("0:-1, 0:1")
		edge := Edge{pts[0], pts[1]}

		// First call should pass.
		dist0, ok := target.updateDistanceToEdge(edge, dist)
		if !ok {
			t.Errorf("target.updateDistanceToEdge(%v, %v) should have succeeded", edge, dist)
		}
		// Second call should fail.
		if dist1, ok := target.updateDistanceToEdge(edge, dist0); ok {
			t.Errorf("target.updateDistanceToEdge(%v, %v) should have failed", edge, dist0)
		}
	*/
}

func TestDistanceTargetMinCellUnionTargetVisitContainingShapes(t *testing.T) {
	// TODO(roberts): Uncomment when implented.
	/*
		index := makeShapeIndex("1:1 # 1:1, 2:2 # 0:0, 0:3, 3:0 | 6:6, 6:9, 9:6 | -1:-1, -1:5, 5:-1")

		// Shapes 2 and 4 contain the leaf cell near 1:1, while shape 3 contains the
		// leaf cell near 7:7.
		targetCellUnion := CellUnion([]CellID{
			cellIDFromPoint(parsePoint("1:1")),
			cellIDFromPoint(parsePoint("7:7")),
		})
		target := NewMinDistanceToCellUnionTarget(targetCellUnion)

		if got, want := containingShapesForTarget(target, index, 1), []int{2}; !reflect.DeepEqual(got, want) {
			t.Errorf("containingShapesForTarget(%v, %q, 1) = %+v, want %+v", targetEdge, shapeIndexDebugString(index), got, want)
		}
		if got, want := containingShapesForTarget(target, index, 5), []int{2, 3, 4}; !reflect.DeepEqual(got, want) {
			t.Errorf("containingShapesForTarget(%v, %q, 5) = %+v, want %+v", targetEdge, shapeIndexDebugString(index), got, want)
		}
	*/
}

func TestDistanceTargetMinEdgeTargetUpdateDistanceToCellWhenEqual(t *testing.T) {
	var minDist minDistance

	targetEdge := parsePoints("1:0, 1:1")
	target := NewMinDistanceToEdgeTarget(Edge{targetEdge[0], targetEdge[1]})
	dist := minDist.infinity()
	cell := CellFromCellID(cellIDFromPoint(parsePoint("0:0")))

	// First call should pass.
	dist0, ok := target.updateDistanceToCell(cell, dist)
	if !ok {
		t.Errorf("target.updateDistanceToCell(%v, %v) should have succeeded", cell, dist)
	}
	// Second call should fail.
	if _, ok := target.updateDistanceToCell(cell, dist0); ok {
		t.Errorf("target.updateDistanceToCell(%v, %v) should have failed", cell, dist0)
	}
}

func TestDistanceTargetMinEdgeTargetUpdateDistanceToEdgeWhenEqual(t *testing.T) {
	var minDist minDistance

	targetEdge := parsePoints("1:0, 1:1")
	target := NewMinDistanceToEdgeTarget(Edge{targetEdge[0], targetEdge[1]})
	dist := minDist.infinity()
	pts := parsePoints("0:-1, 0:1")
	edge := Edge{pts[0], pts[1]}

	// First call should pass.
	dist0, ok := target.updateDistanceToEdge(edge, dist)
	if !ok {
		t.Errorf("target.updateDistanceToEdge(%v, %v) should have succeeded", edge, dist)
	}
	// Second call should fail.
	if _, ok := target.updateDistanceToEdge(edge, dist0); ok {
		t.Errorf("target.updateDistanceToEdge(%v, %v) should have failed", edge, dist0)
	}
}

func TestDistanceTargetMinEdgeTargetVisitContainingShapes(t *testing.T) {
	// Only shapes 2 and 4 should contain the target point.
	index := makeShapeIndex("1:1 # 1:1, 2:2 # 0:0, 0:3, 3:0 | 6:6, 6:9, 9:6 | 0:0, 0:4, 4:0")

	targetEdge := parsePoints("1:2, 2:1")
	target := NewMinDistanceToEdgeTarget(Edge{targetEdge[0], targetEdge[1]})

	if got, want := containingShapesForTarget(target, index, 1), []int{2}; !reflect.DeepEqual(got, want) {
		t.Errorf("containingShapesForTarget(%v, %q, 1) = %+v, want %+v", targetEdge, shapeIndexDebugString(index), got, want)
	}
	if got, want := containingShapesForTarget(target, index, 5), []int{2, 4}; !reflect.DeepEqual(got, want) {
		t.Errorf("containingShapesForTarget(%v, %q, 5) = %+v, want %+v", targetEdge, shapeIndexDebugString(index), got, want)
	}
}

func TestDistanceTargetMinPointTargetUpdateDistanceToCellWhenEqual(t *testing.T) {
	target := NewMinDistanceToPointTarget(parsePoint("1:0"))
	var minDist minDistance
	dist := minDist.infinity()
	cell := CellFromCellID(cellIDFromPoint(parsePoint("0:0")))

	// First call should pass.
	dist1, ok := target.updateDistanceToCell(cell, dist)
	if !ok {
		t.Errorf("target.updateDistanceToCell(%v, %v) should have succeeded", cell, dist)
	}
	// Second call should fail.
	if _, ok := target.updateDistanceToCell(cell, dist1); ok {
		t.Errorf("target.updateDistanceToCell(%v, %v) should have failed", cell, dist)
	}
}

func TestDistanceTargetMinPointTargetUpdateDistanceToEdgeWhenEqual(t *testing.T) {
	target := NewMinDistanceToPointTarget(parsePoint("1:0"))
	var minDist minDistance
	dist := minDist.infinity()
	edge := parsePoints("0:-1, 0:1")

	// First call should pass.
	dist1, ok := target.updateDistanceToEdge(Edge{edge[0], edge[1]}, dist)
	if !ok {
		t.Errorf("target.updateDistanceToEdge(%v, %v) should have succeeded", edge, dist)
	}
	// Second call should fail.
	if _, ok := target.updateDistanceToEdge(Edge{edge[0], edge[1]}, dist1); ok {
		t.Errorf("target.updateDistanceToEdge(%v, %v) should have failed", edge, dist1)
	}
}

func TestDistanceTargetMinPointTargetVisitContainingShapes(t *testing.T) {
	// Only shapes 2 and 4 should contain the target point.
	index := makeShapeIndex("1:1 # 1:1, 2:2 # 0:0, 0:3, 3:0 | 6:6, 6:9, 9:6 | 0:0, 0:4, 4:0")
	point := parsePoint("1:1")
	target := NewMinDistanceToPointTarget(point)

	if got, want := containingShapesForTarget(target, index, 1), []int{2}; !reflect.DeepEqual(got, want) {
		t.Errorf("containingShapesForTarget(%v, %q, 1) = %+v, want %+v", point, shapeIndexDebugString(index), got, want)
	}
	if got, want := containingShapesForTarget(target, index, 5), []int{2, 4}; !reflect.DeepEqual(got, want) {
		t.Errorf("containingShapesForTarget(%v, %q, 5) = %+v, want %+v", point, shapeIndexDebugString(index), got, want)
	}
}

func TestDistanceTargetMinShapeIndexTargetUpdateDistanceToCellWhenEqual(t *testing.T) {
	index := makeShapeIndex("1:0 # #")
	target := NewMinDistanceToShapeIndexTarget(index)
	var minDist minDistance
	dist := minDist.infinity()
	cell := CellFromCellID(cellIDFromPoint(parsePoint("0:0")))

	// First call should pass.
	dist1, ok := target.updateDistanceToCell(cell, dist)
	if !ok {
		t.Errorf("target.updateDistanceToCell(%v, %v) should have succeeded", cell, dist)
	}

	// Repeat call should fail.
	if _, ok := target.updateDistanceToCell(cell, dist1); ok {
		t.Errorf("target.updateDistanceToCell(%v, %v) should have failed", cell, dist1)
	}
}

func TestDistanceTargetMinShapeIndexTargetUpdateDistanceToEdgeWhenEqual(t *testing.T) {
	index := makeShapeIndex("1:0 # #")
	target := NewMinDistanceToShapeIndexTarget(index)
	var minDist minDistance
	dist := minDist.infinity()

	pts := parsePoints("0:-1, 0:1")
	edge := Edge{pts[0], pts[1]}

	// First call should pass.
	dist0, ok := target.updateDistanceToEdge(edge, dist)
	if !ok {
		t.Errorf("target.updateDistanceToEdge(%v, %v) should have succeeded", edge, dist)
	}
	// Second call should fail.
	if _, ok := target.updateDistanceToEdge(edge, dist0); ok {
		t.Errorf("target.updateDistanceToEdge(%v, %v) should have failed", edge, dist0)
	}
}

func TestDistanceTargetMinShapeIndexTargetVisitContainingShapes(t *testing.T) {
	// Create an index containing a repeated grouping of one point, one
	// polyline, and one polygon.
	index := makeShapeIndex("1:1 | 4:4 | 7:7 | 10:10 # " +
		"1:1, 1:2 | 4:4, 4:5 | 7:7, 7:8 | 10:10, 10:11 # " +
		"0:0, 0:3, 3:0 | 3:3, 3:6, 6:3 | 6:6, 6:9, 9:6 | 9:9, 9:12, 12:9")

	// Construct a target consisting of one point, one polyline, and one polygon
	// with two loops where only the second loop is contained by a polygon in
	// the index above.
	targetIndex := makeShapeIndex("1:1 # 4:5, 5:4 # 20:20, 20:21, 21:20; 10:10, 10:11, 11:10")
	target := NewMinDistanceToShapeIndexTarget(targetIndex)

	// These are the shape_ids of the 1st, 2nd, and 4th polygons of "index"
	// (noting that the 4 points are represented by one PointVectorShape).
	if got, want := containingShapesForTarget(target, index, 5), []int{5, 6, 8}; !reflect.DeepEqual(got, want) {
		t.Errorf("containingShapesForTarget(%v, %q, 5) = %+v, want %+v", target, shapeIndexDebugString(index), got, want)
	}
}

func TestDistanceTargetMinShapeIndexTargetVisitContainingShapesEmptyAndFull(t *testing.T) {
	// Verify that VisitContainingShapes never returns empty polygons and always
	// returns full polygons (i.e., those containing the entire sphere).

	// Creating an index containing one empty and one full polygon.
	index := makeShapeIndex("# # empty | full")

	// Check only the full polygon is returned for a point target.
	pointIndex := makeShapeIndex("1:1 # #")
	pointTarget := NewMinDistanceToShapeIndexTarget(pointIndex)
	if got, want := containingShapesForTarget(pointTarget, index, 5), []int{1}; !reflect.DeepEqual(got, want) {
		t.Errorf("containingShapesForTarget(%v, %q, 5) = %+v, want %+v", pointTarget, shapeIndexDebugString(index), got, want)
	}

	// Check only the full polygon is returned for a full polygon target.
	fullPolygonIndex := makeShapeIndex("# # full")
	fullTarget := NewMinDistanceToShapeIndexTarget(fullPolygonIndex)
	if got, want := containingShapesForTarget(fullTarget, index, 5), []int{1}; !reflect.DeepEqual(got, want) {
		t.Errorf("containingShapesForTarget(%v, %q, 5) = %+v, want %+v", fullTarget, shapeIndexDebugString(index), got, want)
	}

	// Check that nothing is returned for an empty polygon target.  (An empty
	// polygon has no connected components and does not intersect anything, so
	// according to the API of GetContainingShapes nothing should be returned.)
	emptyPolygonIndex := makeShapeIndex("# # empty")
	emptyTarget := NewMinDistanceToShapeIndexTarget(emptyPolygonIndex)
	if got, want := containingShapesForTarget(emptyTarget, index, 5), []int(nil); !reflect.DeepEqual(got, want) {
		t.Errorf("containingShapesForTarget(%v, %q, 5) = %+v, want %+v", emptyTarget, shapeIndexDebugString(index), got, want)
	}
}
