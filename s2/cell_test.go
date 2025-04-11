// Copyright 2014 Google Inc. All rights reserved.
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
	"unsafe"

	"github.com/golang/geo/r2"
	"github.com/golang/geo/s1"
)

// maxCellSize is the upper bounds on the number of bytes we want the Cell object to ever be.
const maxCellSize = 48

func TestCellObjectSize(t *testing.T) {
	if sz := unsafe.Sizeof(Cell{}); sz > maxCellSize {
		t.Errorf("Cell struct too big: %d bytes > %d bytes", sz, maxCellSize)
	}
}

func TestCellFaces(t *testing.T) {
	edgeCounts := make(map[Point]int)
	vertexCounts := make(map[Point]int)

	for face := 0; face < 6; face++ {
		id := CellIDFromFace(face)
		cell := CellFromCellID(id)

		if cell.id != id {
			t.Errorf("cell.id != id; %v != %v", cell.id, id)
		}

		if cell.face != int8(face) {
			t.Errorf("cell.face != face: %v != %v", cell.face, face)
		}

		if cell.level != 0 {
			t.Errorf("cell.level != 0: %v != 0", cell.level)
		}

		// Top-level faces have alternating orientations to get RHS coordinates.
		if cell.orientation != int8(face&swapMask) {
			t.Errorf("cell.orientation != orientation: %v != %v", cell.orientation, face&swapMask)
		}

		if cell.IsLeaf() {
			t.Errorf("cell should not be a leaf: IsLeaf = %v", cell.IsLeaf())
		}
		for k := 0; k < 4; k++ {
			edgeCounts[cell.EdgeRaw(k)]++
			vertexCounts[cell.VertexRaw(k)]++
			if d := cell.VertexRaw(k).Dot(cell.EdgeRaw(k).Vector); !float64Eq(0.0, d) {
				t.Errorf("dot product of vertex and edge failed, got %v, want 0", d)
			}
			if d := cell.VertexRaw((k + 1) & 3).Dot(cell.EdgeRaw(k).Vector); !float64Eq(0.0, d) {
				t.Errorf("dot product for edge and next vertex failed, got %v, want 0", d)
			}
			if d := cell.VertexRaw(k).Vector.Cross(cell.VertexRaw((k + 1) & 3).Vector).Normalize().Dot(cell.Edge(k).Vector); !float64Eq(1.0, d) {
				t.Errorf("dot product of cross product for vertices failed, got %v, want 1.0", d)
			}
		}
	}

	// Check that edges have multiplicity 2 and vertices have multiplicity 3.
	for k, v := range edgeCounts {
		if v != 2 {
			t.Errorf("edge %v counts wrong, got %d, want 2", k, v)
		}
	}
	for k, v := range vertexCounts {
		if v != 3 {
			t.Errorf("vertex %v counts wrong, got %d, want 3", k, v)
		}
	}
}

func TestCellUVCoordOfEdge(t *testing.T) {
	// Four cells on face 0 with two boundaries each on 0/0.
	cell0 := []Cell{
		CellFromCellID(CellIDFromToken("0f")),
		CellFromCellID(CellIDFromToken("05")),
		CellFromCellID(CellIDFromToken("1b")),
		CellFromCellID(CellIDFromToken("11")),
	}

	// And four cells on face 4 which is rotated w.r.t face 0.
	cell4 := []Cell{
		CellFromCellID(CellIDFromToken("8f")),
		CellFromCellID(CellIDFromToken("85")),
		CellFromCellID(CellIDFromToken("9b")),
		CellFromCellID(CellIDFromToken("91")),
	}

	for k := 0; k < 4; k++ {
		if got, want := cell0[k].UVCoordOfEdge(k+0), 0.0; !float64Eq(got, want) {
			t.Errorf("%v.UVCoordOfEdge[%d] = %f, want %f", cell4[k], k+0, got, want)
		}
		if got, want := cell0[k].UVCoordOfEdge(k+1), 0.0; !float64Eq(got, want) {
			t.Errorf("%v.UVCoordOfEdge[%d] = %f, want %f", cell4[k], k+1, got, want)
		}
		if got, want := cell4[k].UVCoordOfEdge(k+0), 0.0; !float64Eq(got, want) {
			t.Errorf("%v.UVCoordOfEdge[%d] = %f, want %f", cell4[k], k+1, got, want)
		}
		if got, want := cell4[k].UVCoordOfEdge(k+1), 0.0; !float64Eq(got, want) {
			t.Errorf("%v.UVCoordOfEdge[%d] = %f, want %f", cell4[k], k+1, got, want)
		}
	}
}

func Test2CellIJCoordOfEdge(t *testing.T) {
	for i := 0; i < 100; i++ {
		id := randomCellID()
		cell := CellFromCellID(id)

		// Look up the canonical IJ coordinates of the cell boundary.
		var ij [2]int
		_, ij[0], ij[1], _ = id.faceIJOrientation()

		ijSize := sizeIJ(id.Level())
		var ijBounds r2.Rect
		ijLo := ij[0] & -ijSize
		ijBounds.X.Lo = float64(ijLo)
		ijBounds.X.Hi = float64(ijLo + ijSize)

		ijLo = ij[1] & -ijSize
		ijBounds.Y.Lo = float64(ijLo)
		ijBounds.Y.Hi = float64(ijLo + ijSize)

		// Check that each boundary coordinate is correct.
		for k := 0; k < 4; k++ {
			got := cell.IJCoordOfEdge(k)
			var want int
			if (k+1)%2 == 0 {
				want = int(ijBounds.Vertices()[k].X)
			} else {
				want = int(ijBounds.Vertices()[k].Y)
			}
			if got != want {
				t.Errorf("%v.IJCoordOfEdge(%d) = %v, want %v", cell, k, got, want)
			}
		}
	}
}

func TestCellChildren(t *testing.T) {
	testCellChildren(t, CellFromCellID(CellIDFromFace(0)))
	testCellChildren(t, CellFromCellID(CellIDFromFace(3)))
	testCellChildren(t, CellFromCellID(CellIDFromFace(5)))
}

func testCellChildren(t *testing.T, cell Cell) {
	children, ok := cell.Children()
	if cell.IsLeaf() && !ok {
		return
	}
	if cell.IsLeaf() && ok {
		t.Errorf("leaf cells should not be able to return children. cell %v", cell)
	}

	if !ok {
		t.Errorf("unable to get Children for %v", cell)
		return
	}

	childID := cell.id.ChildBegin()
	for i, ci := range children {
		// Check that the child geometry is consistent with its cell ID.
		if childID != ci.id {
			t.Errorf("%v.child[%d].id = %v, want %v", cell, i, ci.id, childID)
		}

		direct := CellFromCellID(childID)
		if !ci.Center().ApproxEqual(childID.Point()) {
			t.Errorf("%v.Center() = %v, want %v", ci, ci.Center(), childID.Point())
		}
		if ci.face != direct.face {
			t.Errorf("%v.face = %v, want %v", ci, ci.face, direct.face)
		}
		if ci.level != direct.level {
			t.Errorf("%v.level = %v, want %v", ci, ci.level, direct.level)
		}
		if ci.orientation != direct.orientation {
			t.Errorf("%v.orientation = %v, want %v", ci, ci.orientation, direct.orientation)
		}
		if !ci.Center().ApproxEqual(direct.Center()) {
			t.Errorf("%v.Center() = %v, want %v", ci, ci.Center(), direct.Center())
		}

		for k := 0; k < 4; k++ {
			if !direct.VertexRaw(k).ApproxEqual(ci.VertexRaw(k)) {
				t.Errorf("child %d %v.VertexRaw(%d) = %v, want %v", i, ci, k, ci.VertexRaw(k), direct.VertexRaw(k))
			}
			if direct.EdgeRaw(k) != ci.EdgeRaw(k) {
				t.Errorf("child %d %v.EdgeRaw(%d) = %v, want %v", i, ci, k, ci.EdgeRaw(k), direct.EdgeRaw(k))
			}
		}

		// Test ContainsCell() and IntersectsCell().
		if !cell.ContainsCell(ci) {
			t.Errorf("%v.ContainsCell(%v) = false, want true", cell, ci)
		}
		if !cell.IntersectsCell(ci) {
			t.Errorf("%v.IntersectsCell(%v) = false, want true", cell, ci)
		}
		if ci.ContainsCell(cell) {
			t.Errorf("%v.ContainsCell(%v) = true, want false", ci, cell)
		}
		if !cell.ContainsPoint(ci.Center()) {
			t.Errorf("%v.ContainsPoint(%v) = false, want true", cell, ci.Center())
		}
		for j := 0; j < 4; j++ {
			if !cell.ContainsPoint(ci.VertexRaw(j)) {
				t.Errorf("%v.ContainsPoint(%v.VertexRaw(%d)) = false, want true", cell, ci, j)
			}
			if j != i {
				if ci.ContainsPoint(children[j].Center()) {
					t.Errorf("%v.ContainsPoint(%v[%d].Center()) = true, want false", ci, children, j)
				}
				if ci.IntersectsCell(children[j]) {
					t.Errorf("%v.IntersectsCell(%v[%d]) = true, want false", ci, children, j)
				}
			}
		}

		// Test CapBound and RectBound.
		parentCap := cell.CapBound()
		parentRect := cell.RectBound()
		if cell.ContainsPoint(PointFromCoords(0, 0, 1)) || cell.ContainsPoint(PointFromCoords(0, 0, -1)) {
			if !parentRect.Lng.IsFull() {
				t.Errorf("%v.Lng.IsFull() = false, want true", parentRect)
			}
		}
		childCap := ci.CapBound()
		childRect := ci.RectBound()
		if !childCap.ContainsPoint(ci.Center()) {
			t.Errorf("childCap %v.ContainsPoint(%v.Center()) = false, want true", childCap, ci)
		}
		if !childRect.ContainsPoint(ci.Center()) {
			t.Errorf("childRect %v.ContainsPoint(%v.Center()) = false, want true", childRect, ci)
		}
		if !parentCap.ContainsPoint(ci.Center()) {
			t.Errorf("parentCap %v.ContainsPoint(%v.Center()) = false, want true", parentCap, ci)
		}
		if !parentRect.ContainsPoint(ci.Center()) {
			t.Errorf("parentRect %v.ContainsPoint(%v.Center()) = false, want true", parentRect, ci)
		}
		for j := 0; j < 4; j++ {
			if !childCap.ContainsPoint(ci.Vertex(j)) {
				t.Errorf("childCap %v.ContainsPoint(%v.Vertex(%d)) = false, want true", childCap, ci, j)
			}
			if !childRect.ContainsPoint(ci.Vertex(j)) {
				t.Errorf("childRect %v.ContainsPoint(%v.Vertex(%d)) = false, want true", childRect, ci, j)
			}
			if !childRect.ContainsPoint(ci.VertexRaw(j)) {
				t.Errorf("childRect %v.ContainsPoint(%v.VertexRaw(%d)) = false, want true", childRect, ci, j)
			}
			if !parentCap.ContainsPoint(ci.Vertex(j)) {
				t.Errorf("parentCap %v.ContainsPoint(%v.Vertex(%d)) = false, want true", parentCap, ci, j)
			}
			if !parentRect.ContainsPoint(ci.Vertex(j)) {
				t.Errorf("parentRect %v.ContainsPoint(%v.Vertex(%d)) = false, want true", parentRect, ci, j)
			}
			if !parentRect.ContainsPoint(ci.VertexRaw(j)) {
				t.Errorf("parentRect %v.ContainsPoint(%v.VertexRaw(%d)) = false, want true", parentRect, ci, j)
			}
			if j != i {
				// The bounding caps and rectangles should be tight enough so that
				// they exclude at least two vertices of each adjacent cell.
				capCount := 0
				rectCount := 0
				for k := 0; k < 4; k++ {
					if childCap.ContainsPoint(children[j].Vertex(k)) {
						capCount++
					}
					if childRect.ContainsPoint(children[j].VertexRaw(k)) {
						rectCount++
					}
				}
				if capCount > 2 {
					t.Errorf("childs bounding cap should contain no more than 2 points, got %d", capCount)
				}
				if childRect.Lat.Lo > -math.Pi/2 && childRect.Lat.Hi < math.Pi/2 {
					// Bounding rectangles may be too large at the poles
					// because the pole itself has an arbitrary longitude.
					if rectCount > 2 {
						t.Errorf("childs bounding rect should contain no more than 2 points, got %d", rectCount)
					}
				}
			}
		}

		// Check all children for the first few levels, and then sample randomly.
		// We also always subdivide the cells containing a few chosen points so
		// that we have a better chance of sampling the minimum and maximum metric
		// values.  kMaxSizeUV is the absolute value of the u- and v-coordinate
		// where the cell size at a given level is maximal.
		maxSizeUV := 0.3964182625366691
		specialUV := []r2.Point{
			{X: dblEpsilon, Y: dblEpsilon}, // Face center
			{X: dblEpsilon, Y: 1},          // Edge midpoint
			{X: 1, Y: 1},                   // Face corner
			{X: maxSizeUV, Y: maxSizeUV},   // Largest cell area
			{X: dblEpsilon, Y: maxSizeUV},  // Longest edge/diagonal
		}
		forceSubdivide := false
		for _, uv := range specialUV {
			if ci.BoundUV().ContainsPoint(uv) {
				forceSubdivide = true
			}
		}

		// For a more in depth test, add an "|| oneIn(n)" to this condition
		// to cause more children to be tested beyond the ones to level 5.
		if forceSubdivide || cell.level < 5 {
			testCellChildren(t, ci)
		}

		childID = childID.Next()
	}
}

func TestCellAreas(t *testing.T) {
	// relative error bounds for each type of area computation
	var exactError = math.Log(1 + 1e-6)
	var approxError = math.Log(1.03)
	var avgError = math.Log(1 + 1e-15)

	// Test 1. Check the area of a top level cell.
	const level1Cell = CellID(0x1000000000000000)
	const wantArea = 4 * math.Pi / 6
	if area := CellFromCellID(level1Cell).ExactArea(); !float64Eq(area, wantArea) {
		t.Fatalf("Area of a top-level cell %v = %f, want %f", level1Cell, area, wantArea)
	}

	// Test 2. Iterate inwards from this cell, checking at every level that
	// the sum of the areas of the children is equal to the area of the parent.
	childIndex := 1
	for cell := CellID(0x1000000000000000); cell.Level() < 21; cell = cell.Children()[childIndex] {
		var exactArea, approxArea, avgArea float64
		for _, child := range cell.Children() {
			exactArea += CellFromCellID(child).ExactArea()
			approxArea += CellFromCellID(child).ApproxArea()
			avgArea += CellFromCellID(child).AverageArea()
		}

		if area := CellFromCellID(cell).ExactArea(); !float64Eq(exactArea, area) {
			t.Fatalf("Areas of children of a level-%d cell %v don't add up to parent's area. "+
				"This cell: %e, sum of children: %e",
				cell.Level(), cell, area, exactArea)
		}

		childIndex = (childIndex + 1) % 4

		// For ExactArea(), the best relative error we can expect is about 1e-6
		// because the precision of the unit vector coordinates is only about 1e-15
		// and the edge length of a leaf cell is about 1e-9.
		if logExact := math.Abs(math.Log(exactArea / CellFromCellID(cell).ExactArea())); logExact > exactError {
			t.Errorf("The relative error of ExactArea for children of a level-%d "+
				"cell %v should be less than %e, got %e. This cell: %e, children area: %e",
				cell.Level(), cell, exactError, logExact,
				CellFromCellID(cell).ExactArea(), exactArea)
		}
		// For ApproxArea(), the areas are accurate to within a few percent.
		if logApprox := math.Abs(math.Log(approxArea / CellFromCellID(cell).ApproxArea())); logApprox > approxError {
			t.Errorf("The relative error of ApproxArea for children of a level-%d "+
				"cell %v should be within %e%%, got %e. This cell: %e, sum of children: %e",
				cell.Level(), cell, approxError, logApprox,
				CellFromCellID(cell).ExactArea(), exactArea)
		}
		// For AverageArea(), the areas themselves are not very accurate, but
		// the average area of a parent is exactly 4 times the area of a child.
		if logAvg := math.Abs(math.Log(avgArea / CellFromCellID(cell).AverageArea())); logAvg > avgError {
			t.Errorf("The relative error of AverageArea for children of a level-%d "+
				"cell %v should be less than %e, got %e. This cell: %e, sum of children: %e",
				cell.Level(), cell, avgError, logAvg,
				CellFromCellID(cell).AverageArea(), avgArea)
		}
	}
}

func TestCellIntersectsCell(t *testing.T) {
	tests := []struct {
		c    Cell
		oc   Cell
		want bool
	}{
		{
			CellFromCellID(CellIDFromFace(0).ChildBeginAtLevel(2)),
			CellFromCellID(CellIDFromFace(0).ChildBeginAtLevel(2)),
			true,
		},
		{
			CellFromCellID(CellIDFromFace(0).ChildBeginAtLevel(2)),
			CellFromCellID(CellIDFromFace(0).ChildBeginAtLevel(2).ChildBeginAtLevel(5)),
			true,
		},
		{
			CellFromCellID(CellIDFromFace(0).ChildBeginAtLevel(2)),
			CellFromCellID(CellIDFromFace(0).ChildBeginAtLevel(2).Next()),
			false,
		},
	}
	for _, test := range tests {
		if got := test.c.IntersectsCell(test.oc); got != test.want {
			t.Errorf("Cell(%v).IntersectsCell(%v) = %t; want %t", test.c, test.oc, got, test.want)
		}
	}
}

func TestCellContainsCell(t *testing.T) {
	tests := []struct {
		c    Cell
		oc   Cell
		want bool
	}{
		{
			CellFromCellID(CellIDFromFace(0).ChildBeginAtLevel(2)),
			CellFromCellID(CellIDFromFace(0).ChildBeginAtLevel(2)),
			true,
		},
		{
			CellFromCellID(CellIDFromFace(0).ChildBeginAtLevel(2)),
			CellFromCellID(CellIDFromFace(0).ChildBeginAtLevel(2).ChildBeginAtLevel(5)),
			true,
		},
		{
			CellFromCellID(CellIDFromFace(0).ChildBeginAtLevel(2).ChildBeginAtLevel(5)),
			CellFromCellID(CellIDFromFace(0).ChildBeginAtLevel(2)),
			false,
		},
		{
			CellFromCellID(CellIDFromFace(0).ChildBeginAtLevel(2).Next()),
			CellFromCellID(CellIDFromFace(0).ChildBeginAtLevel(2)),
			false,
		},
		{
			CellFromCellID(CellIDFromFace(0).ChildBeginAtLevel(2)),
			CellFromCellID(CellIDFromFace(0).ChildBeginAtLevel(2).Next()),
			false,
		},
	}
	for _, test := range tests {
		if got := test.c.ContainsCell(test.oc); got != test.want {
			t.Errorf("Cell(%v).ContainsCell(%v) = %t; want %t", test.c, test.oc, got, test.want)
		}
	}
}

func TestCellRectBound(t *testing.T) {
	tests := []struct {
		lat float64
		lng float64
	}{
		{50, 50},
		{-50, 50},
		{50, -50},
		{-50, -50},
		{0, 0},
		{0, 180},
		{0, -179},
	}
	for _, test := range tests {
		c := CellFromLatLng(LatLngFromDegrees(test.lat, test.lng))
		rect := c.RectBound()
		for i := 0; i < 4; i++ {
			if !rect.ContainsLatLng(LatLngFromPoint(c.Vertex(i))) {
				t.Errorf("%v should contain %v", rect, c.Vertex(i))
			}
		}
	}
}

func TestCellRectBoundAroundPoleMinLat(t *testing.T) {
	tests := []struct {
		cellID       CellID
		latLng       LatLng
		wantContains bool
	}{
		{
			cellID:       CellIDFromFacePosLevel(2, 0, 0),
			latLng:       LatLngFromDegrees(3, 0),
			wantContains: false,
		},
		{
			cellID:       CellIDFromFacePosLevel(2, 0, 0),
			latLng:       LatLngFromDegrees(50, 0),
			wantContains: true,
		},
		{
			cellID:       CellIDFromFacePosLevel(5, 0, 0),
			latLng:       LatLngFromDegrees(-3, 0),
			wantContains: false,
		},
		{
			cellID:       CellIDFromFacePosLevel(5, 0, 0),
			latLng:       LatLngFromDegrees(-50, 0),
			wantContains: true,
		},
	}
	for _, test := range tests {
		if got := CellFromCellID(test.cellID).RectBound().ContainsLatLng(test.latLng); got != test.wantContains {
			t.Errorf("CellID(%v) contains %v: got %t, want %t", test.cellID, test.latLng, got, test.wantContains)
		}
	}
}

func TestCellCapBound(t *testing.T) {
	c := CellFromCellID(CellIDFromFace(0).ChildBeginAtLevel(20))
	s2Cap := c.CapBound()
	for i := 0; i < 4; i++ {
		if !s2Cap.ContainsPoint(c.Vertex(i)) {
			t.Errorf("%v should contain %v", s2Cap, c.Vertex(i))
		}
	}
}

func TestCellContainsPoint(t *testing.T) {
	tests := []struct {
		c    Cell
		p    Point
		want bool
	}{
		{
			CellFromCellID(CellIDFromFace(0).ChildBeginAtLevel(2)),
			CellFromCellID(CellIDFromFace(0).ChildBeginAtLevel(2).ChildBeginAtLevel(5)).Vertex(1),
			true,
		},
		{
			CellFromCellID(CellIDFromFace(0).ChildBeginAtLevel(2)),
			CellFromCellID(CellIDFromFace(0).ChildBeginAtLevel(2)).Vertex(1),
			true,
		},
		{
			CellFromCellID(CellIDFromFace(0).ChildBeginAtLevel(2).ChildBeginAtLevel(5)),
			CellFromCellID(CellIDFromFace(0).ChildBeginAtLevel(2).Next().ChildBeginAtLevel(5)).Vertex(1),
			false,
		},
	}
	for _, test := range tests {
		if got := test.c.ContainsPoint(test.p); got != test.want {
			t.Errorf("Cell(%v).ContainsPoint(%v) = %t; want %t", test.c, test.p, got, test.want)
		}
	}
}

func TestCellContainsPointConsistentWithS2CellIDFromPoint(t *testing.T) {
	// TODO: Is it still about 1% flaky with a random seed.
	// TODO(rsned): https://github.com/golang/geo/issues/120

	// Construct many points that are nearly on a Cell edge, and verify that
	// CellFromCellID(cellIDFromPoint(p)).Contains(p) is always true.
	for iter := 0; iter < 1000; iter++ {
		cell := CellFromCellID(randomCellID())
		i1 := randomUniformInt(4)
		i2 := (i1 + 1) & 3
		v1 := cell.Vertex(i1)
		v2 := samplePointFromCap(CapFromCenterAngle(cell.Vertex(i2), s1.Angle(epsilon)))
		p := Interpolate(randomUniformFloat64(0, 1.0), v1, v2)
		if !CellFromCellID(cellIDFromPoint(p)).ContainsPoint(p) {
			t.Errorf("For p=%v, CellFromCellID(cellIDFromPoint(p)).ContainsPoint(p) was false", p)
		}
	}
}

func TestCellContainsPointContainsAmbiguousPoint(t *testing.T) {
	// This tests a case where CellID returns the "wrong" cell for a point
	// that is very close to the cell edge. (ConsistentWithS2CellIdFromPoint
	// generates more examples like this.)
	//
	// The Point below should have x = 0, but conversion from LatLng to
	// (x,y,z) gives x = ~6.1e-17. When xyz is converted to uv, this gives
	// u = -6.1e-17. However when converting to st, which has a range of [0,1],
	// the low precision bits of u are lost and we wind up with s = 0.5.
	// cellIDFromPoint then chooses an arbitrary neighboring cell.
	//
	// This tests that Cell.ContainsPoint() expands the cell bounds sufficiently
	// so that the returned cell is still considered to contain p.
	p := PointFromLatLng(LatLngFromDegrees(-2, 90))
	cell := CellFromCellID(cellIDFromPoint(p).Parent(1))
	if !cell.ContainsPoint(p) {
		t.Errorf("For p=%v, CellFromCellID(cellIDFromPoint(p)).ContainsPoint(p) was false", p)
	}
}

func TestCellDistance(t *testing.T) {
	for iter := 0; iter < 1000; iter++ {
		cell := CellFromCellID(randomCellID())
		target := randomPoint()

		expectedToBoundary := minDistanceToPointBruteForce(cell, target).Angle()

		expectedToInterior := expectedToBoundary
		if cell.ContainsPoint(target) {
			expectedToInterior = 0
		}
		expectedMax := maxDistanceToPointBruteForce(cell, target).Angle()

		actualToBoundary := cell.BoundaryDistance(target).Angle()
		actualToInterior := cell.Distance(target).Angle()
		actualMax := cell.MaxDistance(target).Angle()

		// The error has a peak near pi/2 for edge distance, and another peak near
		// pi for vertex distance.
		if !float64Near(expectedToBoundary.Radians(), actualToBoundary.Radians(), 1e-12) {
			t.Errorf("%v.BoundaryDistance(%v) = %v, want %v", cell, target, actualToBoundary, expectedToBoundary)
		}
		if !float64Near(expectedToInterior.Radians(), actualToInterior.Radians(), 1e-12) {
			t.Errorf("%v.Distance(%v) = %v, want %v", cell, target, actualToInterior, expectedToInterior)
		}
		if !float64Near(expectedMax.Radians(), actualMax.Radians(), 1e-12) {
			t.Errorf("%v.MaxDistance(%v) = %v, want %v", cell, target, actualMax, expectedMax)
		}

		if expectedToBoundary.Radians() <= math.Pi/3 {
			if !float64Near(expectedToBoundary.Radians(), actualToBoundary.Radians(), 1e-15) {
				t.Errorf("%v.BoundaryDistance(%v) = %v, want %v", cell, target, actualToBoundary, expectedToBoundary)
			}
			if !float64Near(expectedToInterior.Radians(), actualToInterior.Radians(), 1e-15) {
				t.Errorf("%v.Distance(%v) = %v, want %v", cell, target, actualToInterior, expectedToInterior)
			}
		}

		if expectedMax.Radians() <= math.Pi/3 {
			if !float64Near(expectedMax.Radians(), actualMax.Radians(), 1e-15) {
				t.Errorf("%v.MaxDistance(%v) = %v, want %v", cell, target, actualMax.Radians(), expectedMax.Radians())

			}
		}
	}
}

func chooseEdgeNearCell(cell Cell) (a, b Point) {
	c := cell.CapBound()
	if oneIn(5) {
		// Choose a point anywhere on the sphere.
		a = randomPoint()
	} else {
		// Choose a point inside or somewhere near the cell.
		a = samplePointFromCap(CapFromCenterChordAngle(c.center, 1.5*c.radius))
	}

	// Now choose a maximum edge length ranging from very short to very long
	// relative to the cell size, and choose the other endpoint.
	maxLength := math.Min(100*math.Pow(1e-4, randomFloat64())*c.Radius().Radians(), math.Pi/2)
	b = samplePointFromCap(CapFromCenterAngle(a, s1.Angle(maxLength)))

	// Occasionally replace edge with antipodal edge.
	if oneIn(20) {
		a = Point{a.Mul(-1)}
		b = Point{b.Mul(-1)}
	}

	return a, b
}

func minDistanceToPointBruteForce(cell Cell, target Point) s1.ChordAngle {
	minDistance := s1.InfChordAngle()
	for i := 0; i < 4; i++ {
		minDistance, _ = UpdateMinDistance(target, cell.Vertex(i),
			cell.Vertex((i+1)%4), minDistance)
	}
	return minDistance
}

func maxDistanceToPointBruteForce(cell Cell, target Point) s1.ChordAngle {
	if cell.ContainsPoint(Point{target.Mul(-1)}) {
		return s1.StraightChordAngle
	}
	maxDistance := s1.NegativeChordAngle
	for i := 0; i < 4; i++ {
		maxDistance, _ = UpdateMaxDistance(target, cell.Vertex(i),
			cell.Vertex((i+1)%4), maxDistance)
	}
	return maxDistance
}

func minDistanceToEdgeBruteForce(cell Cell, a, b Point) s1.ChordAngle {
	if cell.ContainsPoint(a) || cell.ContainsPoint(b) {
		return s1.ChordAngle(0)
	}

	minDist := s1.InfChordAngle()
	for i := 0; i < 4; i++ {
		v0 := cell.Vertex(i)
		v1 := cell.Vertex((i + 1) % 4)
		// If the antipodal edge crosses through the cell, min distance is 0.
		if CrossingSign(a, b, v0, v1) != DoNotCross {
			return s1.ChordAngle(0)
		}
		minDist, _ = UpdateMinDistance(a, v0, v1, minDist)
		minDist, _ = UpdateMinDistance(b, v0, v1, minDist)
		minDist, _ = UpdateMinDistance(v0, a, b, minDist)
	}
	return minDist
}

func maxDistanceToEdgeBruteForce(cell Cell, a, b Point) s1.ChordAngle {
	// If any antipodal endpoint is within the cell, the max distance is Pi.
	if cell.ContainsPoint(Point{a.Mul(-1)}) || cell.ContainsPoint(Point{b.Mul(-1)}) {
		return s1.StraightChordAngle
	}

	maxDist := s1.NegativeChordAngle
	for i := 0; i < 4; i++ {
		v0 := cell.Vertex(i)
		v1 := cell.Vertex((i + 1) % 4)
		// If the antipodal edge crosses through the cell, min distance is Pi.
		if CrossingSign(Point{a.Mul(-1)}, Point{b.Mul(-1)}, v0, v1) != DoNotCross {
			return s1.StraightChordAngle
		}
		maxDist, _ = UpdateMaxDistance(a, v0, v1, maxDist)
		maxDist, _ = UpdateMaxDistance(b, v0, v1, maxDist)
		maxDist, _ = UpdateMaxDistance(v0, a, b, maxDist)
	}
	return maxDist
}

func TestCellDistanceToEdge(t *testing.T) {
	// TODO: Is it still about 0.1% flaky with a random seed.
	// TODO(rsned): https://github.com/golang/geo/issues/120

	for iter := 0; iter < 1000; iter++ {
		cell := CellFromCellID(randomCellID())

		a, b := chooseEdgeNearCell(cell)
		expectedMin := minDistanceToEdgeBruteForce(cell, a, b).Angle()
		expectedMax := maxDistanceToEdgeBruteForce(cell, a, b).Angle()
		actualMin := cell.DistanceToEdge(a, b).Angle()
		actualMax := cell.MaxDistanceToEdge(a, b).Angle()

		// The error has a peak near Pi/2 for edge distance, and another peak near
		// Pi for vertex distance.
		expectedError := 1e-12
		if expectedMin.Radians() > math.Pi/2 {
			// Max error for ChordAngle as it approaches Pi is about 3e-8.
			expectedError = 3e-8
		} else if expectedMin.Radians() <= math.Pi/3 {
			expectedError = 1e-15
		}

		if !float64Near(expectedMin.Radians(), actualMin.Radians(), expectedError) {
			t.Errorf("%v.DistanceToEdge(%v, %v) = %v, want %v", cell, a, b, actualMin, expectedMin)
		}

		if !float64Near(expectedMax.Radians(), actualMax.Radians(), 1e-12) {
			t.Errorf("%v.MaxDistanceToEdge(%v, %v) = %v, want %v", cell, a, b, actualMax, expectedMax)
		}
		if expectedMax.Radians() <= math.Pi/3 && !float64Near(expectedMax.Radians(), actualMax.Radians(), 1e-15) {
			t.Errorf("%v.MaxDistanceToEdge(%v, %v) = %v, want %v", cell, a, b, actualMax, expectedMax)
		}
	}
}

func TestCellMaxDistanceToEdge(t *testing.T) {
	// Test an edge for which its antipode crosses the cell. Validates both the
	// standard and brute force implementations for this case.
	cell := CellFromCellID(CellIDFromFacePosLevel(0, 0, 20))
	a := Point{Interpolate(2.0, cell.Center(), cell.Vertex(0)).Mul(-1)}
	b := Point{Interpolate(2.0, cell.Center(), cell.Vertex(2)).Mul(-1)}

	actual := cell.MaxDistanceToEdge(a, b)
	expected := maxDistanceToEdgeBruteForce(cell, a, b)

	if !float64Near(expected.Angle().Radians(), s1.StraightChordAngle.Angle().Radians(), 1e-15) {
		t.Errorf("brute force %v.MaxDistanceToEdge(%v, %v) = %v, want ~%v", cell, a, b, expected, s1.StraightChordAngle.Angle().Radians())
	}
	if !float64Near(actual.Angle().Radians(), s1.StraightChordAngle.Angle().Radians(), 1e-15) {
		t.Errorf("actual %v.MaxDistanceToEdge(%v, %v) = %v, want ~%v", cell, a, b, actual, s1.StraightChordAngle.Angle().Radians())
	}
}

func TestCellMaxDistanceToCellAntipodal(t *testing.T) {
	p := parsePoint("0:0")
	cell := CellFromPoint(p)
	antipodalCell := CellFromPoint(Point{p.Mul(-1)})
	dist := cell.MaxDistanceToCell(antipodalCell)

	if dist != s1.StraightChordAngle {
		t.Errorf("%v.MaxDistanceToCell(%v) = %v, want %v", cell, antipodalCell, dist, s1.StraightChordAngle)
	}
}

func TestCellMaxDistanceToCell(t *testing.T) {
	for i := 0; i < 1000; i++ {
		cell := CellFromCellID(randomCellID())
		testCell := CellFromCellID(randomCellID())
		antipodalLeafID := cellIDFromPoint(Point{testCell.Center().Mul(-1)})
		antipodalTestCell := CellFromCellID(antipodalLeafID.Parent(testCell.Level()))

		distFromMin := s1.StraightChordAngle - cell.DistanceToCell(antipodalTestCell)
		distFromMax := cell.MaxDistanceToCell(testCell)
		if !float64Near(distFromMin.Angle().Radians(), distFromMax.Angle().Radians(), 1e-8) {
			t.Errorf("min distance from antipodal cell: %v - %v.DistanceToCell(%v) = %v, max distance to cell: %v.MaxDistanceToCell(%v) = %v. difference = %v, want < %v", s1.StraightChordAngle, cell, antipodalTestCell, distFromMin, cell, testCell, distFromMax, math.Abs((distFromMin.Angle().Radians() - distFromMax.Angle().Radians())), 1e-8)
		}
	}
}

// TODO(rsned): Differences from C++.
// CellVsLoopRectBound
// RectBoundIsLargeEnough
