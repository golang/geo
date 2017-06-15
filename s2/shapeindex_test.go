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
	"testing"
)

// testShape is a minimal implementation of the Shape interface for use in testing
// until such time as there are other s2 types that implement it.
type testShape struct {
	a, b  Point
	edges int
}

func newTestShape() *testShape                { return &testShape{} }
func (s *testShape) NumEdges() int            { return s.edges }
func (s *testShape) Edge(id int) (a, b Point) { return s.a, s.b }
func (s *testShape) dimension() dimension     { return pointGeometry }
func (s *testShape) numChains() int           { return 0 }
func (s *testShape) chainStart(i int) int     { return 0 }
func (s *testShape) HasInterior() bool        { return false }
func (s *testShape) ContainsOrigin() bool     { return false }

func TestShapeIndexBasics(t *testing.T) {
	index := NewShapeIndex()
	s := newTestShape()

	if index.Len() != 0 {
		t.Errorf("initial index should be empty after creation")
	}
	index.Add(s)

	if index.Len() == 0 {
		t.Errorf("index should not be empty after adding shape")
	}

	index.Reset()
	if index.Len() != 0 {
		t.Errorf("index should be empty after reset, got %v %+v", index.Len(), index)
	}
}

func TestShapeIndexCellBasics(t *testing.T) {
	s := &ShapeIndexCell{}

	if len(s.shapes) != 0 {
		t.Errorf("len(s.shapes) = %v, want %d", len(s.shapes), 0)
	}

	// create some clipped shapes to add.
	c1 := &clippedShape{}
	s.add(c1)

	c2 := newClippedShape(7, 1)
	s.add(c2)

	c3 := &clippedShape{}
	s.add(c3)

	// look up the element at a given index
	if got := s.shapes[1]; got != c2 {
		t.Errorf("%v.shapes[%d] = %v, want %v", s, 1, got, c2)
	}

	// look for the clipped shape that is part of the given shape.
	if got := s.findByShapeID(7); got != c2 {
		t.Errorf("%v.findByShapeID(%v) = %v, want %v", s, 7, got, c2)
	}
}

// validateEdge determines whether or not the edge defined by points A and B should be
// present in that CellID and verify that this matches hasEdge.
func validateEdge(t *testing.T, a, b Point, ci CellID, hasEdge bool) {
	// Expand or shrink the padding slightly to account for errors in the
	// function we use to test for intersection (IntersectsRect).
	padding := cellPadding
	sign := 1.0
	if !hasEdge {
		sign = -1
	}
	padding += sign * intersectsRectErrorUVDist
	bound := ci.boundUV().ExpandedByMargin(padding)
	aUV, bUV, ok := ClipToPaddedFace(a, b, ci.Face(), padding)

	if got := ok && edgeIntersectsRect(aUV, bUV, bound); got != hasEdge {
		t.Errorf("edgeIntersectsRect(%v, %v, %v) = %v && clip = %v, want %v", aUV, bUV, bound, edgeIntersectsRect(aUV, bUV, bound), ok, hasEdge)
	}
}

// validateInterior tests if the given Shape contains the center of the given CellID,
// and that this matches the expected value of indexContainsCenter.
func validateInterior(t *testing.T, shape Shape, ci CellID, indexContainsCenter bool) {
	if shape == nil || !shape.HasInterior() {
		if indexContainsCenter {
			t.Errorf("%v was nil or does not have an interior, but should have", shape)
		}
		return
	}

	a := OriginPoint()
	b := ci.Point()
	crosser := NewEdgeCrosser(a, b)
	containsCenter := shape.ContainsOrigin()
	for e := 0; e < shape.NumEdges(); e++ {
		c, d := shape.Edge(e)
		containsCenter = containsCenter != crosser.EdgeOrVertexCrossing(c, d)
	}

	if containsCenter != indexContainsCenter {
		t.Errorf("validating interior of shape containsCenter = %v, want %v", containsCenter, indexContainsCenter)
	}
}

func testIteratorMethods(t *testing.T, index *ShapeIndex) {
	it := index.Iterator()
	if !it.AtBegin() {
		t.Errorf("new iterator should start positioned at beginning")
	}

	it = index.End()
	if !it.Done() {
		t.Errorf("iterator positioned at end should report as done")
	}

	var ids []CellID
	// "minCellID" is the first CellID in a complete traversal.
	minCellID := CellIDFromFace(0).ChildBeginAtLevel(maxLevel)

	for it.Reset(); !it.Done(); it.Next() {
		// Get the next cell in the iterator.
		ci := it.CellID()
		skipped := CellUnionFromRange(minCellID, ci.RangeMin())

		it2 := index.Iterator()
		for i := 0; i < len(skipped); i++ {
			if it2.LocatePoint(skipped[i].Point()) {
				t.Errorf("iterator should not have been able to find the cell %v wihich was not in the index", skipped[i].Point())
			}

			if got := it2.LocateCellID(skipped[i]); got != Disjoint {
				t.Errorf("CellID location should be Disjoint for non-existent entry, got %v", got)
			}
		}

		if len(ids) != 0 {
			if it.AtBegin() {
				t.Errorf("an iterator from a non-empty set of cells should not be positioned at the beginning")
			}

			it2 = index.Iterator()
			it2.Prev()
			// jump back one spot.
			if ids[len(ids)-1] != it2.CellID() {
				t.Errorf("ShapeIndexIterator should be positioned at the beginning and not equal to last entry")
			}

			it2.Next()
			if ci != it2.CellID() {
				t.Errorf("advancing one spot should put us at the end")
			}

			it2.seek(ids[len(ids)-1])
			if ids[len(ids)-1] != it2.CellID() {
				t.Errorf("seek from beginning for the first entry (%v) should not put us at the end %v", ids[len(ids)-1], it.CellID())
			}

			it2.seekForward(ci)
			if ci != it2.CellID() {
				t.Errorf("%v.seekForward(%v) = %v, want %v", it2, ci, it2.CellID(), ci)
			}

			it2.seekForward(ids[len(ids)-1])
			if ci != it2.CellID() {
				t.Errorf("%v.seekForward(%v) (to last entry) = %v, want %v", it2, ids[len(ids)-1], it2.CellID(), ci)
			}
		}

		it2.Reset()
		if ci.Point() != it.Center() {
			t.Errorf("point at center of current position should equal center of the crrent CellID")
		}

		if !it2.LocatePoint(it.Center()) {
			t.Fatalf("it.LocatePoint(it.Center()) should have been able to locate the point it is currently at")
		}

		if ci != it2.CellID() {
			t.Errorf("CellID of the Point we just located should be equal. got %v, want %v", it2.CellID(), ci)
		}

		it2.Reset()
		if got := it2.LocateCellID(ci); got != Indexed {
			t.Errorf("it.LocateCellID(%v) = %v, want %v", ci, got, Indexed)
		}

		if ci != it2.CellID() {
			t.Errorf("CellID of the CellID we just located should match. got %v, want %v", it2.CellID(), ci)
		}

		if !ci.isFace() {
			it2.Reset()
			if got := it2.LocateCellID(ci.immediateParent()); Subdivided != got {
				t.Errorf("it2.LocateCellID(%v) = %v, want %v", ci.immediateParent(), got, Subdivided)
			}

			if it2.CellID() > ci {
				t.Errorf("CellID of the immediate parent should be above the current cell, got %v, want %v", it2.CellID(), ci)
			}

			if it2.CellID() < ci.immediateParent().RangeMin() {
				t.Errorf("CellID of the current position should fall below the RangeMin of the parent. got %v, want %v", it2.CellID(), ci.immediateParent().RangeMin())
			}
		}

		if !ci.IsLeaf() {
			for i := 0; i < 4; i++ {
				it2.Reset()
				if got, want := it2.LocateCellID(ci.Children()[i]), Indexed; got != want {
					t.Errorf("it2.LocateCellID(%v.Children[%d]) = %v, want %v", ci, i, got, want)
				}

				if ci != it2.CellID() {
					t.Errorf("it2.CellID() = %v, want %v", it2.CellID(), ci)
				}
			}
		}
		// Add this cellID to the set of cells to examine.
		ids = append(ids, ci)
		// Move the minimal CellID to the next CellID past our current position.
		minCellID = ci.RangeMax().Next()
	}
}

func TestShapeIndexNoEdges(t *testing.T) {
	index := NewShapeIndex()
	iter := index.Iterator()

	if !iter.Done() {
		t.Errorf("iterator for empty index should report as done but did not")
	}
	testIteratorMethods(t, index)
}
