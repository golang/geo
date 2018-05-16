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
)

func TestLaxPolylineNoVertices(t *testing.T) {
	shape := Shape(laxPolylineFromPoints([]Point{}))

	if got, want := shape.NumEdges(), 0; got != want {
		t.Errorf("shape.NumEdges() = %v, want %v", got, want)
	}
	if got, want := shape.NumChains(), 0; got != want {
		t.Errorf("shape.NumChains() = %v, want %v", got, want)
	}
	if got, want := shape.Dimension(), 1; got != want {
		t.Errorf("shape.Dimension() = %v, want %v", got, want)
	}
	if !shape.IsEmpty() {
		t.Errorf("shape.IsEmpty() = false, want true")
	}
	if shape.IsFull() {
		t.Errorf("shape.IsFull() = true, want false")
	}
	if shape.ReferencePoint().Contained {
		t.Errorf("shape.ReferencePoint().Contained = true, want false")
	}
}

func TestLaxPolylineOneVertex(t *testing.T) {
	shape := Shape(laxPolylineFromPoints([]Point{PointFromCoords(1, 0, 0)}))
	if got, want := shape.NumEdges(), 0; got != want {
		t.Errorf("shape.NumEdges() = %v, want %v", got, want)
	}
	if got, want := shape.NumChains(), 0; got != want {
		t.Errorf("shape.NumChains() = %v, want %v", got, want)
	}
	if got, want := shape.Dimension(), 1; got != want {
		t.Errorf("shape.Dimension() = %v, want %v", got, want)
	}
	if !shape.IsEmpty() {
		t.Errorf("shape.IsEmpty() = false, want true")
	}
	if shape.IsFull() {
		t.Errorf("shape.IsFull() = true, want false")
	}
}

func TestLaxPolylineEdgeAccess(t *testing.T) {
	vertices := parsePoints("0:0, 0:1, 1:1")
	shape := Shape(laxPolylineFromPoints(vertices))

	if got, want := shape.NumEdges(), 2; got != want {
		t.Errorf("shape.NumEdges() = %v, want %v", got, want)
	}
	if got, want := shape.NumChains(), 1; got != want {
		t.Errorf("shape.NumChains() = %v, want %v", got, want)
	}
	if got, want := shape.Chain(0).Start, 0; got != want {
		t.Errorf("shape.Chain(%d).Start = %d, want 0", got, want)
	}
	if got, want := shape.Chain(0).Length, 2; got != want {
		t.Errorf("shape.Chain(%d).Length = %d, want 2", got, want)
	}
	if got, want := shape.Dimension(), 1; got != want {
		t.Errorf("shape.Dimension() = %v, want %v", got, want)
	}
	if shape.IsEmpty() {
		t.Errorf("shape.IsEmpty() = true, want false")
	}
	if shape.IsFull() {
		t.Errorf("shape.IsFull() = true, want false")
	}

	edge0 := shape.Edge(0)
	if !edge0.V0.ApproxEqual(vertices[0]) {
		t.Errorf("shape.Edge(0).V0 = %v, want %v", edge0.V0, vertices[0])
	}
	if !edge0.V1.ApproxEqual(vertices[1]) {
		t.Errorf("shape.Edge(0).V1 = %v, want %v", edge0.V1, vertices[1])
	}

	edge1 := shape.Edge(1)
	if !edge1.V0.ApproxEqual(vertices[1]) {
		t.Errorf("shape.Edge(1).V0 = %v, want %v", edge1.V0, vertices[1])
	}
	if !edge1.V1.ApproxEqual(vertices[2]) {
		t.Errorf("shape.Edge(1).V1 = %v, want %v", edge1.V1, vertices[2])
	}
}
