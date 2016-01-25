/*
Copyright 2014 Google Inc. All rights reserved.

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
	"unsafe"

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
			edgeCounts[cell.Edge(k)]++
			vertexCounts[cell.Vertex(k)]++
			if d := cell.Vertex(k).Dot(cell.Edge(k).Vector); !float64Eq(0.0, d) {
				t.Errorf("dot product of vertex and edge failed, got %v, want 0", d)
			}
			if d := cell.Vertex((k + 1) & 3).Dot(cell.Edge(k).Vector); !float64Eq(0.0, d) {
				t.Errorf("dot product for edge and next vertex failed, got %v, want 0", d)
			}
			if d := cell.Vertex(k).Vector.Cross(cell.Vertex((k + 1) & 3).Vector).Normalize().Dot(cell.Edge(k).Vector); !float64Eq(1.0, d) {
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

func TestExactArea(t *testing.T) {
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
		childrenArea := 0.0
		for _, child := range cell.Children() {
			childrenArea += CellFromCellID(child).ExactArea()
		}
		if area := CellFromCellID(cell).ExactArea(); !float64Eq(childrenArea, area) {
			t.Fatalf("Areas of children of a level-%d cell %v don't add up to parent's area. "+
				"This cell: %e, sum of children: %e",
				cell.Level(), cell, area, childrenArea)
		}
		childIndex = (childIndex + 1) % 4
	}
}

func TestIntersectsCell(t *testing.T) {
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

func TestContainsCell(t *testing.T) {
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

func TestRectBound(t *testing.T) {
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

func TestRectBoundAroundPoleMinLat(t *testing.T) {
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

func TestCapBound(t *testing.T) {
	c := CellFromCellID(CellIDFromFace(0).ChildBeginAtLevel(20))
	s2Cap := c.CapBound()
	for i := 0; i < 4; i++ {
		if !s2Cap.ContainsPoint(c.Vertex(i)) {
			t.Errorf("%v should contain %v", s2Cap, c.Vertex(i))
		}
	}
}

func TestContainsPoint(t *testing.T) {
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

func TestContainsPointConsistentWithS2CellIDFromPoint(t *testing.T) {
	// Construct many points that are nearly on a Cell edge, and verify that
	// CellFromCellID(cellIDFromPoint(p)).Contains(p) is always true.
	for iter := 0; iter < 1000; iter++ {
		cell := CellFromCellID(randomCellID())
		i1 := randomUniformInt(4)
		i2 := (i1 + 1) & 3
		v1 := cell.Vertex(i1)
		v2 := samplePointFromCap(CapFromCenterAngle(cell.Vertex(i2), s1.Angle(epsilon)))
		p := Interpolate(randomFloat64(), v1, v2)
		if !CellFromCellID(cellIDFromPoint(p)).ContainsPoint(p) {
			t.Errorf("For p=%v, CellFromCellID(cellIDFromPoint(p)).ContainsPoint(p) was false", p)
		}
	}
}

func TestContainsPointContainsAmbiguousPoint(t *testing.T) {
	// This tests a case where S2CellId returns the "wrong" cell for a point
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
