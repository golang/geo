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
	"testing"
)

// allEdgesInShapeIndex returns the full list of edges in the given index in shapeID order.
func allEdgesInShapeIndex(index *ShapeIndex) []Edge {
	var result []Edge
	// Iterator works over the shapes in shape ID order, so we don't just
	// range over the map here because order is not guaranteed.
	for i := 0; i < len(index.shapes); i++ {
		shape := index.shapes[int32(i)]
		if shape == nil {
			continue
		}
		for j := 0; j < shape.NumEdges(); j++ {
			result = append(result, shape.Edge(j))
		}
	}
	return result
}

func verifyEdgeIterator(t *testing.T, index *ShapeIndex) {
	expected := allEdgesInShapeIndex(index)
	i := 0

	for iter := NewEdgeIterator(index); !iter.Done(); iter.Next() {
		if got, want := iter.Edge(), expected[i]; got != want {
			t.Errorf("edge[%d] = %v, want %v", i, got, want)
		}
		i++
	}
}

func TestShapeutilEdgeIteratorEmpty(t *testing.T) {
	index := makeShapeIndex("##")
	verifyEdgeIterator(t, index)
}

func TestShapeutilEdgeIteratorPoints(t *testing.T) {
	index := makeShapeIndex("0:0|1:1##")
	verifyEdgeIterator(t, index)
}

func TestShapeutilEdgeIteratorLines(t *testing.T) {
	index := makeShapeIndex("#0:0,10:10|5:5,5:10|1:2,2:1#")
	verifyEdgeIterator(t, index)
}

func TestShapeutilEdgeIteratorPolygons(t *testing.T) {
	index := makeShapeIndex("##10:10,10:0,0:0|-10:-10,-10:0,0:0,0:-10")
	verifyEdgeIterator(t, index)
}

func TestShapeutilEdgeIteratorCollection(t *testing.T) {
	index := makeShapeIndex("1:1|7:2#1:1,2:2,3:3|2:2,1:7#10:10,10:0,0:0;20:20,20:10,10:10|15:15,15:0,0:0")
	verifyEdgeIterator(t, index)
}

func TestShapeutilEdgeIteratorRemove(t *testing.T) {
	index := makeShapeIndex("1:1|7:2#1:1,2:2,3:3|2:2,1:7#10:10,10:0,0:0;20:20,20:10,10:10|15:15,15:0,0:0")
	index.Remove(index.Shape(0))

	verifyEdgeIterator(t, index)
}

// edgeIteratorEq reports if two edge iterators are the same.
func edgeIteratorEq(a, b *EdgeIterator) bool {
	return a.shapeID == b.shapeID && a.edgeID == b.edgeID && a.index == b.index
}

// copyEdgeIterator copies the state of the given iterator into a new instance.
func copyEdgeIterator(a *EdgeIterator) *EdgeIterator {
	return &EdgeIterator{
		index:    a.index,
		shapeID:  a.shapeID,
		edgeID:   a.edgeID,
		numEdges: a.numEdges,
	}
}

func TestShapeutilEdgeIteratorAssignmentAndEquality(t *testing.T) {
	index1 := makeShapeIndex("1:1|7:2#1:1,2:2,3:3|2:2,1:7#10:10,10:0,0:0;20:20,20:10,10:10|15:15,15:0,0:0")

	index2 := makeShapeIndex("1:1|7:2#1:1,2:2,3:3|2:2,1:7#10:10,10:0,0:0;20:20,20:10,10:10|15:15,15:0,0:0")

	it1 := NewEdgeIterator(index1)
	it2 := NewEdgeIterator(index2)

	// The underlying indexes have the same data, but are not the same index.
	if edgeIteratorEq(it1, it2) {
		t.Errorf("edgeIterators equal but shouldn't be")
	}

	it1 = copyEdgeIterator(it2)
	if !edgeIteratorEq(it1, it2) {
		t.Errorf("edgeIterators not equal but should be")
	}

	it1.Next()
	if edgeIteratorEq(it1, it2) {
		t.Errorf("edgeIterators equal but shouldn't be after one is advanced. \n1: %+v\n2: %+v", it1, it2)
	}

	it2.Next()
	if !edgeIteratorEq(it1, it2) {
		t.Errorf("edgeIterators not equal but should be after both advanced same amount")
	}
}
