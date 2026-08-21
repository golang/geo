// Copyright 2015 Google Inc. All rights reserved.
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

// pointIndexTest is a test helper that tracks both the index and expected contents.
type pointIndexTest struct {
	t        *testing.T
	index    *PointIndex[int]
	contents map[PointData[int]]int // PointData -> occurrence count
}

func newPointIndexTest(t *testing.T) *pointIndexTest {
	return &pointIndexTest{
		t:        t,
		index:    &PointIndex[int]{},
		contents: make(map[PointData[int]]int),
	}
}

func (pt *pointIndexTest) add(point Point, data int) {
	pt.index.Add(point, data)
	pt.contents[PointData[int]{point: point, data: data}]++
}

func (pt *pointIndexTest) remove(point Point, data int) {
	pd := PointData[int]{point: point, data: data}
	pt.contents[pd]--
	if pt.contents[pd] == 0 {
		delete(pt.contents, pd)
	}
	if !pt.index.Remove(point, data) {
		pt.t.Errorf("Remove(%v, %v) returned false, expected true", point, data)
	}
}

func (pt *pointIndexTest) verify() {
	pt.verifyContents()
	pt.verifyIteratorMethods()
}

func (pt *pointIndexTest) verifyContents() {
	remaining := make(map[PointData[int]]int)
	for k, v := range pt.contents {
		remaining[k] = v
	}
	for it := NewPointIndexIterator(pt.index); !it.Done(); it.Next() {
		pd := it.PointData()
		if got := pd.Point(); got != it.Point() {
			pt.t.Errorf("PointData.Point() = %v, want %v", got, it.Point())
		}
		if got := pd.Data(); got != it.Data() {
			pt.t.Errorf("PointData.Data() = %v, want %v", got, it.Data())
		}
		if remaining[pd] <= 0 {
			pt.t.Errorf("point_data %v found in index but not in expected contents", pd)
			continue
		}
		remaining[pd]--
		if remaining[pd] == 0 {
			delete(remaining, pd)
		}
	}
	if len(remaining) > 0 {
		pt.t.Errorf("expected contents not found in index: %v", remaining)
	}
}

func (pt *pointIndexTest) verifyIteratorMethods() {
	it := NewPointIndexIterator(pt.index)
	if it.Prev() {
		pt.t.Error("Prev() returned true on freshly created iterator at position 0")
	}
	it.Finish()
	if !it.Done() {
		pt.t.Error("Done() returned false after Finish()")
	}

	var prevCellID CellID
	minCellID := CellIDFromFace(0).ChildBeginAtLevel(MaxLevel)

	for it.Begin(); !it.Done(); it.Next() {
		cellID := it.CellID()

		if got := cellIDFromPoint(it.Point()); got != cellID {
			pt.t.Errorf("cellIDFromPoint(it.Point()) = %v, want %v", got, cellID)
		}
		if cellID < prevCellID {
			pt.t.Errorf("iterator not in sorted order: %v < %v", cellID, prevCellID)
		}

		it2 := *it
		if cellID == prevCellID {
			it2.Seek(cellID)
		}

		// Verify that seeking to any skipped leaf cell lands at cellID.
		if cellID > prevCellID {
			for _, skipped := range CellUnionFromRange(minCellID, cellID) {
				it2.Seek(skipped)
				if it2.Done() || it2.CellID() != cellID {
					pt.t.Errorf("Seek(%v): got %v, want %v", skipped, it2.CellID(), cellID)
				}
			}
		}

		// Test Prev, Next, and Seek.
		if prevCellID.IsValid() {
			it2 = *it
			it2.Refresh() // decouple cursor from original before cross-boundary Prev
			if !it2.Prev() {
				pt.t.Error("Prev() returned false, expected true")
			}
			if it2.CellID() != prevCellID {
				pt.t.Errorf("after Prev(), CellID() = %v, want %v", it2.CellID(), prevCellID)
			}
			it2.Next()
			if it2.CellID() != cellID {
				pt.t.Errorf("after Next(), CellID() = %v, want %v", it2.CellID(), cellID)
			}
			it2.Seek(prevCellID)
			if it2.CellID() != prevCellID {
				pt.t.Errorf("Seek(%v): CellID() = %v, want %v", prevCellID, it2.CellID(), prevCellID)
			}
		}

		prevCellID = cellID
		minCellID = cellID.Next()
	}
}

func TestPointIndexNoPoints(t *testing.T) {
	pt := newPointIndexTest(t)
	pt.verify()
}

func TestPointIndexDuplicatePoints(t *testing.T) {
	pt := newPointIndexTest(t)
	p := PointFromCoords(1, 0, 0)
	for range 10 {
		pt.add(p, 123)
	}
	pt.verify()
	for range 5 {
		pt.remove(p, 123)
	}
	pt.verify()

	// Remove with wrong data value — point is present but data does not match.
	if pt.index.Remove(p, 456) {
		t.Error("Remove(p, 456) = true, want false: data 456 was never added")
	}
	// Remove with a point not in the index at all.
	absent := PointFromCoords(0, 1, 0)
	if pt.index.Remove(absent, 123) {
		t.Error("Remove(absent, 123) = true, want false: point was never added")
	}
}

func TestPointIndexRandomPoints(t *testing.T) {
	pt := newPointIndexTest(t)
	for range 100 {
		pt.add(randomPoint(), randomUniformInt(100))
	}
	pt.verify()

	// Remove some points via iterator traversal to a random leaf cell.
	for range 10 {
		it := NewPointIndexIterator(pt.index)
		found := false
		for range 100 {
			it.Seek(randomCellIDForLevel(MaxLevel))
			if !it.Done() {
				found = true
				break
			}
		}
		if !found {
			t.Fatal("failed to find a non-empty position after 100 seeks")
		}
		pt.remove(it.Point(), it.Data())
		pt.verify()
	}
}

func TestPointIndexNumPoints(t *testing.T) {
	index := &PointIndex[int]{}
	if got := index.NumPoints(); got != 0 {
		t.Errorf("NumPoints() = %d, want 0 for empty index", got)
	}
	p := PointFromCoords(1, 0, 0)
	for i := range 5 {
		index.Add(p, i)
	}
	if got := index.NumPoints(); got != 5 {
		t.Errorf("NumPoints() = %d, want 5 after 5 adds", got)
	}
	index.Remove(p, 2)
	if got := index.NumPoints(); got != 4 {
		t.Errorf("NumPoints() = %d, want 4 after one removal", got)
	}
}

func TestPointIndexClear(t *testing.T) {
	index := &PointIndex[int]{}
	for i := range 10 {
		index.Add(randomPoint(), i)
	}
	index.Clear()
	if got := index.NumPoints(); got != 0 {
		t.Errorf("NumPoints() = %d, want 0 after Clear", got)
	}
	it := NewPointIndexIterator(index)
	if !it.Done() {
		t.Error("iterator not Done() immediately after Clear")
	}
	// Verify the index is usable after clearing.
	p := PointFromCoords(0, 1, 0)
	index.Add(p, 99)
	if got := index.NumPoints(); got != 1 {
		t.Errorf("NumPoints() = %d, want 1 after Add following Clear", got)
	}
}

func TestPointIndexLocatePoint(t *testing.T) {
	index := &PointIndex[int]{}
	// Three points on distinct S2 faces to guarantee distinct leaf CellIDs.
	points := []Point{
		PointFromCoords(1, 0, 0),
		PointFromCoords(0, 1, 0),
		PointFromCoords(0, 0, 1),
	}
	for i, p := range points {
		index.Add(p, i)
	}

	it := NewPointIndexIterator(index)
	for i, p := range points {
		if !it.LocatePoint(p) {
			t.Errorf("LocatePoint(points[%d]) = false, want true", i)
			continue
		}
		if got := it.Data(); got != i {
			t.Errorf("LocatePoint(points[%d]): Data() = %d, want %d", i, got, i)
		}
	}

	absent := PointFromCoords(1, 1, 1)
	if it.LocatePoint(absent) {
		t.Errorf("LocatePoint(%v) = true, want false for absent point", absent)
	}
}

func TestPointIndexLocateCellID(t *testing.T) {
	index := &PointIndex[int]{}
	p := PointFromCoords(1, 0, 0)
	index.Add(p, 42)
	leafID := cellIDFromPoint(p)

	it := NewPointIndexIterator(index)

	// Exact leaf cell in the index → Indexed, iterator at that cell.
	if got := it.LocateCellID(leafID); got != Indexed {
		t.Errorf("LocateCellID(leaf) = %v, want Indexed", got)
	}
	if it.CellID() != leafID {
		t.Errorf("after LocateCellID(leaf): CellID() = %v, want %v", it.CellID(), leafID)
	}

	// Parent cell containing the leaf → Subdivided, iterator at the leaf.
	parent := leafID.Parent(MaxLevel - 1)
	if got := it.LocateCellID(parent); got != Subdivided {
		t.Errorf("LocateCellID(parent) = %v, want Subdivided", got)
	}
	if it.CellID() != leafID {
		t.Errorf("after LocateCellID(parent): CellID() = %v, want %v", it.CellID(), leafID)
	}

	// Level-0 cell on a different face → Disjoint.
	otherFace := CellIDFromFace(2)
	if got := it.LocateCellID(otherFace); got != Disjoint {
		t.Errorf("LocateCellID(otherFace) = %v, want Disjoint", got)
	}
}
