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

import "github.com/tidwall/btree"

// PointData holds a Point and its associated data.
type PointData[Data comparable] struct {
	point Point
	data  Data
}

// Point returns the point.
func (pd PointData[Data]) Point() Point { return pd.point }

// Data returns the associated data.
func (pd PointData[Data]) Data() Data { return pd.data }

// PointIndex maintains an index of points sorted by leaf CellID using a B-tree.
// Each point can optionally store auxiliary data such as an integer or pointer.
// This can be used to map results back to client data structures.
//
// The index supports adding or removing points dynamically and provides a
// seekable iterator interface for navigating the index.
//
// You can use this class in conjunction with ClosestPointQuery to find the
// closest index points to a given query point. For example,
//
//		index := &PointIndex[int]{}
//		for i, p := range indexPoints {
//		    index.Add(p, i)
//		}
//	 TODO(fmeurisse): Implement ClosestPointQuery integration and update example.
//	  S2ClosestPointQuery<int> query(&index);
//	  query.mutable_options()->set_max_results(5);
//	  for (const S2Point& target_point : target_points) {
//	    S2ClosestPointQueryPointTarget target(target_point);
//	    for (const auto& result : query.FindClosestPoints(&target)) {
//	      // The Result class contains the following methods:
//	      //   distance() is the distance to the target.
//	      //   point() is the indexed point.
//	      //   data() is the auxiliary data.
//	      DoSomething(target_point, result);
//	    }
//	  }
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
// or Remove(). However when the index is modified, you must call Init() on
// each iterator before using it again (or simply create a new iterator):
//
//	index.Add(newPoint, 123456)
//	it.Init(index)
//	it.Seek(target.RangeMin())
//
// PointIndex is not safe for concurrent use without external synchronization.
type PointIndex[Data comparable] struct {
	// tree maps each leaf CellID to the slice of PointData values at that cell.
	// Multiple points at the same CellID are stored together in one entry.
	tree      btree.Map[CellID, []PointData[Data]]
	numPoints int
}

// NumPoints returns the number of points in the index.
func (p *PointIndex[Data]) NumPoints() int { return p.numPoints }

// Add adds the given point with associated data to the index. Invalidates all iterators.
func (p *PointIndex[Data]) Add(point Point, data Data) {
	id := cellIDFromPoint(point)
	slice, _ := p.tree.Get(id)
	p.tree.Set(id, append(slice, PointData[Data]{point: point, data: data}))
	p.numPoints++
}

// Remove removes one occurrence of the given point and data from the index.
// Returns false if no matching entry was found. Invalidates all iterators.
func (p *PointIndex[Data]) Remove(point Point, data Data) bool {
	id := cellIDFromPoint(point)
	pd := PointData[Data]{point: point, data: data}
	slice, found := p.tree.Get(id)
	if !found {
		return false
	}
	for i, existing := range slice {
		if existing == pd {
			slice = append(slice[:i], slice[i+1:]...)
			if len(slice) == 0 {
				p.tree.Delete(id)
			} else {
				p.tree.Set(id, slice)
			}
			p.numPoints--
			return true
		}
	}
	return false
}

// Clear resets the index to its original empty state. Invalidates all iterators.
func (p *PointIndex[Data]) Clear() {
	p.tree.Clear()
	p.numPoints = 0
}

// PointIndexIterator is a seekable iterator for a PointIndex.
//
// Points at the same CellID are yielded consecutively. The iterator is safe
// to copy for save/restore of position:
//
//	it2 := *it
//
// After a copy, call Init(), Begin(), or Seek() on the copy before calling
// Next() or Prev() across a CellID boundary — these methods assign a fresh
// internal iterator, decoupling the copy from the original. Calling Next() or
// Prev() across a CellID boundary on a raw copy without a prior repositioning
// call is undefined behaviour.
//
// After any Add or Remove call, call Init() to make the iterator valid again,
// or create a new iterator.
type PointIndexIterator[Data comparable] struct {
	index        *PointIndex[Data]
	iter         btree.MapIter[CellID, []PointData[Data]] // live cursor on map keys
	currentID    CellID
	currentSlice []PointData[Data] // points slice of the current map entry
	sliceIdx     int               // position within currentSlice
	valid        bool
	atEnd        bool // true when positioned logically past the last entry
}

// NewPointIndexIterator creates a new iterator for the given PointIndex,
// positioned at the first entry (if any).
func NewPointIndexIterator[Data comparable](index *PointIndex[Data]) *PointIndexIterator[Data] {
	var it PointIndexIterator[Data]
	it.Init(index)
	return &it
}

// Init (re)initializes the iterator for the given index, positioning it at
// the first entry if any. This may be called multiple times, e.g. to make an
// iterator valid again after the index is modified.
func (it *PointIndexIterator[Data]) Init(index *PointIndex[Data]) {
	it.index = index
	it.iter = index.tree.Iter() // fresh iter: independent backing array
	it.valid = it.iter.First()
	it.atEnd = false
	it.sliceIdx = 0
	if it.valid {
		it.currentID = it.iter.Key()
		it.currentSlice = it.iter.Value()
	}
}

// Done reports whether the iterator is positioned past the last entry.
func (it *PointIndexIterator[Data]) Done() bool { return !it.valid }

// CellID returns the CellID of the current entry.
// Requires: !Done()
func (it *PointIndexIterator[Data]) CellID() CellID { return it.currentID }

// Point returns the point of the current entry.
// Requires: !Done()
func (it *PointIndexIterator[Data]) Point() Point { return it.currentSlice[it.sliceIdx].point }

// Data returns the data of the current entry.
// Requires: !Done()
func (it *PointIndexIterator[Data]) Data() Data { return it.currentSlice[it.sliceIdx].data }

// PointData returns the (Point, Data) pair for the current entry.
// Requires: !Done()
func (it *PointIndexIterator[Data]) PointData() PointData[Data] {
	return it.currentSlice[it.sliceIdx]
}

// Refresh gives the iterator a fresh internal cursor positioned at the current
// entry without changing the logical position (CellID or slice index). Call
// this after copying an iterator (it2 = *it) and before calling Next() or
// Prev() across a CellID boundary, to decouple the copy's cursor from the
// original.
func (it *PointIndexIterator[Data]) Refresh() {
	if !it.valid {
		return
	}
	it.iter = it.index.tree.Iter()
	it.iter.Seek(it.currentID)
}

// Begin positions the iterator at the first entry (if any).
func (it *PointIndexIterator[Data]) Begin() { it.Init(it.index) }

// Finish positions the iterator so that Done() is true.
func (it *PointIndexIterator[Data]) Finish() {
	it.valid = false
	it.atEnd = true
}

// Next advances to the next entry.
//
// Next uses the live internal cursor directly (O(1) per step). After copying
// an iterator (it2 = *it), call it2.Refresh() before the first cross-CellID
// Next() call to decouple the copy's cursor from the original.
// Requires: !Done()
func (it *PointIndexIterator[Data]) Next() {
	// Fast path: advance within the current CellID's group.
	if it.sliceIdx+1 < len(it.currentSlice) {
		it.sliceIdx++
		return
	}
	// Slow path: advance the live cursor to the next map entry — O(1).
	it.valid = it.iter.Next()
	it.atEnd = !it.valid
	it.sliceIdx = 0
	if it.valid {
		it.currentID = it.iter.Key()
		it.currentSlice = it.iter.Value()
	}
}

// Prev moves to the previous entry and reports whether the iterator was not
// already at the first entry. If Done() is true (e.g. after Seek past the end),
// Prev navigates to the last entry.
//
// Prev uses the live internal cursor directly (O(1) per step). After copying
// an iterator (it2 = *it), call it2.Refresh() before the first cross-CellID
// Prev() call to decouple the copy's cursor from the original.
func (it *PointIndexIterator[Data]) Prev() bool {
	// Fast path: go back within the current CellID's group.
	if it.valid && it.sliceIdx > 0 {
		it.sliceIdx--
		return true
	}
	if it.valid {
		if it.iter.Prev() {
			it.currentID = it.iter.Key()
			it.currentSlice = it.iter.Value()
			it.sliceIdx = len(it.currentSlice) - 1
			it.atEnd = false
			return true
		}
		// Already at the first entry; restore the cursor so Next() still works.
		it.iter.Seek(it.currentID)
		return false
	}
	if it.atEnd {
		if it.iter.Last() {
			it.currentID = it.iter.Key()
			it.currentSlice = it.iter.Value()
			it.sliceIdx = len(it.currentSlice) - 1
			it.valid = true
			it.atEnd = false
			return true
		}
		return false
	}
	return false
}

// Seek positions the iterator at the first entry with CellID >= target,
// or at Done if no such entry exists.
func (it *PointIndexIterator[Data]) Seek(target CellID) {
	it.iter = it.index.tree.Iter() // fresh iter: decouples from any copy
	it.valid = it.iter.Seek(target)
	it.atEnd = !it.valid
	it.sliceIdx = 0
	if it.valid {
		it.currentID = it.iter.Key()
		it.currentSlice = it.iter.Value()
	}
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
