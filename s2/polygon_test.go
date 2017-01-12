/*
Copyright 2015 Google Inc. All rights reserved.

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

func TestPolygonEmptyAndFull(t *testing.T) {
	emptyPolygon := &Polygon{}

	if !emptyPolygon.IsEmpty() {
		t.Errorf("empty polygon should be empty")
	}
	if emptyPolygon.IsFull() {
		t.Errorf("empty polygon should not be full")
	}
	/*
		// TODO(roberts): Uncomment when Polygon finishes the Shape interface.
		if emptyPolygon.ContainsOrigin() {
			t.Errorf("emptyPolygon.ContainsOrigin() = true, want false")
		}
		if got, want := emptyPolygon.NumEdges(), 0; got != want {
			t.Errorf("emptyPolygon.NumEdges() = %v, want %v", got, want)
		}
	*/
	if got := emptyPolygon.dimension(); got != polygonGeometry {
		t.Errorf("emptyPolygon.dimension() = %v, want %v", got, polygonGeometry)
	}
	if got, want := emptyPolygon.numChains(), 0; got != want {
		t.Errorf("emptyPolygon.numChains() = %v, want %v", got, want)
	}

	fullPolygon := FullPolygon()
	if fullPolygon.IsEmpty() {
		t.Errorf("full polygon should not be emtpy")
	}
	if !fullPolygon.IsFull() {
		t.Errorf("full polygon should be full")
	}
	/*
		// TODO(roberts): Uncomment when Polygon finishes the Shape interface.
		if !fullPolygon.ContainsOrigin() {
			t.Errorf("fullPolygon.ContainsOrigin() = false, want true")
		}
		if got, want := fullPolygon.NumEdges(), 0; got != want {
			t.Errorf("fullPolygon.NumEdges() = %v, want %v", got, want)
		}
	*/
	if got := fullPolygon.dimension(); got != polygonGeometry {
		t.Errorf("emptyPolygon.dimension() = %v, want %v", got, polygonGeometry)
	}
	if got, want := fullPolygon.numChains(), 0; got != want {
		t.Errorf("emptyPolygon.numChains() = %v, want %v", got, want)
	}
}

func TestPolygonShape(t *testing.T) {
	// TODO(roberts): Once Polygon implements Shape uncomment this test.
	/*
		p := &Polygon{}
		shape := Shape(p)
		if p.NumVertices() != shape.NumEdges() {
			t.Errorf("the number of vertices in a polygon should equal the number of edges")
		}
		if p.NumLoops() != shape.numChains() {
			t.Errorf("the number of loops in a polygon should equal the number of chains")
		}
		e := 0
		for i, l := range p.loops {
			if e != shape.chainStart(i) {
				t.Errorf("the edge if of the start of loop(%d) should equal the sum of vertices so far in the polygon. got %d, want %d", i, shape.chainStart(i), e)
			}
			for j := 0; j < len(l.Vertices()); j++ {
				v0, v1 := shape.Edge(e)
				// TODO(roberts): Update once Loop implements orientedVertex.
				//if l.orientedVertex(j) != v0 {
				if l.Vertex(j) != v0 {
					t.Errorf("l.Vertex(%d) = %v, want %v", j, l.Vertex(j), v0)
				}
				// TODO(roberts): Update once Loop implements orientedVertex.
				//if l.orientedVertex(j+1) != v1 {
				if l.Vertex(j+1) != v1 {
					t.Errorf("l.Vertex(%d) = %v, want %v", j+1, l.Vertex(j+1), v1)
				}
				e++
			}
			if e != shape.chainStart(i+1) {
				t.Errorf("the edge id of the start of the next loop(%d+1) should equal the sum of vertices so far in the polygon. got %d, want %d", i, shape.chainStart(i+1), e)
			}
		}
		if shape.dimension() != polygonGeometry {
			t.Errorf("polygon.dimension() = %v, want %v", shape.dimension() , polygonGeometry)
		}
		if !shape.HasInterior() {
			t.Errorf("polygons should always have interiors")
		}
	*/
}

func TestPolygonLoop(t *testing.T) {
	full := FullPolygon()
	if full.NumLoops() != 1 {
		t.Errorf("full polygon should have one loop")
	}

	l := &Loop{}
	p1 := PolygonFromLoops([]*Loop{l})
	if p1.NumLoops() != 1 {
		t.Errorf("polygon with one loop should have one loop")
	}
	if p1.Loop(0) != l {
		t.Errorf("polygon with one loop should return it")
	}

	// TODO: When multiple loops are supported, add more test cases.
}

func TestPolygonParent(t *testing.T) {
	p1 := PolygonFromLoops([]*Loop{&Loop{}})
	tests := []struct {
		p    *Polygon
		have int
		want int
		ok   bool
	}{
		{FullPolygon(), 0, -1, false},
		{p1, 0, -1, false},

		// TODO: When multiple loops are supported, add more test cases to
		// more fully show the parent levels.
	}

	for _, test := range tests {
		if got, ok := test.p.Parent(test.have); ok != test.ok || got != test.want {
			t.Errorf("%v.Parent(%d) = %d,%v, want %d,%v", test.p, test.have, got, ok, test.want, test.ok)
		}
	}
}

func TestPolygonLastDescendant(t *testing.T) {
	p1 := PolygonFromLoops([]*Loop{&Loop{}})

	tests := []struct {
		p    *Polygon
		have int
		want int
	}{
		{FullPolygon(), 0, 0},
		{FullPolygon(), -1, 0},

		{p1, 0, 0},
		{p1, -1, 0},

		// TODO: When multiple loops are supported, add more test cases.
	}

	for _, test := range tests {
		if got := test.p.LastDescendant(test.have); got != test.want {
			t.Errorf("%v.LastDescendant(%d) = %d, want %d", test.p, test.have, got, test.want)
		}
	}
}

func TestPolygonLoopIsHoleAndLoopSign(t *testing.T) {
	if FullPolygon().loopIsHole(0) {
		t.Errorf("the full polygons only loop should not be a hole")
	}
	if FullPolygon().loopSign(0) != 1 {
		t.Errorf("the full polygons only loop should be postitive")
	}

	loop := LoopFromPoints(parsePoints("30:20, 40:20, 39:43, 33:35"))
	p := PolygonFromLoops([]*Loop{loop})

	if p.loopIsHole(0) {
		t.Errorf("first loop in a polygon should not start out as a hole")
	}
	if p.loopSign(0) != 1 {
		t.Errorf("first loop in a polygon should start out as positive")
	}

	// TODO: When multiple loops are supported, add more test cases to
	// more fully show the parent levels.
}
