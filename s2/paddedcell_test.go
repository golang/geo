/*
Copyright 2016 Google Inc. All rights reserved.

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

	"github.com/golang/geo/r2"
)

func TestPaddedCellMethods(t *testing.T) {
	for i := 0; i < 1000; i++ {
		cid := randomCellID()
		padding := math.Pow(1e-15, randomFloat64())
		cell := CellFromCellID(cid)
		pCell := PaddedCellFromCellID(cid, padding)

		if cell.id != pCell.id {
			t.Errorf("%v.id = %v, want %v", pCell, pCell.id, cell.id)
		}
		if cell.id.Level() != pCell.Level() {
			t.Errorf("%v.Level() = %v, want %v", pCell, pCell.Level(), cell.id.Level())
		}

		if padding != pCell.Padding() {
			t.Errorf("%v.Padding() = %v, want %v", pCell, pCell.Padding(), padding)
		}

		// TODO(roberts): When Cell has BoundUV, uncomment this.
		//if cell.BoundUV().Expanded(padding) != pCell.BoundUV() {
		//	t.Errorf("%v.BoundUV() = %v, want %v", pCell, pCell.BoundUV(), cell.BoundUV().Expanded(padding))
		//}

		r := r2.RectFromPoints(cell.id.centerUV()).ExpandedByMargin(padding)
		if r != pCell.Middle() {
			t.Errorf("%v.Middle() = %v, want %v", pCell, pCell.Middle(), r)
		}

		if cell.id.Point() != pCell.Center() {
			t.Errorf("%v.Center() = %v, want %v", pCell, pCell.Center(), cell.id.Point())
		}
		// TODO(roberts): When Cell has Children/Subdivide method,
		// add the remainder of this test section.
	}
}

func TestEntryExitVertices(t *testing.T) {
	for i := 0; i < 1000; i++ {
		id := randomCellID()
		unpadded := PaddedCellFromCellID(id, 0)
		padded := PaddedCellFromCellID(id, 0.5)

		// Check that entry/exit vertices do not depend on padding.
		if unpadded.EntryVertex() != padded.EntryVertex() {
			t.Errorf("entry vertex should not depend on padding; %v != %v", unpadded.EntryVertex(), padded.EntryVertex())
		}

		if unpadded.ExitVertex() != padded.ExitVertex() {
			t.Errorf("exit vertex should not depend on padding; %v != %v", unpadded.ExitVertex(), padded.ExitVertex())
		}

		// Check that the exit vertex of one cell is the same as the entry vertex
		// of the immediately following cell. This also tests wrapping from the
		// end to the start of the CellID curve with high probability.
		if got := PaddedCellFromCellID(id.NextWrap(), 0).EntryVertex(); unpadded.ExitVertex() != got {
			t.Errorf("PaddedCellFromCellID(%v.NextWrap(), 0).EntryVertex() = %v, want %v", id, got, unpadded.ExitVertex())
		}

		// Check that the entry vertex of a cell is the same as the entry vertex
		// of its first child, and similarly for the exit vertex.
		if id.IsLeaf() {
			continue
		}
		if got := PaddedCellFromCellID(id.Children()[0], 0).EntryVertex(); unpadded.EntryVertex() != got {
			t.Errorf("PaddedCellFromCellID(%v.Children()[0], 0).EntryVertex() = %v, want %v", id, got, unpadded.EntryVertex())
		}
		if got := PaddedCellFromCellID(id.Children()[3], 0).ExitVertex(); unpadded.ExitVertex() != got {
			t.Errorf("PaddedCellFromCellID(%v.Children()[3], 0).ExitVertex() = %v, want %v", id, got, unpadded.ExitVertex())
		}
	}
}
