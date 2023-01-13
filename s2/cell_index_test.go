// Copyright 2020 Google Inc. All rights reserved.
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
)

func cellIndexQuadraticValidate(t *testing.T, desc string, index *CellIndex, contents []cellIndexNode) {
	// Verifies that the index computes the correct set of (cell_id, label) pairs
	// for every possible leaf cell.  The running time of this function is
	// quadratic in the size of the index.
	index.Build()
	verifyCellIndexCellIterator(t, desc, index)
	verifyCellIndexRangeIterators(t, desc, index)
	verifyCellIndexContents(t, desc, index)
}

// less reports whether this node is less than the other.
func (c cellIndexNode) less(other cellIndexNode) bool {
	if c.cellID != other.cellID {
		return c.cellID < other.cellID
	}
	if c.label != other.label {
		return c.label < other.label
	}
	return c.parent < other.parent
}

func cellIndexNodesEqual(a, b []cellIndexNode) bool {
	sort.Slice(a, func(i, j int) bool {
		return a[i].less(a[j])
	})
	sort.Slice(b, func(i, j int) bool {
		return b[i].less(b[j])
	})
	return reflect.DeepEqual(a, b)
}

// copyCellIndexNodes creates a copy of the nodes so that sorting and other tests
// don't alter the instance in a given CellIndex.
func copyCellIndexNodes(in []cellIndexNode) []cellIndexNode {
	out := make([]cellIndexNode, len(in))
	copy(out, in)
	return out
}

func verifyCellIndexCellIterator(t *testing.T, desc string, index *CellIndex) {
	// TODO(roberts): Once the plain iterator is implemented, add this check.
	/*
		var actual []cellIndexNode
		iter := NewCellIndexIterator(index)
		for iter.Begin(); !iter.Done(); iter.Next() {
			actual = append(actual, cellIndexNode{iter.StartID(), iter.Label())
		}

		want := copyCellIndexNodes(index.cellTree)
		if !cellIndexNodesEqual(actual, want) {
			t.Errorf("%s: cellIndexNodes not equal but should be.  %v != %v", desc, actual, want)
		}
	*/
}

func verifyCellIndexRangeIterators(t *testing.T, desc string, index *CellIndex) {
	// tests Finish(), which is not otherwise tested below.
	it := NewCellIndexRangeIterator(index)
	it.Begin()
	it.Finish()
	if !it.Done() {
		t.Errorf("%s: positioning iterator to finished should be done, but was not", desc)
	}

	// And also for non-empty ranges.
	nonEmpty := NewCellIndexNonEmptyRangeIterator(index)
	nonEmpty.Begin()
	nonEmpty.Finish()
	if !nonEmpty.Done() {
		t.Errorf("%s: positioning non-empty iterator to finished should be done, but was not", desc)
	}

	// Iterate through all the ranges in the index.  We simultaneously iterate
	// through the non-empty ranges and check that the correct ranges are found.
	prevStart := CellID(0)
	nonEmptyPrevStart := CellID(0)

	it.Begin()
	nonEmpty.Begin()
	for ; !it.Done(); it.Next() {
		// Check that seeking in the current range takes us to this range.
		it2 := NewCellIndexRangeIterator(index)
		start := it.StartID()
		it2.Seek(it.StartID())
		if start != it2.StartID() {
			t.Errorf("%s: id: %v. id2 start: %v\nit: %+v\nit2: %+v", desc, start, it2.StartID(), it, it2)
		}
		it2.Seek(it.LimitID().Prev())
		if start != it2.StartID() {
			t.Errorf("%s: it2.Seek(%v) = %v, want %v", desc, it.LimitID().Prev(), it2.StartID(), start)
		}

		// And also for non-empty ranges.
		nonEmpty2 := NewCellIndexNonEmptyRangeIterator(index)
		nonEmptyStart := nonEmpty.StartID()
		nonEmpty2.Seek(it.StartID())
		if nonEmptyStart != nonEmpty2.StartID() {
			t.Errorf("%s: nonEmpty2.StartID() = %v, want %v", desc, nonEmpty2.StartID(), nonEmptyStart)
		}
		nonEmpty2.Seek(it.LimitID().Prev())
		if nonEmptyStart != nonEmpty2.StartID() {
			t.Errorf("%s: nonEmpty2.StartID() = %v, want %v", desc, nonEmpty2.StartID(), nonEmptyStart)
		}

		// Test Prev() and Next().
		if it2.Prev() {
			if prevStart != it2.StartID() {
				t.Errorf("%s: it2.StartID() = %v, want %v", desc, it2.StartID(), prevStart)
			}
			it2.Next()
			if start != it2.StartID() {
				t.Errorf("%s: it2.StartID() = %v, want %v", desc, it2.StartID(), start)
			}
		} else {
			if start != it2.StartID() {
				t.Errorf("%s: it2.StartID() = %v, want %v", desc, it2.StartID(), start)
			}
			if 0 != prevStart {
				t.Errorf("%s: prevStart = %v, want %v", desc, prevStart, 0)
			}
		}

		// And also for non-empty ranges.
		if nonEmpty2.Prev() {
			if nonEmptyPrevStart != nonEmpty2.StartID() {
				t.Errorf("%s: nonEmpty2.StartID() = %v, want %v", desc, nonEmpty2.StartID(), nonEmptyPrevStart)
			}
			nonEmpty2.Next()
			if nonEmptyStart != nonEmpty2.StartID() {
				t.Errorf("%s: nonEmpty2.StartID() = %v, want %v", desc, nonEmpty2.StartID(), nonEmptyStart)
			}
		} else {
			if nonEmptyStart != nonEmpty2.StartID() {
				t.Errorf("%s: nonEmpty2.StartID() = %v, want %v", desc, nonEmpty2.StartID(), nonEmptyStart)
			}
			if nonEmptyPrevStart != 0 {
				t.Errorf("%s: nonEmptyPrevStart = %v, want 0", desc, nonEmptyPrevStart)
			}
		}

		// Keep the non-empty iterator synchronized with the regular one.
		if !it.IsEmpty() {
			if it.StartID() != nonEmpty.StartID() {
				t.Errorf("%s: it.StartID = %v, want %v", desc, it.StartID(), nonEmpty.StartID())
			}
			if it.LimitID() != nonEmpty.LimitID() {
				t.Errorf("%s: it.LimitID = %v, want %v", desc, it.LimitID(), nonEmpty.LimitID())
			}
			if nonEmpty.Done() {
				t.Errorf("%s: nonEmpty iterator should not be done but was", desc)
			}
			nonEmptyPrevStart = nonEmptyStart
			nonEmpty.Next()
		}
		prevStart = start
	}

	// Verify that the NonEmptyRangeIterator is also finished.
	if !nonEmpty.Done() {
		t.Errorf("%s: non empty iterator should have also finished", desc)
	}
}

// verifies that RangeIterator and ContentsIterator can be used to determine
// the exact set of (s2cell_id, label) pairs that contain any leaf cell.
func verifyCellIndexContents(t *testing.T, desc string, index *CellIndex) {
	// "minCellID" is the first CellID that has not been validated yet.
	minCellID := CellIDFromFace(0).ChildBeginAtLevel(maxLevel)
	r := NewCellIndexRangeIterator(index)
	for r.Begin(); !r.Done(); r.Next() {
		if minCellID != r.StartID() {
			t.Errorf("%s: minCellID should match the previous ending cellID. got %v, want %v", desc, r.StartID(), minCellID)
		}
		if minCellID >= r.LimitID() {
			t.Errorf("%s: minCellID should be >= the end of the current range. got %v, want %v", desc, r.LimitID(), minCellID)
		}
		if !r.LimitID().IsLeaf() {
			t.Errorf("%s: ending range cell ID should not be a leaf, but was", desc)
		}

		minCellID = r.LimitID()

		// Build a list of expected (CellID, label) for this range.
		var expected []cellIndexNode
		for _, x := range index.cellTree {
			// The cell contains the entire range.
			if x.cellID.RangeMin() <= r.StartID() &&
				x.cellID.RangeMax().Next() >= r.LimitID() {
				expected = append(expected, x)
			} else {
				// Verify that the cell does not intersect the range.
				if x.cellID.RangeMin() <= r.LimitID().Prev() &&
					x.cellID.RangeMax() >= r.StartID() {
					t.Errorf("%s: CellID does not interect the current range: %v <= %v && %v >= %v", desc, x.cellID.RangeMin(), r.LimitID().Prev(), x.cellID.RangeMax(), r.StartID())
				}
			}
		}
		var actual []cellIndexNode
		cIter := NewCellIndexContentsIterator(index)
		for cIter.StartUnion(r); !cIter.Done(); cIter.Next() {
			actual = append(actual, cIter.node)
		}

		if !cellIndexNodesEqual(expected, actual) {
			t.Errorf("%s: comparing contents iterator contents to this range: got %+v, want %+v", desc, actual, expected)
		}
	}

	if CellIDFromFace(5).ChildEndAtLevel(maxLevel) != minCellID {
		t.Errorf("%s: the final cell should be the sentinel value, got %v", desc, minCellID)
	}
}

func TestCellIndex(t *testing.T) {
	type cellIndexTestInput struct {
		cellID string
		label  int32
	}
	tests := []struct {
		label string
		have  []cellIndexTestInput
	}{
		{
			label: "Empty",
		},
		{
			label: "One face cell",
			have: []cellIndexTestInput{
				{"0/", 0},
			},
		},

		{
			label: "One Leaf Cell",
			have: []cellIndexTestInput{
				{"1/012301230123012301230123012301", 12},
			},
		},
		{
			label: "Duplicate Values",
			have: []cellIndexTestInput{
				{"0/", 0},
				{"0/", 0},
				{"0/", 1},
				{"0/", 17},
			},
		},
		{
			label: "Disjoint Cells",
			have: []cellIndexTestInput{
				{"0/", 0},
				{"3/", 0},
			},
		},
		{
			// Tests nested cells, including cases where several cells have the same
			// RangeMin or RangeMax and with randomly ordered labels.
			label: "Nested Cells",
			have: []cellIndexTestInput{
				{"1/", 3},
				{"1/0", 15},
				{"1/000", 9},
				{"1/00000", 11},
				{"1/012", 6},
				{"1/01212", 5},
				{"1/312", 17},
				{"1/31200", 4},
				{"1/3120000", 10},
				{"1/333", 20},
				{"1/333333", 18},
				{"5/", 3},
				{"5/3", 31},
				{"5/3333", 27},
			},
		},
		{
			// Checks that the contents iterator stops reporting values
			// once it reaches a node of the cell tree that was visited
			// by the previous call to Begin().
			label: "Contents Iterator Suppresses Duplicates",
			have: []cellIndexTestInput{
				{"2/1", 1},
				{"2/1", 2},
				{"2/10", 3},
				{"2/100", 4},
				{"2/102", 5},
				{"2/1023", 6},
				{"2/31", 7},
				{"2/313", 8},
				{"2/3132", 9},
				{"3/1", 10},
				{"3/12", 11},
				{"3/13", 12},
			},
		},
	}

	for _, test := range tests {
		index := &CellIndex{}
		for _, v := range test.have {
			index.Add(cellIDFromString(v.cellID), v.label)
		}
		cellIndexQuadraticValidate(t, test.label, index, nil)
	}
}

func TestCellIndexRandomCellUnions(t *testing.T) {
	// Construct cell unions from random CellIDs at random levels. Note that
	// because the cell level is chosen uniformly, there is a very high
	// likelihood that the cell unions will overlap.
	index := &CellIndex{}
	for i := int32(0); i < 100; i++ {
		index.AddCellUnion(randomCellUnion(10), i)
	}
	cellIndexQuadraticValidate(t, "Random Cell Unions", index, nil)
}

func TestCellIndexIntersectionOptimization(t *testing.T) {
	type cellIndexTestInput struct {
		cellID string
		label  int32
	}
	tests := []struct {
		label string
		have  []cellIndexTestInput
	}{
		{
			// Tests various corner cases for the binary search optimization in
			// VisitIntersectingCells.
			label: "Intersection Optimization",
			have: []cellIndexTestInput{
				{"1/001", 1},
				{"1/333", 2},
				{"2/00", 3},
				{"2/0232", 4},
			},
		},
	}

	for _, test := range tests {
		index := &CellIndex{}
		for _, v := range test.have {
			index.Add(cellIDFromString(v.cellID), v.label)
		}
		index.Build()
		checkIntersection(t, test.label, makeCellUnion("1/010", "1/3"), index)
		checkIntersection(t, test.label, makeCellUnion("2/010", "2/011", "2/02"), index)
	}
}

func TestCellIndexIntersectionRandomCellUnions(t *testing.T) {
	// Construct cell unions from random CellIDs at random levels. Note that
	// because the cell level is chosen uniformly, there is a very high
	// likelihood that the cell unions will overlap.
	index := &CellIndex{}
	for i := int32(0); i < 100; i++ {
		index.AddCellUnion(randomCellUnion(10), i)
	}
	index.Build()
	for i := 0; i < 200; i++ {
		checkIntersection(t, "", randomCellUnion(10), index)
	}
}

func TestCellIndexIntersectionSemiRandomCellUnions(t *testing.T) {
	for i := 0; i < 200; i++ {
		index := &CellIndex{}
		id := cellIDFromString("1/0123012301230123")
		var target CellUnion
		for j := 0; j < 100; j++ {
			switch {
			case oneIn(10):
				index.Add(id, int32(j))
			case oneIn(4):
				target = append(target, id)
			case oneIn(2):
				id = id.NextWrap()
			case oneIn(6) && !id.isFace():
				id = id.immediateParent()
			case oneIn(6) && !id.IsLeaf():
				id = id.ChildBegin()
			}
		}
		target.Normalize()
		index.Build()
		checkIntersection(t, "", target, index)
	}
}

func checkIntersection(t *testing.T, desc string, target CellUnion, index *CellIndex) {
	var expected, actual []int32
	for it := NewCellIndexIterator(index); !it.Done(); it.Next() {
		if target.IntersectsCellID(it.CellID()) {
			expected = append(expected, it.Label())
		}
	}

	index.VisitIntersectingCells(target, func(cellID CellID, label int32) bool {
		actual = append(actual, label)
		return true
	})

	if !labelsEqual(actual, expected) {
		t.Errorf("%s: labels not equal but should be.  %v != %v", desc, actual, expected)
	}
}

func labelsEqual(a, b []int32) bool {
	sort.Slice(a, func(i, j int) bool {
		return a[i] < a[j]
	})
	sort.Slice(b, func(i, j int) bool {
		return b[i] < b[j]
	})
	return reflect.DeepEqual(a, b)
}

// TODO(roberts): Differences from C++
//
// Add remainder of TestCellIndexContentsIteratorSuppressesDuplicates
//
// additional Iterator related parts
