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

import "sort"

// PointData holds a Point and its associated data.
type PointData[Data comparable] struct {
	point Point
	data  Data
}

// Point returns the point.
func (pd PointData[Data]) Point() Point { return pd.point }

// Data returns the associated data.
func (pd PointData[Data]) Data() Data { return pd.data }

// pointIndexEntry is a single entry in the sorted PointIndex.
type pointIndexEntry[Data comparable] struct {
	id        CellID
	pointData PointData[Data]
}

// PointIndex maintains an index of points sorted by leaf CellID. Each point
// can optionally store auxiliary data such as an integer or pointer. This can
// be used to map results back to client data structures.
//
// The index supports adding or removing points dynamically, and provides a
// seekable iterator interface for navigating the index.
//
// You can use this class in conjunction with ClosestPointQuery to find the
// closest index points to a given query point. For example:
//
//	index := &PointIndex[int]{}
//	for i, p := range indexPoints {
//	    index.Add(p, i)
//	}
//
// You can also access the index directly using the iterator interface. For
// example, here is how to iterate through all the points in a given CellID
// target:
//
//	it := NewPointIndexIterator(index)
//	for it.Seek(target.RangeMin()); !it.Done() && it.CellID() <= target.RangeMax(); it.Next() {
//	    DoSomething(it.CellID(), it.Point(), it.Data())
//	}
//
// Points can be added or removed from the index at any time by calling Add()
// or Remove(). However when the index is modified, any existing iterator's
// position may refer to the wrong entry; create a new iterator to resume
// traversal safely.
//
// Note: Add and Remove maintain sorted order by shifting elements, so they
// run in O(n) time. This index is suitable for building once and querying
// many times, or for small dynamic datasets.
type PointIndex[Data comparable] struct {
	entries []pointIndexEntry[Data]
}

// NumPoints returns the number of points in the index.
func (p *PointIndex[Data]) NumPoints() int { return len(p.entries) }

// Add adds the given point with associated data to the index. Invalidates all iterators.
func (p *PointIndex[Data]) Add(point Point, data Data) {
	id := cellIDFromPoint(point)
	pos := sort.Search(len(p.entries), func(i int) bool {
		return p.entries[i].id >= id
	})
	entry := pointIndexEntry[Data]{id: id, pointData: PointData[Data]{point: point, data: data}}
	p.entries = append(p.entries, pointIndexEntry[Data]{})
	copy(p.entries[pos+1:], p.entries[pos:])
	p.entries[pos] = entry
}

// Remove removes the given point and data from the index. Returns false if the
// given point was not present. Invalidates all iterators.
func (p *PointIndex[Data]) Remove(point Point, data Data) bool {
	id := cellIDFromPoint(point)
	pd := PointData[Data]{point: point, data: data}
	pos := sort.Search(len(p.entries), func(i int) bool {
		return p.entries[i].id >= id
	})
	for pos < len(p.entries) && p.entries[pos].id == id {
		if p.entries[pos].pointData == pd {
			p.entries = append(p.entries[:pos], p.entries[pos+1:]...)
			return true
		}
		pos++
	}
	return false
}

// Clear resets the index to its original empty state. Invalidates all iterators.
func (p *PointIndex[Data]) Clear() {
	p.entries = nil
}

// PointIndexIterator is a seekable iterator for a PointIndex.
//
// The iterator holds a pointer to the index, so the underlying data is always
// live. However, mutations to the index (Add/Remove) may shift entries and
// leave the iterator's position pointing at the wrong entry. Create a new
// iterator after any mutation.
type PointIndexIterator[Data comparable] struct {
	index    *PointIndex[Data]
	position int
}

// NewPointIndexIterator creates a new iterator for the given PointIndex.
// If the index is non-empty, the iterator is positioned at the first entry.
func NewPointIndexIterator[Data comparable](index *PointIndex[Data]) *PointIndexIterator[Data] {
	return &PointIndexIterator[Data]{index: index}
}

// CellID returns the CellID for the current index entry.
// Requires: !Done()
func (it *PointIndexIterator[Data]) CellID() CellID {
	return it.index.entries[it.position].id
}

// Point returns the point associated with the current index entry.
// Requires: !Done()
func (it *PointIndexIterator[Data]) Point() Point {
	return it.index.entries[it.position].pointData.point
}

// Data returns the data associated with the current index entry.
// Requires: !Done()
func (it *PointIndexIterator[Data]) Data() Data {
	return it.index.entries[it.position].pointData.data
}

// PointData returns the (Point, Data) pair for the current index entry.
// Requires: !Done()
func (it *PointIndexIterator[Data]) PointData() PointData[Data] {
	return it.index.entries[it.position].pointData
}

// Done reports if the iterator is positioned past the last index entry.
func (it *PointIndexIterator[Data]) Done() bool {
	return it.position >= len(it.index.entries)
}

// Begin positions the iterator at the first index entry (if any).
func (it *PointIndexIterator[Data]) Begin() {
	it.position = 0
}

// Finish positions the iterator so that Done() is true.
func (it *PointIndexIterator[Data]) Finish() {
	it.position = len(it.index.entries)
}

// Next advances the iterator to the next index entry.
// Requires: !Done()
func (it *PointIndexIterator[Data]) Next() {
	it.position++
}

// Prev positions the iterator at the previous entry and reports whether the
// iterator was not already positioned at the beginning.
func (it *PointIndexIterator[Data]) Prev() bool {
	if it.position == 0 {
		return false
	}
	it.position--
	return true
}

// Seek positions the iterator at the first entry with CellID() >= target, or
// at the end of the index if no such entry exists.
func (it *PointIndexIterator[Data]) Seek(target CellID) {
	it.position = sort.Search(len(it.index.entries), func(i int) bool {
		return it.index.entries[i].id >= target
	})
}

// LocatePoint positions the iterator at the entry for the cell containing the
// given point. Returns true if such an entry exists.
func (it *PointIndexIterator[Data]) LocatePoint(target Point) bool {
	id := cellIDFromPoint(target)
	it.Seek(id)
	if !it.Done() && it.CellID().RangeMin() <= id {
		return true
	}
	if it.Prev() && it.CellID().RangeMax() >= id {
		return true
	}
	return false
}

// LocateCellID positions the iterator given the target CellID. Let T be the
// target CellID. If T is contained by some index cell I (including equality),
// the iterator is positioned at I and Indexed is returned. Otherwise if T
// contains one or more (smaller) index cells, the iterator is positioned at
// the first such cell and Subdivided is returned. Otherwise Disjoint is
// returned and the iterator position is unspecified.
func (it *PointIndexIterator[Data]) LocateCellID(target CellID) CellRelation {
	it.Seek(target.RangeMin())
	if !it.Done() {
		if it.CellID() >= target && it.CellID().RangeMin() <= target {
			return Indexed
		}
		if it.CellID() <= target.RangeMax() {
			return Subdivided
		}
	}
	if it.Prev() && it.CellID().RangeMax() >= target {
		return Indexed
	}
	return Disjoint
}
