// Copyright 2025 The S2 Geometry Project Authors. All rights reserved.
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

// TestIndexCellDataDimensionFilteringWorks verifies that we can filter
// shapes by dimension when loading cells.
func testIndexCellDataDimensionFilteringWorks(t *testing.T) {
	tests := []struct {
		dimWanted [3]bool
		dimEmpty  [3]bool
	}{
		{
			// Check that we get all dimensions by default.
			dimWanted: [3]bool{true, true, true},
			dimEmpty:  [3]bool{false, false, false},
		},
		{
			// No dimensions should work too, we just don't decode edges.
			dimWanted: [3]bool{false, false, false},
			dimEmpty:  [3]bool{true, true, true},
		},
		// Should be able to get ranges even if a dimension is off.
		{
			dimWanted: [3]bool{false, true, true},
			dimEmpty:  [3]bool{true, false, false},
		},
		{
			dimWanted: [3]bool{false, true, false},
			dimEmpty:  [3]bool{true, false, true},
		},
		{
			dimWanted: [3]bool{true, false, true},
			dimEmpty:  [3]bool{false, true, false},
		},
		{
			dimWanted: [3]bool{true, false, false},
			dimEmpty:  [3]bool{false, true, true},
		},
		{
			dimWanted: [3]bool{false, false, true},
			dimEmpty:  [3]bool{true, true, false},
		},
	}

	for i, test := range tests {
		index := makeShapeIndex("0:0" +
			"#1:1, 2:2" +
			"#1:0, 0:1, -1:0, 0:-1")

		iter := index.Iterator()
		iter.Begin()
		data := newIndexCellData()
		for j := 0; j < 3; j++ {
			data.setDimWanted(j, test.dimWanted[j])
		}
		data.loadCell(index, iter.CellID(), iter.cell)

		for dim := 0; dim < 2; dim++ {
			empty := len(data.dimEdges(dim)) == 0
			// Didn't get the answer we expected.
			if empty != test.dimEmpty[dim] {
				if test.dimEmpty[dim] {
					t.Errorf("%d. Expected empty dimEdges(%d), but got non-empty", i, dim)
				} else {
					t.Errorf("%d. Expected non-empty dimEdges(%d), but got empty", i, dim)
				}
			}
		}
	}

	// Finally, test dimRangeEdges as well.
	index := makeShapeIndex("0:0" +
		"#1:1, 2:2" +
		"#1:0, 0:1, -1:0, 0:-1")

	iter := index.Iterator()
	iter.Begin()
	data := newIndexCellData()
	data.setDimWanted(0, false)
	data.setDimWanted(1, true)
	data.setDimWanted(2, true)
	data.loadCell(index, iter.CellID(), iter.cell)
	if len(data.dimRangeEdges(0, 0)) != 0 {
		t.Error("Expected empty dimRangeEdges(0, 0)")
	}
	if len(data.dimRangeEdges(0, 2)) == 0 {
		t.Error("Expected non-empty dimRangeEdges(0, 2)")
	}
}

// TestIndexCellDataCellAndCenterRecomputed verifies that when we load a
// new unique cell, the cached cell and center values are updated if we
// access them.
func testIndexCellDataCellAndCenterRecomputed(t *testing.T) {
	// A line between two faces will guarantee we get at least two cells.
	index := makeShapeIndex("# 0:0, 0:-90 #")

	// Get the first cell from the index
	iter := index.Iterator()
	iter.Begin()

	data := newIndexCellData()
	data.loadCell(index, iter.CellID(), iter.cell)

	center0 := data.center()
	cell0 := data.cell()

	// Move to the next cell
	iter.Next()
	if iter.Done() {
		t.Errorf("Expected at least two cells in the index")
	}

	// Load the second cell and check that the cached values change
	data.loadCell(index, iter.CellID(), iter.cell)
	center1 := data.center()
	cell1 := data.cell()

	if cell0 == cell1 {
		t.Error("Expected different cells")
	}
	if center0 == center1 {
		t.Error("Expected different cell centers")
	}

	// Load the same cell again, nothing should change
	data.loadCell(index, iter.CellID(), iter.cell)
	center2 := data.center()
	cell2 := data.cell()

	if cell1 != cell2 {
		t.Error("Expected same cells when reloading the same cell")
	}
	if center1 != center2 {
		t.Error("Expected same cell centers when reloading the same cell")
	}
}

// TODO(rsned): Test shapeContains
