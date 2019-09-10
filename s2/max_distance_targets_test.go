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
	"sort"
	"testing"

	"github.com/golang/geo/s1"
)

func TestDistanceTargetMaxCellTargetCapBound(t *testing.T) {
	var md maxDistance
	zero := md.zero()

	for i := 0; i < 100; i++ {
		cell := CellFromCellID(randomCellID())
		target := NewMaxDistanceToCellTarget(cell)
		c := target.capBound()

		for j := 0; j < 100; j++ {
			pTest := randomPoint()
			// Check points outside of cap to be away from maxDistance's zero().
			if !c.ContainsPoint(pTest) {
				if got := cell.MaxDistance(pTest); !zero.less(maxDistance(got)) {
					t.Errorf("%v.MaxDistance(%v) = %v, want < %v", cell, pTest, got, zero)
				}
			}
		}
	}
}

func TestDistanceTargetMaxCellTargetUpdateDistance(t *testing.T) {
	var ok bool

	targetCell := CellFromCellID(cellIDFromPoint(parsePoint("0:1")))
	target := NewMaxDistanceToCellTarget(targetCell)

	dist0 := maxDistance(0)
	dist10 := maxDistance(s1.ChordAngleFromAngle(s1.Angle(10) * s1.Degree))

	// Update max distance target to point.
	p := parsePoint("0:0")
	if _, ok = target.updateDistanceToPoint(p, dist0); !ok {
		t.Errorf("target.updateDistanceToPoint(%v, %v) should have succeeded", p, dist0)
	}
	if _, ok = target.updateDistanceToPoint(p, dist10); ok {
		t.Errorf("target.updateDistanceToPoint(%v, %v) should have failed", p, dist10)
	}

	// Reset dist0 which was updated.
	dist0 = maxDistance(0)
	// Test for edges.
	pts := parsePoints("0:2, 0:3")
	edge := Edge{pts[0], pts[1]}
	if _, ok := target.updateDistanceToEdge(edge, dist0); !ok {
		t.Errorf("target.updateDistanceToEdge(%v, %v) should have succeeded", edge, dist0)
	}
	// Second call should fail.
	if _, ok := target.updateDistanceToEdge(edge, dist10); ok {
		t.Errorf("target.updateDistanceToEdge(%v, %v) should have failed", edge, dist10)
	}

	// Reset dist0 which was updated.
	dist0 = maxDistance(0)
	// Test for cell.
	cell := CellFromCellID(cellIDFromPoint(parsePoint("0:0")))
	if _, ok = target.updateDistanceToCell(cell, dist0); !ok {
		t.Errorf("target.updateDistanceToCell(%v, %v) should have succeeded", cell, dist0)
	}
	// Leaf cell will be tiny compared to 10 degrees - expect no update.
	if _, ok = target.updateDistanceToCell(cell, dist10); ok {
		t.Errorf("target.updateDistanceToCell(%v, %v) should have failed", cell, dist10)
	}
}

func TestDistanceTargetMaxCellTargetUpdateDistanceToCellAntipodal(t *testing.T) {
	var maxDist maxDistance

	p := parsePoint("0:0")
	targetCell := CellFromCellID(cellIDFromPoint(p))
	target := NewMaxDistanceToCellTarget(targetCell)
	dist := maxDist.infinity()
	cell := CellFromCellID(cellIDFromPoint(Point{p.Mul(-1)}))

	// First call should pass.
	dist0, ok := target.updateDistanceToCell(cell, dist)
	if !ok {
		t.Errorf("target.updateDistanceToCell(%v, %v) should have succeeded", cell, dist)
	}
	if dist0.chordAngle() != s1.StraightChordAngle {
		t.Errorf("target.updateDistanceToCell() = %v, want %v", dist0.chordAngle(), s1.StraightChordAngle)
	}
	// Second call should fail.
	if _, ok := target.updateDistanceToCell(cell, dist0); ok {
		t.Errorf("target.updateDistanceToCell(%v, %v) should have failed", cell, dist0)
	}
}

func TestDistanceTargetMaxCellTargetUpdateDistanceToCellWhenEqual(t *testing.T) {
	var maxDist maxDistance

	targetCell := CellFromCellID(cellIDFromPoint(parsePoint("0:1")))
	target := NewMaxDistanceToCellTarget(targetCell)
	dist := maxDist.infinity()
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

func TestDistanceTargetMaxCellTargetUpdateDistanceToEdgeWhenEqual(t *testing.T) {
	var maxDist maxDistance

	targetCell := CellFromCellID(cellIDFromPoint(parsePoint("0:1")))
	target := NewMaxDistanceToCellTarget(targetCell)
	dist := maxDist.infinity()
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

func TestDistanceTargetMaxCellTargetVisitContainingShapes(t *testing.T) {
	index := makeShapeIndex("1:1 # 1:1, 2:2 # 0:0, 0:3, 3:0 | 6:6, 6:9, 9:6 | -1:-1, -1:5, 5:-1")
	// Only shapes 2 and 4 should contain a very small cell near
	// the antipode of 1:1.
	targetCell := CellFromCellID(cellIDFromPoint(Point{parsePoint("1:1").Mul(-1)}))
	target := NewMaxDistanceToCellTarget(targetCell)

	if got, want := containingShapesForTarget(target, index, 1), []int{2}; !reflect.DeepEqual(got, want) {
		t.Errorf("containingShapesForTarget(%v, %q, 1) = %+v, want %+v", target, shapeIndexDebugString(index), got, want)
	}
	if got, want := containingShapesForTarget(target, index, 5), []int{2, 4}; !reflect.DeepEqual(got, want) {
		t.Errorf("containingShapesForTarget(%v, %q, 5) = %+v, want %+v", target, shapeIndexDebugString(index), got, want)
	}

	// For a larger antipodal cell that properly contains one or more index
	// cells, all shapes that intersect the first such cell in S2CellId order are
	// returned.  In the test below, this happens to again be the 1st and 3rd
	// polygons (whose shape_ids are 2 and 4).
	target2 := NewMaxDistanceToCellTarget(CellFromCellID(targetCell.ID().Parent(5)))
	if got, want := containingShapesForTarget(target2, index, 5), []int{2, 4}; !reflect.DeepEqual(got, want) {
		t.Errorf("containingShapesForTarget(%v, %q, 5) = %+v, want %+v", target2, shapeIndexDebugString(index), got, want)
	}
}

func TestDistanceTargetMaxPointTargetUpdateDistance(t *testing.T) {
	var ok bool
	var dist0, dist10 distance
	target := NewMaxDistanceToPointTarget(parsePoint("0:0"))
	dist0 = maxDistance(0)
	dist10 = maxDistance(s1.ChordAngleFromAngle(s1.Angle(10) * s1.Degree))

	// Update max distance target to point.
	p := parsePoint("1:0")
	if dist0, ok = target.updateDistanceToPoint(p, dist0); !ok {
		t.Errorf("target.updateDistanceToPoint(%v, %v) should have succeeded", p, dist0)
	}
	if got, want := dist0.chordAngle().Angle().Degrees(), 1.0; !float64Near(got, want, epsilon) {
		t.Errorf("target.updateDistanceToPoint(%v, %v) = %v, want ~%v", p, dist0.chordAngle(), got, want)
	}
	if _, ok = target.updateDistanceToPoint(p, dist10); ok {
		t.Errorf("target.updateDistanceToPoint(%v, %v) should have failed", p, dist0)

	}

	// Reset dist0 which was updated.
	dist0 = maxDistance(0)
	// Test for edges.
	pts := parsePoints("0:-1, 0:1")
	edge := Edge{pts[0], pts[1]}
	if dist0, ok = target.updateDistanceToEdge(edge, dist0); !ok {
		t.Errorf("target.updateDistanceToEdge(%v, %v) should have succeeded", edge, dist0)
	}
	if got, want := dist0.chordAngle().Angle().Degrees(), 1.0; !float64Near(got, want, epsilon) {
		t.Errorf("target.updateDistanceToEdge(%v, %v) = %v, want ~%v", edge, dist0.chordAngle(), got, want)
	}
	if _, ok = target.updateDistanceToEdge(edge, dist10); ok {
		t.Errorf("target.updateDistanceToEdge(%v, %v) should have failed", edge, dist10)
	}

	// Reset dist0 which was updated.
	dist0 = maxDistance(0)
	// Test for cell.
	cell := CellFromCellID(cellIDFromPoint(parsePoint("0:0")))
	if _, ok = target.updateDistanceToCell(cell, dist0); !ok {
		t.Errorf("target.updateDistanceToCell(%v, %v) should have succeeded", cell, dist0)
	}
	// Leaf cell will be tiny compared to 10 degrees - expect no update.
	if _, ok = target.updateDistanceToCell(cell, dist10); ok {
		t.Errorf("target.updateDistanceToCell(%v, %v) should have failed", cell, dist10)
	}
}

func TestDistanceTargetMaxPointTargetUpdateDistanceToCellWhenEqual(t *testing.T) {
	var maxDist maxDistance

	target := NewMaxDistanceToPointTarget(parsePoint("1:0"))
	dist := maxDist.infinity()
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

func TestDistanceTargetMaxPointTargetUpdateDistanceToEdgeWhenEqual(t *testing.T) {
	var maxDist maxDistance

	// Verifies that UpdateDistance only returns true when the new distance
	// is less than the old distance (not less than or equal to).
	target := NewMaxDistanceToPointTarget(parsePoint("1:0"))
	dist := maxDist.infinity()
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

func containingShapesForTarget(target distanceTarget, index *ShapeIndex, maxShapes int) []int {
	shapeIDs := map[int32]bool{}
	target.visitContainingShapes(index,
		func(containingShape Shape, targetPoint Point) bool {
			// TODO(roberts): Update this if Shapes get an ID.
			shapeIDs[index.idForShape(containingShape)] = true
			return len(shapeIDs) < maxShapes
		})
	var ids []int
	for k := range shapeIDs {
		ids = append(ids, int(k))
	}
	sort.Ints(ids)
	return ids
}

func TestDistanceTargetMaxPointTargetVisitContainingShapes(t *testing.T) {
	// Only shapes 2 and 4 should contain the target point.
	index := makeShapeIndex("1:1 # 1:1, 2:2 # 0:0, 0:3, 3:0 | 6:6, 6:9, 9:6 | 0:0, 0:4, 4:0")

	// Test against antipodal point.
	point := Point{parsePoint("1:1").Mul(-1)}
	target := NewMaxDistanceToPointTarget(point)

	if got, want := containingShapesForTarget(target, index, 1), []int{2}; !reflect.DeepEqual(got, want) {
		t.Errorf("containingShapesForTarget(%v, %q, 1) = %+v, want %+v", point, shapeIndexDebugString(index), got, want)
	}
	if got, want := containingShapesForTarget(target, index, 5), []int{2, 4}; !reflect.DeepEqual(got, want) {
		t.Errorf("containingShapesForTarget(%v, %q, 5) = %+v, want %+v", point, shapeIndexDebugString(index), got, want)
	}
}

func TestDistanceTargetMaxEdgeTargetUpdateDistance(t *testing.T) {
	var ok bool
	var dist0, dist10 distance

	targetPts := parsePoints("0:-1, 0:1")
	targetEdge := Edge{targetPts[0], targetPts[1]}
	target := NewMaxDistanceToEdgeTarget(targetEdge)

	dist0 = maxDistance(0)
	dist10 = maxDistance(s1.ChordAngleFromAngle(s1.Angle(10) * s1.Degree))

	// Update max distance target to point.
	p := parsePoint("0:2")
	if dist0, ok = target.updateDistanceToPoint(p, dist0); !ok {
		t.Errorf("target.updateDistanceToPoint(%v, %v) should have succeeded", p, dist0)
	}
	if got, want := dist0.chordAngle().Angle().Degrees(), 3.0; !float64Near(got, want, epsilon) {
		t.Errorf("target.updateDistanceToPoint(%v, %v) = %v, want ~%v", p, dist0.chordAngle(), got, want)
	}
	if _, ok = target.updateDistanceToPoint(p, dist10); ok {
		t.Errorf("target.updateDistanceToPoint(%v, %v) should have failed", p, dist10)
	}

	// Reset dist0 which was updated.
	dist0 = maxDistance(0)
	// Test for edges.
	pts := parsePoints("0:2, 0:3")
	edge := Edge{pts[0], pts[1]}
	if dist0, ok = target.updateDistanceToEdge(edge, dist0); !ok {
		t.Errorf("target.updateDistanceToEdge(%v, %v) should have succeeded", edge, dist0)
	}
	if got, want := dist0.chordAngle().Angle().Degrees(), 4.0; !float64Near(got, want, epsilon) {
		t.Errorf("target.updateDistanceToEdge(%v, %v) = %v, want ~%v", p, dist0.chordAngle(), got, want)
	}
	if _, ok = target.updateDistanceToEdge(edge, dist10); ok {
		t.Errorf("target.updateDistanceToEdge(%v, %v) should have failed", edge, dist10)
	}

	// Reset dist0 which was updated.
	dist0 = maxDistance(0)
	// Test for cell.
	cell := CellFromCellID(cellIDFromPoint(parsePoint("0:0")))
	if _, ok = target.updateDistanceToCell(cell, dist0); !ok {
		t.Errorf("target.updateDistanceToCell(%v, %v) should have succeeded", cell, dist0)
	}
	// Leaf cell will be tiny compared to 10 degrees - expect no update.
	if _, ok = target.updateDistanceToCell(cell, dist10); ok {
		t.Errorf("target.updateDistanceToCell(%v, %v) should have failed", cell, dist10)
	}
}

func TestDistanceTargetMaxEdgeTargetUpdateDistanceToCellWhenEqual(t *testing.T) {
	var maxDist maxDistance

	targetEdge := parsePoints("1:0, 1:1")
	target := NewMaxDistanceToEdgeTarget(Edge{targetEdge[0], targetEdge[1]})
	dist := maxDist.infinity()
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

func TestDistanceTargetMaxEdgeTargetUpdateDistanceToEdgeAntipodal(t *testing.T) {
	var maxDist maxDistance

	targetPts := parsePoints("0:89, 0:91")
	targetEdge := Edge{targetPts[0], targetPts[1]}
	target := NewMaxDistanceToEdgeTarget(targetEdge)
	dist := maxDist.infinity()
	pts := parsePoints("1:-90, -1:-90")
	edge := Edge{pts[0], pts[1]}

	// First call should pass.
	dist0, ok := target.updateDistanceToEdge(edge, dist)
	if !ok {
		t.Errorf("target.updateDistanceToEdge(%v, %v) should have succeeded", edge, dist)
	}

	if dist0.chordAngle() != s1.StraightChordAngle {
		t.Errorf("target.updateDistanceToPoint(%v, %v) = %v, want %v", edge, dist0, dist0, s1.StraightChordAngle)
	}
}

func TestDistanceTargetMaxEdgeTargetUpdateDistanceToEdgeWhenEqual(t *testing.T) {
	var maxDist maxDistance

	targetEdge := parsePoints("1:0, 1:1")
	target := NewMaxDistanceToEdgeTarget(Edge{targetEdge[0], targetEdge[1]})
	dist := maxDist.infinity()
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

func TestDistanceTargetMaxEdgeTargetVisitContainingShapes(t *testing.T) {
	// Only shapes 2 and 4 should contain the target edge.
	index := makeShapeIndex("1:1 # 1:1, 2:2 # 0:0, 0:3, 3:0 | 6:6, 6:9, 9:6 | 0:0, 0:4, 4:0")

	// Test against antipodal edge.
	pts := parsePoints("1:2, 2:1")
	edge := Edge{Point{pts[0].Mul(-1)}, Point{pts[1].Mul(-1)}}
	target := NewMaxDistanceToEdgeTarget(edge)

	if got, want := containingShapesForTarget(target, index, 1), []int{2}; !reflect.DeepEqual(got, want) {
		t.Errorf("containingShapesForTarget(%v, %q, 1) = %+v, want %+v", target, shapeIndexDebugString(index), got, want)
	}
	if got, want := containingShapesForTarget(target, index, 5), []int{2, 4}; !reflect.DeepEqual(got, want) {
		t.Errorf("containingShapesForTarget(%v, %q, 5) = %+v, want %+v", target, shapeIndexDebugString(index), got, want)
	}
}

func TestDistanceTargetMaxShapeIndexTargetCapBound(t *testing.T) {
	// TODO(roberts): Uncomment when ShapeIndexRegion is implemented.
	/*
		var md maxDistance
		zero := md.zero()
		inf := md.infinity()

		index := NewShapeIndex()
		index.Add(PolygonFromCell(CellFromCellID(randomCellID())))
		pv := PointVector([]Point{randomPoint()})
		index.Add(Shape(&pv))
		target := NewMaxDistanceToShapeIndexTarget(index)
		c := target.capBound()

		for j := 0; j < 100; j++ {
			pTest := randomPoint()
			// Check points outside of cap to be away from maxDistance's zero().
			if !c.ContainsPoint(pTest) {
				var curDist distance = inf
				var ok bool
				if curDist, ok = target.updateDistanceToPoint(pTest, curDist); !ok {
					t.Errorf("updateDistanceToPoint failed, but should have succeeeded")
					continue
				}
				if !zero.less(curDist) {
					t.Errorf("point %v outside of cap should be less than %v distance, but were %v", pTest, zero, curDist)
				}
			}
		}
	*/
}

func TestDistanceTargetMaxShapeIndexTargetUpdateDistanceToCellWhenEqual(t *testing.T) {
	var maxDist maxDistance

	index := makeShapeIndex("1:0 # #")
	target := NewMaxDistanceToShapeIndexTarget(index)
	dist := maxDist.infinity()
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

func TestDistanceTargetMaxShapeIndexTargetUpdateDistanceToEdgeWhenEqual(t *testing.T) {
	var maxDist maxDistance

	index := makeShapeIndex("1:0 # #")
	target := NewMaxDistanceToShapeIndexTarget(index)
	dist := maxDist.infinity()
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

// Negates S2 points to reflect them through the sphere.
func reflectPoints(pts []Point) []Point {
	var negativePts []Point
	for _, p := range pts {
		negativePts = append(negativePts, Point{p.Mul(-1)})
	}
	return negativePts
}

func TestDistanceTargetMaxShapeIndexTargetVisitContainingShapes(t *testing.T) {
	// Create an index containing a repeated grouping of one point, one
	// polyline, and one polygon.
	index := makeShapeIndex("1:1 | 4:4 | 7:7 | 10:10 # " +
		"1:1, 1:2 | 4:4, 4:5 | 7:7, 7:8 | 10:10, 10:11 # " +
		"0:0, 0:3, 3:0 | 3:3, 3:6, 6:3 | 6:6, 6:9, 9:6 | 9:9, 9:12, 12:9")

	// Construct a target consisting of one point, one polyline, and one polygon
	// with two loops where only the second loop is contained by a polygon in
	// the index above.
	targetIndex := NewShapeIndex()

	pts := PointVector(reflectPoints(parsePoints("1:1")))
	targetIndex.Add(&pts)

	line := Polyline(reflectPoints(parsePoints("4:5, 5:4")))
	targetIndex.Add(&line)

	loops := [][]Point{
		reflectPoints(parsePoints("20:20, 20:21, 21:20")),
		reflectPoints(parsePoints("10:10, 10:11, 11:10")),
	}
	laxPoly := laxPolygonFromPoints(loops)
	targetIndex.Add(laxPoly)

	target := NewMaxDistanceToShapeIndexTarget(targetIndex)

	// These are the shape_ids of the 1st, 2nd, and 4th polygons of "index"
	// (noting that the 4 points are represented by one S2PointVectorShape).
	if got, want := containingShapesForTarget(target, index, 5), []int{5, 6, 8}; !reflect.DeepEqual(got, want) {
		t.Errorf("containingShapesForTarget(%v, %q, 5) = %+v, want %+v", target, shapeIndexDebugString(index), got, want)
	}
}

func TestDistanceTargetMaxShapeIndexTargetVisitContainingShapesEmptyAndFull(t *testing.T) {
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
