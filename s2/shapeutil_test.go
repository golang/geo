// Copyright 2017 Google Inc. All rights reserved.
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

	"github.com/golang/geo/s1"
)

func TestShapeutilContainsBruteForceNoInterior(t *testing.T) {
	// Defines a polyline that almost entirely encloses the point 0:0.
	polyline := makeLaxPolyline("0:0, 0:1, 1:-1, -1:-1, -1e9:1")
	point := parsePoint("0:0")
	if containsBruteForce(polyline, point) {
		t.Errorf("containsBruteForce(%v, %v) = true, want false", polyline, point)
	}
}

func TestShapeutilContainsBruteForceContainsReferencePoint(t *testing.T) {
	// Checks that containsBruteForce agrees with ReferencePoint.
	polygon := makeLaxPolygon("0:0, 0:1, 1:-1, -1:-1, -1e9:1")
	ref := polygon.ReferencePoint()
	if got := containsBruteForce(polygon, ref.Point); got != ref.Contained {
		t.Errorf("containsBruteForce(%v, %v) = %v, want %v", polygon, ref.Point, got, ref.Contained)
	}
}

func TestShapeutilContainsBruteForceConsistentWithLoop(t *testing.T) {
	// Checks that containsBruteForce agrees with Loop Contains.
	loop := RegularLoop(parsePoint("89:-179"), s1.Angle(10)*s1.Degree, 100)
	for i := 0; i < loop.NumVertices(); i++ {
		if got, want := loop.ContainsPoint(loop.Vertex(i)),
			containsBruteForce(loop, loop.Vertex(i)); got != want {
			t.Errorf("loop.ContainsPoint(%v) = %v, containsBruteForce(shape, %v) = %v, should be the same", loop.Vertex(i), got, loop.Vertex(i), want)
		}
	}
}

func TestShapeutilRangeIteratorNext(t *testing.T) {
	// Create an index with one point each on CellID faces 0, 1, and 2.
	index := makeShapeIndex("0:0 | 0:90 | 90:0 # #")
	it := newRangeIterator(index)
	itCount := 0
	next := func() {
		itCount++
		it.next()
		if it.done() {
			t.Errorf("There should be 3 items in the index but there were only %d", itCount)
		}
	}

	if got, want := it.cellID().Face(), 0; got != want {
		t.Errorf("it.CellID().Face() = %v, want %v", got, want)
	}
	next()

	if got, want := it.cellID().Face(), 1; got != want {
		t.Errorf("it.CellID().Face() = %v, want %v", got, want)
	}
	next()

	if got, want := it.cellID().Face(), 2; got != want {
		t.Errorf("it.CellID().Face() = %v, want %v", got, want)
	}

	it.next()
	if !it.done() {
		t.Errorf("iterator over index of three items should be done after 3 calls to next")
	}
	if got, want := it.cellID(), SentinelCellID; got != want {
		t.Errorf("it.CellID() = %v, want %v", got, want)
	}
}

func TestShapeutilRangeIteratorEmptyIndex(t *testing.T) {
	empty := makeShapeIndex("# #")
	nonEmpty := makeShapeIndex("0:0 # #")

	emptyIter := newRangeIterator(empty)
	nonEmptyIter := newRangeIterator(nonEmpty)

	if !emptyIter.done() {
		t.Errorf("the rangeIterator on an empty ShapeIndex should be done at creation")
	}

	emptyIter.seekTo(nonEmptyIter)
	if !emptyIter.done() {
		t.Errorf("seeking in the range iterator on an empty index to a cell should hit the end")
	}

	emptyIter.seekBeyond(nonEmptyIter)
	if !emptyIter.done() {
		t.Errorf("seeking in the range iterator on an empty index beyond a cell should hit the end")
	}

	emptyIter.seekTo(emptyIter)
	if !emptyIter.done() {
		t.Errorf("seeking in the range iterator on an empty index to a its current position should hit the end")
	}

	emptyIter.seekBeyond(emptyIter)
	if !emptyIter.done() {
		t.Errorf("seeking in the range iterator on an empty index beyond itself should hit the end")
	}
}
