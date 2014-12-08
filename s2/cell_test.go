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
