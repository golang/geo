// Copyright 2018 Google Inc. All rights reserved.
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
	"testing"

	"github.com/golang/geo/s1"
)

func TestContainsPointQueryVertexModelOpen(t *testing.T) {
	index := makeShapeIndex("0:0 # -1:1, 1:1 # 0:5, 0:7, 2:6")
	q := NewContainsPointQuery(index, VertexModelOpen)

	tests := []struct {
		pt   Point
		want bool
	}{
		{pt: parsePoint("0:0"), want: false},
		{pt: parsePoint("-1:1"), want: false},
		{pt: parsePoint("1:1"), want: false},
		{pt: parsePoint("0:2"), want: false},
		{pt: parsePoint("0:3"), want: false},
		{pt: parsePoint("0:5"), want: false},
		{pt: parsePoint("0:7"), want: false},
		{pt: parsePoint("2:6"), want: false},
		{pt: parsePoint("1:6"), want: true},
		{pt: parsePoint("10:10"), want: false},
	}
	for _, test := range tests {
		if got := q.Contains(test.pt); got != test.want {
			t.Errorf("query.ContainsPoint(%v) = %v, want %v", test.pt, got, test.want)
		}
	}

	if s, p := index.Shape(1), parsePoint("1:6"); q.ShapeContains(s, p) {
		t.Errorf("query.ShapeContains(%v) = %v, want %v", p, true, false)
	}
	if s, p := index.Shape(2), parsePoint("1:6"); !q.ShapeContains(s, p) {
		t.Errorf("query.ShapeContains(%v) = %v, want %v", p, false, true)
	}
	if s, p := index.Shape(2), parsePoint("0:5"); q.ShapeContains(s, p) {
		t.Errorf("query.ShapeContains(%v) = %v, want %v", p, true, false)
	}
	if s, p := index.Shape(2), parsePoint("0:7"); q.ShapeContains(s, p) {
		t.Errorf("query.ShapeContains(%v) = %v, want %v", p, true, false)
	}
}

func TestContainsPointQueryVertexModelSemiOpen(t *testing.T) {
	index := makeShapeIndex("0:0 # -1:1, 1:1 # 0:5, 0:7, 2:6")
	q := NewContainsPointQuery(index, VertexModelSemiOpen)

	tests := []struct {
		pt   Point
		want bool
	}{
		{pt: parsePoint("0:0"), want: false},
		{pt: parsePoint("-1:1"), want: false},
		{pt: parsePoint("1:1"), want: false},
		{pt: parsePoint("0:2"), want: false},
		{pt: parsePoint("0:5"), want: false},
		{pt: parsePoint("0:7"), want: true}, // contained vertex
		{pt: parsePoint("2:6"), want: false},
		{pt: parsePoint("1:6"), want: true},
		{pt: parsePoint("10:10"), want: false},
	}
	for _, test := range tests {
		if got := q.Contains(test.pt); got != test.want {
			t.Errorf("query.ContainsPoint(%v) = %v, want %v", test.pt, got, test.want)
		}
	}

	if s, p := index.Shape(1), parsePoint("1:6"); q.ShapeContains(s, p) {
		t.Errorf("query.ShapeContains(%v) = %v, want %v", p, true, false)
	}
	if s, p := index.Shape(2), parsePoint("1:6"); !q.ShapeContains(s, p) {
		t.Errorf("query.ShapeContains(%v) = %v, want %v", p, false, true)
	}
	if s, p := index.Shape(2), parsePoint("0:5"); q.ShapeContains(s, p) {
		t.Errorf("query.ShapeContains(%v) = %v, want %v", p, true, false)
	}
	if s, p := index.Shape(2), parsePoint("0:7"); !q.ShapeContains(s, p) {
		t.Errorf("query.ShapeContains(%v) = %v, want %v", p, false, true)
	}
}

func TestContainsPointQueryVertexModelClosed(t *testing.T) {
	index := makeShapeIndex("0:0 # -1:1, 1:1 # 0:5, 0:7, 2:6")
	q := NewContainsPointQuery(index, VertexModelClosed)

	tests := []struct {
		pt   Point
		want bool
	}{
		{pt: parsePoint("0:0"), want: true},
		{pt: parsePoint("-1:1"), want: true},
		{pt: parsePoint("1:1"), want: true},
		{pt: parsePoint("0:2"), want: false},
		{pt: parsePoint("0:5"), want: true},
		{pt: parsePoint("0:7"), want: true},
		{pt: parsePoint("2:6"), want: true},
		{pt: parsePoint("1:6"), want: true},
		{pt: parsePoint("10:10"), want: false},
	}
	for _, test := range tests {
		if got := q.Contains(test.pt); got != test.want {
			t.Errorf("query.ContainsPoint(%v) = %v, want %v", test.pt, got, test.want)
		}
	}

	if s, p := index.Shape(1), parsePoint("1:6"); q.ShapeContains(s, p) {
		t.Errorf("query.ShapeContains(%v) = %v, want %v", p, true, false)
	}
	if s, p := index.Shape(2), parsePoint("1:6"); !q.ShapeContains(s, p) {
		t.Errorf("query.ShapeContains(%v) = %v, want %v", p, false, true)
	}
	if s, p := index.Shape(2), parsePoint("0:5"); !q.ShapeContains(s, p) {
		t.Errorf("query.ShapeContains(%v) = %v, want %v", p, false, true)
	}
	if s, p := index.Shape(2), parsePoint("0:7"); !q.ShapeContains(s, p) {
		t.Errorf("query.ShapeContains(%v) = %v, want %v", p, false, true)
	}
}

func TestContainsPointQueryContainingShapes(t *testing.T) {
	const numVerticesPerLoop = 10
	maxLoopRadius := kmToAngle(10)
	centerCap := CapFromCenterAngle(randomPoint(), maxLoopRadius)
	index := NewShapeIndex()

	for i := 0; i < 100; i++ {
		index.Add(RegularLoop(samplePointFromCap(centerCap), s1.Angle(randomFloat64())*maxLoopRadius, numVerticesPerLoop))
	}

	query := NewContainsPointQuery(index, VertexModelSemiOpen)

	for i := 0; i < 100; i++ {
		p := samplePointFromCap(centerCap)
		var want []Shape

		for j := int32(0); j < int32(len(index.shapes)); j++ {
			shape := index.Shape(j)
			// All the shapes we added were of type loop.
			loop := shape.(*Loop)
			if loop.ContainsPoint(p) {
				if !query.ShapeContains(shape, p) {
					t.Errorf("index.Shape(%d).ContainsPoint(%v) = true, but query.ShapeContains(%v) = false", j, p, p)
				}
				want = append(want, shape)
			} else {
				if query.ShapeContains(shape, p) {
					t.Errorf("query.ShapeContains(shape, %v) = true, but the original loop does not contain the point.", p)
				}
			}
		}
		got := query.ContainingShapes(p)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("%d query.ContainingShapes(%v) = %+v, want %+v", i, p, got, want)
		}
	}
}

// TODO(roberts): Remaining tests
// TestContainsPointQueryVisitIncidentEdges
