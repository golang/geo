// Copyright 2023 Google Inc. All rights reserved.
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
)

// set padding to at least twice the maximum error for reliable results.
const shapeIndexCellPadding = 2 * (faceClipErrorUVCoord + intersectsRectErrorUVDist)

func padCell(id CellID, paddingUV float64) Shape {
	face, i, j, _ := id.faceIJOrientation()

	uv := ijLevelToBoundUV(i, j, id.Level()).ExpandedByMargin(paddingUV)

	vertices := make([]Point, 4)
	for i, v := range uv.Vertices() {
		vertices[i] = Point{faceUVToXYZ(face, v.X, v.Y).Normalize()}
	}

	return LaxLoopFromPoints(vertices)
}

func TestShapeIndexRegionCapBound(t *testing.T) {
	id := CellIDFromString("3/0123012301230123012301230123")

	// Add a polygon that is slightly smaller than the cell being tested.
	index := NewShapeIndex()
	index.Add(padCell(id, -shapeIndexCellPadding))

	cellBound := CellFromCellID(id).CapBound()
	indexBound := index.Region().CapBound()
	if !indexBound.Contains(cellBound) {
		t.Errorf("%v.Contains(%v) = false, want true", indexBound, cellBound)
	}

	// Note that CellUnion.CapBound returns a slightly larger bound than
	// Cell.CapBound even when the cell union consists of a single CellID.
	if got, want := indexBound.Radius(), 1.00001*cellBound.Radius(); got > want {
		t.Errorf("%v.CapBound.Radius() = %v, want %v", index, got, want)
	}
}

func TestShapeIndexRegionRectBound(t *testing.T) {
	id := CellIDFromString("3/0123012301230123012301230123")

	// Add a polygon that is slightly smaller than the cell being tested.
	index := NewShapeIndex()
	index.Add(padCell(id, -shapeIndexCellPadding))
	cellBound := CellFromCellID(id).RectBound()
	indexBound := index.Region().RectBound()

	if indexBound != cellBound {
		t.Errorf("%v.RectBound() = %v, want %v", index, indexBound, cellBound)
	}
}

func TestShapeIndexRegionCellUnionBoundMultipleFaces(t *testing.T) {
	have := []CellID{
		CellIDFromString("3/00123"),
		CellIDFromString("2/11200013"),
	}

	index := NewShapeIndex()
	for _, id := range have {
		index.Add(padCell(id, -shapeIndexCellPadding))
	}

	got := index.Region().CellUnionBound()

	sortCellIDs(have)

	if !CellUnion(have).Equal(CellUnion(got)) {
		t.Errorf("%v.CellUnionBound() = %v, want %v", index, got, have)
	}
}

func TestShapeIndexRegionCellUnionBoundOneFace(t *testing.T) {
	// This tests consists of 3 pairs of CellIDs.  Each pair is located within
	// one of the children of face 5, namely the cells 5/0, 5/1, and 5/3.
	// We expect CellUnionBound to compute the smallest cell that bounds the
	// pair on each face.
	have := []CellID{
		CellIDFromString("5/010"),
		CellIDFromString("5/0211030"),
		CellIDFromString("5/110230123"),
		CellIDFromString("5/11023021133"),
		CellIDFromString("5/311020003003030303"),
		CellIDFromString("5/311020023"),
	}

	want := []CellID{
		CellIDFromString("5/0"),
		CellIDFromString("5/110230"),
		CellIDFromString("5/3110200"),
	}

	index := NewShapeIndex()
	for _, id := range have {
		// Add each shape 3 times to ensure that the ShapeIndex subdivides.
		index.Add(padCell(id, -shapeIndexCellPadding))
		index.Add(padCell(id, -shapeIndexCellPadding))
		index.Add(padCell(id, -shapeIndexCellPadding))
	}

	sortCellIDs(have)

	got := index.Region().CellUnionBound()
	if !CellUnion(want).Equal(CellUnion(got)) {
		t.Errorf("%v.CellUnionBound() = %v, want %v", index, got, want)
	}
}

// TODO(roberts): remaining tests
// func TestShapeIndexRegionContainsCellMultipleShapes(t *testing.T) { }
// func TestShapeIndexRegionIntersectsShrunkenCell(t *testing.T){ }
// func TestShapeIndexRegionIntersectsExactCell(t *testing.T){ }
// Add VisitIntersectingShapes tests
// Benchmarks
